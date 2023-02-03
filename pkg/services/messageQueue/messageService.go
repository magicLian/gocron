package messagequeue

import (
	"context"
	"time"

	"github.com/magicLian/logx"
	amqp "github.com/rabbitmq/amqp091-go"
)

type MessageService struct {
	uri            string
	exchangeName   string
	registerQueue  string
	taskReplyQueue string
	taskQueue      string

	log     logx.Logger
	conn    *amqp.Connection
	channel *amqp.Channel
	errChan chan error //internal err chan
}

// connect connects channel and exchange and 3 kinds of queues.
func (q *MessageService) connect() error {
	return nil
}

func (q *MessageService) KeepAlive(ctx context.Context) {
	for {
		err := q.channel.PublishWithContext(ctx, "", "heartbeat", false, false,
			amqp.Publishing{
				ContentType: "text/plain",
				Body:        []byte("replyMessageData"),
			},
		)
		if err != nil {
			q.log.Info(err)
			q.errChan <- err
		}

		select {
		case <-ctx.Done():
			q.log.Info("Exit keep alive go routine")
			return
		default:
			q.log.Info("Send heartbeat message")
		}

		time.Sleep(30 * time.Second)
	}
}

func (q *MessageService) Send(ctx context.Context, queueName, content string) error {
	if err := q.channel.PublishWithContext(ctx, "", queueName, false, false, amqp.Publishing{
		DeliveryMode: amqp.Persistent,
		ContentType:  "text/plain",
		Body:         []byte(content),
	},
	); err != nil {
		q.log.Info(err.Error())
		return err
	}

	return nil
}

func (q *MessageService) Close() error {
	if err := q.channel.Close(); err != nil {
		q.log.Debugf("close rmq channel raise error,[%s]", err.Error())
	}
	if err := q.conn.Close(); err != nil {
		q.log.Debugf("close rmq connection raise error,[%s]", err.Error())
	}
	q.log.Info("Rabbitmq is closed")

	return nil
}

func (q *MessageService) Receive(ctx context.Context, queueName string) error {
	return nil
}
