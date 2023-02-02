package messagequeue

import (
	"context"
	"time"

	"github.com/magicLian/logx"
	amqp "github.com/rabbitmq/amqp091-go"
)

type MessageService struct {
	uri           string
	exchangeName  string
	registerQueue string
	errReplyQueue string
	workQueue     string

	log     logx.Logger
	conn    *amqp.Connection
	channel *amqp.Channel
	errChan chan error //internal err chan
}

// connect connects channel and exchange and 3 kinds of queues.
func (q *MessageService) connect() error {
	cfg := amqp.Config{Properties: amqp.NewConnectionProperties()}
	conn, err := amqp.DialConfig(q.uri, cfg)
	if err != nil {
		q.log.Errorf("connect to rmq failed, [%s]", err.Error())
		return err
	}
	q.conn = conn

	ch, err := conn.Channel()
	if err != nil {
		q.log.Errorf("rmq channel create failed, [%s]", err.Error())
		return err
	}
	q.channel = ch

	if err := ch.ExchangeDeclare(q.exchangeName, amqp.ExchangeDirect, true, false, false, false, nil); err != nil {
		if err := ch.Close(); err != nil {
			q.log.Debugf("close rmq channel raise error, [%s]", err.Error())
			return err
		}
		return err
	}

	//register queue
	rq, err := ch.QueueDeclare(q.registerQueue, false, false, false, false, nil)
	if err != nil {
		q.log.Errorf("declare register queue failed, [%s]", err.Error())
	}

	if err := ch.QueueBind(rq.Name, q.registerQueue, q.exchangeName, false, nil); err != nil {
		return err
	}

	//work queue
	wq, err := ch.QueueDeclare(q.workQueue, false, false, false, false, nil)
	if err != nil {
		q.log.Errorf("declare work queue failed, [%s]", err.Error())
	}

	if err := ch.QueueBind(wq.Name, q.workQueue, q.exchangeName, false, nil); err != nil {
		return err
	}

	//err reply queue
	erq, err := ch.QueueDeclare(q.errReplyQueue, false, false, false, false, nil)
	if err != nil {
		q.log.Errorf("declare err reply queue failed, [%s]", err.Error())
	}

	if err := ch.QueueBind(erq.Name, q.errReplyQueue, q.exchangeName, false, nil); err != nil {
		return err
	}

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
