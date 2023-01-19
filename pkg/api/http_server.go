package api

import (
	"context"
	"net/http"
	"time"

	"github.com/magicLian/gocron/pkg/api/middleware"
	"github.com/magicLian/gocron/pkg/setting"

	"github.com/gin-gonic/gin"
	"github.com/magicLian/logx"
)

type HTTPServer struct {
	log       logx.Logger
	context   context.Context
	ginEngine *gin.Engine

	cfg        *setting.Cfg
	middleware *middleware.MiddleWare
}

func ProvideHttpServer(cfg *setting.Cfg, middleware *middleware.MiddleWare) *HTTPServer {
	httpServer := &HTTPServer{
		log:        logx.NewLogx("http.server", setting.LogxLevel),
		cfg:        cfg,
		middleware: middleware,
		ginEngine:  gin.Default(),
	}
	return httpServer
}

func (hs *HTTPServer) Run(ctx context.Context) error {
	hs.context = ctx

	hs.ginEngine.Use(middleware.CrossRequestCheck())

	hs.apiRegister()

	httpSrv := &http.Server{
		Addr:    ":" + setting.HttpPort,
		Handler: hs.ginEngine,
	}

	// Initializing the server in a goroutine so that
	// it won't block the graceful shutdown handling below
	go func() {
		if err := httpSrv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			hs.log.Errorf("Could not start listener:[%s]", err.Error())
		}
	}()

loop:
	for {
		select {
		case <-ctx.Done():
			break loop
		default:
			time.Sleep(1000)
		}
	}
	if err := httpSrv.Shutdown(ctx); err != nil {
		hs.log.Errorf("Server forced to shutdown:[%s]", err.Error())
	}
	return nil
}
