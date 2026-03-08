package store

import "time"

type (
	GameStatus string
	RunKind    string
)

const (
	StatusNew        GameStatus = "new"
	StatusInProgress GameStatus = "in_progress"
	StatusCompleted  GameStatus = "completed"
	StatusAbandoned  GameStatus = "abandoned"

	RunKindNormal RunKind = "normal"
	RunKindDaily  RunKind = "daily"
	RunKindWeekly RunKind = "weekly"
	RunKindSeeded RunKind = "seeded"
)

type GameRecord struct {
	ID           int64
	Name         string
	GameID       string
	GameType     string
	ModeID       string
	Mode         string
	InitialState string
	SaveState    string
	Status       GameStatus
	RunKind      RunKind
	RunDate      *time.Time
	WeekYear     int
	WeekNumber   int
	WeekIndex    int
	SeedText     string
	CreatedAt    time.Time
	UpdatedAt    time.Time
	CompletedAt  *time.Time
}
