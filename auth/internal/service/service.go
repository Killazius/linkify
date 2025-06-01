package service

import (
	"context"
	"github.com/Killazius/linkify-proto/pkg/api"
	"google.golang.org/grpc"
)

type Repository interface {
	Register(ctx context.Context, email, password string) (userID int64, err error)
	Login(ctx context.Context, email, password string) (token string, err error)
	IsAdmin(ctx context.Context, userID int64) (isAdmin bool, err error)
	Logout(ctx context.Context, token string) (success bool, err error)
}
type Service struct {
	repo Repository
	api.UnimplementedAuthServer
}

func New(repo Repository) *Service {
	return &Service{repo: repo}
}

func Register(gRPC *grpc.Server, service *Service) {
	api.RegisterAuthServer(gRPC, service)
}

func (s *Service) Register(ctx context.Context, req *api.RegisterRequest) (*api.RegisterResponse, error) {
	userID, err := s.repo.Register(ctx, req.Email, req.Password)
	if err != nil {
		return nil, err
	}
	return &api.RegisterResponse{UserId: userID}, nil
}
func (s *Service) Login(ctx context.Context, req *api.LoginRequest) (*api.LoginResponse, error) {
	token, err := s.repo.Login(ctx, req.Email, req.Password)
	if err != nil {
		return nil, err
	}
	return &api.LoginResponse{Token: token}, nil
}
func (s *Service) IsAdmin(ctx context.Context, req *api.IsAdminRequest) (*api.IsAdminResponse, error) {
	isAdmin, err := s.repo.IsAdmin(ctx, req.UserId)
	if err != nil {
		return nil, err
	}
	return &api.IsAdminResponse{IsAdmin: isAdmin}, nil

}
func (s *Service) Logout(ctx context.Context, req *api.LogoutRequest) (*api.LogoutResponse, error) {
	success, err := s.repo.Logout(ctx, req.Token)
	if err != nil {
		return nil, err
	}
	return &api.LogoutResponse{Success: success}, nil
}
