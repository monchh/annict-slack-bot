package repository

import (
	"context"
	"fmt"
	"log/slog"
	"reflect"
	"sort"
	"time"

	"github.com/monchh/annict-slack-bot/domain/entity"
	"github.com/monchh/annict-slack-bot/infrastructure/annict"
	"github.com/monchh/annict-slack-bot/pkg/jst"
	"github.com/monchh/annict-slack-bot/usecase"
)

// annictRepository implements the ProgramRepository interface using the Annict GraphQL API.
type annictRepository struct {
	annictAPIClient *annict.Client
	logger          *slog.Logger
}

// NewAnnictRepository creates a new repository instance.
func NewAnnictRepository(client *annict.Client, logger *slog.Logger) usecase.ProgramRepository {
	return &annictRepository{
		annictAPIClient: client,
		logger:          logger,
	}
}

func (r *annictRepository) FetchTodayPrograms(ctx context.Context) ([]*entity.Program, error) {
	r.logger.DebugContext(ctx, "Fetching unwatched programs from Annict API")
	resp, err := r.annictAPIClient.GetPrograms(ctx)
	if err != nil {
		r.logger.ErrorContext(ctx, "Failed to call GetPrograms", slog.String("error", err.Error()))
		return nil, fmt.Errorf("annictAPIClient.GetPrograms failed: %w", err)
	}

	if resp == nil || resp.Viewer == nil || resp.Viewer.Programs == nil {
		r.logger.InfoContext(ctx, "No unwatched programs data returned from Annict API or viewer/programs is nil")
		return []*entity.Program{}, nil
	}

	var programs []*entity.Program
	for _, programNode := range resp.Viewer.Programs.Nodes {
		if programNode == nil {
			continue
		}

		// mapping *annict.GetPrograms_Viewer_Programs_Nodes to *entity.Program
		domainProgram := mapAnnictProgramToDomainProgram(programNode)
		if domainProgram == nil {
			continue
		}
		programs = append(programs, domainProgram)
	}
	r.logger.InfoContext(ctx, "Successfully fetched unwatched programs", slog.Int("count", len(programs)))
	return programs, nil
}

func mapAnnictProgramToDomainProgram(node *annict.GetPrograms_Viewer_Programs_Nodes) *entity.Program {
	if node == nil {
		return nil
	}
	prog := &entity.Program{}
	if !reflect.ValueOf(node.Work).IsZero() {
		prog.Work.Title = node.Work.GetTitle()
		if node.Work.GetOfficialSiteURL() != nil && *node.Work.GetOfficialSiteURL() != "" {
			prog.Work.OfficialSiteURL = node.Work.GetOfficialSiteURL()
		}
		if node.Work.Image != nil {
			// Prefer RecommendedImageURL, if not available use FacebookOgImageURL
			if node.Work.Image.GetRecommendedImageURL() != nil && *node.Work.Image.GetRecommendedImageURL() != "" {
				prog.Work.ImageURL = node.Work.Image.GetRecommendedImageURL()
			} else if node.Work.Image.GetFacebookOgImageURL() != nil {
				prog.Work.ImageURL = node.Work.Image.GetFacebookOgImageURL()
			}
		}
	}
	if !reflect.ValueOf(node.Episode).IsZero() {
		if node.Episode.GetNumberText() != nil && *node.Episode.GetNumberText() != "" {
			prog.Episode.NumberText = *node.Episode.GetNumberText()
		} else if node.Episode.GetNumber() != nil {
			prog.Episode.NumberText = fmt.Sprintf("第%d話", *node.Episode.GetNumber())
		} else {
			prog.Episode.NumberText = "不明"
		}
		if node.Episode.GetTitle() != nil {
			prog.Episode.Title = node.Episode.GetTitle()
		}
	}
	if !reflect.ValueOf(node.Channel).IsZero() {
		prog.Channel.Name = node.Channel.GetName()
	}
	if node.StartedAt != "" {
		parsedTime, err := time.Parse(time.RFC3339, node.StartedAt)
		if err == nil {
			prog.StartTime = parsedTime
		}
	}
	return prog
}

