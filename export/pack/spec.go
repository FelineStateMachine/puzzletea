package packexport

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"time"

	"github.com/FelineStateMachine/puzzletea/difficulty"
	"github.com/FelineStateMachine/puzzletea/export/builtinprint"
	"github.com/FelineStateMachine/puzzletea/export/pdf"
	"github.com/FelineStateMachine/puzzletea/game"
	"github.com/FelineStateMachine/puzzletea/namegen"
	"github.com/FelineStateMachine/puzzletea/puzzle"
	"github.com/FelineStateMachine/puzzletea/registry"
	"github.com/FelineStateMachine/puzzletea/resolve"
)

// Version is embedded at build time.
var Version = "dev"

var nowFn = time.Now

const PackSpecSchemaV2 = "puzzletea.pack.v2"

const (
	DefaultTitle       = "PuzzleTea Sampler"
	DefaultHeader      = "A small mixed pack that highlights a few PuzzleTea puzzle styles."
	DefaultAdvert      = "Find more puzzles at github.com/FelineStateMachine/puzzletea"
	DefaultSheetLayout = "half-letter"
)

type Spec struct {
	Schema          string                                  `json:"schema,omitempty"`
	Title           string                                  `json:"title,omitempty"`
	Header          string                                  `json:"header,omitempty"`
	Advert          string                                  `json:"advert,omitempty"`
	Volume          int                                     `json:"volume,omitempty"`
	SheetLayout     string                                  `json:"sheet_layout,omitempty"`
	Seed            string                                  `json:"seed,omitempty"`
	PDFOutputPath   string                                  `json:"pdf_output_path"`
	JSONLOutputPath string                                  `json:"jsonl_output_path,omitempty"`
	Counts          map[puzzle.GameID]map[puzzle.ModeID]int `json:"counts"`
}

type Result struct {
	TotalCount      int
	PDFOutputPath   string
	JSONLOutputPath string
}

type GameCatalog struct {
	GameID   puzzle.GameID
	GameType string
	Modes    []ModeCatalog
}

type ModeCatalog struct {
	GameID    puzzle.GameID
	GameType  string
	ModeID    puzzle.ModeID
	ModeTitle string
	Seeded    bool
	PresetElo *int
}

type modeSelection struct {
	gameID        puzzle.GameID
	gameType      string
	modeID        puzzle.ModeID
	modeTitle     string
	count         int
	spawner       game.Spawner
	seededSpawner game.SeededSpawner
	eloSpawner    game.EloSpawner
	targetElo     *difficulty.Elo
}

type runPlan struct {
	spec       Spec
	selections []modeSelection
}

func DefaultSpec(cwd string) Spec {
	if strings.TrimSpace(cwd) == "" {
		cwd = "."
	}

	spec := Spec{
		Schema:        PackSpecSchemaV2,
		Title:         DefaultTitle,
		Header:        DefaultHeader,
		Advert:        DefaultAdvert,
		Volume:        1,
		SheetLayout:   DefaultSheetLayout,
		PDFOutputPath: filepath.Join(cwd, "out", "puzzletea-export.pdf"),
		Counts:        zeroCounts(),
	}

	setDefaultCount(spec.Counts, "Sudoku", "Sudoku", 2)
	setDefaultCount(spec.Counts, "Word Search", "Word Search", 2)
	setDefaultCount(spec.Counts, "Hitori", "Hitori", 1)
	setDefaultCount(spec.Counts, "Nonogram", "Nonogram", 1)
	setDefaultCount(spec.Counts, "Nurikabe", "Nurikabe", 1)
	setDefaultCount(spec.Counts, "Takuzu", "Takuzu", 1)
	setDefaultCount(spec.Counts, "Shikaku", "Shikaku", 1)
	setDefaultCount(spec.Counts, "Spell Puzzle", "Spell Puzzle", 1)

	return spec
}

