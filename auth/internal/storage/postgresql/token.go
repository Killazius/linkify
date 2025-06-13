package postgresql

import (
	"auth/internal/lib/jwt"
	"auth/internal/storage"
	"context"
	"errors"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"time"
)

func (s *Storage) StoreRefreshToken(
	ctx context.Context,
	userID string,
	token string,
	expiresAt time.Time,
) error {
	hash, err := jwt.HashToken(token)
	if err != nil {
		return errors.Join(storage.ErrTokenProcessing, err)
	}

	query := `
		INSERT INTO auth_schema.refresh_tokens 
		(token_hash, user_id, expires_at) 
		VALUES ($1, $2, $3)
		ON CONFLICT (token_hash) DO NOTHING`

	_, err = s.db.Exec(ctx, query, hash, userID, expiresAt)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23503" {
			return storage.ErrUserNotFound
		}
		return err
	}

	return nil
}

func (s *Storage) GetRefreshToken(ctx context.Context, tokenHash string) (*storage.RefreshToken, error) {
	query := `
		SELECT 
			rt.token_hash, 
			rt.user_id,
			u.email,
			rt.expires_at, 
			rt.created_at
		FROM 
			auth_schema.refresh_tokens rt
		JOIN 
			auth_schema.users u ON rt.user_id = u.id
		WHERE 
			rt.token_hash = $1`

	var rt storage.RefreshToken
	err := s.db.QueryRow(ctx, query, tokenHash).Scan(
		&rt.TokenHash,
		&rt.UserID,
		&rt.Email,
		&rt.ExpiresAt,
		&rt.CreatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, storage.ErrTokenNotFound
		}
		return nil, err
	}

	return &rt, nil
}

func (s *Storage) ValidateRefreshToken(ctx context.Context, token string) (*storage.RefreshToken, error) {
	hash, err := jwt.HashToken(token)
	if err != nil {
		return nil, errors.Join(storage.ErrTokenProcessing, err)
	}

	rt, err := s.GetRefreshToken(ctx, hash)
	if err != nil {
		return nil, err
	}

	if time.Now().After(rt.ExpiresAt) {
		return nil, storage.ErrTokenExpired
	}

	return rt, nil
}

func (s *Storage) DeleteRefreshToken(ctx context.Context, tokenHash string) error {
	query := `
		DELETE FROM auth_schema.refresh_tokens
		WHERE token_hash = $1`

	_, err := s.db.Exec(ctx, query, tokenHash)
	return err
}

func (s *Storage) DeleteRefreshTokenByUserID(ctx context.Context, userID int64) error {
	query := `
		DELETE FROM auth_schema.refresh_tokens
		WHERE user_id = $1`

	_, err := s.db.Exec(ctx, query, userID)
	return err
}

func (s *Storage) DeleteExpiredRefreshTokens(ctx context.Context) error {
	query := `
		DELETE FROM auth_schema.refresh_tokens
		WHERE expires_at < NOW()`

	_, err := s.db.Exec(ctx, query)
	return err
}
