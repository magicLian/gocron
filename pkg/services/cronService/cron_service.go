package cronservice

import (
	"context"
	"os"
	"path"

	"github.com/magicLian/gocron/pkg/setting"
	"github.com/magicLian/gocron/pkg/utils"
	"github.com/magicLian/logx"
	"github.com/robfig/cron/v3"
)

type CronService struct {
	log      logx.Logger
	cfg      *setting.Cfg
	timeZone string
}

func ProvideCronService(cfg *setting.Cfg) *CronService {
	service := &CronService{
		log:      logx.NewLogx("cron service", setting.LogxLevel),
		cfg:      cfg,
		timeZone: utils.SetDefaultString(utils.GetEnvOrIniValue(cfg.Raw, "server", "time_zone"), "Asia/Shanghai"),
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
	parser := cron.NewParser(cron.Second | cron.Minute |
		cron.Hour | cron.Dom | cron.Month | cron.DowOptional | cron.Descriptor)
	cronTasks := cron.New(cron.WithParser(parser), cron.WithLocation(location))

	cronTasks.Start()

	return nil
}

func (c *CronService) setTimeZoneEnv() {
	if err := os.Setenv("ZONEINFO", path.Join(setting.HomePath, "public/tzdata/data.zip")); err != nil {
		c.log.Error("init os timezone file failed")
	}
}
