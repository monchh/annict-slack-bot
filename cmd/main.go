package main

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	// Domain
	"github.com/monchh/annict-slack-bot/usecase"

	// Interfaces (Adapters)
	"github.com/monchh/annict-slack-bot/interfaces/presenter"
	"github.com/monchh/annict-slack-bot/interfaces/repository"
	"github.com/monchh/annict-slack-bot/interfaces/validator"

	// Infrastructure
	"github.com/monchh/annict-slack-bot/infrastructure/annict"
	"github.com/monchh/annict-slack-bot/infrastructure/config"
	"github.com/monchh/annict-slack-bot/infrastructure/httpclient"
	"github.com/monchh/annict-slack-bot/infrastructure/slack"
)

func main() {
	// Configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("FATAL: Error loading configuration: %v", err)
	}
	// Setup structured logger (slog)
	var logLevel slog.Level
	switch cfg.LogLevel {
	case "DEBUG":
		logLevel = slog.LevelDebug
	case "INFO":
		logLevel = slog.LevelInfo
	default:
		logLevel = slog.LevelInfo
	}
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: logLevel, AddSource: cfg.IsDevelopment}))
	slog.SetDefault(logger) // Set as default logger for the application
	slog.Info("Configuration loaded successfully", slog.String("logLevel", cfg.LogLevel), slog.Bool("isDevelopment", cfg.IsDevelopment))

	// Infrastructure Layer Instances
	slog.Info("Initializing Infrastructure...")
	annictAuthHttpClient := &http.Client{
		Transport: &config.AnnictAuthTransport{
			Token:     cfg.AnnictToken,
			Transport: http.DefaultTransport,
		},
		Timeout: 30 * time.Second,
	}
	annictClient := annict.NewClient(annictAuthHttpClient, cfg.AnnictEndpoint, nil)
	httpClient := httpclient.NewClient(cfg.ImageCheckTimeout)

	// Interfaces Layer Instances (Adapters)
	slog.Info("Initializing Interfaces...")
	annictRepo := repository.NewAnnictRepository(annictClient, logger)

	// Validator (shared)
	httpValidator := validator.NewHTTPImageValidator(httpClient)
	// Presenter (handles combined output)
	slackPresenter := presenter.NewSlackProgramPresenter(cfg.AnnictLimitNumToDisplay)

	// Domain Layer Instances (Use Cases)
	slog.Info("Initializing Domain...")
	// Use case for today's programs
	annictInfo := usecase.NewAnnictInfoGetter(annictRepo, httpValidator)

	// Slack Bot (Infrastructure, orchestrates everything)
	slog.Info("Initializing Slack Bot...")
	slackBot, err := slack.NewBot(
		cfg.SlackBotToken,
		cfg.SlackAppToken,
		annictInfo,
		slackPresenter,
		cfg.IsDevelopment,
	)
	if err != nil {
		slog.Error(fmt.Sprintf("Error creating Slack bot: %s", err.Error()))
	}

	// Graceful Shutdown Setup
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	// Start the Bot
	slog.Info("Starting bot...")
	err = slackBot.Run(ctx)

	// Handle Shutdown
	if err != nil && ctx.Err() == nil {
		slog.Error(fmt.Sprintf("Bot execution stopped with error: %s", err.Error()))
	} else if ctx.Err() != nil {
		slog.Info("Bot shutdown requested via signal.")
	}
	slog.Info("Shutdown complete.")
}
