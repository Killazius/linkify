package postgreSQL

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Storage struct {
	conn *pgxpool.Pool
}

func NewStorage(url string) (*Storage, error) {
	const op = "storage.postgreSQL.NewStorage"
	conn, err := pgxpool.New(context.Background(), url)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	query := `CREATE TABLE IF NOT EXISTS url(
    	id INTEGER PRIMARY KEY,
    	alias TEXT NOT NULL UNIQUE,
    	url TEXT NOT NULL);
	CREATE INDEX IF NOT EXISTS idx_alias ON url(alias);`
	_, err = conn.Exec(context.Background(), query)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &Storage{conn: conn}, nil
}

func (s *Storage) SaveURL(urlToSave string, alias string) error {
	const op = "storage.postgreSQL.SaveURL"

	query := `INSERT INTO url(alias, url) VALUES ($1, $2)`
	_, err := s.conn.Exec(context.Background(), query, alias, urlToSave)
	if err != nil {
		// TODO: Сделать проверку, если уже есть такой alias
		return fmt.Errorf("%s: %w", op, err)
	}
	return nil
}
