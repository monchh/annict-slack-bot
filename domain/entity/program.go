package entity

import "time"

// Work represents an anime work.
type Work struct {
	Title           string
	OfficialSiteURL *string // Nullable
	ImageURL        *string // Nullable and validated
}

// Episode represents an anime episode.
type Episode struct {
	NumberText string // Formatted number, e.g., "第1話"
	Title      *string // Nullable
}

// Channel represents a broadcast channel.
type Channel struct {
	Name string
}

// Program represents a scheduled broadcast of an episode.
// This is the core entity for our use case.
type Program struct {
	Work      Work
	Episode   Episode
	Channel   Channel
	StartTime time.Time // Always in JST
}
