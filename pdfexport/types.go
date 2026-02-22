package pdfexport

import "time"

type PackMetadata struct {
	GeneratedRaw   string
	GeneratedAt    time.Time
	Version        string
	Category       string
	ModeSelection  string
	Count          int
	Seed           string
	Format         string
	SourceFileName string
}

type PackDocument struct {
	SourcePath string
	Metadata   PackMetadata
	Puzzles    []Puzzle
}

type DifficultyConfidence string

const (
	DifficultyConfidenceHigh   DifficultyConfidence = "high"
	DifficultyConfidenceMedium DifficultyConfidence = "medium"
)

type Puzzle struct {
	SourcePath           string
	SourceFileName       string
	Category             string
	ModeSelection        string
	Name                 string
	Index                int
	Body                 string
	Nonogram             *NonogramData
	Table                *GridTable
	DifficultyScore      float64
	DifficultyConfidence DifficultyConfidence
	DifficultySource     string
}

type NonogramData struct {
	Width    int
	Height   int
	RowHints [][]int
	ColHints [][]int
	Grid     [][]string
}

type GridTable struct {
	Rows         [][]string
	HasHeaderRow bool
	HasHeaderCol bool
}

type RenderConfig struct {
	Title       string
	AdvertText  string
	GeneratedAt time.Time
	ShuffleSeed string
}
