package main

import (
	"context"
	"go.uber.org/zap"
	_ "linkify/docs"
	"linkify/internal/apiserver"
	"linkify/internal/config"
	"linkify/internal/lib/logger"
	"linkify/internal/metrics"
	"linkify/internal/storage/cache"
	"linkify/internal/storage/postgresql"
	"os"
	"os/signal"
	"syscall"
	"time"
)

// @title           Linkify
// @version         1.4
// @description     Link shortening service.

// @contact.name Telegram Developer
// @contact.url https://t.me/killazDev

// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html

// @host localhost:8080
// @BasePath /
func main() {
	cfg := config.MustLoad()

	log, err := logger.LoadLoggerConfig(cfg.LoggerPath)
	if err != nil || log == nil {
		os.Exit(1)
	}
	repo, err := postgresql.New(cfg.StorageURL)
	if err != nil {
		log.Fatal("failed to initialize storage", zap.Error(err))
	}
	redisCache, err := cache.New(cfg.Redis.Address, cfg.Redis.Password, cfg.Redis.DB)
	if err != nil {
		log.Fatal("failed to initialize cache", zap.Error(err))
	}
	metricsCollector := metrics.New(cfg.Prometheus, log)
	srv := apiserver.New(cfg.HTTPServer, log, repo, redisCache, metricsCollector)

	go srv.MustRun()
	log.Infow("starting server", "address", cfg.HTTPServer.Address)
	go metricsCollector.MustRun()
	log.Infow("starting metricsCollector", "address", cfg.Prometheus.Address)

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	<-stop
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	srv.Stop(ctx)
	metricsCollector.Stop(ctx)
	log.Info("server shutting down")
}
