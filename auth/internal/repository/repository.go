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

type Storage interface {
	SaveUser(ctx context.Context, email string, passHash []byte) (int64, error)
	LoginUser(ctx context.Context, email string) (*domain.User, error)
	IsAdmin(ctx context.Context, userID int64) (bool, error)
	LogoutUser(ctx context.Context, token string) (bool, error)
}
type Repository struct {
	log      *zap.SugaredLogger
	storage  Storage
	tokenTTL time.Duration
}

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
)

func New(log *zap.SugaredLogger, storage Storage, tokenTTL time.Duration) *Repository {
	return &Repository{
		log:      log,
		storage:  storage,
		tokenTTL: tokenTTL,
	}
}

func (r *Repository) Register(ctx context.Context, email, password string) (int64, error) {
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return 0, fmt.Errorf("failed to hash password: %w", err)
	}
	userID, err := r.storage.SaveUser(ctx, email, passwordHash)
	if err != nil {
		if errors.Is(err, storage.ErrUserExists) {
			return 0, ErrInvalidCredentials
		}
		return 0, fmt.Errorf("failed to save user: %w", err)
	}

	return userID, nil
}
func (r *Repository) Login(ctx context.Context, email, password string) (string, error) {
	user, err := r.storage.LoginUser(ctx, email)
	if err != nil {
		if errors.Is(err, storage.ErrUserNotFound) {
			return "", ErrInvalidCredentials
		}
		return "", fmt.Errorf("failed to login: %w", err)
	}
	if err = bcrypt.CompareHashAndPassword(user.PassHash, []byte(password)); err != nil {
		return "", fmt.Errorf("failed to compare password: %w", err)
	}
	token, err := jwt.NewToken(user, r.tokenTTL)
	if err != nil {
		return "", fmt.Errorf("failed to generate token: %w", err)
	}
	return token, nil
}
func (r *Repository) IsAdmin(ctx context.Context, userID int64) (bool, error) {
	isAdmin, err := r.storage.IsAdmin(ctx, userID)
	if err != nil {
		return false, fmt.Errorf("failed to check if user is an admin: %w", err)
	}
	return isAdmin, nil
}

func (r *Repository) Logout(ctx context.Context, token string) (bool, error) {
	success, err := r.storage.LogoutUser(ctx, token)
	if err != nil {
		if errors.Is(err, storage.ErrUserNotFound) {
			return false, ErrInvalidCredentials
		}
		return false, fmt.Errorf("failed to logout: %w", err)
	}
	return success, nil
}
