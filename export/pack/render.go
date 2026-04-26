package packexport

import (
	"fmt"
	"strings"
	"time"

	"github.com/FelineStateMachine/puzzletea/export/pdf"
	"github.com/FelineStateMachine/puzzletea/puzzle"
)

type RenderOptions struct {
	Title       string
	Header      string
	Advert      string
	Volume      int
	SheetLayout string
	ShuffleSeed string
	GeneratedAt time.Time
}

func BuildRenderConfig(options RenderOptions, _ []pdfexport.PackDocument) (pdfexport.RenderConfig, error) {
	if options.Volume < 1 {
		return pdfexport.RenderConfig{}, fmt.Errorf("volume must be >= 1")
	}

	layout, err := parseSheetLayout(options.SheetLayout)
	if err != nil {
		return pdfexport.RenderConfig{}, err
	}

	subtitle := strings.TrimSpace(options.Title)
	if subtitle == "" {
		subtitle = "PuzzleTea Mixed Puzzle Pack"
	}

	cfg := pdfexport.RenderConfig{
		CoverSubtitle: subtitle,
		HeaderText:    strings.TrimSpace(options.Header),
		VolumeNumber:  options.Volume,
		AdvertText:    strings.TrimSpace(options.Advert),
		GeneratedAt:   options.GeneratedAt,
		ShuffleSeed:   strings.TrimSpace(options.ShuffleSeed),
		SheetLayout:   layout,
	}
	if cfg.AdvertText == "" {
		cfg.AdvertText = DefaultAdvert
	}
	return cfg, nil
}

func AnnotatePuzzlesForPrint(puzzles []pdfexport.Puzzle, definitions []puzzle.Definition) {
	lookup := buildModeDifficultyLookup(definitions)
	for i := range puzzles {
		if puzzles[i].ActualDifficultyElo != nil || puzzles[i].TargetDifficultyElo != nil {
			continue
		}

		mode := normalizeDifficultyToken(puzzles[i].ModeSelection)
		if mode == "" || strings.Contains(mode, "mixed modes") {
			puzzles[i].DifficultyScore = 0.5
			puzzles[i].DifficultyConfidence = pdfexport.DifficultyConfidenceMedium
			puzzles[i].DifficultySource = "mixed-mode fallback"
			continue
		}

		category := normalizeDifficultyToken(puzzles[i].Category)
		modes, ok := lookup[category]
		if !ok {
			puzzles[i].DifficultyScore = 0.5
			puzzles[i].DifficultyConfidence = pdfexport.DifficultyConfidenceMedium
			puzzles[i].DifficultySource = "category lookup fallback"
			continue
		}

		score, ok := modes[mode]
		if !ok {
			puzzles[i].DifficultyScore = 0.5
			puzzles[i].DifficultyConfidence = pdfexport.DifficultyConfidenceMedium
			puzzles[i].DifficultySource = "mode lookup fallback"
			continue
		}

		puzzles[i].DifficultyScore = score
		puzzles[i].DifficultyConfidence = pdfexport.DifficultyConfidenceHigh
		puzzles[i].DifficultySource = "mode-order"
	}
}

func parseSheetLayout(raw string) (pdfexport.SheetLayout, error) {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "", "half-letter":
		return pdfexport.SheetLayoutHalfLetter, nil
	case "duplex-booklet":
		return pdfexport.SheetLayoutDuplexBooklet, nil
	default:
		return pdfexport.SheetLayoutHalfLetter, fmt.Errorf("--sheet-layout must be half-letter or duplex-booklet")
	}
}

func buildModeDifficultyLookup(definitions []puzzle.Definition) map[string]map[string]float64 {
	lookup := make(map[string]map[string]float64, len(definitions))
	for _, def := range definitions {
		scores := make(map[string]float64)
		variantTitles := []string{}
		for _, variant := range def.Variants {
			title := strings.TrimSpace(variant.Title)
			if title == "" {
				continue
			}
			variantTitles = append(variantTitles, title)
		}
		addModeOrderScores(scores, variantTitles)
		addLegacyModeOrderScores(scores, def)
		if len(scores) == 0 {
			continue
		}
		lookup[normalizeDifficultyToken(def.Name)] = scores
	}
	return lookup
}

func addLegacyModeOrderScores(scores map[string]float64, def puzzle.Definition) {
	if len(def.LegacyModes) == 0 {
		titles := make([]string, 0, len(def.Modes))
		for _, mode := range def.Modes {
			title := strings.TrimSpace(mode.Title)
			if title != "" {
				titles = append(titles, title)
			}
		}
		addModeOrderScores(scores, titles)
		return
	}

	modeIndex := make(map[puzzle.ModeID]int, len(def.Modes))
	for i, mode := range def.Modes {
		modeIndex[mode.ID] = i
	}
	for _, alias := range def.LegacyModes {
		title := strings.TrimSpace(alias.Title)
		if title == "" {
			continue
		}
		i, ok := modeIndex[alias.ID]
		if !ok || len(def.Modes) <= 1 {
			scores[normalizeDifficultyToken(title)] = 0.5
			continue
		}
		scores[normalizeDifficultyToken(title)] = float64(i) / float64(len(def.Modes)-1)
	}
}

func addModeOrderScores(scores map[string]float64, titles []string) {
	switch len(titles) {
	case 0:
		return
	case 1:
		scores[normalizeDifficultyToken(titles[0])] = 0.5
	default:
		for i, title := range titles {
			scores[normalizeDifficultyToken(title)] = float64(i) / float64(len(titles)-1)
		}
	}
}
