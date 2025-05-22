package main

import (
	"context"
	_ "linkify/docs"
	"linkify/internal/apiserver"
	"linkify/internal/config"
	"linkify/internal/lib/logger"
	"linkify/internal/lib/logger/sl"
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
	log := logger.SetupLogger(cfg.Env)
	storage, err := postgresql.NewStorage(cfg.StorageURL)
	if err != nil {
		log.Error("failed to initialize storage", sl.Err(err))
		os.Exit(1)
	}
	redis, err := cache.NewStorage(cfg.Redis.Address, cfg.Redis.Password, cfg.Redis.DB)
	if err != nil {
		log.Error("failed to initialize cache", sl.Err(err))
		os.Exit(1)
	}
	metricsCollector := metrics.New(cfg.Prometheus, log)
	srv := apiserver.New(cfg.HTTPServer, log, storage, redis, metricsCollector)

	go srv.MustRun()
	log.Info("starting server", "address", cfg.HTTPServer.Address)
	go metricsCollector.MustRun(log)
	log.Info("starting metricsCollector", "address", cfg.Prometheus.Address)

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	<-stop
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	srv.Stop(ctx)
	metricsCollector.Stop(ctx)
	log.Info("server shutting down")
}
