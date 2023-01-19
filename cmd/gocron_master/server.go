package main

import (
	"context"
	"errors"
	"fmt"
	"reflect"

	"github.com/magicLian/gocron/cmd/gocron_master/backgroundSvc"
	"github.com/magicLian/gocron/pkg/setting"
	"github.com/magicLian/logx"
	"golang.org/x/sync/errgroup"
)

type GoCronMasterServer struct {
	context       context.Context
	shutdownFn    context.CancelFunc
	childRoutines *errgroup.Group
	log           logx.Logger
	cfg           *setting.Cfg
	services      []backgroundSvc.BackgroundService
}

func NewGoCronMasterServer(cfg *setting.Cfg, bgs *backgroundSvc.BackgroundServiceRegistry) *GoCronMasterServer {
	rootCtx, shutdownFn := context.WithCancel(context.Background())
	childRoutines, childCtx := errgroup.WithContext(rootCtx)

	return &GoCronMasterServer{
		context:       childCtx,
		shutdownFn:    shutdownFn,
		childRoutines: childRoutines,
		log:           logx.NewLogx(fmt.Sprintf("server-%s", setting.ROLE), setting.LogxLevel),
		cfg:           cfg,
		services:      bgs.GetServices(),
	}
}

func (ar *GoCronMasterServer) Run() error {
	services := ar.services
	for _, svc := range services {
		select {
		case <-ar.context.Done():
			return ar.context.Err()
		default:
		}
		service := svc
		serviceName := reflect.TypeOf(svc).String()
		ar.childRoutines.Go(func() error {
			select {
			case <-ar.context.Done():
				return ar.context.Err()
			default:
			}
			ar.log.Debug("Starting background service", "service", serviceName)
			err := service.Run(ar.context)
			if !errors.Is(err, context.Canceled) && err != nil {
				ar.log.Error("Stopped ", "reason", err)
				return fmt.Errorf("%s run error: %s", serviceName, err.Error())
			}
			return nil
		})
	}

	return ar.childRoutines.Wait()

}

func (i *GoCronMasterServer) Shutdown(reason string) {
	i.log.Info("Shutdown started", "reason", reason)
	// call cancel func on root context
	i.shutdownFn()

	// wait for child routines
	i.childRoutines.Wait()
}

func (i *GoCronMasterServer) Exit(reason error) int {
	// default exit code is 1
	code := 1

	i.log.Error("Server shutdown", "reason", reason)
	return code
}
