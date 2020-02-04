package channel

import (
	"io"
	"time"
)

type Channel interface {
	io.Closer

	Push(topic string, message []byte) error

	Take(topic string, timeout time.Duration) ([]byte, error)

	Listen(topic string, listener Listener)
}

type Listener interface {
	OnNext(message []byte)

	OnError(err error)

	OnComplete()
}
