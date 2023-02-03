package models

import (
	"fmt"
	"time"
)

const (
	MSG_TYPE_REGISTER_MSG   = "registerMsg"
	MSG_TYPE_ERR_REPLY_MSG  = "errReplyMsg"
	MSG_TYPE_TASK_MSG       = "taskMsg"
	MSG_TYPE_TASK_REPLY_MSG = "taskReplyMsg"
)

type GenericReceiveMsg struct {
	Type     string `json:"type"`
	WorkerId string `json:"workerId"`
	Content  string `json:"content"`
}

type SendMsg struct {
	Type    string `json:"type"`
	Content string `json:"content"`
}

type RegisterMsg struct {
	Name       string    `json:"workerName"`
	Ip         string    `json:"ip"`
	OnlineTime time.Time `json:"onlineTime"`
}

func (rm *RegisterMsg) String() string {
	return fmt.Sprintf("worker:[%s:%s],onlineTime:[%s]", rm.Name, rm.Ip, rm.OnlineTime.Format(time.RFC3339))
}

type TaskReplyMsg struct {
	TaskId     string    `json:"taskId"`
	WorkerId   string    `json:"workerId"`
	Status     string    `json:"status"`
	StartedAt  time.Time `json:"startedAt"`
	FinishedAt time.Time `json:"finishedAt"`
	ErrMsg     string    `json:"errMsg"`
}
