package messagequeue

import "context"

type Messager interface {
	KeepAlive(ctx context.Context)
	Send(topic, content string) error
	Receive(topic string) ([]byte, error)
	Close() error
}