func ExportCatalog() []GameCatalog {
	entries := registry.Entries()
	catalog := make([]GameCatalog, 0, len(entries))
	for _, entry := range entries {
		if entry.Print == nil {
			continue
		}

		gameModes := make([]ModeCatalog, 0, len(entry.Variants))
		for _, variant := range entry.Variants {
			elo := int(variant.Definition.DefaultElo)
			gameModes = append(gameModes, ModeCatalog{
				GameID:    entry.Definition.ID,
				GameType:  entry.Definition.Name,
				ModeID:    puzzle.ModeID(variant.Definition.ID),
				ModeTitle: variant.Definition.Title,
				Seeded:    variant.Seeded != nil,
				PresetElo: &elo,
			})
		}

		catalog = append(catalog, GameCatalog{
			GameID:   entry.Definition.ID,
			GameType: entry.Definition.Name,
			Modes:    gameModes,
		})
	}
	return catalog
}

func LoadSpecFile(path string) (Spec, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Spec{}, fmt.Errorf("read export spec: %w", err)
	}

	var spec Spec
	if err := json.Unmarshal(data, &spec); err != nil {
		return Spec{}, fmt.Errorf("decode export spec: %w", err)
	}

	return spec, nil
}

func ValidateSpec(spec Spec) error {
	_, err := buildRunPlan(spec)
	return err
}

func Run(ctx context.Context, spec Spec) (Result, error) {
	if ctx == nil {
		ctx = context.Background()
	}

	builtinprint.Register()

	plan, err := buildRunPlan(spec)
	if err != nil {
		return Result{}, err
	}

	generatedAt := nowFn()
	records, err := generateRecords(ctx, plan.selections, strings.TrimSpace(plan.spec.Seed), generatedAt)
	if err != nil {
		return Result{}, err
	}

	result := Result{
		TotalCount: len(records),
	}
	if err := ctx.Err(); err != nil {
		return Result{}, err
	}

	var stagedJSONL stagedOutput
	hasJSONL := strings.TrimSpace(plan.spec.JSONLOutputPath) != ""
	if hasJSONL {
		stagedJSONL, err = stageJSONL(plan.spec.JSONLOutputPath, records)
		if err != nil {
			return Result{}, err
		}
		defer stagedJSONL.Cleanup()
	}

	if err := ctx.Err(); err != nil {
		return Result{}, err
	}

	doc, err := pdfexport.ParseJSONLRecords("packexport-generated.jsonl", records)
	if err != nil {
		return Result{}, err
	}

	docs := []pdfexport.PackDocument{doc}
	puzzles := flattenPuzzles(docs)
	AnnotatePuzzlesForPrint(puzzles, registry.Definitions())

	shuffleSeed := strings.TrimSpace(plan.spec.Seed)
	if shuffleSeed == "" {
		shuffleSeed = generatedAt.Format(time.RFC3339Nano)
	}
	ordered := pdfexport.OrderPuzzlesForPrint(puzzles, shuffleSeed)

	cfg, err := BuildRenderConfig(RenderOptions{
		Title:       plan.spec.Title,
		Header:      plan.spec.Header,
		Advert:      plan.spec.Advert,
		Volume:      plan.spec.Volume,
		SheetLayout: plan.spec.SheetLayout,
		ShuffleSeed: shuffleSeed,
		GeneratedAt: generatedAt,
	}, docs)
	if err != nil {
		return Result{}, err
	}

	stagedPDF, err := stagePDF(plan.spec.PDFOutputPath, docs, ordered, cfg)
	if err != nil {
		return Result{}, err
	}
	defer stagedPDF.Cleanup()

	if err := ctx.Err(); err != nil {
		return Result{}, err
	}
	if err := stagedPDF.Publish(); err != nil {
		return Result{}, err
	}
	result.PDFOutputPath = plan.spec.PDFOutputPath

	if hasJSONL {
		if err := stagedJSONL.Publish(); err != nil {
			return Result{}, err
		}
		result.JSONLOutputPath = plan.spec.JSONLOutputPath
	}

	return result, nil
}

func zeroCounts() map[puzzle.GameID]map[puzzle.ModeID]int {
	counts := make(map[puzzle.GameID]map[puzzle.ModeID]int)
	for _, game := range ExportCatalog() {
		gameCounts := make(map[puzzle.ModeID]int, len(game.Modes))
		for _, mode := range game.Modes {
			gameCounts[mode.ModeID] = 0
		}
		counts[game.GameID] = gameCounts
	}
	return counts
}

