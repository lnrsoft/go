package pipeline

import (
	"github.com/stellar/go/xdr"
	"github.com/stellar/go/ingest/io"
)

const bufferSize = 2000

func (b *bufferedStateReadWriteCloser) init() {
	b.buffer = make(chan xdr.LedgerEntry, bufferSize)
}

func (b *bufferedStateReadWriteCloser) GetSequence() uint32 {
	return 0
}

func (b *bufferedStateReadWriteCloser) Read() (xdr.LedgerEntry, error) {
	b.initOnce.Do(b.init)

	entry, more := <-b.buffer
	if more {
		return entry, nil
	} else {
		return xdr.LedgerEntry{}, io.EOF
	}
}

func (b *bufferedStateReadWriteCloser) Write(entry xdr.LedgerEntry) error {
	b.initOnce.Do(b.init)
	b.buffer <- entry
	return nil
}

func (b *bufferedStateReadWriteCloser) Close() error {
	b.initOnce.Do(b.init)
	close(b.buffer)
	return nil
}

var _ io.StateReader = &bufferedStateReadWriteCloser{}
var _ io.StateWriteCloser = &bufferedStateReadWriteCloser{}
