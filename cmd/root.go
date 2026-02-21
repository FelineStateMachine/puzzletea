// Package cmd defines the CLI commands for PuzzleTea using Cobra.
package cmd

import (
	"fmt"
	"log"
	"strings"

	"github.com/FelineStateMachine/puzzletea/app"
	"github.com/FelineStateMachine/puzzletea/config"
	"github.com/FelineStateMachine/puzzletea/stats"
	"github.com/FelineStateMachine/puzzletea/store"
	"github.com/FelineStateMachine/puzzletea/theme"

	tea "charm.land/bubbletea/v2"
	"github.com/spf13/cobra"
)

// Version is set at build time via ldflags:
//
//	go build -ldflags "-X github.com/FelineStateMachine/puzzletea/cmd.Version=v1.0.0"
var Version = "dev"

var (
	flagNew      string
	flagContinue string
	flagTheme    string
	// flagSetSeed is declared in new.go and shared across root and new commands.
)

// RootCmd is the top-level Cobra command.
var RootCmd = &cobra.Command{
	Use:     "puzzletea",
	Version: Version,
	Short:   "A terminal-based puzzle game framework",
	Long:    "PuzzleTea is a terminal-based puzzle game framework featuring Nonogram, Nurikabe, Sudoku, Word Search, Hashiwokakero, and Lights Out.",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg := loadConfig()

		if flagSetSeed != "" {
			if flagNew != "" || flagContinue != "" {
				return fmt.Errorf("--set-seed cannot be combined with --new or --continue")
			}
			return launchSeededGame(flagSetSeed, cfg)
		}
		if flagNew != "" {
			parts := strings.SplitN(flagNew, ":", 2)
			gameArg := parts[0]
			var modeArg string
			if len(parts) > 1 {
				modeArg = parts[1]
			}
			return launchNewGame(gameArg, modeArg, cfg)
		}
		if flagContinue != "" {
			return continueGame(flagContinue, cfg)
		}
		// Default: launch TUI menu.
		s, err := store.Open(store.DefaultDBPath())
		if err != nil {
			return err
		}
		defer s.Close()
		stats.InitModeXP(app.Categories)
		p := tea.NewProgram(app.InitialModel(s, cfg))
		_, err = p.Run()
		return err
	},
}

func init() {
	RootCmd.Flags().StringVar(&flagNew, "new", "", "start a new game (game:mode)")
	RootCmd.Flags().StringVar(&flagContinue, "continue", "", "resume a saved game by name")
	RootCmd.Flags().StringVar(&flagSetSeed, "set-seed", "", "seed string for deterministic puzzle selection and generation")
	RootCmd.PersistentFlags().StringVar(&flagTheme, "theme", "", "color theme name (overrides config)")

	RootCmd.AddCommand(newCmd, continueCmd, listCmd)
}

// loadConfig reads the config file and applies the active theme. The --theme
// flag takes precedence over the persisted setting.
func loadConfig() *config.Config {
	cfg, err := config.Load(config.DefaultPath())
	if err != nil {
		log.Printf("warning: %v (using defaults)", err)
		cfg = config.Default()
	}

	themeName := cfg.Theme
	if flagTheme != "" {
		themeName = flagTheme
	}
	if err := theme.Apply(themeName); err != nil {
		log.Printf("warning: %v (using default theme)", err)
	}
	return cfg
}
