package pipeline

import (
	"fmt"
	"testing"
	"time"
	"sync"
	"strings"

	"github.com/stellar/go/ingest/io"
	"github.com/stellar/go/xdr"
	"github.com/stellar/go/keypair"
)

func AccountLedgerEntry() xdr.LedgerEntry {
	random, err := keypair.Random()
	if err != nil {
		panic(err)
	}

	id := xdr.AccountId{}
	id.SetAddress(random.Address())

	return xdr.LedgerEntry{
		LastModifiedLedgerSeq: 0,
		Data: xdr.LedgerEntryData{
			Type: xdr.LedgerEntryTypeAccount,
			Account: &xdr.AccountEntry{
				AccountId: id,
			},
		},
	}
}

func TrustLineLedgerEntry() xdr.LedgerEntry {
	random, err := keypair.Random()
	if err != nil {
		panic(err)
	}

	id := xdr.AccountId{}
	id.SetAddress(random.Address())

	return xdr.LedgerEntry{
		LastModifiedLedgerSeq: 0,
		Data: xdr.LedgerEntryData{
			Type: xdr.LedgerEntryTypeTrustline,
			TrustLine: &xdr.TrustLineEntry{
				AccountId: id,
			},
		},
	}
}

func TestStore(t *testing.T) {
	var s Store

	s.Lock()
	s.Put("value", 0)
	s.Unlock()

	s.Lock()
	v := s.Get("value")
	s.Put("value", v.(int)+1)
	s.Unlock()

	fmt.Println(s.Get("value"))
}

func TestBuffer(t *testing.T) {
	buffer := &bufferedStateReadWriteCloser{}

	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		for {
			entry, err := buffer.Read()
			if err != nil {
				if err == io.EOF {
					break
				} else {
					panic(err)
				}
			}
			fmt.Println("Read", entry.Data.Account.AccountId.Address())
			time.Sleep(4*time.Second)
		}
	}()

	go func() {
		defer wg.Done()
		for i := 0; i < 20; i++ {
			buffer.Write(AccountLedgerEntry())
			fmt.Println("Wrote")
			time.Sleep(time.Second)
		}
		buffer.Close()
	}()

	wg.Wait()
}

func TestPipeline(t *testing.T) {
	pipeline := &Pipeline{}

	passthroughProcessor := &PassthroughProcessor{}
	accountsOnlyFilter := &AccountsOnlyFilter{}
	printProcessor := &PrintProcessor{}

	pipeline.AddStateProcessorTree(
		pipeline.Node(passthroughProcessor).
			Pipe(
				pipeline.Node(accountsOnlyFilter).
					Pipe(
						pipeline.Node(&CountPrefixProcessor{Prefix: "GA"}).
							Pipe(pipeline.Node(printProcessor)),
						pipeline.Node(&CountPrefixProcessor{Prefix: "GB"}).
							Pipe(pipeline.Node(printProcessor)),
						pipeline.Node(&CountPrefixProcessor{Prefix: "GC"}).
							Pipe(pipeline.Node(printProcessor)),
						pipeline.Node(&CountPrefixProcessor{Prefix: "GD"}).
							Pipe(pipeline.Node(printProcessor)),
					),
				),
	)

	buffer := &bufferedStateReadWriteCloser{}
	
	go func() {
		for i := 0; i < 10000; i++ {
			buffer.Write(AccountLedgerEntry())
			buffer.Write(TrustLineLedgerEntry())
		}
		buffer.Close()
	}()

	done := pipeline.ProcessState(buffer)
	<-done
}

type SimpleProcessor struct{
	sync.Mutex
	callCount int
}

func (n *SimpleProcessor) IsConcurrent() bool {
	return false
}

func (n *SimpleProcessor) RequiresInput() bool {
	return true
}

func (n *SimpleProcessor) CallCount() int {
	n.Lock()
	defer n.Unlock()
	n.callCount++
	return n.callCount
}

type PassthroughProcessor struct {
	SimpleProcessor
}

func (p *PassthroughProcessor) ProcessState(store *Store, r io.StateReader, w io.StateWriteCloser) error {
	for {
		entry, err := r.Read()
		if err != nil {
			if err == io.EOF {
				break
			} else {
				return err
			}
		}

		w.Write(entry)
	}

	w.Close()
	return nil
}

func (p *PassthroughProcessor) Name() string {
	return "PassthroughProcessor"
}

type AccountsOnlyFilter struct {
	SimpleProcessor
}

func (p *AccountsOnlyFilter) ProcessState(store *Store, r io.StateReader, w io.StateWriteCloser) error {
	for {
		entry, err := r.Read()
		if err != nil {
			if err == io.EOF {
				break
			} else {
				return err
			}
		}

		if entry.Data.Type == xdr.LedgerEntryTypeAccount {
			w.Write(entry)
		}
	}

	w.Close()
	return nil
}

func (p *AccountsOnlyFilter) Name() string {
	return "AccountsOnlyFilter"
}

type CountPrefixProcessor struct {
	SimpleProcessor
	Prefix string
}

func (p *CountPrefixProcessor) ProcessState(store *Store, r io.StateReader, w io.StateWriteCloser) error {
	// Close writer when we're done
	defer w.Close()

	count := 0

	for {
		entry, err := r.Read()
		if err != nil {
			if err == io.EOF {
				break
			} else {
				return err
			}
		}

		address := entry.Data.Account.AccountId.Address()

		if strings.HasPrefix(address, p.Prefix) {
			count++
		}
	}

	store.Lock()
	prevCount := store.Get("count"+p.Prefix)
	if prevCount != nil {
		count += prevCount.(int)
	}
	store.Put("count"+p.Prefix, count)
	store.Unlock()
	
	return nil
}

func (p *CountPrefixProcessor) IsConcurrent() bool {
	return true
}

func (p *CountPrefixProcessor) Name() string {
	return "CountPrefixProcessor"
}

type PrintProcessor struct {
	SimpleProcessor
}

func (p *PrintProcessor) ProcessState(store *Store, r io.StateReader, w io.StateWriteCloser) error {
	defer w.Close()

	// This should be a helper function or a method on `io.StateReader`.
	for {
		_, err := r.Read()
		if err != nil {
			if err == io.EOF {
				break
			} else {
				return err
			}
		}
	}

	if p.CallCount() != 4 {
		return nil
	}

	store.Lock()
	fmt.Println("countGA", store.Get("countGA"))
	fmt.Println("countGB", store.Get("countGB"))
	fmt.Println("countGC", store.Get("countGC"))
	fmt.Println("countGD", store.Get("countGD"))
	store.Unlock()

	return nil
}

func (p *PrintProcessor) Name() string {
	return "PrintProcessor"
}
