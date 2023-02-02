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

type MasterMessageService struct {
	MessageService

	sendToWorkerJobsChan    chan *models.TaskDto
	receiveRegisterInfoChan chan *models.RegisterMsg
	receiveErrRelyChan      chan *models.ErrReplyMsg
}

func ProvideMasterMsgService(cfg *setting.Cfg) (Messager, error) {
	mmq := &MasterMessageService{
		MessageService: MessageService{
			uri:           utils.SetDefaultString(utils.GetEnvOrIniValue(cfg.Raw, "rabbitmq", "uri"), "amqp://root:123@127.0.0.1:5672/"),
			exchangeName:  utils.SetDefaultString(utils.GetEnvOrIniValue(cfg.Raw, "rabbitmq", "exchange_name"), "gocron_exchange"),
			registerQueue: utils.SetDefaultString(utils.GetEnvOrIniValue(cfg.Raw, "rabbitmq", "register_queue"), "register_q"),
			errReplyQueue: utils.SetDefaultString(utils.GetEnvOrIniValue(cfg.Raw, "rabbitmq", "err_reply_queue"), "err_r_q"),
			workQueue:     utils.SetDefaultString(utils.GetEnvOrIniValue(cfg.Raw, "rabbitmq", "work_queue"), "work_q"),

			log:     logx.NewLogx("msg-service-master", setting.LogxLevel),
			errChan: make(chan error),
		},

		sendToWorkerJobsChan:    make(chan *models.TaskDto),
		receiveRegisterInfoChan: make(chan *models.RegisterMsg),
		receiveErrRelyChan:      make(chan *models.ErrReplyMsg),
	}

	go mmq.startMaster()

	return mmq, nil
}

func (mq *MasterMessageService) startMaster() {
	mq.log.Info("Starting to connect rmq server...")
	for {
		if err := mq.connect(); err != nil {
			mq.log.Errorf("rmq connect failed, [%s]", err.Error())
			time.Sleep(10 * time.Second)
			continue
		}

		ctx, cancel := context.WithCancel(context.Background())

		go mq.KeepAlive(ctx)
		go mq.Receive(ctx, mq.registerQueue)
		go mq.Receive(ctx, mq.errReplyQueue)
		go mq.HandleSendToWorkers(ctx)

		<-mq.errChan
		mq.Close()
		cancel()
	}
}

func (mq *MasterMessageService) HandleRegisterMsg(ctx context.Context) error {
	return nil
}

func (mq *MasterMessageService) HandleErrReplyMsg(ctx context.Context) error {
	return nil
}

func (mq *MasterMessageService) HandleSendToWorkers(ctx context.Context) {
	ticker := time.NewTicker(60 * time.Second)
	defer ticker.Stop()
	defer mq.log.Info("Exit send message to worker go routine")

	for {
		select {
		case <-ctx.Done():
			mq.log.Info("Exit keepalive routine")
			return

		case job := <-mq.sendToWorkerJobsChan:
			messageData, err := json.Marshal(job)
			if err != nil {
				mq.log.Error(err.Error())
				continue
			}

			sendMsg := &models.SendMsg{
				Type:    models.MSG_TYPE_TASK_MSG,
				Content: string(messageData),
			}

			msg, err := json.Marshal(sendMsg)
			if err != nil {
				mq.log.Error(err.Error())
				continue
			}

			mq.log.Debugf("Send worker jobs message: [%s]", string(msg))

			if err = mq.Send(ctx, string(msg), ""); err != nil {
				mq.log.Error(err.Error())
				mq.errChan <- err
			}
		case <-ticker.C:
			mq.log.Info("Check unhandled message, ready to resend")

		}
	}
}

func (mq *MasterMessageService) Receive(ctx context.Context, queueName string) error {
	msgs, err := mq.channel.Consume(queueName, "", false, false, false, false, nil)
	if err != nil {
		return err
	}

	for d := range msgs {
		mq.log.Debugf("receive msg [%s]", string(d.Body))

		commonMsg := &models.ReceiveMsg{}
		if err := json.Unmarshal(d.Body, commonMsg); err != nil {
			mq.log.Debugf("Unmarshal receive msg failed, [%s]", err.Error())
			continue
		}

		switch commonMsg.Type {
		case models.MSG_TYPE_REGISTER_MSG:
			registerMsg := &models.RegisterMsg{}
			if err := json.Unmarshal([]byte(commonMsg.Content), registerMsg); err != nil {
				mq.log.Debugf("Unmarshal receive msg to register msg failed, [%s]", err.Error())
				continue
			}
			mq.receiveRegisterInfoChan <- registerMsg

			d.Ack(false)
		case models.MSG_TYPE_ERR_REPLY_MSG:
			errReplyMsg := &models.ErrReplyMsg{}
			if err := json.Unmarshal([]byte(commonMsg.Content), errReplyMsg); err != nil {
				mq.log.Debugf("Unmarshal receive msg to errReplyMsg msg failed, [%s]", err.Error())
				continue
			}
			mq.receiveErrRelyChan <- errReplyMsg

			d.Ack(false)
		default:
			mq.log.Debugf("Unmarshal receive msg failed, unsupport msg type")
			continue
		}
	}

	return nil
}
