package nodemanager

import "github.com/magicLian/gocron/pkg/models"

type NodeManager interface {
	GetNodeById(id string) (*models.Node, error)
	GetNodes(name, ip string) ([]*models.Node, error)
	ExistNode(name, ip string) (bool, error)
	CreateNode(createNode *models.CreateNode) error
	UpdateNode(updateNode *models.UpdateNode) error
	DeleteNode(id string) error
}
