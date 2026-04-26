package cmd

import (
	"fmt"
	"strings"

	"github.com/FelineStateMachine/puzzletea/config"
	"github.com/FelineStateMachine/puzzletea/difficulty"
	"github.com/FelineStateMachine/puzzletea/game"
	"github.com/FelineStateMachine/puzzletea/registry"
	"github.com/FelineStateMachine/puzzletea/resolve"
	sessionflow "github.com/FelineStateMachine/puzzletea/session"
	"github.com/FelineStateMachine/puzzletea/store"

	"github.com/spf13/cobra"
)

var (
	flagSetSeed       string
	flagWithSeed      string
	flagDifficulty    int
	flagDifficultySet bool
	flagExport        int
	flagOutput        string
)

var newCmd = &cobra.Command{
	Use:   "new <game> [mode]",
	Short: "Start a new puzzle game",
	Long: fmt.Sprintf("Start a new puzzle game, optionally specifying the difficulty mode.\n"+
		"Use --set-seed to generate a deterministic puzzle from a seed string.\n\nAvailable games:\n  %s",
		strings.Join(registry.Names(), "\n  ")),
	Args: cobra.RangeArgs(0, 2),
	RunE: func(cmd *cobra.Command, args []string) error {
		flagDifficultySet = cmd.Flags().Changed("difficulty")
		if flagOutput != "" || cmd.Flags().Changed("export") {
			return runNewExport(cmd, args)
		}

		cfg := loadActiveConfig()

		if flagSetSeed != "" {
			if len(args) > 0 {
				return fmt.Errorf("cannot specify game/mode arguments with --set-seed; the seed determines the puzzle type")
			}
			return launchSeededGameFn(flagSetSeed, cfg)
		}
		if len(args) == 0 {
			return fmt.Errorf("requires at least 1 arg(s), only received 0")
		}
		gameArg := args[0]
		var modeArg string
		if len(args) > 1 {
			modeArg = args[1]
		}
		return launchNewGameFn(gameArg, modeArg, flagWithSeed, cfg)
	},
}

func init() {
	newCmd.Flags().StringVar(&flagSetSeed, "set-seed", "", "seed string for deterministic puzzle selection and generation")
	newCmd.Flags().StringVar(&flagWithSeed, "with-seed", "", "seed string for deterministic puzzle generation within the selected game/mode")
	newCmd.Flags().IntVar(&flagDifficulty, "difficulty", -1, "target Elo difficulty for Elo-capable generation (0..3000)")
	newCmd.Flags().IntVarP(&flagExport, "export", "e", 1, "number of puzzles to export")
	newCmd.Flags().StringVarP(&flagOutput, "output", "o", "", "write puzzles to a jsonl file (defaults to stdout)")
}

// launchNewGame resolves the game/mode, spawns a new game, and launches the TUI.
func launchNewGame(gameArg, modeArg, seed string, cfg *config.Config) error {
	entry, ok := registry.Resolve(gameArg)
	if !ok {
		err := fmt.Errorf("unknown game %q", gameArg)
		return err
	}

	selection, err := resolve.VariantEntry(entry, modeArg)
	if err != nil {
		return err
	}

	targetElo, err := difficultyFlag()
	if err != nil {
		return err
	}
	if targetElo == nil {
		targetElo = selection.ExplicitElo
	}

	s, err := openStoreFn(cfg.DBPath)
	if err != nil {
		return err
	}
	defer s.Close()

	name := sessionflow.GenerateUniqueName(s)

	g, report, err := spawnFromVariant(selection.Variant, seed, targetElo, name)
	if err != nil {
		return fmt.Errorf("failed to spawn game: %w", err)
	}
	g = g.SetTitle(name)

	meta := difficultyMetadataFromReport(report)
	rec, err := sessionflow.CreateRecordWithDifficulty(s, g, name, entry.Definition.Name, selection.DisplayTitle, store.NormalRunMetadata(), meta)
	if err != nil {
		return err
	}

	return runGameProgramFn(s, cfg, activeConfigPath(), g, rec.ID, false)
}

