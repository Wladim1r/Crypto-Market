package config

import (
	"log/slog"
	"os"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
	"github.com/joho/godotenv"
)

type Config struct {
	Env      string      `yaml:"env" env:"ENV" env-default:"local"`
	gRPC     GRPCConfig  `yaml:"grpc"`
	Database DBConfig    `yaml:"database"`
	Token    TokenConfig `yaml:"token"`
}

type GRPCConfig struct {
	Port    int           `yaml:"port" env:"GRPC_PORT" env-default:"50051"`
	Timeout time.Duration `yaml:"timeout" env:"GRPC_TIMEOUT" env-default:"1h"`
}

type DBConfig struct {
	Host     string `yaml:"host" env:"POSTGRES_HOST" env-default:"localhost"`
	Port     int    `yaml:"port" env:"POSTGRES_PORT" env-default:"5432"`
	User     string `yaml:"user" env:"POSTGRES_USER" env-default:"postgres"`
	Password string `yaml:"password" env:"POSTGRES_PASSWORD" env-default:"postgres"`
	DBName   string `yaml:"dbname" env:"POSTGRES_DB" env-default:"users"`
}

type TokenConfig struct {
	Secret       string        `yaml:"secret" env:"KEY_SECRET" env-required:"true"`
	AccessToken  time.Duration `yaml:"access_token_ttl" env:"ACCESS_TOKEN_TTL" env-default:"15m"`
	RefreshToken time.Duration `yaml:"refresh_token_ttl" env:"REFRESH_TOKEN_TTL" env-default:"7d"`
}

func MustLoad() *Config {
	if err := godotenv.Load(); err != nil {
		slog.Info("no .env file found, reading from environment variables")
	}

	var cfg Config

	if err := cleanenv.ReadEnv(&cfg); err != nil {
		slog.Error("failed to read environment variables", "error", err)
		os.Exit(1)
	}

	return &cfg
}
