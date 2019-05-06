package pipeline

import (
	"sync"
)

func (p *Pipeline) Node(processor StateProcessor) *PipelineNode {
	return &PipelineNode{
		Processor: processor,
	}
}

func (p *Pipeline) AddStateProcessorTree(rootProcessor *PipelineNode) {
	p.rootStateProcessor = rootProcessor
}

func (p *Pipeline) ProcessState(reader StateReader) (done chan error) {
	return p.processStateNode(&Store{}, p.rootStateProcessor, reader)
}

func (p *Pipeline) processStateNode(store *Store, node *PipelineNode, reader StateReader) chan error {
	outputs := make([]StateWriteCloser, len(node.Children))

	for i := range outputs {
		outputs[i] = &BufferedStateReadWriteCloser{}
	}

	writer := &multiWriteCloser{writers: outputs}

	var wg sync.WaitGroup
	wg.Add(1)

	go func(reader StateReader, writer StateWriteCloser) {
		defer wg.Done()
		err := node.Processor.ProcessState(store, reader, writer)
		if err != nil {
			panic(err)
		}
	}(reader, writer)

	for i, child := range node.Children {
		wg.Add(1)
		go func(i int, child *PipelineNode) {
			defer wg.Done()
			done := p.processStateNode(store, child, outputs[i].(*BufferedStateReadWriteCloser))
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
