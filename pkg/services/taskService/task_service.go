package taskservice

import (
	"github.com/magicLian/gocron/pkg/models"
	"github.com/magicLian/gocron/pkg/services/sqlstore"
	"github.com/magicLian/gocron/pkg/setting"
	"github.com/magicLian/logx"
)

type TaskService struct {
	cfg      *setting.Cfg
	log      logx.Logger
	sqlstore sqlstore.SqlStoreInterface
}

func ProvideNodeService(cfg *setting.Cfg, sqlstore sqlstore.SqlStoreInterface) TaskManager {
	return &TaskService{
		cfg:      cfg,
		sqlstore: sqlstore,
		log:      logx.NewLogx("task service", setting.LogxLevel),
	}
}

// GetNodeById implements NodeManager
func (t *TaskService) GetTaskById(id string) (*models.TaskDto, error) {
	return t.sqlstore.GetTaskById(id)
}

// GetNodes implements NodeManager
func (t *TaskService) GetTasks() ([]*models.TaskDto, error) {
	return t.sqlstore.GetTasks()
}
