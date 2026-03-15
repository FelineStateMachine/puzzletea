// Package cmd defines the CLI commands for PuzzleTea using Cobra.
package cmd

import (
	"fmt"
	"log"
	"strings"

	"github.com/FelineStateMachine/puzzletea/app"
	"github.com/FelineStateMachine/puzzletea/config"
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
	flagNew        string
	flagContinue   string
	flagTheme      string
	flagConfigPath string
	// flagSetSeed is declared in new.go and shared across root and new commands.
)

var (
	launchNewGameFn    = launchNewGame
	launchSeededGameFn = launchSeededGame
	continueGameFn     = continueGame
	openStoreFn        = store.Open
)

// RootCmd is the top-level Cobra command.
var RootCmd = &cobra.Command{
	Use:     "puzzletea",
	Version: Version,
	Short:   "A terminal-based puzzle game framework",
	Long:    "PuzzleTea is a terminal-based puzzle game framework featuring Nonogram, Nurikabe, Sudoku, Word Search, Hashiwokakero, and Lights Out.",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg := loadActiveConfig()

		if flagSetSeed != "" {
			if flagNew != "" || flagContinue != "" {
				return fmt.Errorf("--set-seed cannot be combined with --new or --continue")
			}
			return launchSeededGameFn(flagSetSeed, cfg)
		}
		if flagNew != "" {
			parts := strings.SplitN(flagNew, ":", 2)
			gameArg := parts[0]
			var modeArg string
			if len(parts) > 1 {
				modeArg = parts[1]
			}
			return launchNewGameFn(gameArg, modeArg, "", cfg)
		}
		if flagContinue != "" {
			return continueGameFn(flagContinue, cfg)
		}
		// Default: launch TUI menu.
		s, err := openStoreFn(cfg.DBPath)
		if err != nil {
			return err
		}
		defer s.Close()
		p := tea.NewProgram(app.InitialModel(s, cfg))
		_, err = p.Run()
		return err
	},
}

func init() {
	RootCmd.Flags().StringVar(&flagNew, "new", "", "start a new game (game:mode)")
	RootCmd.Flags().StringVar(&flagContinue, "continue", "", "resume a saved game by name")
	RootCmd.Flags().StringVar(&flagSetSeed, "set-seed", "", "seed string for deterministic puzzle selection and generation")
	RootCmd.PersistentFlags().StringVar(&flagConfigPath, "config", "", "path to config file (default: ~/.puzzletea/config.json)")
	RootCmd.PersistentFlags().StringVar(&flagTheme, "theme", "", "color theme name (overrides config)")

	RootCmd.AddCommand(newCmd, continueCmd, listCmd, exportPDFCmd, testCmd)
}

func loadActiveConfig() *config.Config {
	return loadConfig(flagConfigPath)
}

// loadConfig reads the config file and applies the active theme. The --theme
// flag takes precedence over the persisted setting.
func loadConfig(configPath string) *config.Config {
	if configPath == "" {
		configPath = config.DefaultPath()
	}

	cfg, err := config.Load(configPath)
	if err != nil {
		log.Printf("warning: %v (using defaults)", err)
		cfg = config.Default()
	}

	themeName := cfg.Theme
	if flagTheme != "" {
		themeName = flagTheme
	}
	applyTheme(themeName)
	return cfg
}

func applyTheme(themeName string) {
	if err := theme.Apply(themeName); err != nil {
		log.Printf("warning: %v (using default theme)", err)
	}
}
