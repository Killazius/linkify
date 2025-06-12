package transport

import "context"

type Repository interface {
	Register(ctx context.Context, email, password string) (userID int64, err error)
	Login(ctx context.Context, email, password string) (access, refresh string, err error)
	IsAdmin(ctx context.Context, userID int64) (isAdmin bool, err error)
	RefreshTokens(ctx context.Context, refreshToken string) (newAccessToken, newRefreshToken string, err error)
	Logout(ctx context.Context, token string) (err error)
	DeleteAccount(ctx context.Context, userID int64) error
}