func setDefaultCount(counts map[puzzle.GameID]map[puzzle.ModeID]int, gameName, modeTitle string, count int) {
	gameID := puzzle.CanonicalGameID(gameName)
	modeID := puzzle.CanonicalModeID(modeTitle)
	gameCounts, ok := counts[gameID]
	if !ok {
		return
	}
	if _, ok := gameCounts[modeID]; !ok {
		return
	}
	gameCounts[modeID] = count
}

func buildRunPlan(spec Spec) (runPlan, error) {
	spec = normalizeSpec(spec)

	if spec.Schema != PackSpecSchemaV2 {
		return runPlan{}, fmt.Errorf("unsupported pack spec schema %q", spec.Schema)
	}
	if !strings.EqualFold(filepath.Ext(spec.PDFOutputPath), ".pdf") {
		return runPlan{}, fmt.Errorf("pdf output path must use a .pdf extension")
	}
	if strings.TrimSpace(spec.JSONLOutputPath) != "" && !strings.EqualFold(filepath.Ext(spec.JSONLOutputPath), ".jsonl") {
		return runPlan{}, fmt.Errorf("jsonl output path must use a .jsonl extension")
	}
	if spec.Volume < 1 {
		return runPlan{}, fmt.Errorf("volume must be >= 1")
	}

	entries := registry.Entries()
	entryByID := make(map[puzzle.GameID]registry.Entry, len(entries))
	for _, entry := range entries {
		entryByID[entry.Definition.ID] = entry
	}

	selections := []modeSelection{}
	total := 0
	for gameID, modes := range spec.Counts {
		entry, ok := entryByID[gameID]
		if !ok {
			for modeID, count := range modes {
				if count != 0 {
					return runPlan{}, fmt.Errorf("%s / %s is not a known export target", gameID, modeID)
				}
			}
			continue
		}
		if entry.Print == nil {
			for modeID, count := range modes {
				if count != 0 {
					return runPlan{}, fmt.Errorf("%s / %s does not support export", entry.Definition.Name, modeID)
				}
			}
			continue
		}

		variantByID := make(map[puzzle.ModeID]registry.VariantEntry, len(entry.Variants))
		for _, variant := range entry.Variants {
			variantByID[puzzle.ModeID(variant.Definition.ID)] = variant
		}
		legacyByID := make(map[puzzle.ModeID]puzzle.LegacyModeAlias, len(entry.LegacyModes))
		for _, alias := range entry.LegacyModes {
			legacyByID[alias.ID] = alias
		}

		for modeID, count := range modes {
			if count < 0 {
				return runPlan{}, fmt.Errorf("%s / %s count must be >= 0", entry.Definition.Name, modeID)
			}
			if count == 0 {
				continue
			}

			variant, targetElo, ok := exportVariantForModeID(entry, variantByID, legacyByID, modeID)
			if !ok {
				return runPlan{}, fmt.Errorf("%s / %s is not a known export variant", entry.Definition.Name, modeID)
			}
			if strings.TrimSpace(spec.Seed) != "" && variant.Seeded == nil {
				return runPlan{}, fmt.Errorf("%s / %s does not support seeded export", entry.Definition.Name, variant.Definition.Title)
			}

			total += count
			selections = append(selections, modeSelection{
				gameID:        entry.Definition.ID,
				gameType:      entry.Definition.Name,
				modeID:        puzzle.ModeID(variant.Definition.ID),
				modeTitle:     variant.Definition.Title,
				count:         count,
				spawner:       variant.Seeded,
				seededSpawner: variant.Seeded,
				eloSpawner:    variant.Elo,
				targetElo:     &targetElo,
			})
		}
	}

	if total == 0 {
		return runPlan{}, fmt.Errorf("total exported puzzle count must be > 0")
	}

	slices.SortFunc(selections, func(a, b modeSelection) int {
		if c := strings.Compare(puzzle.NormalizeName(a.gameType), puzzle.NormalizeName(b.gameType)); c != 0 {
			return c
		}
		return strings.Compare(puzzle.NormalizeName(a.modeTitle), puzzle.NormalizeName(b.modeTitle))
	})

	return runPlan{
		spec:       spec,
		selections: selections,
	}, nil
}

