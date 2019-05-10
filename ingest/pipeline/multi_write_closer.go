package pipeline

import (
	"github.com/stellar/go/xdr"
)

func (t *multiWriteCloser) Write(entry xdr.LedgerEntry) error {
	for _, w := range t.writers {
		err := w.Write(entry)
		if err != nil {
			return err
		}
	}
	return nil
}

func (m *multiWriteCloser) Close() error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	m.closeAfter--
	if m.closeAfter > 0 {
		return nil
	}

	for _, w := range m.writers {
		err := w.Close()
		if err != nil {
			return err
		}
	}

	return nil
}
