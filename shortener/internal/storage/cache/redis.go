package cache

import (
	"context"
	"fmt"
	"github.com/redis/go-redis/v9"
	"linkify/internal/storage"
	"time"
)

type Storage struct {
	client *redis.Client
}

func New(addr, password string, db int) (*Storage, error) {
	const op = "storage.cache.New"
	client := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       db,
	})
	if err := client.Ping(context.Background()).Err(); err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	return &Storage{client: client}, nil
}

func (s *Storage) Set(ctx context.Context, key string, value string, expiration time.Duration) error {
	const op = "storage.cache.Set"
	exists, err := s.client.Exists(ctx, key).Result()
	if err != nil {
		return fmt.Errorf("%s: failed to check key existence: %w", op, err)
	}
	if exists > 0 {
		return fmt.Errorf("%s: %w", op, storage.ErrAliasExists)
	}
	err = s.client.Set(ctx, key, value, expiration).Err()
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	return nil
}

func (s *Storage) Get(ctx context.Context, key string) (string, error) {
	const op = "storage.cache.Get"
	exists, err := s.client.Exists(ctx, key).Result()
	if err != nil {
		return "", fmt.Errorf("%s: failed to check key existence: %w", op, err)
	}
	if exists == 0 {
		return "", fmt.Errorf("%s: %w", op, storage.ErrAliasNotFound)
	}
	res, err := s.client.Get(ctx, key).Result()
	if err != nil {
		return "", fmt.Errorf("%s: %w", op, err)
	}
	err = s.client.Expire(ctx, key, 1*time.Hour).Err()
	if err != nil {
		return "", fmt.Errorf("%s: failed to extend key expiration: %w", op, err)
	}
	return res, nil
}
func (s *Storage) Delete(ctx context.Context, key string) error {
	const op = "storage.cache.Delete"
	err := s.client.Del(ctx, key).Err()
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	return nil
}
func (s *Storage) Stop() error {
	if s.client != nil {
		err := s.client.Close()
		if err != nil {
			return fmt.Errorf("failed to close redis client: %w", err)
		}
	}
	return nil
}
