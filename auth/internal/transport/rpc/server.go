package rpc

import (
	"auth/internal/lib/jwt"
	"auth/internal/transport"
	"context"
	"github.com/Killazius/linkify-proto/pkg/api"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Service struct {
	repo transport.Repository
	api.UnimplementedAuthServer
}

func New(repo transport.Repository) *Service {
	return &Service{repo: repo}
}

func Register(gRPC *grpc.Server, service *Service) {
	api.RegisterAuthServer(gRPC, service)
}

func (s *Service) ValidateToken(_ context.Context, req *api.TokenRequest) (*api.TokenResponse, error) {
	user, err := jwt.VerifyToken(req.Token)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid token")
	}
	return &api.TokenResponse{
		Valid:  true,
		UserId: user.ID,
		Email:  user.Email,
		Error:  "",
	}, nil

}
