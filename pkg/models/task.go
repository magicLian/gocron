package models

import (
	"database/sql/driver"
	"encoding/json"
	"time"
)

type TaskType string

const (
	TASK_TYPE_SHELL   = TaskType("shell")
	TASK_TYPE_GO_FUNC = TaskType("go-func")
)

type Task struct {
	Id           string          `json:"id" gorm:"column:id;primary_key;not null;type:varchar(255)"`
	Name         string          `json:"name"`
	Expr         string          `json:"expr"`
	TaskType     TaskType        `json:"taskType"`
	ShellJobInfo *ShellTaskInfo  `json:"shellInfo"`
	GoFuncInfo   *GoTaskFuncInfo `json:"goFuncInfo"`
}

type ShellTaskInfo struct {
	Command    string   `json:"command"`
	Parameters []string `json:"parameters"`
}

func (sti *ShellTaskInfo) Scan(value interface{}) error {
	return json.Unmarshal([]byte(value.(string)), sti)
}

func (sti ShellTaskInfo) Value() (driver.Value, error) {
	b, err := json.Marshal(sti)
	return string(b), err
}

type GoTaskFuncInfo struct {
	FuncName   string        `json:"funcName"`
	Parameters []interface{} `json:"parameters"`
}

func (gtfi *GoTaskFuncInfo) Scan(value interface{}) error {
	return json.Unmarshal([]byte(value.(string)), gtfi)
}

func (gtfi GoTaskFuncInfo) Value() (driver.Value, error) {
	b, err := json.Marshal(gtfi)
	return string(b), err
}

type TaskDto struct {
	*Task
	NextRunTime time.Time
}
