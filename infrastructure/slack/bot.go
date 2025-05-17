package slack

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"strings"
	"time"

	"github.com/monchh/annict-slack-bot/domain/entity"
	"github.com/monchh/annict-slack-bot/usecase"
	"github.com/monchh/annict-slack-bot/usecase/annictcmd"

	"github.com/monchh/annict-slack-bot/pkg/jst"
	"github.com/slack-go/slack"
	"github.com/slack-go/slack/slackevents"
	"github.com/slack-go/slack/socketmode"
)

// Bot handles Slack interactions and orchestrates the use case execution.
type Bot struct {
	slackClient      *slack.Client
	socketClient     *socketmode.Client
	annictInfoGetter AnnictInfoGetter
	presenter        ProgramPresenter
	botUserID        string
}

// AnnictInfoGetter defines the method needed from the domain use case.
type AnnictInfoGetter interface {
	Execute(ctx context.Context) (*usecase.AnnictInfoGetterOutput, error)
}

// ProgramPresenter defines the methods needed from the presenter.
type ProgramPresenter interface {
	FormatCombinedPrograms(todaysPrograms []*entity.Program, unwatchedPrograms []*entity.Program, date time.Time) []slack.Block
	FormatError(err error) string
}

// NewBot creates a new Slack Bot instance.
func NewBot(
	slackBotToken, slackAppToken string,
	annictInfoGetter AnnictInfoGetter,
	presenter ProgramPresenter,
	debug bool,
) (*Bot, error) {

	apiClientOpts := []slack.Option{
		slack.OptionAppLevelToken(slackAppToken),
		slack.OptionLog(log.New(log.Writer(), "slack-api: ", log.Lshortfile|log.LstdFlags)),
	}
	if debug {
		apiClientOpts = append(apiClientOpts, slack.OptionDebug(true))
	}
	apiClient := slack.New(slackBotToken, apiClientOpts...)

	authTestResponse, err := apiClient.AuthTest()
	if err != nil {
		return nil, fmt.Errorf("slack AuthTest failed: %w", err)
	}
	botUserID := authTestResponse.UserID
	slog.Info(fmt.Sprintf("Slack Bot User ID: %s", botUserID))

	socketClientOpts := []socketmode.Option{
		socketmode.OptionLog(log.New(log.Writer(), "socketmode: ", log.Lshortfile|log.LstdFlags)),
	}
	if debug {
		socketClientOpts = append(socketClientOpts, socketmode.OptionDebug(true))
	}
	socketClient := socketmode.New(apiClient, socketClientOpts...)

	return &Bot{
		slackClient:      apiClient,
		socketClient:     socketClient,
		annictInfoGetter: annictInfoGetter,
		presenter:        presenter,
		botUserID:        botUserID,
	}, nil
}

// Run starts the bot's event loop.
func (b *Bot) Run(ctx context.Context) error {
	go b.eventHandler(ctx)

	slog.Info("Starting Slack Socket Mode client...")
	err := b.socketClient.RunContext(ctx)
	if err != nil && ctx.Err() == nil {
		slog.Info(fmt.Sprintf("Socket Mode client RunContext error: %v", err))
	}
	return err
}

// eventHandler processes events from Slack.
func (b *Bot) eventHandler(ctx context.Context) {
	slog.Info("Starting event handler loop...")
	for {
		select {
		case <-ctx.Done():
			slog.Info("Event handler loop shutting down...")
			return
		case socketEvent := <-b.socketClient.Events:
			slog.Info("Received event from Slack Socket Mode client")
			b.processEvent(ctx, socketEvent)
		}
	}
}

