package cmd

import (
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
		return fmt.Errorf("game %q does not support markdown export", cat.Name)
	}

	modeArg := ""
	if len(args) > 1 {
		modeArg = args[1]
	}

	entries, modeSelection, err := collectExportModes(cat, modeArg)
	if err != nil {
		return err
	}

	sections, err := buildExportSections(cat.Name, entries, flagExport, flagWithSeed)
	if err != nil {
		return err
	}

	doc := markdownexport.BuildDocument(markdownexport.DocumentConfig{
		Version:       Version,
		Category:      cat.Name,
		ModeSelection: modeSelection,
		Count:         flagExport,
		Seed:          flagWithSeed,
		GeneratedAt:   exportNow(),
	}, sections)

	if err := writeExportMarkdown(cmd, flagOutput, doc); err != nil {
		return err
	}

	return nil
}

func validateNewExportFlags(cmd *cobra.Command, args []string) error {
	if flagExport < 1 {
		return fmt.Errorf("--export must be at least 1")
	}
	if strings.TrimSpace(flagOutput) != "" && !strings.EqualFold(filepath.Ext(flagOutput), ".md") {
		return fmt.Errorf("--output must use a .md extension")
	}
	if flagSetSeed != "" {
		return fmt.Errorf("--set-seed cannot be combined with markdown export (--export/--output)")
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

func buildExportSections(gameType string, entries []exportModeEntry, count int, seed string) ([]markdownexport.PuzzleSection, error) {
	var rng *rand.Rand
	if seed != "" {
		rng = resolve.RNGFromString(seed)
	}

	sections := make([]markdownexport.PuzzleSection, 0, count)
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

		snippet, err := markdownexport.RenderPuzzleSnippet(gameType, entry.mode, save)
		if err != nil {
			if errors.Is(err, markdownexport.ErrUnsupportedGame) {
				return nil, fmt.Errorf("game %q does not support markdown export", gameType)
			}
			return nil, fmt.Errorf("render puzzle %d: %w", i+1, err)
		}

		sections = append(sections, markdownexport.PuzzleSection{
			Index:    i + 1,
			GameType: gameType,
			Mode:     entry.mode,
			Body:     snippet,
		})
	}

	return sections, nil
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

func writeExportMarkdown(cmd *cobra.Command, path, content string) error {
	if strings.TrimSpace(path) == "" {
		if _, err := io.WriteString(cmd.OutOrStdout(), content); err != nil {
			return fmt.Errorf("write export markdown to stdout: %w", err)
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
		return fmt.Errorf("write output markdown: %w", err)
	}
	return nil
}
