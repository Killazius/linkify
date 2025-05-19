package config

import (
	"github.com/ilyakaznacheev/cleanenv"
	"github.com/joho/godotenv"
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
	Timeout     time.Duration `yaml:"timeout" env:"HTTP_TIMEOUT" env-default:"4s"`
	IdleTimeout time.Duration `yaml:"IdleTimeout" env:"HTTP_IDLE_TIMEOUT" env-default:"60s"`
}

func MustLoad() *Config {
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file: %s", err)
	}
	configPath := os.Getenv("CONFIG_PATH")
	if configPath == "" {
		configPath = "config/config.yaml"
	}

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		log.Fatalf("CONFIG_PATH does not exist: %s", configPath)
	}

	var cfg Config
	if err := cleanenv.ReadConfig(configPath, &cfg); err != nil {
		log.Fatalf("Error loading config: %s", err)
	}
	return &cfg
}
