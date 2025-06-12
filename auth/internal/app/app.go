package app

import (
	"auth/internal/app/grpcapp"
	"auth/internal/config"
	"auth/internal/repository"
	"auth/internal/storage/postgresql"
	"auth/internal/transport/grpcapi"
	"auth/internal/transport/httpapi"
	"context"
	"go.uber.org/zap"
)

type App struct {
	log        *zap.SugaredLogger
	GRPCServer *grpcapp.App
	HTTPServer *httpapi.Server
	storage    *postgresql.Storage
}

func New(log *zap.SugaredLogger, cfg *config.Config) *App {
	storage, err := postgresql.New(cfg.StorageURL, cfg.MigrationsPath)
	if err != nil {
		log.Fatal(err)
	}
	repo := repository.New(log, storage, storage, cfg.AccessTokenTTL, cfg.RefreshTokenTTL)
	s := grpcapi.New(repo)
	GRPCServer := grpcapp.New(cfg.GRPCServer.Port, s)
	HTTPServer := httpapi.NewServer(log, repo, cfg.HTTPServer)
	return &App{
		log:        log,
		GRPCServer: GRPCServer,
		HTTPServer: HTTPServer,
		storage:    storage,
	}
}

func (a *App) Stop(ctx context.Context) {
	a.GRPCServer.Stop()
	a.HTTPServer.Stop(ctx)
	a.storage.Stop()
}
