package pipeline

import (
	"io"
	"sync"
)

// Proof of concept types
type (
	StateReader      = io.Reader
	StateWriteCloser = io.WriteCloser
)

type BufferedStateReadWriteCloser struct {
	initOnce sync.Once
	closed   chan bool
	buffer   chan byte
}

type Pipeline struct {
	rootStateProcessor *PipelineNode
	done               bool
}

type multiWriteCloser struct {
	writers []StateWriteCloser
}

type PipelineNode struct {
	Processor StateProcessor
	Children  []*PipelineNode
}

// StateProcessor defines methods required by state processing pipeline.
type StateProcessor interface {
	// ProcessState ...
	ProcessState(store *Store, reader StateReader, writeCloser StateWriteCloser) (err error)
	// IsConcurent defines if processing pipeline should start a single instance
	// of the processor or multiple instances. Multiple instances will read
	// from the same StateReader and write to the same StateWriter.
	// Example: you can calculate number of asset holders in a single processor but
	// you can also start multiple processors that sum asset holders in a shared
	// variable to calculate it faster.
	IsConcurent() bool
	// RequiresInput defines if processor requires input data (StateReader). If not,
	// it will receive empty reader, it's parent process will write to "void" and
	// writes to `writer` will go to "void".
	// This is useful for processors resposible for saving aggregated data that don't
	// need state objects.
	RequiresInput() bool
	// Returns processor name. Helpful for errors, debuging and reports.
	Name() string
}

type Store struct {
	sync.Mutex
	initOnce sync.Once
	values   map[string]interface{}
}

// ReduceStateProcessor forwards the final produced by applying all the
// ledger entries to the writer.
// Let's say that there are 3 ledger entries:
//     - Create account A (XLM balance = 20)
//     - Update XLM balance of A to 5
//     - Update XLM balance of A to 15
// Running ReduceStateProcessor will add a single ledger entry:
//     - Create account A (XLM balance = 15)
// to the writer.
type ReduceStateProcessor struct {
	//
}
