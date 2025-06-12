package service

import (
	"auth/internal/lib/jwt"
	"context"
	"github.com/Killazius/linkify-proto/pkg/api"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Repository interface {
	Register(ctx context.Context, email, password string) (userID int64, err error)
	Login(ctx context.Context, email, password string) (access, refresh string, err error)
	IsAdmin(ctx context.Context, userID int64) (isAdmin bool, err error)
	RefreshTokens(ctx context.Context, refreshToken string) (newAccessToken, newRefreshToken string, err error)
	//Logout(ctx context.Context, token string) (success bool, err error)
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
