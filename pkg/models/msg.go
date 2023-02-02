package models

import "time"

const (
	MSG_TYPE_REGISTER_MSG  = "registerMsg"
	MSG_TYPE_ERR_REPLY_MSG = "errReplyMsg"
)

type ReceiveMsg struct {
	Type    string `json:"type"`
	Content string `json:"content"`
}

type RegisterMsg struct {
	WorkerId   string    `json:"workerId"`
	OnlineTime time.Time `json:"onlineTime"`
}

type ErrReplyMsg struct {
	TaskId string `json:"taskId"`
	ErrMsg string `json:"errMsg"`
}
