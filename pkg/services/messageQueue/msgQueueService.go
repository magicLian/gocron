package messagequeue

import (
	"context"
	"encoding/json"
	"time"

	"github.com/magicLian/gocron/pkg/models"
	"github.com/magicLian/gocron/pkg/setting"
	"github.com/magicLian/gocron/pkg/utils"
	"github.com/magicLian/logx"
	"github.com/streadway/amqp"
)

type MessageQueueService struct {
	uri           string
	exchangeName  string
	registerQueue string
	errReplyQueue string
	workQueue     string

	log            logx.Logger
	conn           *amqp.Connection
	channel        *amqp.Channel
	errChan        chan error
	workerJobsChan chan *models.TaskDto
}

func ProvideMsgQueue(cfg *setting.Cfg) (Messager, error) {
	mq := &MessageQueueService{
		uri:           utils.SetDefaultString(utils.GetEnvOrIniValue(cfg.Raw, "rabbitmq", "uri"), "amqp://root:123@127.0.0.1:5672/"),
		exchangeName:  utils.SetDefaultString(utils.GetEnvOrIniValue(cfg.Raw, "rabbitmq", "exchange_name"), "gocron_exchange"),
		registerQueue: utils.SetDefaultString(utils.GetEnvOrIniValue(cfg.Raw, "rabbitmq", "register_queue"), "register_q"),
		errReplyQueue: utils.SetDefaultString(utils.GetEnvOrIniValue(cfg.Raw, "rabbitmq", "err_reply_queue"), "err_r_q"),
		workQueue:     utils.SetDefaultString(utils.GetEnvOrIniValue(cfg.Raw, "rabbitmq", "work_queue"), "work_q"),

		log:            logx.NewLogx("mq", setting.LogxLevel),
		errChan:        make(chan error),
		workerJobsChan: make(chan *models.TaskDto),
	}

	go mq.start()

	return mq, nil
}

func (q *MessageQueueService) start() {
	q.log.Info("Starting to connect rmq server...")
	for {
		if err := q.connect(); err != nil {
			q.log.Errorf("rmq connect failed, [%s]", err.Error())
			time.Sleep(10 * time.Second)
			continue

		}

		ctx, cancel := context.WithCancel(context.Background())

		go q.KeepAlive(ctx)
		go q.HandleSendToWorkers(ctx)

		<-q.errChan
		q.Close()
		cancel()
	}
}

func (q *MessageQueueService) connect() error {
	conn, err := amqp.Dial(q.uri)
	if err != nil {
		q.log.Errorf("connect to rmq failed, [%s]", err.Error())
		return err
	}
	q.conn = conn

	channel, err := conn.Channel()
	if err != nil {
		q.log.Errorf("rmq channel create failed, [%s]", err.Error())
		return err
	}
	q.channel = channel

	if err := channel.ExchangeDeclare(q.exchangeName, "topic", true, false, false, false, nil); err != nil {
		if err := channel.Close(); err != nil {
			q.log.Debugf("close rmq channel raise error, [%s]", err.Error())
			return err
		}
		return err
	}

	return nil
}

func (q *MessageQueueService) KeepAlive(ctx context.Context) {
	for {
		err := q.channel.Publish(q.exchangeName, "heartbeat", false, false,
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

func (q *MessageQueueService) HandleSendToWorkers(ctx context.Context) {
	ticker := time.NewTicker(60 * time.Second)
	defer ticker.Stop()
	defer q.log.Info("Exit send message to worker go routine")

	for {
		select {
		case <-ctx.Done():
			q.log.Info("Exit keepalive routine")
			return

		case job := <-q.workerJobsChan:
			messageData, err := json.Marshal(job)
			if err != nil {
				q.log.Error(err.Error())
				continue
			}

			q.log.Debugf("Send worker jobs message: [%s]", string(messageData))
			if err = q.Send(string(messageData), ""); err != nil {
				q.log.Error(err.Error())
				q.errChan <- err
			}
		case <-ticker.C:
			q.log.Info("Check unhandled message, ready to resend")

		}
	}
}

func (q *MessageQueueService) Send(topic, content string) error {
	if err := q.channel.Publish(q.exchangeName, topic, false, false,
		amqp.Publishing{
			ContentType: "text/plain",
			Body:        []byte(content),
		},
	); err != nil {
		q.log.Info(err.Error())
		return err
	}

	return nil
}

func (q *MessageQueueService) Receive(topic string) ([]byte, error) {
	return nil, nil
}

func (q *MessageQueueService) Close() error {
	if err := q.channel.Close(); err != nil {
		q.log.Debugf("close rmq channel raise error,[%s]", err.Error())
	}
	if err := q.conn.Close(); err != nil {
		q.log.Debugf("close rmq connection raise error,[%s]", err.Error())
	}
	q.log.Info("Rabbitmq is closed")

	return nil
}
