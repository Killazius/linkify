package config

import (
	"fmt"
	"github.com/ilyakaznacheev/cleanenv"
	"log"
	"os"
	"time"
)

type Config struct {
	LoggerPath     string `yaml:"logger_path"`
	StorageURL     string
	GRPC           GRPCConfig    `yaml:"grpc"`
	MigrationsPath string        `yaml:"migrations_path"`
	TokenTTL       time.Duration `yaml:"token_ttl" env-default:"1h"`
}

type GRPCConfig struct {
	Port    int           `yaml:"port"`
	Timeout time.Duration `yaml:"timeout"`
}

func MustLoad() *Config {
	configPath := os.Getenv("CONFIG_PATH")
	if configPath == "" {
		configPath = "config/config.yaml"
	}

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		log.Fatalf("CONFIG_PATH does not exist: %s", configPath)
	}

	var cfg Config
	cfg.StorageURL = buildPostgresURL()
	if err := cleanenv.ReadConfig(configPath, &cfg); err != nil {
		log.Fatalf("Error loading config: %s", err)
	}
	return &cfg
}
func buildPostgresURL() string {
	user := os.Getenv("POSTGRES_USER")
	password := os.Getenv("POSTGRES_PASSWORD")
	host := os.Getenv("POSTGRES_HOST")
	if host == "" {
		host = "postgres"
	}
	port := os.Getenv("POSTGRES_PORT")
	if port == "" {
		port = "5432"
	}
	db := os.Getenv("POSTGRES_DB")
	maxConnections := os.Getenv("POSTGRES_MAX_CONNECTIONS")
	if maxConnections == "" {
		maxConnections = "10"
	}
	minConnections := os.Getenv("POSTGRES_MIN_CONNECTIONS")
	if minConnections == "" {
		minConnections = "5"
	}

	return fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable", // &pool_max_conns=%s&pool_min_conns=%s
		user, password, host, port, db)
}
