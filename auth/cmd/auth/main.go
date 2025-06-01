package main

import (
	"auth/internal/app"
	"auth/internal/config"
	"auth/pkg/logger"
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	cfg := config.MustLoad()
	log, err := logger.LoadLoggerConfig(cfg.LoggerPath)
	if err != nil || log == nil {
		os.Exit(1)
	}
	application := app.New(log, cfg)
	go application.GRPCServer.MustRun()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop
	log.Info("Shutting down...")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	application.Stop(ctx)
}
