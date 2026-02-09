package store

import "time"

type GameStatus string

const (
	StatusNew        GameStatus = "new"
	StatusInProgress GameStatus = "in_progress"
	StatusCompleted  GameStatus = "completed"
	StatusAbandoned  GameStatus = "abandoned"
)

type GameRecord struct {
	ID           int64
	Name         string
	GameType     string
	Mode         string
	InitialState string
	SaveState    string
	Status       GameStatus
	CreatedAt    time.Time
	UpdatedAt    time.Time
	CompletedAt  *time.Time
}
