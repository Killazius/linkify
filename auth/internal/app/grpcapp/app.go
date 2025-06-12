package grpcapp

import (
	"auth/internal/transport/rpc"
	"fmt"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"net"
)

type App struct {
	server *grpc.Server
	port   int
}

func New(port int, repo *rpc.Service) *App {
	grpcServer := grpc.NewServer()
	rpc.Register(grpcServer, repo)
	return &App{
		server: grpcServer,
		port:   port,
	}
}

func (a *App) MustRun() {
	if err := a.Run(); err != nil {
		zap.L().Fatal(err.Error())
	}
}

func (a *App) Run() error {
	const op = "grpcApp.Run"
	lis, err := net.Listen("tcp", fmt.Sprintf("0.0.0.0:%d", a.port))
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	zap.L().Info("rpc server started", zap.String("addr", lis.Addr().String()))
	if err := a.server.Serve(lis); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	return nil
}

func (a *App) Stop() {
	zap.L().Info("rpc server stopped")
	a.server.GracefulStop()
}