func spawnFromVariant(variant registry.VariantEntry, seed string, targetElo *difficulty.Elo, fallbackEloSeed string) (game.Gamer, difficulty.Report, error) {
	effectiveElo := targetElo
	if effectiveElo == nil {
		elo := variant.Definition.DefaultElo
		effectiveElo = &elo
	}
	if variant.Elo == nil {
		return nil, difficulty.Report{}, fmt.Errorf("variant does not support Elo difficulty")
	}
	eloSeed := seed
	if eloSeed == "" {
		eloSeed = fallbackEloSeed
	}
	g, report, err := variant.Elo.SpawnElo(eloSeed, *effectiveElo)
	if err != nil {
		return nil, difficulty.Report{}, err
	}
	return g, report, nil
}

func spawnFromMode(mode registry.ModeEntry, seed string, targetElo *difficulty.Elo, fallbackEloSeed string) (game.Gamer, difficulty.Report, error) {
	effectiveElo := targetElo
	if effectiveElo == nil {
		effectiveElo = mode.Definition.PresetElo
	}

	if effectiveElo != nil {
		if mode.Elo == nil {
			return nil, difficulty.Report{}, fmt.Errorf("mode does not support Elo difficulty")
		}
		eloSeed := seed
		if eloSeed == "" {
			eloSeed = fallbackEloSeed
		}
		g, report, err := mode.Elo.SpawnElo(eloSeed, *effectiveElo)
		if err != nil {
			return nil, difficulty.Report{}, err
		}
		return g, report, nil
	}

	if seed == "" {
		g, err := mode.Spawner.Spawn()
		return g, difficulty.Report{}, err
	}

	if mode.Seeded == nil {
		return nil, difficulty.Report{}, fmt.Errorf("mode does not support deterministic spawning")
	}
	g, err := mode.Seeded.SpawnSeeded(resolve.RNGFromString(seed))
	return g, difficulty.Report{}, err
}

func difficultyFlag() (*difficulty.Elo, error) {
	if !flagDifficultySet {
		return nil, nil
	}
	elo := difficulty.Elo(flagDifficulty)
	if err := difficulty.ValidateElo(elo); err != nil {
		return nil, err
	}
	return &elo, nil
}

func difficultyMetadataFromReport(report difficulty.Report) sessionflow.DifficultyMetadata {
	if report.Confidence == "" {
		return sessionflow.DifficultyMetadata{}
	}
	target := report.TargetElo
	actual := report.ActualElo
	return sessionflow.DifficultyMetadata{
		TargetElo:  &target,
		ActualElo:  &actual,
		Confidence: report.Confidence,
	}
}

// launchSeededGame uses an arbitrary seed string to deterministically select
// a game type, mode, and puzzle. The same seed always produces the same puzzle.
// If a game with the same seed-derived name already exists, it is resumed.
//
// Name generation and mode selection each use independent hashing so that
// changes to namegen word lists or mode lists don't cascade. The shared
// RNG is reserved purely for puzzle content generation.
func launchSeededGame(seed string, cfg *config.Config) error {
	seed = sessionflow.NormalizeSeed(seed)
	name := sessionflow.SeededName(seed)

	s, err := openStoreFn(cfg.DBPath)
	if err != nil {
		return err
	}
	defer s.Close()

	// If a game with this name already exists, resume it (including
	// abandoned games — seeded puzzles are deterministic and should
	// always be resumable rather than duplicated).
	if rec, err := s.GetDailyGame(name); err == nil && rec != nil {
		g, err := sessionflow.ImportRecord(rec)
		if err != nil {
			return err
		}
		completed := rec.Status == store.StatusCompleted

		if err := sessionflow.ResumeAbandonedDeterministicRecord(s, rec); err != nil {
			return err
		}

		return runGameProgramFn(s, cfg, activeConfigPath(), g, rec.ID, completed)
	}

	// Mode selection uses rendezvous hashing (independent of RNG).
	spawner, gameType, modeTitle, err := resolve.SeededMode(seed, registry.Entries())
	if err != nil {
		return err
	}

	// RNG is reserved purely for puzzle content generation.
	rng := resolve.RNGFromString(seed)
	g, err := spawner.SpawnSeeded(rng)
	if err != nil {
		return fmt.Errorf("failed to spawn game: %w", err)
	}
	g = g.SetTitle(name)

	rec, err := sessionflow.CreateRecord(s, g, name, gameType, modeTitle, store.SeededRunMetadata(seed))
	if err != nil {
		return err
	}

	return runGameProgramFn(s, cfg, activeConfigPath(), g, rec.ID, false)
}