// processEvent routes incoming Socket Mode events.
func (b *Bot) processEvent(ctx context.Context, socketEvent socketmode.Event) {
	switch socketEvent.Type {
	case socketmode.EventTypeConnecting:
		slog.Info("Connecting to Slack...")
	case socketmode.EventTypeConnectionError:
		slog.Info("Connection failed. Retrying...")
	case socketmode.EventTypeConnected:
		slog.Info("Connected to Slack.")
	case socketmode.EventTypeHello:
		slog.Info("Received HELLO from Slack.")
	case socketmode.EventTypeDisconnect:
		slog.Info("Socket Mode client disconnected.")
	case socketmode.EventTypeEventsAPI:
		slog.Info("Socket Mode client event api.")
		eventsAPIEvent, ok := socketEvent.Data.(slackevents.EventsAPIEvent)
		if !ok {
			slog.Warn(fmt.Sprintf("Ignored unexpected EventsAPI data type: %T", socketEvent.Data))
			return
		}
		b.socketClient.Ack(*socketEvent.Request)
		b.handleEventsAPI(ctx, eventsAPIEvent)
	default:
		slog.Debug(fmt.Sprintf("Skipped event type: %s", socketEvent.Type))
	}
}

// handleEventsAPI processes specific Events API events.
func (b *Bot) handleEventsAPI(ctx context.Context, event slackevents.EventsAPIEvent) {
	switch event.Type {
	case slackevents.CallbackEvent:
		innerEvent := event.InnerEvent
		switch ev := innerEvent.Data.(type) {
		case *slackevents.AppMentionEvent:
			b.handleAppMention(ctx, ev)
		default: // Ignore other callback events
		}
	default: // Ignore other Events API types
	}
}

// handleAppMention processes mentions to the bot.
func (b *Bot) handleAppMention(ctx context.Context, event *slackevents.AppMentionEvent) {
	slog.Info(fmt.Sprintf("Received mention from user %s in channel %s with text: %q", event.User, event.Channel, event.Text))
	if event.User == b.botUserID {
		return // Ignore self
	}

	textContent := strings.ToLower(strings.TrimSpace(event.Text))
	if strings.Contains(textContent, annictcmd.ANNICT_TODAY) {
		slog.Info(fmt.Sprintf("Received command: '%s'", annictcmd.ANNICT_TODAY))

		// Execute both use cases
		var todayPrograms []*entity.Program
		var libraryEntries []*entity.Program
		var combinedErr error

		// Fetch Annict
		annictInfo, err := b.annictInfoGetter.Execute(ctx)
		if err != nil {
			slog.Info(fmt.Sprintf("Error fetching today's programs: %v", err))
			combinedErr = fmt.Errorf("annictからの情報取得エラー: %w", err)
		} else if annictInfo != nil {
			todayPrograms = annictInfo.Programs
			libraryEntries = annictInfo.LibraryEntries
		}

		// Present the results
		if combinedErr != nil && len(libraryEntries) == 0 && len(todayPrograms) == 0 {
			errorMsg := b.presenter.FormatError(combinedErr)
			b.postTextMessage(ctx, event.Channel, errorMsg)
			return
		}

		// Use the updated presenter method
		blocks := b.presenter.FormatCombinedPrograms(todayPrograms, libraryEntries, jst.Now())
		fallbackText := fmt.Sprintf("%s のアニメ情報 + 未視聴", jst.FormatDate(jst.Now()))
		b.postBlockMessage(ctx, event.Channel, fallbackText, blocks)

		if combinedErr != nil {
			// Optionally log or notify about partial errors
			slog.Info(fmt.Sprintf("Partial error occurred during fetch: %v", combinedErr))
		}
	} else {
		slog.Info(fmt.Sprintf("Receive unknown command: %s", textContent))
		b.postTextMessage(ctx, event.Channel, fmt.Sprintf("Receive unknown command: %s", textContent))
	}
}

// postTextMessage and postBlockMessage remain the same
func (b *Bot) postTextMessage(ctx context.Context, channelID, text string) {
	_, _, err := b.slackClient.PostMessageContext(ctx, channelID, slack.MsgOptionText(text, false))
	if err != nil {
		slog.Info(fmt.Sprintf("Error posting text message to channel %s: %v", channelID, err))
	}
}

func (b *Bot) postBlockMessage(ctx context.Context, channelID, fallbackText string, blocks []slack.Block) {
	_, _, err := b.slackClient.PostMessageContext(
		ctx,
		channelID,
		slack.MsgOptionBlocks(blocks...),
		slack.MsgOptionText(fallbackText, false),
	)
	if err != nil {
		slog.Info(fmt.Sprintf("Error posting block message to channel %s: %v", channelID, err))
	}
}
