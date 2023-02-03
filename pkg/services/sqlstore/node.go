package sqlstore

import (
	"github.com/magicLian/gocron/pkg/models"
	"gorm.io/gorm"
)

type NodeSqlStore interface {
	GetNodeById(id string) (*models.Node, error)
	GetNodes(name, ip string) ([]*models.Node, error)
	ExistNode(name, ip string) (bool, error)
	CreateNode(node *models.CreateNode) error
	UpdateNode(node *models.UpdateNode) error
	DeleteNode(id string) error
}

func (s *SqlStore) GetNodeById(id string) (*models.Node, error) {
	node := &models.Node{}
	if err := s.db.Model(&models.Node{}).Where("id = ?", id).First(&node).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}

		return nil, err
	}

	return node, nil
}

func (s *SqlStore) GetNodes(name, ip string) ([]*models.Node, error) {
	nodes := make([]*models.Node, 0)

	db := s.db
	if name != "" {
		db = db.Where("name = ?", name)
	}
	if ip != "" {
		db = db.Where("ip = ?", ip)
	}

	if err := db.Model(&models.Node{}).Find(&nodes).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}

		return nil, err
	}

	return nodes, nil
}

func (s *SqlStore) ExistNode(name, ip string) (bool, error) {
	node := &models.Node{}
	if err := s.db.Model(&models.Node{}).Where("name = ? and ip = ?", name, ip).First(&node).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return false, nil
		}

		return false, err
	}

	return true, nil
}

func (s *SqlStore) CreateNode(node *models.CreateNode) error {
	return s.db.Model(&models.Node{}).Create(&node).Error
}

func (s *SqlStore) UpdateNode(node *models.UpdateNode) error {
	return s.db.Model(&models.Node{}).Updates(&node).Error
}

func (s *SqlStore) DeleteNode(id string) error {
	return s.db.Model(&models.Node{}).Where("id = ?", id).Delete(&models.Node{}).Error
}
