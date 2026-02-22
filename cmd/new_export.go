package cmd

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/rand/v2"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/FelineStateMachine/puzzletea/app"
	"github.com/FelineStateMachine/puzzletea/game"
	"github.com/FelineStateMachine/puzzletea/markdownexport"
	"github.com/FelineStateMachine/puzzletea/namegen"
	"github.com/FelineStateMachine/puzzletea/pdfexport"
	"github.com/FelineStateMachine/puzzletea/resolve"

	"github.com/spf13/cobra"
)

var exportNow = time.Now

type exportModeEntry struct {
	spawner game.Spawner
	mode    string
}

func runNewExport(cmd *cobra.Command, args []string) error {
	if err := validateNewExportFlags(cmd, args); err != nil {
		return err
	}

	cat, err := resolve.Category(args[0], app.GameCategories)
	if err != nil {
		return err
	}
	if !markdownexport.SupportsGameType(cat.Name) {
		return fmt.Errorf("game %q does not support export", cat.Name)
	}

	modeArg := ""
	if len(args) > 1 {
		modeArg = args[1]
	}

	entries, modeSelection, err := collectExportModes(cat, modeArg)
	if err != nil {
		return err
	}

	generatedAt := exportNow()
	records, err := buildExportRecords(cat.Name, modeSelection, entries, flagExport, flagWithSeed, generatedAt)
	if err != nil {
		return err
	}

	if err := writeExportJSONL(cmd, flagOutput, records); err != nil {
		return err
	}

	return nil
}

func validateNewExportFlags(cmd *cobra.Command, args []string) error {
	if flagExport < 1 {
		return fmt.Errorf("--export must be at least 1")
	}
	if strings.TrimSpace(flagOutput) != "" && !strings.EqualFold(filepath.Ext(flagOutput), ".jsonl") {
		return fmt.Errorf("--output must use a .jsonl extension")
	}
	if flagSetSeed != "" {
		return fmt.Errorf("--set-seed cannot be combined with export (--export/--output)")
	}
	if len(args) == 0 {
		return fmt.Errorf("requires at least 1 arg(s), only received 0")
	}
	if len(args) > 2 {
		return fmt.Errorf("accepts at most 2 arg(s), received %d", len(args))
	}
	return nil
}

func collectExportModes(cat game.Category, modeArg string) ([]exportModeEntry, string, error) {
	if modeArg != "" {
		spawner, modeTitle, err := resolve.Mode(cat, modeArg)
		if err != nil {
			return nil, "", err
		}
		return []exportModeEntry{{spawner: spawner, mode: modeTitle}}, modeTitle, nil
	}

	entries := make([]exportModeEntry, 0, len(cat.Modes))
	for _, item := range cat.Modes {
		mode, ok := item.(game.Mode)
		if !ok {
			continue
		}
		spawner, ok := item.(game.Spawner)
		if !ok {
			continue
		}
		entries = append(entries, exportModeEntry{
			spawner: spawner,
			mode:    mode.Title(),
		})
	}
	if len(entries) == 0 {
		return nil, "", fmt.Errorf("game %q has no exportable modes", cat.Name)
	}

	return entries, "mixed modes", nil
}

func buildExportRecords(
	gameType, modeSelection string,
	entries []exportModeEntry,
	count int,
	seed string,
	generatedAt time.Time,
) ([]pdfexport.JSONLRecord, error) {
	var rng *rand.Rand
	if seed != "" {
		rng = resolve.RNGFromString(seed)
	}

	nameSeed := seed
	if strings.TrimSpace(nameSeed) == "" {
		nameSeed = generatedAt.Format(time.RFC3339Nano)
	}
	nameRNG := resolve.RNGFromString("export-names:" + nameSeed)

	records := make([]pdfexport.JSONLRecord, 0, count)
	for i := 0; i < count; i++ {
		entry := entries[0]
		if len(entries) > 1 {
			var modeIndex int
			if rng != nil {
				modeIndex = rng.IntN(len(entries))
			} else {
				modeIndex = rand.IntN(len(entries))
			}
			entry = entries[modeIndex]
		}

		puzzle, err := spawnExportPuzzle(entry.spawner, rng)
		if err != nil {
			return nil, fmt.Errorf("generate puzzle %d: %w", i+1, err)
		}

		save, err := puzzle.GetSave()
		if err != nil {
			return nil, fmt.Errorf("serialize puzzle %d: %w", i+1, err)
		}
		if !json.Valid(save) {
			return nil, fmt.Errorf("serialize puzzle %d: save payload is not valid JSON", i+1)
		}

		snippet, err := markdownexport.RenderPuzzleSnippet(gameType, entry.mode, save)
		if err != nil {
			if errors.Is(err, markdownexport.ErrUnsupportedGame) {
				return nil, fmt.Errorf("game %q does not support export", gameType)
			}
			return nil, fmt.Errorf("render puzzle %d: %w", i+1, err)
		}

		records = append(records, pdfexport.JSONLRecord{
			Schema: pdfexport.ExportSchemaV1,
			Pack: pdfexport.JSONLPackMeta{
				Generated:     generatedAt.Format(time.RFC3339),
				Version:       Version,
				Category:      gameType,
				ModeSelection: modeSelection,
				Count:         count,
				Seed:          seed,
			},
			Puzzle: pdfexport.JSONLPuzzle{
				Index:   i + 1,
				Name:    namegen.GenerateSeeded(nameRNG),
				Game:    gameType,
				Mode:    entry.mode,
				Save:    json.RawMessage(save),
				Snippet: snippet,
			},
		})
	}

	return records, nil
}

func spawnExportPuzzle(spawner game.Spawner, rng *rand.Rand) (game.Gamer, error) {
	if rng == nil {
		return spawner.Spawn()
	}

	seeded, ok := spawner.(game.SeededSpawner)
	if !ok {
		return nil, fmt.Errorf("mode does not support deterministic spawning")
	}
	return seeded.SpawnSeeded(rng)
}

func writeExportJSONL(cmd *cobra.Command, path string, records []pdfexport.JSONLRecord) error {
	var b strings.Builder
	for _, record := range records {
		data, err := json.Marshal(record)
		if err != nil {
			return fmt.Errorf("encode jsonl record: %w", err)
		}
		b.Write(data)
		b.WriteByte('\n')
	}

	content := b.String()
	if strings.TrimSpace(path) == "" {
		if _, err := io.WriteString(cmd.OutOrStdout(), content); err != nil {
			return fmt.Errorf("write export jsonl to stdout: %w", err)
		}
		return nil
	}

	dir := filepath.Dir(path)
	if dir != "" && dir != "." {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return fmt.Errorf("create output directory: %w", err)
		}
	}

	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		return fmt.Errorf("write output jsonl: %w", err)
	}
	return nil
}
