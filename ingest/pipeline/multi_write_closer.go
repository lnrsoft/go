package pipeline

import (
	"io"

	"github.com/stellar/go/support/errors"
)

func (t *multiWriteCloser) Write(p []byte) (n int, err error) {
	for _, w := range t.writers {
		// BufferedStateReadWriteCloser supports writing only one byte
		// at a time so loop over more bytes
		for _, rb := range p {
			n, err = w.Write([]byte{rb})
			if err != nil {
				return
			}

			if n != 1 {
				err = errors.Wrap(io.ErrShortWrite, "multiWriteCloser")
				return
			}
		}
	}
	return len(p), nil
}

func (m *multiWriteCloser) Close() error {
	for _, w := range m.writers {
		err := w.Close()
		if err != nil {
			return err
		}
	}

	return nil
}
