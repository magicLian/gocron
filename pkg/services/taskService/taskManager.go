package taskservice

import "github.com/magicLian/gocron/pkg/models"

type TaskManager interface {
	GetTaskById(id string) (*models.TaskDto, error)
	GetTasks() ([]*models.TaskDto, error)
}
