package app

import (
	"auth/internal/app/grpcapp"
	"auth/internal/config"
	"auth/internal/repository"
	"auth/internal/service"
	"auth/internal/storage/postgresql"
	"context"
	"go.uber.org/zap"
)

type App struct {
	log        *zap.SugaredLogger
	GRPCServer *grpcapp.App
}

func New(log *zap.SugaredLogger, cfg *config.Config) *App {
	storage, err := postgresql.New(cfg.StorageURL, cfg.MigrationsPath)
	if err != nil {
		log.Fatal(err)
	}
	repo := repository.New(log, storage, cfg.TokenTTL)
	s := service.New(repo)
	GRPCServer := grpcapp.New(cfg.GRPC.Port, s)
	return &App{
		log:        log,
		GRPCServer: GRPCServer,
	}
}

func (a *App) Stop(ctx context.Context) {
	a.GRPCServer.Stop()
}
