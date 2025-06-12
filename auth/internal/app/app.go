package app

import (
	"auth/internal/app/grpcapp"
	"auth/internal/config"
	"auth/internal/repository"
	"auth/internal/service"
	"auth/internal/storage/postgresql"
	"auth/internal/transport"
	"context"
	"go.uber.org/zap"
)

type App struct {
	log        *zap.SugaredLogger
	GRPCServer *grpcapp.App
	HTTPServer *transport.Server
}

func New(log *zap.SugaredLogger, cfg *config.Config) *App {
	storage, err := postgresql.New(cfg.StorageURL, cfg.MigrationsPath)
	if err != nil {
		log.Fatal(err)
	}
	repo := repository.New(log, storage, storage, cfg.AccessTokenTTL, cfg.RefreshTokenTTL)
	s := service.New(repo)
	GRPCServer := grpcapp.New(cfg.GRPCServer.Port, s)
	HTTPServer := transport.New(cfg.HTTPServer, repo, log)
	return &App{
		log:        log,
		GRPCServer: GRPCServer,
		HTTPServer: HTTPServer,
	}
}

func (a *App) Stop(ctx context.Context) {
	a.GRPCServer.Stop()
	a.HTTPServer.Stop(ctx)
}
