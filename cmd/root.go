// Package cmd defines the CLI commands for PuzzleTea using Cobra.
package cmd

import (
	"strings"

	"github.com/FelineStateMachine/puzzletea/app"
	"github.com/FelineStateMachine/puzzletea/stats"
	"github.com/FelineStateMachine/puzzletea/store"

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
)

// RootCmd is the top-level Cobra command.
var RootCmd = &cobra.Command{
	Use:     "puzzletea",
	Version: Version,
	Short:   "A terminal-based puzzle game framework",
	Long:    "PuzzleTea is a terminal-based puzzle game framework featuring Nonogram, Sudoku, Word Search, Hashiwokakero, and Lights Out.",
	RunE: func(cmd *cobra.Command, args []string) error {
		if flagNew != "" {
			parts := strings.SplitN(flagNew, ":", 2)
			gameArg := parts[0]
			var modeArg string
			if len(parts) > 1 {
				modeArg = parts[1]
			}
			return launchNewGame(gameArg, modeArg)
		}
		if flagContinue != "" {
			return continueGame(flagContinue)
		}
		// Default: launch TUI menu.
		s, err := store.Open(store.DefaultDBPath())
		if err != nil {
			return err
		}
		defer s.Close()
		stats.InitModeXP(app.GameCategories)
		p := tea.NewProgram(app.InitialModel(s))
		_, err = p.Run()
		return err
	},
}

func init() {
	RootCmd.Flags().StringVar(&flagNew, "new", "", "start a new game (game:mode)")
	RootCmd.Flags().StringVar(&flagContinue, "continue", "", "resume a saved game by name")

	RootCmd.AddCommand(newCmd, continueCmd, listCmd)
}
