package postgresql

import (
	"errors"
	"fmt"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"shorturl/internal/storage"
	"time"
)

type Storage struct {
	db *gorm.DB
}
type URL struct {
	ID        uint      `gorm:"primaryKey"`
	Alias     string    `gorm:"unique;not null"`
	URL       string    `gorm:"not null"`
	CreatedAt time.Time `gorm:"not null;default:now()"`
}

func NewStorage(url string) (*Storage, error) {
	const op = "storage.postgresql.NewStorage"
	db, err := gorm.Open(postgres.Open(url), &gorm.Config{}) //sql.Open("pgx", url)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	err = db.AutoMigrate(&URL{})
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	return &Storage{db: db}, nil
}

func (s *Storage) SaveURL(urlToSave string, alias string, createdAt time.Time) error {
	const op = "storage.postgresql.SaveURL"
	url := URL{
		Alias:     alias,
		URL:       urlToSave,
		CreatedAt: createdAt,
	}
	result := s.db.Create(&url)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrDuplicatedKey) {
			return storage.ErrURLExists
		}
		return fmt.Errorf("%s: %w", op, result.Error)
	}
	return nil
}

func (s *Storage) GetURL(alias string) (string, error) {
	const op = "storage.postgresql.GetURL"
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

func (s *Storage) DeleteURL(alias string) error {
	const op = "storage.postgresql.DeleteURL"
	result := s.db.Where("alias = ?", alias).Delete(&URL{})
	if result.Error != nil {
		return fmt.Errorf("%s: %w", op, result.Error)
	}
	if result.RowsAffected == 0 {
		return storage.ErrURLNotFound
	}
	return nil
}
