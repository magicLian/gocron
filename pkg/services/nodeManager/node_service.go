package nodemanager

import (
	"github.com/magicLian/gocron/pkg/models"
	"github.com/magicLian/gocron/pkg/services/sqlstore"
	"github.com/magicLian/gocron/pkg/setting"
	"github.com/magicLian/logx"
)

type NodeService struct {
	cfg      *setting.Cfg
	log      logx.Logger
	sqlstore sqlstore.SqlStoreInterface
}

func ProvideNodeService(cfg *setting.Cfg, sqlstore sqlstore.SqlStoreInterface) NodeManager {
	return &NodeService{
		cfg:      cfg,
		sqlstore: sqlstore,
		log:      logx.NewLogx("node service", setting.LogxLevel),
	}
}

// GetNodeById implements NodeManager
func (n *NodeService) GetNodeById(id string) (*models.Node, error) {
	return n.sqlstore.GetNodeById(id)
}

// GetNodes implements NodeManager
func (n *NodeService) GetNodes(name, ip string) ([]*models.Node, error) {
	return n.sqlstore.GetNodes(name, ip)
}

// ExistNode implements NodeManager
func (n *NodeService) ExistNode(name, ip string) (bool, error) {
	return n.sqlstore.ExistNode(name, ip)
}

// CreateNode implements NodeManager
func (n *NodeService) CreateNode(createNode *models.CreateNode) error {
	return n.sqlstore.CreateNode(createNode)
}

// UpdateNode implements NodeManager
func (n *NodeService) UpdateNode(updateNode *models.UpdateNode) error {
	return n.sqlstore.UpdateNode(updateNode)
}

// DeleteNode implements NodeManager
func (n *NodeService) DeleteNode(id string) error {
	return n.sqlstore.DeleteNode(id)
}
