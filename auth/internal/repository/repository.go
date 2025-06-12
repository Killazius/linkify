package repository

import (
	"auth/internal/domain"
	"auth/internal/lib/jwt"
	"auth/internal/storage"
	"context"
	"errors"
	"fmt"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
	"time"
)

type UserStorage interface {
	SaveUser(ctx context.Context, email string, passHash []byte) (int64, error)
	LoginUser(ctx context.Context, email string) (*domain.User, error)
	IsAdmin(ctx context.Context, userID int64) (bool, error)
	DeleteAccount(ctx context.Context, userID int64) error
}
type RefreshTokenStorage interface {
	StoreRefreshToken(ctx context.Context, userID string, token string, expiresAt time.Time) error
	GetRefreshToken(ctx context.Context, tokenHash string) (*storage.RefreshToken, error)
	ValidateRefreshToken(ctx context.Context, token string) (*storage.RefreshToken, error)
	DeleteRefreshToken(ctx context.Context, tokenHash string) error
	DeleteRefreshTokenByUserID(ctx context.Context, userID int64) error
	DeleteExpiredRefreshTokens(ctx context.Context) error
}

type Repository struct {
	log             *zap.SugaredLogger
	userStorage     UserStorage
	tokenStorage    RefreshTokenStorage
	AccessTokenTTL  time.Duration
	RefreshTokenTTL time.Duration
}

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
)

func New(log *zap.SugaredLogger, userStorage UserStorage, tokenStorage RefreshTokenStorage, AccessTokenTTL time.Duration, RefreshTokenTTL time.Duration) *Repository {
	return &Repository{
		log:             log,
		userStorage:     userStorage,
		tokenStorage:    tokenStorage,
		AccessTokenTTL:  AccessTokenTTL,
		RefreshTokenTTL: RefreshTokenTTL,
	}
}

func (r *Repository) Register(ctx context.Context, email, password string) (int64, error) {
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return 0, fmt.Errorf("failed to hash password: %w", err)
	}
	userID, err := r.userStorage.SaveUser(ctx, email, passwordHash)
	if err != nil {
		if errors.Is(err, storage.ErrUserExists) {
			return 0, ErrInvalidCredentials
		}
		return 0, fmt.Errorf("failed to save user: %w", err)
	}

	return userID, nil
}
func (r *Repository) Login(ctx context.Context, email, password string) (string, string, error) {
	user, err := r.userStorage.LoginUser(ctx, email)
	if err != nil {
		if errors.Is(err, storage.ErrUserNotFound) {
			return "", "", ErrInvalidCredentials
		}
		return "", "", fmt.Errorf("failed to login: %w", err)
	}
	if err = bcrypt.CompareHashAndPassword(user.PassHash, []byte(password)); err != nil {
		return "", "", fmt.Errorf("failed to compare password: %w", err)
	}
	accessToken, err := jwt.NewToken(user, r.AccessTokenTTL)
	if err != nil {
		return "", "", fmt.Errorf("failed to generate token: %w", err)
	}
	refreshToken, err := jwt.NewToken(user, r.RefreshTokenTTL)
	if err != nil {
		return "", "", fmt.Errorf("failed to generate token: %w", err)
	}
	err = r.tokenStorage.StoreRefreshToken(ctx, user.ID, refreshToken, time.Now().Add(r.RefreshTokenTTL))
	if err != nil {
		return "", "", fmt.Errorf("failed to store refresh token: %w", err)
	}
	return accessToken, refreshToken, nil
}
func (r *Repository) RefreshTokens(ctx context.Context, refreshToken string) (newAccessToken, newRefreshToken string, err error) {
	rt, err := r.tokenStorage.ValidateRefreshToken(ctx, refreshToken)
	if err != nil {
		return "", "", fmt.Errorf("failed to validate refresh token: %w", err)
	}

	user, err := r.userStorage.LoginUser(ctx, rt.Email)
	if err != nil {
		return "", "", fmt.Errorf("failed to get user: %w", err)
	}

	newAccessToken, err = jwt.NewToken(user, r.AccessTokenTTL)
	if err != nil {
		return "", "", fmt.Errorf("failed to generate new access token: %w", err)
	}

	newRefreshToken, err = jwt.NewToken(user, r.RefreshTokenTTL)
	if err != nil {
		return "", "", fmt.Errorf("failed to generate new refresh token: %w", err)
	}

	if err = r.tokenStorage.DeleteRefreshToken(ctx, rt.TokenHash); err != nil {
		return "", "", fmt.Errorf("failed to delete old refresh token: %w", err)
	}

	err = r.tokenStorage.StoreRefreshToken(ctx, user.ID, newRefreshToken, time.Now().Add(r.RefreshTokenTTL))
	if err != nil {
		return "", "", fmt.Errorf("failed to store new refresh token: %w", err)
	}

	return newAccessToken, newRefreshToken, nil
}
func (r *Repository) IsAdmin(ctx context.Context, userID int64) (bool, error) {
	isAdmin, err := r.userStorage.IsAdmin(ctx, userID)
	if err != nil {
		if errors.Is(err, storage.ErrUserNotFound) {
			return false, ErrInvalidCredentials
		}
		return false, fmt.Errorf("failed to check if user is an admin: %w", err)
	}
	return isAdmin, nil
}

func (r *Repository) Logout(ctx context.Context, token string) error {
	hash, err := jwt.HashToken(token)
	if err != nil {
		return fmt.Errorf("failed to hash token: %w", err)
	}

	rt, err := r.tokenStorage.GetRefreshToken(ctx, hash)
	if err != nil {
		return fmt.Errorf("failed to get refresh token: %w", err)
	}
	if rt == nil {
		return fmt.Errorf("token not found")
	}

	if err := r.tokenStorage.DeleteRefreshToken(ctx, hash); err != nil {
		return fmt.Errorf("failed to delete refresh token: %w", err)
	}

	return nil
}

func (r *Repository) DeleteAccount(ctx context.Context, userID int64) error {
	err := r.userStorage.DeleteAccount(ctx, userID)
	if err != nil {
		return fmt.Errorf("failed to delete account: %w", err)
	}
	return nil
}
