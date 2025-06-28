package app

import (
	"auth/internal/app/grpcapp"
	"auth/internal/config"
	"auth/internal/repository"
	"auth/internal/storage/postgresql"
	"auth/internal/transport/rest"
	"auth/internal/transport/rpc"
	"context"
	"go.uber.org/zap"
	"net"
)

type App struct {
	log        *zap.SugaredLogger
	GRPCServer *grpcapp.App
	HTTPServer *rest.Server
	storage    *postgresql.Storage
}

func New(log *zap.SugaredLogger, cfg *config.Config) *App {
	storage, err := postgresql.New(cfg.StorageURL, cfg.MigrationsPath)
	if err != nil {
		log.Fatal(err)
	}
	repo := repository.New(log, storage, storage, cfg.AccessTokenTTL, cfg.RefreshTokenTTL)
	s := rpc.New(repo)
	GRPCServer := grpcapp.New(net.JoinHostPort(cfg.GRPCServer.Host, cfg.GRPCServer.Port), s)
	HTTPServer := rest.NewServer(log, repo, cfg.HTTPServer)
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
