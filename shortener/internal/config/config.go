package config

import (
	"fmt"
	"github.com/ilyakaznacheev/cleanenv"
	"log"
	"os"
	"time"
)

type Config struct {
	StorageURL string
	LoggerPath string     `yaml:"logger_path"`
	HTTPServer HTTPServer `yaml:"http_server"`
	Redis      Redis      `yaml:"redis"`
	Prometheus Prometheus `yaml:"prometheus"`
}

type Redis struct {
	Address  string `env:"REDIS_ADDR" env-required:"true"`
	Password string `env:"REDIS_PASSWORD" env-required:"true"`
	DB       int    `env:"REDIS_DB" env-default:"0"`
}

type HTTPServer struct {
	Address     string        `yaml:"address" env:"HTTP_ADDRESS" env-default:"8080"`
	IP          string        `env:"SERVER_IP" env-default:"localhost"`
	Timeout     time.Duration `yaml:"timeout" env:"HTTP_TIMEOUT" env-default:"4s"`
	IdleTimeout time.Duration `yaml:"idle_timeout" env:"HTTP_IDLE_TIMEOUT" env-default:"60s"`
	AliasLength int           `yaml:"alias_length"`
}
type Prometheus struct {
	Address     string        `yaml:"address" env:"PROMETHEUS_ADDRESS" env-default:"8080"`
	Timeout     time.Duration `yaml:"timeout" env-default:"4s"`
	IdleTimeout time.Duration `yaml:"idle_timeout" env-default:"60s"`
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

	return fmt.Sprintf("postgres://%s:%s@%s:%s/%s",
		user, password, host, port, db)
}
