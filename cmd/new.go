package cmd

import (
	"fmt"
	"strings"

	"github.com/FelineStateMachine/puzzletea/app"
	"github.com/FelineStateMachine/puzzletea/config"
	"github.com/FelineStateMachine/puzzletea/game"
	"github.com/FelineStateMachine/puzzletea/namegen"
	"github.com/FelineStateMachine/puzzletea/resolve"
	"github.com/FelineStateMachine/puzzletea/stats"
	"github.com/FelineStateMachine/puzzletea/store"

	"github.com/spf13/cobra"
)

var (
	flagSetSeed  string
	flagWithSeed string
)

var newCmd = &cobra.Command{
	Use:   "new <game> [mode]",
	Short: "Start a new puzzle game",
	Long: fmt.Sprintf("Start a new puzzle game, optionally specifying the difficulty mode.\n"+
		"Use --set-seed to generate a deterministic puzzle from a seed string.\n\nAvailable games:\n  %s",
		strings.Join(resolve.CategoryNames(app.GameCategories), "\n  ")),
	Args: cobra.RangeArgs(0, 2),
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg := loadConfig()

		if flagSetSeed != "" {
			if len(args) > 0 {
				return fmt.Errorf("cannot specify game/mode arguments with --set-seed; the seed determines the puzzle type")
			}
			return launchSeededGame(flagSetSeed, cfg)
		}
		if len(args) == 0 {
			return fmt.Errorf("requires at least 1 arg(s), only received 0")
		}
		gameArg := args[0]
		var modeArg string
		if len(args) > 1 {
			modeArg = args[1]
		}
		return launchNewGame(gameArg, modeArg, flagWithSeed, cfg)
	},
}

func init() {
	newCmd.Flags().StringVar(&flagSetSeed, "set-seed", "", "seed string for deterministic puzzle selection and generation")
	newCmd.Flags().StringVar(&flagWithSeed, "with-seed", "", "seed string for deterministic puzzle generation within the selected game/mode")
}

// launchNewGame resolves the game/mode, spawns a new game, and launches the TUI.
func launchNewGame(gameArg, modeArg, seed string, cfg *config.Config) error {
	cat, err := resolve.Category(gameArg, app.GameCategories)
	if err != nil {
		return err
	}

	spawner, modeTitle, err := resolve.Mode(cat, modeArg)
	if err != nil {
		return err
	}

	g, err := spawnFromMode(spawner, seed)
	if err != nil {
		return fmt.Errorf("failed to spawn game: %w", err)
	}

	s, err := store.Open(store.DefaultDBPath())
	if err != nil {
		return err
	}
	defer s.Close()

	stats.InitModeXP(app.Categories)

	name := app.GenerateUniqueName(s)
	g = g.SetTitle(name)

	rec, err := createGameRecord(s, g, name, cat.Name, modeTitle)
	if err != nil {
		return err
	}

	return runGameProgram(s, cfg, g, rec.ID, false)
}

func spawnFromMode(spawner game.Spawner, seed string) (game.Gamer, error) {
	if seed == "" {
		return spawner.Spawn()
	}

	seeded, ok := spawner.(game.SeededSpawner)
	if !ok {
		return nil, fmt.Errorf("mode does not support deterministic spawning")
	}
	return seeded.SpawnSeeded(resolve.RNGFromString(seed))
}

// launchSeededGame uses an arbitrary seed string to deterministically select
// a game type, mode, and puzzle. The same seed always produces the same puzzle.
// If a game with the same seed-derived name already exists, it is resumed.
//
// Name generation and mode selection each use independent hashing so that
// changes to namegen word lists or mode lists don't cascade. The shared
// RNG is reserved purely for puzzle content generation.
func launchSeededGame(seed string, cfg *config.Config) error {
	// Prevent seeded puzzles from mimicking daily puzzle names by
	// silently lowercasing a "Daily" prefix so titles never start
	// with the real daily prefix "Daily ".
	if strings.HasPrefix(strings.ToLower(seed), "daily") {
		seed = strings.ToLower(seed[:len("daily")]) + seed[len("daily"):]
	}

	// Name uses its own sub-RNG so namegen changes can't affect mode
	// selection or puzzle generation.
	nameRNG := resolve.RNGFromString("name:" + seed)
	name := seed + " - " + namegen.GenerateSeeded(nameRNG)

	s, err := store.Open(store.DefaultDBPath())
	if err != nil {
		return err
	}
	defer s.Close()

	stats.InitModeXP(app.Categories)

	// If a game with this name already exists, resume it (including
	// abandoned games — seeded puzzles are deterministic and should
	// always be resumable rather than duplicated).
	if rec, err := s.GetDailyGame(name); err == nil && rec != nil {
		g, err := importSavedGame(rec)
		if err != nil {
			return err
		}
		completed := rec.Status == store.StatusCompleted

		// Resume abandoned seeded games by resetting their status.
		if rec.Status == store.StatusAbandoned {
			if err := s.UpdateStatus(rec.ID, store.StatusInProgress); err != nil {
				return fmt.Errorf("failed to mark seeded game in progress: %w", err)
			}
		}

		return runGameProgram(s, cfg, g, rec.ID, completed)
	}

	// Mode selection uses rendezvous hashing (independent of RNG).
	spawner, gameType, modeTitle, err := resolve.SeededMode(seed, app.GameCategories)
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

	rec, err := createGameRecord(s, g, name, gameType, modeTitle)
	if err != nil {
		return err
	}

	return runGameProgram(s, cfg, g, rec.ID, false)
}
