package models

import (
	"time"

	"gorm.io/gorm"
)

const (
	NODE_STATUS_ONLINE  = "online"
	NODE_STATUS_OFFLINE = "offline"
)

type Node struct {
	Id         string         `json:"id" gorm:"column:id;primary_key;not null;type:varchar(255)"`
	Name       string         `json:"name"`
	Ip         string         `json:"ip"`
	Status     string         `json:"status"`
	CreatedAt  time.Time      `json:"createdAt"`
	OnlineTime time.Time      `json:"onlineTime"`
	UpdatedAt  time.Time      `json:"updatedAt"`
	DeletedAt  gorm.DeletedAt `json:"deletedAt"`
}

func (n *Node) IsAlive() bool {
	return n.Status == NODE_STATUS_ONLINE
}

type CreateNode struct {
	Name       string    `json:"name"`
	Ip         string    `json:"ip"`
	OnlineTime time.Time `json:"onlineTime"`
}

type UpdateNode struct {
	Id         string    `json:"id"`
	Name       string    `json:"name"`
	Ip         string    `json:"ip"`
	OnlineTime time.Time `json:"onlineTime"`
}
