package pipeline

import (
	"sync"

	"github.com/stellar/go/ingest/io"
)

func (p *Pipeline) Node(processor StateProcessor) *PipelineNode {
	return &PipelineNode{
		Processor: processor,
	}
}

func (p *Pipeline) AddStateProcessorTree(rootProcessor *PipelineNode) {
	p.rootStateProcessor = rootProcessor
}

func (p *Pipeline) ProcessState(reader io.StateReader) (done chan error) {
	return p.processStateNode(&Store{}, p.rootStateProcessor, reader)
}

func (p *Pipeline) processStateNode(store *Store, node *PipelineNode, reader io.StateReader) chan error {
	outputs := make([]io.StateWriteCloser, len(node.Children))

	for i := range outputs {
		outputs[i] = &bufferedStateReadWriteCloser{}
	}

	var wg sync.WaitGroup
	
	jobs := 1
	if node.Processor.IsConcurrent() {
		jobs = 10
	}

	writer := &multiWriteCloser{
		writers: outputs,
		closeAfter: jobs,
	}

	for i := 1; i <= jobs; i++ {
		wg.Add(1)
		go func(reader io.StateReader, writer io.StateWriteCloser) {
			defer wg.Done()
			err := node.Processor.ProcessState(store, reader, writer)
			if err != nil {
				// TODO return to pipeline error channel
				panic(err)
			}
		}(reader, writer)
	}

	for i, child := range node.Children {
		wg.Add(1)
		go func(i int, child *PipelineNode) {
			defer wg.Done()
			done := p.processStateNode(store, child, outputs[i].(*bufferedStateReadWriteCloser))
			<-done
		}(i, child)
	}

	done := make(chan error)

	go func() {
		wg.Wait()
		done <- nil
	}()

	return done
}
