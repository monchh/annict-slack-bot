package usecase

import (
	"context"
	"fmt"

	"github.com/monchh/annict-slack-bot/domain/entity"
	"github.com/monchh/annict-slack-bot/pkg/jst"
)

// ProgramRepository defines the interface for fetching program data.
// The implementation will reside in the interfaces layer.
type ProgramRepository interface {
	// FetchTodayPrograms fetches unwatched programs. (Added to satisfy both use cases with one repository implementation)
	FetchTodayPrograms(ctx context.Context) ([]*entity.Program, error)
	// FetchLibraryEntries fetches programs, typically for the current season or based on library status.
	FetchLibraryEntries(ctx context.Context) ([]*entity.Program, error)
}

// AnnictInfoGetter defines the use case for fetching today's programs.
type AnnictInfoGetter struct {
	repo      ProgramRepository
	validator ImageValidationService
}

// ImageValidationService defines the interface for validating image URLs.
type ImageValidationService interface {
	// ValidateURL checks if the URL points to a valid, non-redirecting image.
	ValidateURL(ctx context.Context, url string) (isValid bool, validatedURL string)
}

// AnnictInfoGetterOutput holds the output of the use case.
type AnnictInfoGetterOutput struct {
	LibraryEntries []*entity.Program
	Programs       []*entity.Program
}

// NewAnnictInfoGetter creates a new instance of the use case.
func NewAnnictInfoGetter(repo ProgramRepository, validator ImageValidationService) *AnnictInfoGetter {
	return &AnnictInfoGetter{
		repo:      repo,
		validator: validator,
	}
}

// Execute runs the use case logic.
func (ag *AnnictInfoGetter) Execute(ctx context.Context) (*AnnictInfoGetterOutput, error) {
	// Call the repository method to get today programs.
	programs, err := ag.repo.FetchTodayPrograms(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to find recent unwatched programs: %w", err)
	}

	// Validate image URLs concurrently (optional optimization)
	// For simplicity, we'll do it sequentially here.
	var validatedPrograms []*entity.Program
	for _, p := range programs {
		if p.StartTime.IsZero() || !jst.IsSameDate(p.StartTime, jst.Now()) {
			continue
		}
		progToValidate := p // Create a copy to modify
		if progToValidate.Work.ImageURL != nil && *progToValidate.Work.ImageURL != "" {
			isValid, _ := ag.validator.ValidateURL(ctx, *progToValidate.Work.ImageURL)
			if !isValid {
				progToValidate.Work.ImageURL = nil // Invalidate the URL if check fails
			}
		}
		validatedPrograms = append(validatedPrograms, progToValidate)
	}

	// Call the repository method to get unwatched programs.
	libraryEntries, err := ag.repo.FetchLibraryEntries(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to find programs: %w", err)
	}

	var validatedLibraryEntries []*entity.Program
	for _, l := range libraryEntries {
		// Validate image URL
		// If p is a pointer, progToValidate still points to the original.
		// If modification of ImageURL is intended only for the output,
		// a deep copy of 'p' or at least 'p.Work' would be needed here.
		// For simplicity, assuming modification of the slice element is acceptable for now.
		progToValidate := l
		if progToValidate.Work.ImageURL != nil && *progToValidate.Work.ImageURL != "" {
			isValid, _ := ag.validator.ValidateURL(ctx, *progToValidate.Work.ImageURL)
			if !isValid {
				progToValidate.Work.ImageURL = nil
			}
		}
		validatedLibraryEntries = append(validatedLibraryEntries, progToValidate)
	}

	output := &AnnictInfoGetterOutput{
		Programs:       validatedPrograms,
		LibraryEntries: validatedLibraryEntries,
	}

	return output, nil
}
