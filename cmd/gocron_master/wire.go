//go:build wireinject
// +build wireinject

package main

import (
	"github.com/google/wire"
	"github.com/magicLian/gocron/cmd/gocron_master/backgroundSvc"
	"github.com/magicLian/gocron/pkg/api"
	"github.com/magicLian/gocron/pkg/api/middleware"
	cronService "github.com/magicLian/gocron/pkg/services/cronService"
	nodeManager "github.com/magicLian/gocron/pkg/services/nodeManager"
	"github.com/magicLian/gocron/pkg/services/sqlstore"
	taskService "github.com/magicLian/gocron/pkg/services/taskService"
	"github.com/magicLian/gocron/pkg/setting"
)

var wireSet = wire.NewSet(
	NewGoCronMasterServer,
	backgroundSvc.ProviceBackgroupServiceRegistry,
	setting.ProvideSettingCfg,
	sqlstore.ProvideSqlStore,
	api.ProvideHttpServer,
	middleware.ProvideMiddleWare,
	nodeManager.ProvideNodeService,
	cronService.ProvideCronService,
	taskService.ProvideTaskService,
)

func InitGoCronMasterWire(cmd *setting.CommandLineArgs) (*GoCronMasterServer, error) {
	wire.Build(wireSet)
	return &GoCronMasterServer{}, nil
}
