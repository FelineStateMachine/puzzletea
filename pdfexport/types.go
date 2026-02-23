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
	SaveData             []byte
	PrintPayload         any
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

type NurikabeData struct {
	Width  int
	Height int
	Clues  [][]int
}

type ShikakuData struct {
	Width  int
	Height int
	Clues  [][]int
}

type HashiIsland struct {
	X        int
	Y        int
	Required int
}

type HashiData struct {
	Width   int
	Height  int
	Islands []HashiIsland
}

type HitoriData struct {
	Size    int
	Numbers [][]string
}

type TakuzuData struct {
	Size          int
	Givens        [][]string
	GroupEveryTwo bool
}

type SudokuData struct {
	Givens [9][9]int `json:"givens"`
}

type WordSearchData struct {
	Width  int        `json:"width"`
	Height int        `json:"height"`
	Grid   [][]string `json:"grid"`
	Words  []string   `json:"words"`
}

type GridTable struct {
	Rows         [][]string
	HasHeaderRow bool
	HasHeaderCol bool
}

type RGB struct{ R, G, B uint8 }

type RenderConfig struct {
	Title         string
	CoverSubtitle string
	HeaderText    string
	VolumeNumber  int
	AdvertText    string
	GeneratedAt   time.Time
	ShuffleSeed   string
	CoverColor    *RGB // nil = random vibrant nature tone
}
