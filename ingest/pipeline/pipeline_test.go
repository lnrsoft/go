package pipeline

import (
	"fmt"
	"io"
	"io/ioutil"
	"strings"
	"testing"
	"time"
)

func TestBuffer(t *testing.T) {
	buffer := &BufferedStateReadWriteCloser{}

	go func() {
		read, err := ioutil.ReadAll(buffer)
		if err != nil {
			panic(err)
		}
		fmt.Println(read)
	}()

	buffer.WriteCloseString("test")
	time.Sleep(time.Second)
}

func TestAbc(t *testing.T) {
	pipeline := &Pipeline{}

	passthroughProcessor := &PassthroughProcessor{}
	uppercaseProcessor := &UppercaseProcessor{}
	lowercaseProcessor := &LowercaseProcessor{}
	printProcessor := &PrintProcessor{}

	pipeline.AddStateProcessorTree(
		pipeline.Node(passthroughProcessor).
			Pipe(
				pipeline.Node(lowercaseProcessor).
					Pipe(pipeline.Node(printProcessor)),
				pipeline.Node(uppercaseProcessor).
					Pipe(pipeline.Node(printProcessor)),
			),
	)

	buffer := &BufferedStateReadWriteCloser{}
	go buffer.WriteCloseString("testTEST")
	done := pipeline.ProcessState(buffer)
	<-done
}

type SimpleProcessor struct{}

func (n *SimpleProcessor) IsConcurent() bool {
	return false
}

func (n *SimpleProcessor) RequiresInput() bool {
	return true
}

type PassthroughProcessor struct {
	SimpleProcessor
}

func (p *PassthroughProcessor) ProcessState(store *Store, r StateReader, w StateWriteCloser) error {
	_, err := io.Copy(w, r)
	if err != nil {
		return err
	}

	w.Close()
	return nil
}

func (p *PassthroughProcessor) Name() string {
	return "PassthroughProcessor"
}

type UppercaseProcessor struct {
	SimpleProcessor
}

func (p *UppercaseProcessor) ProcessState(store *Store, r StateReader, w StateWriteCloser) error {
	defer w.Close()

	lettersCount := make(map[byte]int)
	for {
		b := make([]byte, 1)
		rn, rerr := r.Read(b)
		if rn == 1 {
			lettersCount[b[0]]++

			newLetter := b[0]
			if b[0] >= 97 && b[0] <= 122 {
				newLetter -= 32
			}

			_, werr := w.Write([]byte{newLetter})
			if werr != nil {
				return werr
			}
		}

		if rerr != nil {
			if rerr == io.EOF {
				break
			} else {
				return rerr
			}
		}
	}

	store.Lock()
	store.Put("letterCount", lettersCount)
	store.Unlock()

	return nil
}

func (p *UppercaseProcessor) Name() string {
	return "UppercaseProcessor"
}

type LowercaseProcessor struct {
	SimpleProcessor
}

func (p *LowercaseProcessor) ProcessState(store *Store, r StateReader, w StateWriteCloser) error {
	// This will read all into memory. See UppercaseProcessor for streaming
	// example.
	read, err := ioutil.ReadAll(r)
	if err != nil {
		return err
	}

	n := strings.ToLower(string(read))

	defer w.Close()
	_, err = fmt.Fprint(w, n)
	if err != nil {
		return err
	}

	return nil
}

func (p *LowercaseProcessor) Name() string {
	return "LowercaseProcessor"
}

type PrintProcessor struct {
	SimpleProcessor
}

func (p *PrintProcessor) ProcessState(store *Store, r StateReader, w StateWriteCloser) error {
	defer w.Close()

	read, err := ioutil.ReadAll(r)
	if err != nil {
		return err
	}

	store.Lock()
	fmt.Println(string(read), store.Get("letterCount"))
	store.Unlock()
	return nil
}

func (p *PrintProcessor) Name() string {
	return "PrintProcessor"
}
