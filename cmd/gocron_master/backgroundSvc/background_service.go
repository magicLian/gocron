package backgroundSvc

import (
	"context"

	"github.com/magicLian/gocron/pkg/api"
	cronservice "github.com/magicLian/gocron/pkg/services/cronService"
)

type BackgroundService interface {
	Run(ctx context.Context) error
}

type BackgroundServiceRegistry struct {
	services []BackgroundService
}

func ProviceBackgroupServiceRegistry(httpServer *api.HTTPServer, cronService *cronservice.CronService) *BackgroundServiceRegistry {
	return NewBackgroundServiceRegistry(
		httpServer,
		cronService,
	)
}

func NewBackgroundServiceRegistry(services ...BackgroundService) *BackgroundServiceRegistry {
	return &BackgroundServiceRegistry{
		services: services,
	}
}

func (r *BackgroundServiceRegistry) GetServices() []BackgroundService {
	return r.services
}
