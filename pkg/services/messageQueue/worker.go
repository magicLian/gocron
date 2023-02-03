package messagequeue

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/magicLian/gocron/pkg/models"
	"github.com/magicLian/gocron/pkg/setting"
	"github.com/magicLian/gocron/pkg/utils"
	"github.com/magicLian/logx"
)

type WorkerMessageService struct {
	MessageService

	cfg *setting.Cfg

	registerTopic  string
	tasksTopic     string
	taskReplyTopic string

	workerId   string
	workerIP   string
	workerName string

	receiveWorkerTasksChan chan *models.TaskDto
	sendTaskRelyChan       chan *models.TaskReplyMsg
}

func ProvideWorkerMsgService(cfg *setting.Cfg) (Messager, error) {
	wmq := &WorkerMessageService{
		cfg: cfg,
		MessageService: MessageService{
			uri:            utils.SetDefaultString(utils.GetEnvOrIniValue(cfg.Raw, "rabbitmq", "uri"), "amqp://root:123@127.0.0.1:5672/"),
			exchangeName:   utils.SetDefaultString(utils.GetEnvOrIniValue(cfg.Raw, "rabbitmq", "exchange_name"), "gocron_exchange"),
			registerQueue:  utils.SetDefaultString(utils.GetEnvOrIniValue(cfg.Raw, "rabbitmq", "register_queue"), "register_q"),
			taskReplyQueue: utils.SetDefaultString(utils.GetEnvOrIniValue(cfg.Raw, "rabbitmq", "task_reply_queue"), "task_r_q"),
			taskQueue:      utils.SetDefaultString(utils.GetEnvOrIniValue(cfg.Raw, "rabbitmq", "work_queue"), "work_q"),

			log:     logx.NewLogx("msg-service-worker", setting.LogxLevel),
			errChan: make(chan error),
		},

		receiveWorkerTasksChan: make(chan *models.TaskDto),
		sendTaskRelyChan:       make(chan *models.TaskReplyMsg),
	}

	workerName := wmq.GenerateWorkerName()
	wmq.workerName = workerName

	workerIP, err := wmq.GetWorkerIp()
	if err != nil {
		return nil, err
	}
	wmq.workerId = workerIP
	wmq.registerTopic = fmt.Sprintf("%s.register_t", wmq.workerName)
	wmq.tasksTopic = fmt.Sprintf("%s.tasks_t", wmq.workerName)
	wmq.taskReplyTopic = fmt.Sprintf("%s.task_r_t", wmq.workerName)

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

		wq.SendRegisterMsg(ctx)

		go wq.KeepAlive(ctx)
		go wq.Receive(ctx, wq.taskQueue)
		go wq.HandleTasks(ctx)

		<-wq.errChan
		wq.Close()
		cancel()
	}
}

func (wq *WorkerMessageService) GenerateWorkerName() string {
	name := utils.GetEnvOrIniValue(wq.cfg.Raw, "worker", "name")
	if name != "" {
		return name
	}

	hostname := os.Getenv("hostname")
	return hostname
}

func (wq *WorkerMessageService) GetWorkerIp() (string, error) {
	return utils.GetOutBoundIP()
}

func (wq *WorkerMessageService) SendRegisterMsg(ctx context.Context) error {
	ip, err := wq.GetWorkerIp()
	if err != nil {
		return err
	}
	wq.workerIP = ip

	workerName := wq.GenerateWorkerName()
	wq.workerName = workerName

	registerMsg := &models.RegisterMsg{
		Name:       workerName,
		Ip:         ip,
		OnlineTime: time.Now(),
	}

	msg, err := json.Marshal(registerMsg)
	if err != nil {
		return err
	}

	if err := wq.Send(ctx, wq.registerQueue, string(msg)); err != nil {
		return err
	}

	return nil
}

func (wq *WorkerMessageService) HandleRegisterReply(ctx context.Context) {

}

func (wq *WorkerMessageService) SendTaskReplyMsg(ctx context.Context) error {
	return nil
}

func (wq *WorkerMessageService) HandleTasks(ctx context.Context) {
	for task := range wq.receiveWorkerTasksChan {
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

		commonMsg := &models.GenericReceiveMsg{}
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
			wq.receiveWorkerTasksChan <- taskMsg

			d.Ack(false)
		default:
			wq.log.Debugf("Unmarshal receive msg failed, unsupport msg type")
			continue
		}
	}

	return nil
}