func (r *annictRepository) FetchLibraryEntries(ctx context.Context) ([]*entity.Program, error) {
	targetSeason := jst.GetAnnictSeason(jst.Now())
	seasons := []string{targetSeason}
	r.logger.DebugContext(ctx, "Fetching library entries from Annict API", slog.String("season", targetSeason))

	resp, err := r.annictAPIClient.GetLibraryEntries(ctx, seasons)
	if err != nil {
		r.logger.ErrorContext(ctx, "Failed to call GetLibraryEntries", slog.String("error", err.Error()))
		return nil, fmt.Errorf("annictAPIClient.GetLibraryEntries failed: %w", err)
	}

	if resp == nil || resp.Viewer == nil || resp.Viewer.LibraryEntries == nil {
		r.logger.InfoContext(ctx, "No library entries data returned from Annict API or viewer/libraryEntries is nil")
		return []*entity.Program{}, nil
	}

	var programs []*entity.Program
	for _, entryNode := range resp.Viewer.LibraryEntries.Nodes {
		if entryNode == nil {
			continue
		}
		// mapping *annict.GetLibraryEntries_Viewer_LibraryEntries_Nodes to entity.Program
		domainProgram := mapAnnictLibraryEntriesToDomainProgram(entryNode)
		if domainProgram == nil {
			continue
		}

		programs = append(programs, domainProgram)
	}
	sort.Slice(programs, func(i, j int) bool {
		if programs[i].StartTime.IsZero() {
			return true
		}
		if programs[j].StartTime.IsZero() {
			return false
		}
		return programs[i].StartTime.After(programs[j].StartTime)
	})
	r.logger.InfoContext(ctx, "Successfully fetched and filtered library entries for today", slog.Int("count", len(programs)))
	return programs, nil
}

func mapAnnictLibraryEntriesToDomainProgram(node *annict.GetLibraryEntries_Viewer_LibraryEntries_Nodes) *entity.Program {
	if node == nil {
		return nil
	}
	prog := &entity.Program{}
	if !reflect.ValueOf(node.Work).IsZero() {
		prog.Work.Title = node.Work.GetTitle()
		if node.Work.GetOfficialSiteURL() != nil && *node.Work.GetOfficialSiteURL() != "" {
			prog.Work.OfficialSiteURL = node.Work.GetOfficialSiteURL()
		}
		if node.Work.Image != nil {
			// Prefer RecommendedImageURL, if not available use FacebookOgImageURL
			if node.Work.Image.GetRecommendedImageURL() != nil && *node.Work.Image.GetRecommendedImageURL() != "" {
				prog.Work.ImageURL = node.Work.Image.GetRecommendedImageURL()
			} else if node.Work.Image.GetFacebookOgImageURL() != nil {
				prog.Work.ImageURL = node.Work.Image.GetFacebookOgImageURL()
			}
		}
	}
	if node.NextEpisode != nil {
		if node.NextEpisode.GetNumberText() != nil && *node.NextEpisode.GetNumberText() != "" {
			prog.Episode.NumberText = *node.NextEpisode.GetNumberText()
		} else {
			prog.Episode.NumberText = "不明"
		}
		if node.NextEpisode.GetTitle() != nil {
			prog.Episode.Title = node.NextEpisode.GetTitle()
		}
	}
	if node.NextProgram != nil {
		if !reflect.ValueOf(node.NextProgram.Channel).IsZero() {
			prog.Channel.Name = node.NextProgram.Channel.GetName()
		}
		if node.NextProgram.StartedAt != "" {
			parsedTime, err := time.Parse(time.RFC3339, node.NextProgram.StartedAt)
			if err == nil {
				prog.StartTime = parsedTime
			}
		}
	}
	return prog
}
