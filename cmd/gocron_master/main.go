package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/magicLian/gocron/pkg/setting"
)

func main() {
	server, err := InitGoCronMasterWire(&setting.CommandLineArgs{
		Args: nil,
	})
	if err != nil {
		panic(err.Error())
	}
	go listenToSystemSignals(server)

	if err := server.Run(); err != nil {
		code := server.Exit(err)
		os.Exit(code)
	}
}

func listenToSystemSignals(server *GoCronMasterServer) {
	signalChan := make(chan os.Signal, 1)
	sighupChan := make(chan os.Signal, 1)

	signal.Notify(sighupChan, syscall.SIGHUP)
	signal.Notify(signalChan, os.Interrupt, syscall.SIGTERM)

	for {
		select {
		case <-sighupChan:
		case sig := <-signalChan:
			server.Shutdown(fmt.Sprintf("System signal: %s", sig))
		}
	}
}
