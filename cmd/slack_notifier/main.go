package main

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"os"

	"github.com/monchh/annict-slack-bot/infrastructure/config"
	"github.com/monchh/annict-slack-bot/usecase/annictcmd"
	"github.com/slack-go/slack"
)

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("FATAL: Error loading configuration: %v", err)
	}

	setupLogger(cfg)

	if err := run(context.Background(), cfg); err != nil {
		slog.Error("Execution failed", "error", err)
		os.Exit(1)
	}
}

func setupLogger(cfg *config.Config) {
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
	slog.SetDefault(logger)
}

func run(ctx context.Context, cfg *config.Config) error {
	if cfg.ScheduleChannelID == "" {
		return fmt.Errorf("SCHEDULE_CHANNEL_ID is not set")
	}

	// Prepare Slack Client
	apiClient := slack.New(cfg.SlackHomeIotToken)
	annictSlackApp := slack.New(cfg.SlackBotToken)

	authTest, err := annictSlackApp.AuthTest()
	if err != nil {
		return fmt.Errorf("slack AuthTest failed: %w", err)
	}

	// 4. Format and Post Message
	return postNotification(ctx, apiClient, cfg.ScheduleChannelID, authTest.UserID)
}

func postNotification(ctx context.Context, api *slack.Client, channelID, botUserID string) error {
	mentionText := fmt.Sprintf("<@%s> %s", botUserID, annictcmd.ANNICT_TODAY)

	_, _, err := api.PostMessageContext(ctx, channelID,
		slack.MsgOptionText(mentionText, false),
	)
	if err != nil {
		return fmt.Errorf("failed to post message: %w", err)
	}
	slog.Info("Successfully posted notification")
	return nil
}
