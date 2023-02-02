package messagequeue

import "context"

type Messager interface {
	KeepAlive(ctx context.Context)
	Send(ctx context.Context, queueName, content string) error
	Receive(ctx context.Context, queueName string) error
	Close() error
}
