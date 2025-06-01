package grpcapp

import (
	"auth/internal/service"
	"fmt"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"net"
)

type App struct {
	server *grpc.Server
	port   int
}

func New(port int, repo *service.Service) *App {
	grpcServer := grpc.NewServer()
	service.Register(grpcServer, repo)
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
	zap.L().Info("grpc server started", zap.String("addr", lis.Addr().String()))
	if err := a.server.Serve(lis); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	return nil
}

func (a *App) Stop() {

	zap.L().Info("grpc server stopped")
	a.server.GracefulStop()
}
