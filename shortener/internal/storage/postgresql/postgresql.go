package postgresql

import (
	"database/sql"
	"errors"
	"fmt"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"linkify/internal/storage"
	"time"
)

type Storage struct {
	db   *gorm.DB
	conn *sql.DB
}
type URL struct {
	ID        uint      `gorm:"primaryKey"`
	Alias     string    `gorm:"unique;not null"`
	URL       string    `gorm:"not null"`
	CreatedAt time.Time `gorm:"not null;default:now()"`
}

func New(url string) (*Storage, error) {
	const op = "storage.postgresql.New"
	db, err := gorm.Open(postgres.Open(url), &gorm.Config{}) //sql.Open("pgx", url)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	conn, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("%s: failed to get database connection: %w", op, err)
	}
	conn.SetMaxOpenConns(25)
	conn.SetMaxIdleConns(25)
	conn.SetConnMaxLifetime(5 * time.Minute)

	err = db.AutoMigrate(&URL{})
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	return &Storage{db: db, conn: conn}, nil
}

func (s *Storage) Save(urlToSave string, alias string, createdAt time.Time) error {
	const op = "storage.postgresql.Save"
	url := URL{
		Alias:     alias,
		URL:       urlToSave,
		CreatedAt: createdAt,
	}
	result := s.db.Create(&url)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrDuplicatedKey) {
			return storage.ErrAliasExists
		}
		return fmt.Errorf("%s: %w", op, result.Error)
	}
	return nil
}

func (s *Storage) Get(alias string) (string, error) {
	const op = "storage.postgresql.Get"
	var url URL
	result := s.db.Where("alias = ?", alias).First(&url)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return "", storage.ErrURLNotFound
		}
		return "", fmt.Errorf("%s: %w", op, result.Error)
	}
	return url.URL, nil
}

func (s *Storage) Delete(alias string) error {
	const op = "storage.postgresql.Delete"
	result := s.db.Where("alias = ?", alias).Delete(&URL{})
	if result.Error != nil {
		return fmt.Errorf("%s: %w", op, result.Error)
	}
	if result.RowsAffected == 0 {
		return storage.ErrURLNotFound
	}
	return nil
}

func (s *Storage) Stop() error {
	if s.conn != nil {
		err := s.conn.Close()
		if err != nil {
			return fmt.Errorf("failed to close database connection: %w", err)
		}
	}
	return nil
}
