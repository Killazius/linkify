package postgresql

import (
	"auth/internal/domain"
	"auth/internal/storage"
	"context"
	"errors"
	"fmt"
	"github.com/golang-migrate/migrate"
	_ "github.com/golang-migrate/migrate/database/postgres" // db
	_ "github.com/golang-migrate/migrate/source/file"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Storage struct {
	db *pgxpool.Pool
}

func New(dbURL, migrationPath string) (*Storage, error) {
	conn, err := pgxpool.New(context.Background(), dbURL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}
	if err = Migrate(dbURL, migrationPath); err != nil {
		return nil, fmt.Errorf("failed to migrate database: %w", err)
	}
	return &Storage{db: conn}, nil
}
func Migrate(url, migrationPath string) error {
	sourceURL := "file://" + migrationPath
	m, err := migrate.New(sourceURL, url)
	if err != nil {
		return fmt.Errorf("failed to create migrate instance: %w", err)
	}
	if err = m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("failed to apply migrations: %w", err)
	}
	return nil
}

func (s *Storage) Stop() {
	s.db.Close()
}

func (s *Storage) SaveUser(ctx context.Context, email string, passHash []byte) (int64, error) {
	query := `INSERT INTO auth_schema.users (email, pass_hash) VALUES ($1, $2) RETURNING id`

	var id int64
	err := s.db.QueryRow(ctx, query, email, passHash).Scan(&id)
	if err != nil {
		var pgxErr *pgconn.PgError
		if errors.As(err, &pgxErr) && pgxErr.Code == "23505" {
			return 0, storage.ErrUserExists
		}
		return 0, err
	}
	return id, nil
}
func (s *Storage) LoginUser(ctx context.Context, email string) (*domain.User, error) {
	query := `SELECT id,email,pass_hash FROM auth_schema.users WHERE email = $1`
	user := &domain.User{}
	err := s.db.QueryRow(ctx, query, email).Scan(&user.ID, &user.Email, &user.PassHash)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, storage.ErrUserNotFound
		}
		return nil, err
	}
	return user, nil
}
func (s *Storage) IsAdmin(ctx context.Context, userID int64) (bool, error) {
	query := `SELECT is_admin FROM auth_schema.users WHERE id = $1`
	var isAdmin bool
	err := s.db.QueryRow(ctx, query, userID).Scan(&isAdmin)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return false, storage.ErrUserNotFound
		}
		return false, fmt.Errorf("failed to check for admin user: %w", err)
	}
	return isAdmin, nil
}

func (s *Storage) DeleteAccount(ctx context.Context, userID int64) error {
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer func(tx pgx.Tx, ctx context.Context) {
		err := tx.Rollback(ctx)
		if err != nil {
			return
		}
	}(tx, ctx)

	if _, err := tx.Exec(ctx, "DELETE FROM auth_schema.refresh_tokens WHERE user_id = $1", userID); err != nil {
		return fmt.Errorf("failed to delete refresh tokens: %w", err)
	}

	if _, err := tx.Exec(ctx, "DELETE FROM auth_schema.users WHERE id = $1", userID); err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}

	return tx.Commit(ctx)
}
