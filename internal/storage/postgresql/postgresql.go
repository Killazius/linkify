package postgresql

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	_ "github.com/jackc/pgx/v5/stdlib" // driver postgresql
	"shorturl/internal/storage"
	"time"
)

type Storage struct {
	db *sql.DB
}

func NewStorage(url string) (*Storage, error) {
	const op = "storage.postgresql.NewStorage"
	db, err := sql.Open("pgx", url)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	stmt, err := db.Prepare(`CREATE TABLE IF NOT EXISTS url (
    id SERIAL PRIMARY KEY,
    alias TEXT NOT NULL UNIQUE,
    url TEXT NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
	);`)
	if err != nil {
		return nil, err
	}
	defer stmt.Close()
	_, err = stmt.Exec()
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	stmt, err = db.Prepare(`CREATE INDEX IF NOT EXISTS idx_alias ON url(alias);`)
	if err != nil {
		return nil, err
	}
	defer stmt.Close()
	_, err = stmt.Exec()
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	return &Storage{db: db}, nil
}

func (s *Storage) SaveURL(urlToSave string, alias string, createdAt time.Time) error {
	const op = "storage.postgresql.SaveURL"
	stmt, err := s.db.Prepare(`INSERT INTO url(alias, url, created_at) VALUES ($1, $2, $3)`)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	defer stmt.Close()
	_, err = stmt.Exec(alias, urlToSave, createdAt)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return storage.ErrURLExists
		}
		return fmt.Errorf("%s: %w", op, err)
	}
	return nil
}

func (s *Storage) GetURL(alias string) (string, error) {
	const op = "storage.postgresql.GetURL"

	stmt, err := s.db.Prepare(`SELECT url FROM url WHERE alias = $1`)
	if err != nil {
		return "", fmt.Errorf("%s: %w", op, err)
	}
	defer stmt.Close()

	var url string
	err = stmt.QueryRow(alias).Scan(&url)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", storage.ErrURLNotFound
		}
		return "", fmt.Errorf("%s: %w", op, err)
	}
	return url, nil
}

func (s *Storage) DeleteURL(alias string) error {
	const op = "storage.postgresql.DeleteURL"

	stmt, err := s.db.Prepare(`DELETE FROM url WHERE alias = $1`)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	defer stmt.Close()
	result, err := stmt.Exec(alias)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	if rowsAffected == 0 {
		return storage.ErrURLNotFound
	}

	return nil
}
