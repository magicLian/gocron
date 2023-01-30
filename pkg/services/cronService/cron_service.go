package cronservice

import (
	"context"
	"fmt"
	"os"
	"path"

	"github.com/magicLian/gocron/pkg/models"
	taskservice "github.com/magicLian/gocron/pkg/services/taskService"
	"github.com/magicLian/gocron/pkg/setting"
	"github.com/magicLian/gocron/pkg/utils"
	"github.com/magicLian/logx"
	"github.com/robfig/cron/v3"
)

type CronService struct {
	cron *cron.Cron

	cfg      *setting.Cfg
	taskSvc  taskservice.TaskManager
	log      logx.Logger
	timeZone string
}

func ProvideCronService(cfg *setting.Cfg, taskManager taskservice.TaskManager) *CronService {
	service := &CronService{
		cfg:      cfg,
		taskSvc:  taskManager,
		log:      logx.NewLogx("cron manager", setting.LogxLevel),
		timeZone: utils.SetDefaultString(utils.GetEnvOrIniValue(cfg.Raw, "server", "time_zone"), "Asia/Tokyo"),
	}

	return service
}

func (c *CronService) Run(ctx context.Context) error {
	c.log.Info("init cron tasks...")

	c.setTimeZoneEnv()

	location, err := utils.GetTimeLocation(c.timeZone)
	if err != nil {
		return err
	}

	parser := cron.NewParser(cron.SecondOptional | cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.DowOptional | cron.Descriptor)
	cronSvc := cron.New(cron.WithParser(parser), cron.WithLocation(location))
	c.cron = cronSvc

	if c.initJobs(); err != nil {
		return err
	}

	cronSvc.Start()

	return nil
}

func (c *CronService) setTimeZoneEnv() {
	if err := os.Setenv("ZONEINFO", path.Join(setting.HomePath, "public/tzdata/data.zip")); err != nil {
		c.log.Fatal("init os timezone file failed")
	}
}

func (c *CronService) initJobs() error {
	tasks, err := c.taskSvc.GetTasks()
	if err != nil {
		return err
	}

	for _, task := range tasks {
		switch task.TaskType {
		case models.TASK_TYPE_SHELL:
			c.initShellCronJob(task)
		case models.TASK_TYPE_GO_FUNC:
			c.initGoFuncCronJob(task)
		default:
			c.initShellCronJob(task)
		}
	}

	return nil
}

func (c *CronService) initShellCronJob(task *models.TaskDto) error {
	if task.ShellJobInfo.Command == "" {
		return fmt.Errorf("shell task command is empty")
	}

	entityId, err := c.cron.AddFunc(task.Expr, func() {
		ret, err := utils.ExecCommand(task.ShellJobInfo.Command, task.ShellJobInfo.Parameters)
		if err != nil {
			task.ErrMsg = err.Error()
			c.log.Errorf("exec task[%s] command failed, [%s]", task.Name, err.Error())
		}
		task.Result = string(ret)

		if err := c.taskSvc.UpdateTaskResult(task.Id, task.Result, task.ErrMsg); err != nil {
			c.log.Error(err.Error())
		}
	})
	if err != nil {
		return err
	}

	c.log.Infof("init shell job [%s] entityId[%d]", task.Name, entityId)
	return nil
}

func (c *CronService) initGoFuncCronJob(task *models.TaskDto) error {
	return nil
}
