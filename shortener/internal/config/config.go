package config

import (
	"fmt"
	"github.com/ilyakaznacheev/cleanenv"
	"log"
	"os"
	"time"
)

type Config struct {
	Env         string `yaml:"env" env:"ENV" env-default:"local"`
	StorageURL  string `env:"STORAGE_URL" env-required:"true"`
	AliasLength int    `yaml:"alias_length" env:"ALIAS_LENGTH" env-default:"6"`
	HTTPServer  `yaml:"http_server"`
	Redis       `yaml:"redis"`
}

type Redis struct {
	Addr     string `env:"REDIS_ADDR" env-required:"true"`
	Password string `env:"REDIS_PASSWORD" env-required:"true"`
	DB       int    `env:"REDIS_DB" env-default:"0"`
}

type HTTPServer struct {
	Address     string        `yaml:"address" env:"HTTP_ADDRESS" env-default:"8080"`
	IP          string        `env:"SERVER_IP" env-default:"localhost"`
	Timeout     time.Duration `yaml:"timeout" env:"HTTP_TIMEOUT" env-default:"4s"`
	IdleTimeout time.Duration `yaml:"IdleTimeout" env:"HTTP_IDLE_TIMEOUT" env-default:"60s"`
}

func MustLoad() *Config {
	configPath := os.Getenv("CONFIG_PATH")
	if configPath == "" {
		configPath = "config.yaml"
	}

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		log.Fatalf("CONFIG_PATH does not exist: %s", configPath)
	}
	if err := os.Setenv("STORAGE_URL", buildPostgresURL()); err != nil {
		log.Printf("failed to set STORAGE_URL: %v", err)
	}

	var cfg Config
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

	return fmt.Sprintf("postgres://%s:%s@%s:%s/%s",
		user, password, host, port, db)
}
