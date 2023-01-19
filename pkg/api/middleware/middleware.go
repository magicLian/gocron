package middleware

import (
	"github.com/magicLian/gocron/pkg/setting"
	"github.com/magicLian/logx"
)

type MiddleWare struct {
	log logx.Logger

	cfg *setting.Cfg
}

func ProvideMiddleWare(cfg *setting.Cfg) *MiddleWare {
	return &MiddleWare{
		cfg: cfg,
		log: logx.NewLogx("middleware", setting.LogxLevel),
	}
}
