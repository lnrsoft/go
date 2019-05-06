package pipeline

import (
	"io"
)

const bufferSize = 4

func (b *BufferedStateReadWriteCloser) init() {
	b.buffer = make(chan byte, bufferSize)
	b.closed = make(chan bool)
}

func (b *BufferedStateReadWriteCloser) Read(p []byte) (n int, err error) {
	b.initOnce.Do(b.init)

	// This is to make sure to drain channel first if b.closed is ready.
	// TODO move `case` contents to another method
	select {
	case rb := <-b.buffer:
		p[0] = rb
		return 1, nil
	default:
	}

	select {
	case rb := <-b.buffer:
		p[0] = rb
		return 1, nil
	case <-b.closed:
		return 0, io.EOF
	}
}

func (b *BufferedStateReadWriteCloser) Write(p []byte) (n int, err error) {
	b.initOnce.Do(b.init)
	b.buffer <- p[0]
	return 1, nil
}

func (b *BufferedStateReadWriteCloser) Close() error {
	b.initOnce.Do(b.init)
	b.closed <- true
	close(b.closed)
	close(b.buffer)
	return nil
}

func (b *BufferedStateReadWriteCloser) WriteCloseString(s string) {
	b.initOnce.Do(b.init)

	for _, rb := range s {
		_, err := b.Write([]byte{byte(rb)})
		if err != nil {
			panic(err)
		}
	}

	err := b.Close()
	if err != nil {
		panic(err)
	}
}
