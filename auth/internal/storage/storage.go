package storage

import (
	"errors"
	"time"
)

type RefreshToken struct {
	TokenHash string    `json:"token_hash"`
	UserID    int64     `json:"user_id"`
	Email     string    `json:"email"`
	ExpiresAt time.Time `json:"expires_at"`
	CreatedAt time.Time `json:"created_at"`
}

var (
	ErrUserExists   = errors.New("user already exists")
	ErrUserNotFound = errors.New("user not found")
)
