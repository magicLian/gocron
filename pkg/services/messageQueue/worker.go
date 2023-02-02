package messagequeue

import (
	"context"
	"encoding/json"
	"time"

	"github.com/magicLian/gocron/pkg/models"
	"github.com/magicLian/gocron/pkg/setting"
	"github.com/magicLian/gocron/pkg/utils"
	"github.com/magicLian/logx"
)

type WorkerMessageService struct {
	MessageService

	receiveWorkerJobsChan chan *models.TaskDto
	sendRegisterInfoChan  chan *models.RegisterMsg
	sendErrRelyChan       chan *models.ErrReplyMsg
}

func ProvideWorkerMsgService(cfg *setting.Cfg) (Messager, error) {
	wmq := &WorkerMessageService{
		MessageService: MessageService{
			uri:           utils.SetDefaultString(utils.GetEnvOrIniValue(cfg.Raw, "rabbitmq", "uri"), "amqp://root:123@127.0.0.1:5672/"),
			exchangeName:  utils.SetDefaultString(utils.GetEnvOrIniValue(cfg.Raw, "rabbitmq", "exchange_name"), "gocron_exchange"),
			registerQueue: utils.SetDefaultString(utils.GetEnvOrIniValue(cfg.Raw, "rabbitmq", "register_queue"), "register_q"),
			errReplyQueue: utils.SetDefaultString(utils.GetEnvOrIniValue(cfg.Raw, "rabbitmq", "err_reply_queue"), "err_r_q"),
			workQueue:     utils.SetDefaultString(utils.GetEnvOrIniValue(cfg.Raw, "rabbitmq", "work_queue"), "work_q"),

			log:     logx.NewLogx("msg-service-worker", setting.LogxLevel),
			errChan: make(chan error),
		},

		receiveWorkerJobsChan: make(chan *models.TaskDto),
		sendRegisterInfoChan:  make(chan *models.RegisterMsg),
		sendErrRelyChan:       make(chan *models.ErrReplyMsg),
	}

	go wmq.startWorker()

	return wmq, nil
}

func (wq *WorkerMessageService) startWorker() {
	wq.log.Info("Starting to connect rmq server...")
	for {
		if err := wq.connect(); err != nil {
			wq.log.Errorf("rmq connect failed, [%s]", err.Error())
			time.Sleep(10 * time.Second)
			continue
		}

		ctx, cancel := context.WithCancel(context.Background())

		go wq.KeepAlive(ctx)
		go wq.Receive(ctx, wq.workQueue)
		go wq.HandleTasks(ctx)

		<-wq.errChan
		wq.Close()
		cancel()
	}
}

func (wq *WorkerMessageService) SendRegisterMsg(ctx context.Context) error {
	return nil
}

func (wq *WorkerMessageService) SendErrReplyMsg(ctx context.Context) error {
	return nil
}

func (wq *WorkerMessageService) HandleTasks(ctx context.Context) {
	for task := range wq.receiveWorkerJobsChan {
		wq.log.Debugf("start processing task[%s]", task.Name)
		time.Sleep(3 * time.Second)
	}
}

func (wq *WorkerMessageService) Receive(ctx context.Context, queueName string) error {
	msgs, err := wq.channel.Consume(queueName, "", false, false, false, false, nil)
	if err != nil {
		return err
	}

	for d := range msgs {
		wq.log.Debugf("receive msg [%s]", string(d.Body))

		commonMsg := &models.ReceiveMsg{}
		if err := json.Unmarshal(d.Body, commonMsg); err != nil {
			wq.log.Debugf("Unmarshal receive msg failed, [%s]", err.Error())
			continue
		}

		switch commonMsg.Type {
		case models.MSG_TYPE_TASK_MSG:
			taskMsg := &models.TaskDto{}
			if err := json.Unmarshal([]byte(commonMsg.Content), taskMsg); err != nil {
				wq.log.Debugf("Unmarshal receive msg to task info failed, [%s]", err.Error())
				continue
			}
			wq.receiveWorkerJobsChan <- taskMsg

			d.Ack(false)
		default:
			wq.log.Debugf("Unmarshal receive msg failed, unsupport msg type")
			continue
		}
	}

	return nil
}
