package presenter

import (
	"fmt"
	"strings"
	"time"

	"github.com/monchh/annict-slack-bot/domain/entity"
	"github.com/monchh/annict-slack-bot/pkg/jst"
	"github.com/slack-go/slack"
)

// SlackProgramPresenter formats domain entities into Slack Block Kit blocks.
type SlackProgramPresenter struct {
	annictLimitNumToDisplay int
}

// NewSlackProgramPresenter creates a new presenter.
func NewSlackProgramPresenter(limit int) *SlackProgramPresenter {
	return &SlackProgramPresenter{
		annictLimitNumToDisplay: limit,
	}
}

// FormatCombinedPrograms formats both today's and unwatched programs.
func (p *SlackProgramPresenter) FormatCombinedPrograms(
	todaysPrograms []*entity.Program,
	unwatchedPrograms []*entity.Program,
	date time.Time,
) []slack.Block {
	todayStr := jst.FormatDate(date)
	var blocks []slack.Block

	// Header for Today's Programs
	headerText := fmt.Sprintf(":calendar: %s 放送予定のアニメ", todayStr)
	headerBlock := slack.NewHeaderBlock(slack.NewTextBlockObject(slack.PlainTextType, headerText, true, false))
	blocks = append(blocks, headerBlock)

	if len(todaysPrograms) == 0 {
		noResultsBlock := slack.NewSectionBlock(
			slack.NewTextBlockObject(slack.MarkdownType, "本日の放送予定は見つかりませんでした。", false, false),
			nil, nil,
		)
		blocks = append(blocks, noResultsBlock)
	} else {
		blocks = append(blocks, p.formatProgramList(todaysPrograms)...)
	}

	// Divider
	blocks = append(blocks, slack.NewDividerBlock())

	// Header for Unwatched Programs
	unwatchedHeaderText := ":eyes: 未視聴のアニメ"
	unwatchedHeaderBlock := slack.NewHeaderBlock(slack.NewTextBlockObject(slack.PlainTextType, unwatchedHeaderText, true, false))
	blocks = append(blocks, unwatchedHeaderBlock)

	if len(unwatchedPrograms) == 0 {
		noUnwatchedBlock := slack.NewSectionBlock(
			slack.NewTextBlockObject(slack.MarkdownType, "未視聴のアニメは見つかりませんでした。", false, false),
			nil, nil,
		)
		blocks = append(blocks, noUnwatchedBlock)
	} else {
		// Limit the number of unwatched recommendations shown.
		limit := p.annictLimitNumToDisplay
		if len(unwatchedPrograms) > limit {
			unwatchedPrograms = unwatchedPrograms[:limit]
			// Add a note that more exist?
		}
		blocks = append(blocks, p.formatProgramList(unwatchedPrograms)...)
	}

	return blocks
}

// formatProgramList formats a list of programs into blocks (used by FormatCombinedPrograms).
func (p *SlackProgramPresenter) formatProgramList(programs []*entity.Program) []slack.Block {
	var blocks []slack.Block
	for i, program := range programs {
		if i > 0 { // Add divider only within this list if needed (optional)
			// blocks = append(blocks, slack.NewDividerBlock())
		}

		// Build Text for Section Block
		var textBuilder strings.Builder
		var title string
		if program.Work.OfficialSiteURL != nil && *program.Work.OfficialSiteURL != "" {
			title = fmt.Sprintf("<%s|%s>", *program.Work.OfficialSiteURL, program.Work.Title)
		} else {
			title = fmt.Sprintf("*%s*", program.Work.Title)
		}
		textBuilder.WriteString(fmt.Sprintf("%s\n", title))

		episodeTitleStr := ""
		if program.Episode.Title != nil && *program.Episode.Title != "" {
			episodeTitleStr = fmt.Sprintf("「%s」", *program.Episode.Title)
		}
		// For unwatched, show air date as well?
		airDateTimeStr := fmt.Sprintf("%s %s", jst.FormatDate(program.StartTime), jst.FormatTime(program.StartTime))
		textBuilder.WriteString(fmt.Sprintf(" • %s %s\n",
			program.Episode.NumberText,
			episodeTitleStr,
		))
		textBuilder.WriteString(fmt.Sprintf(" • :tv: %s %s", program.Channel.Name, airDateTimeStr)) // Channel might be less relevant but okay

		sectionText := slack.NewTextBlockObject(slack.MarkdownType, textBuilder.String(), false, false)
		sectionBlock := slack.NewSectionBlock(sectionText, nil, nil)

		// Optional Image Block
		var imageBlock *slack.ImageBlock
		if program.Work.ImageURL != nil {
			imageBlock = slack.NewImageBlock(
				*program.Work.ImageURL,
				fmt.Sprintf("%s image", program.Work.Title),
				"", nil,
			)
		}

		blocks = append(blocks, sectionBlock)
		if imageBlock != nil {
			blocks = append(blocks, imageBlock)
		}
	}
	return blocks
}

// FormatError formats an error message for Slack.
func (p *SlackProgramPresenter) FormatError(err error) string {
	return fmt.Sprintf(":warning: エラーが発生しました:\n```%v```", err)
}