func exportVariantForModeID(
	entry registry.Entry,
	variantByID map[puzzle.ModeID]registry.VariantEntry,
	legacyByID map[puzzle.ModeID]puzzle.LegacyModeAlias,
	modeID puzzle.ModeID,
) (registry.VariantEntry, difficulty.Elo, bool) {
	if variant, ok := variantByID[modeID]; ok {
		return variant, variant.Definition.DefaultElo, true
	}
	alias, ok := legacyByID[modeID]
	if !ok {
		return registry.VariantEntry{}, 0, false
	}
	for _, variant := range entry.Variants {
		if variant.Definition.ID == alias.TargetVariantID {
			return variant, alias.PresetElo, true
		}
	}
	return registry.VariantEntry{}, 0, false
}

func normalizeSpec(spec Spec) Spec {
	spec.Schema = strings.TrimSpace(spec.Schema)
	spec.Title = strings.TrimSpace(spec.Title)
	spec.Header = strings.TrimSpace(spec.Header)
	spec.Advert = strings.TrimSpace(spec.Advert)
	spec.Seed = strings.TrimSpace(spec.Seed)
	spec.PDFOutputPath = strings.TrimSpace(spec.PDFOutputPath)
	spec.JSONLOutputPath = strings.TrimSpace(spec.JSONLOutputPath)

	if spec.Schema == "" {
		spec.Schema = PackSpecSchemaV2
	}
	if spec.Volume == 0 {
		spec.Volume = 1
	}
	if spec.SheetLayout == "" {
		spec.SheetLayout = DefaultSheetLayout
	}
	if spec.Advert == "" {
		spec.Advert = DefaultAdvert
	}
	if spec.Counts == nil {
		spec.Counts = zeroCounts()
	}
	return spec
}

func generateRecords(
	ctx context.Context,
	selections []modeSelection,
	seed string,
	generatedAt time.Time,
) ([]pdfexport.JSONLRecord, error) {
	var rng *randState
	if seed != "" {
		rng = newRandState(seed)
	}

	nameSeed := seed
	if strings.TrimSpace(nameSeed) == "" {
		nameSeed = generatedAt.Format(time.RFC3339Nano)
	}
	nameRNG := resolve.RNGFromString("export-names:" + nameSeed)

	totalCount := 0
	for _, selection := range selections {
		totalCount += selection.count
	}

	records := make([]pdfexport.JSONLRecord, 0, totalCount)
	recordIndex := 0
	for _, selection := range selections {
		for i := 0; i < selection.count; i++ {
			if err := ctx.Err(); err != nil {
				return nil, err
			}
			recordIndex++

			eloSeed := scopedPackEloSeed(seed, generatedAt, selection.gameType, selection.modeTitle, recordIndex)
			generated, err := spawnSaveData(ctx, selection, rng, eloSeed)
			if err != nil {
				return nil, fmt.Errorf("generate %s / %s puzzle %d: %w", selection.gameType, selection.modeTitle, i+1, err)
			}

			records = append(records, pdfexport.JSONLRecord{
				Schema: pdfexport.ExportSchemaV2,
				Pack: pdfexport.JSONLPackMeta{
					Generated:     generatedAt.Format(time.RFC3339),
					Version:       Version,
					Category:      selection.gameType,
					ModeSelection: selection.modeTitle,
					Count:         totalCount,
					Seed:          seed,
				},
				Puzzle: pdfexport.JSONLPuzzle{
					Index:                len(records) + 1,
					Name:                 namegen.GenerateSeeded(nameRNG),
					Game:                 selection.gameType,
					Mode:                 selection.modeTitle,
					TargetDifficultyElo:  intPtrFromReportElo(generated.Report.TargetElo, generated.Report.Confidence),
					ActualDifficultyElo:  intPtrFromReportElo(generated.Report.ActualElo, generated.Report.Confidence),
					DifficultyConfidence: string(generated.Report.Confidence),
					Save:                 generated.SaveData,
				},
			})
		}
	}

	return records, nil
}

func intPtrFromReportElo(elo difficulty.Elo, confidence difficulty.Confidence) *int {
	if confidence == "" {
		return nil
	}
	v := int(elo)
	return &v
}

func scopedPackEloSeed(seed string, generatedAt time.Time, gameType, mode string, index int) string {
	base := strings.TrimSpace(seed)
	if base == "" {
		base = generatedAt.Format(time.RFC3339Nano)
	}
	return fmt.Sprintf("packexport:%s:%s:%s:%d", base, gameType, mode, index)
}
