package sqlstore

import (
	"github.com/magicLian/gocron/pkg/models"
	"gorm.io/gorm"
)

type TaskSqlStore interface {
	GetTaskById(id string) (*models.TaskDto, error)
	GetTasks() ([]*models.TaskDto, error)
	ExistTask(name string) (bool, error)
	CreateTask(task *models.Task) error
	UpdateTask(task *models.Task) error
	DeleteTask(id string) error
}

func (s *SqlStore) GetTaskById(id string) (*models.TaskDto, error) {
	task := &models.TaskDto{}
	if err := s.db.Model(&models.Task{}).Where("id = ?", id).First(&task).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}

		return nil, err
	}

	return task, nil
}

func (s *SqlStore) GetTasks() ([]*models.TaskDto, error) {
	tasks := make([]*models.TaskDto, 0)
	if err := s.db.Model(&models.Task{}).Find(&tasks).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}

		return nil, err
	}

	return tasks, nil
}

func (s *SqlStore) ExistTask(name string) (bool, error) {
	task := &models.TaskDto{}
	if err := s.db.Model(&models.Task{}).Where("name = ?", name).First(&task).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return false, nil
		}

		return false, err
	}

	return true, nil
}

func (s *SqlStore) CreateTask(task *models.Task) error {
	return s.db.Model(&models.Task{}).Create(&task).Error
}

func (s *SqlStore) UpdateTask(task *models.Task) error {
	return s.db.Model(&models.Task{}).Updates(&task).Error
}

func (s *SqlStore) DeleteTask(id string) error {
	return s.db.Model(&models.Task{}).Where("id = ?", id).Delete(&models.Task{}).Error
}
