package main

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/monchh/annict-slack-bot/infrastructure/annict"
	"github.com/monchh/annict-slack-bot/infrastructure/config"
	"github.com/monchh/annict-slack-bot/infrastructure/httpclient"
	"github.com/monchh/annict-slack-bot/interfaces/presenter"
	"github.com/monchh/annict-slack-bot/interfaces/repository"
	"github.com/monchh/annict-slack-bot/interfaces/validator"
	"github.com/monchh/annict-slack-bot/pkg/jst"
	"github.com/monchh/annict-slack-bot/usecase"
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

	// 1. Dependency Injection
	annictAuthHttpClient := &http.Client{
		Transport: &config.AnnictAuthTransport{
			Token: cfg.AnnictToken,
			Transport: http.DefaultTransport,
		},
		Timeout:   30 * time.Second,
	}
	annictClient := annict.NewClient(annictAuthHttpClient, cfg.AnnictEndpoint, nil)
	annictRepo := repository.NewAnnictRepository(annictClient, slog.Default())
	httpValidator := validator.NewHTTPImageValidator(httpclient.NewClient(cfg.ImageCheckTimeout))
	slackPresenter := presenter.NewSlackProgramPresenter(cfg.AnnictLimitNumToDisplay)
	annictInfoGetter := usecase.NewAnnictInfoGetter(annictRepo, httpValidator)

	// 2. Prepare Slack Client
	apiClient := slack.New(cfg.SlackBotToken)
	authTest, err := apiClient.AuthTest()
	if err != nil {
		return fmt.Errorf("slack AuthTest failed: %w", err)
	}

	// 3. Execute Business Logic
	nowJST := jst.Now()
	slog.Info("Executing scheduled notification", "channel", cfg.ScheduleChannelID)
	output, err := annictInfoGetter.Execute(ctx)
	if err != nil {
		return fmt.Errorf("error executing usecase: %w", err)
	}

	// 4. Format and Post Message
	return postNotification(ctx, apiClient, cfg.ScheduleChannelID, authTest.UserID, output, slackPresenter, nowJST)
}

func postNotification(ctx context.Context, api *slack.Client, channelID, botUserID string, output *usecase.AnnictInfoGetterOutput, p *presenter.SlackProgramPresenter, date time.Time) error {
	blocks := p.FormatCombinedPrograms(output.Programs, output.LibraryEntries, date)
	mentionText := fmt.Sprintf("<@%s> %s", botUserID, annictcmd.ANNICT_TODAY)

	_, _, err := api.PostMessageContext(ctx, channelID,
		slack.MsgOptionText(mentionText, false),
		slack.MsgOptionBlocks(blocks...),
	)
	if err != nil {
		return fmt.Errorf("failed to post message: %w", err)
	}
	slog.Info("Successfully posted notification")
	return nil
}
