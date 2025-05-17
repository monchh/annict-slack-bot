package config

import (
	"fmt"
	"log/slog"
	"os"
	"strings"
	"time"

	"github.com/joho/godotenv"
	"github.com/kelseyhightower/envconfig"
)

// Config holds application configuration.
type Config struct {
	SlackBotToken           string        `envconfig:"SLACK_BOT_TOKEN" required:"true"`
	SlackAppToken           string        `envconfig:"SLACK_APP_TOKEN" required:"true"`
	AnnictToken             string        `envconfig:"ANNICT_ACCESS_TOKEN" required:"true"`
	AnnictEndpoint          string        `envconfig:"ANNICT_ENDPOINT" default:"https://api.annict.com/graphql"`
	AnnictLimitNumToDisplay int           `envconfig:"ANNICT_LIMIT_NUM_TO_DISPLAY" default:"5"`
	LogLevel                string        `envconfig:"LOG_LEVEL" default:"info"`
	IsDevelopment           bool          `envconfig:"IS_DEVELOPMENT" default:"false"`
	ImageCheckTimeout       time.Duration `envconfig:"IMAGE_CHECK_TIMEOUT" default:"5s"`
}

// LoadConfig loads configuration from environment variables (.env fallback).
func LoadConfig() (*Config, error) {
	if err := godotenv.Load(); err != nil && !os.IsNotExist(err) {
		slog.Warn(fmt.Sprintf("Error loading .env file: %v", err))
	}

	var cfg Config
	err := envconfig.Process("", &cfg)
	if err != nil {
		return nil, err
	}
	cfg.LogLevel = strings.ToLower(cfg.LogLevel)
	return &cfg, nil
}
