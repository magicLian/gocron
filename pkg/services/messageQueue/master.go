package messagequeue

import (
	"context"
	"encoding/json"
	"time"

	"github.com/magicLian/gocron/pkg/models"
	nodemanager "github.com/magicLian/gocron/pkg/services/nodeManager"
	"github.com/magicLian/gocron/pkg/setting"
	"github.com/magicLian/gocron/pkg/utils"
	"github.com/magicLian/logx"

	amqp "github.com/rabbitmq/amqp091-go"
)

type MasterMessageService struct {
	MessageService

	registerTopic  string
	tasksTopic     string
	taskReplyTopic string

	cfg                     *setting.Cfg
	nodeService             nodemanager.NodeManager
	sendToWorkerTasksChan   chan *models.TaskDto
	receiveRegisterInfoChan chan *models.RegisterMsg
	receiveTaskReplyChan    chan *models.TaskReplyMsg
}

func ProvideMasterMsgService(cfg *setting.Cfg, nodeService nodemanager.NodeManager) (Messager, error) {
	mmq := &MasterMessageService{
		registerTopic:  utils.SetDefaultString(utils.GetEnvOrIniValue(cfg.Raw, "rabbitmq", "register_queue"), "*.register_t"),
		tasksTopic:     utils.SetDefaultString(utils.GetEnvOrIniValue(cfg.Raw, "rabbitmq", "task_topic"), "*.tasks_t"),
		taskReplyTopic: utils.SetDefaultString(utils.GetEnvOrIniValue(cfg.Raw, "rabbitmq", "task_reply_queue"), "*.task_r_t"),

		cfg:         cfg,
		nodeService: nodeService,
		MessageService: MessageService{
			uri:            utils.SetDefaultString(utils.GetEnvOrIniValue(cfg.Raw, "rabbitmq", "uri"), "amqp://root:123@127.0.0.1:5672/"),
			exchangeName:   utils.SetDefaultString(utils.GetEnvOrIniValue(cfg.Raw, "rabbitmq", "exchange_name"), "gocron_exchange"),
			registerQueue:  utils.SetDefaultString(utils.GetEnvOrIniValue(cfg.Raw, "rabbitmq", "register_queue"), "register_q"),
			taskQueue:      utils.SetDefaultString(utils.GetEnvOrIniValue(cfg.Raw, "rabbitmq", "task_queue"), "task_q"),
			taskReplyQueue: utils.SetDefaultString(utils.GetEnvOrIniValue(cfg.Raw, "rabbitmq", "task_reply_queue"), "task_r_q"),
			log:            logx.NewLogx("msg-service-master", setting.LogxLevel),
			errChan:        make(chan error),
		},

		sendToWorkerTasksChan:   make(chan *models.TaskDto),
		receiveRegisterInfoChan: make(chan *models.RegisterMsg),
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
		go mq.Receive(ctx, mq.taskReplyQueue)
		go mq.HandleRegisterMsg(ctx)
		go mq.HandleSendToWorkers(ctx)

		<-mq.errChan
		mq.Close()
		cancel()
	}
}

// connect
func (q *MasterMessageService) connect() error {
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

	if err := ch.ExchangeDeclare(q.exchangeName, amqp.ExchangeTopic, true, false, false, false, nil); err != nil {
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

	if err := ch.QueueBind(rq.Name, q.registerTopic, q.exchangeName, false, nil); err != nil {
		return err
	}

	//task queue
	wq, err := ch.QueueDeclare(q.taskQueue, false, false, false, false, nil)
	if err != nil {
		q.log.Errorf("declare task queue failed, [%s]", err.Error())
	}

	if err := ch.QueueBind(wq.Name, q.tasksTopic, q.exchangeName, false, nil); err != nil {
		return err
	}

	//err reply queue
	erq, err := ch.QueueDeclare(q.taskReplyQueue, false, false, false, false, nil)
	if err != nil {
		q.log.Errorf("declare err reply queue failed, [%s]", err.Error())
	}

	if err := ch.QueueBind(erq.Name, q.taskReplyTopic, q.exchangeName, false, nil); err != nil {
		return err
	}

	return nil
}

func (mq *MasterMessageService) HandleRegisterMsg(ctx context.Context) {
	defer mq.log.Info("Exit register worker go routine")

	for {
		select {
		case <-ctx.Done():
			mq.log.Info("Exit keepalive routine")
			return
		case registerMsg := <-mq.receiveRegisterInfoChan:
			mq.log.Debugf("Receive worker info: [%s]", registerMsg.String())

			b, err := mq.nodeService.ExistNode(registerMsg.Name, registerMsg.Ip)
			if err != nil {
				mq.log.Errorf("check worker existence failed, [%s]", err.Error())
				continue
			}
			if !b {
				if err := mq.nodeService.CreateNode(
					&models.CreateNode{Name: registerMsg.Name, Ip: registerMsg.Ip, OnlineTime: registerMsg.OnlineTime}); err != nil {
					mq.log.Errorf("create worker failed, [%s]", err.Error())
					continue
				}
			} else {
				nodes, err := mq.nodeService.GetNodes(registerMsg.Name, registerMsg.Ip)
				if err != nil {
					mq.log.Errorf("query worker info failed, [%s]", err.Error())
					continue
				}
				if err := mq.nodeService.UpdateNode(&models.UpdateNode{Id: nodes[0].Id, OnlineTime: registerMsg.OnlineTime}); err != nil {
					mq.log.Errorf("update worker info failed, [%s]", err.Error())
					continue
				}
			}
		}
	}
}

func (mq *MasterMessageService) HandleTaskReplyMsg(ctx context.Context) error {
	return nil
}

func (mq *MasterMessageService) HandleSendToWorkers(ctx context.Context) {
	defer mq.log.Info("Exit send message to worker go routine")

	for {
		select {
		case <-ctx.Done():
			mq.log.Info("Exit keepalive routine")
			return
		case job := <-mq.sendToWorkerTasksChan:
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

			mq.log.Debugf("Send worker tasks message: [%s]", string(msg))

			if err = mq.Send(ctx, string(msg), ""); err != nil {
				mq.log.Error(err.Error())
				mq.errChan <- err
			}
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

		genericMsg := &models.GenericReceiveMsg{}
		if err := json.Unmarshal(d.Body, genericMsg); err != nil {
			mq.log.Debugf("Unmarshal receive msg failed, [%s]", err.Error())
			continue
		}

		switch genericMsg.Type {
		case models.MSG_TYPE_REGISTER_MSG:
			registerMsg := &models.RegisterMsg{}
			if err := json.Unmarshal([]byte(genericMsg.Content), registerMsg); err != nil {
				mq.log.Debugf("Unmarshal receive msg to register msg failed, [%s]", err.Error())
				continue
			}
			mq.receiveRegisterInfoChan <- registerMsg

			d.Ack(false)
		case models.MSG_TYPE_TASK_REPLY_MSG:
			taskReplyMsg := &models.TaskReplyMsg{}
			if err := json.Unmarshal([]byte(genericMsg.Content), taskReplyMsg); err != nil {
				mq.log.Debugf("Unmarshal receive msg to task reply msg failed, [%s]", err.Error())
				continue
			}
			mq.receiveTaskReplyChan <- taskReplyMsg

			d.Ack(false)
		default:
			mq.log.Debugf("Unmarshal receive msg failed, unsupport msg type")
			continue
		}
	}

	return nil
}
