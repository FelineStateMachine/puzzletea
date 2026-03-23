package cmd

import (
	"fmt"

	"github.com/FelineStateMachine/puzzletea/config"
	sessionflow "github.com/FelineStateMachine/puzzletea/session"
	"github.com/FelineStateMachine/puzzletea/store"

	"github.com/spf13/cobra"
)

var continueCmd = &cobra.Command{
	Use:   "continue <name>",
	Short: "Resume a saved game by name",
	Long:  "Resume a previously saved game using its unique name.\nUse 'puzzletea list' to see available saved games.",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return continueGameFn(args[0], loadActiveConfig())
	},
}

// continueGame looks up a saved game by name and launches the TUI.
func continueGame(name string, cfg *config.Config) error {
	s, err := openStoreFn(cfg.DBPath)
	if err != nil {
		return err
	}
	defer s.Close()

	rec, err := s.GetGameByName(name)
	if err != nil {
		return err
	}
	if rec == nil {
		return fmt.Errorf("no saved game found with name %q\nRun 'puzzletea list' to see available games.", name)
	}

	g, err := sessionflow.ImportRecord(rec)
	if err != nil {
		return err
	}

	completed := rec.Status == store.StatusCompleted
	return runGameProgramFn(s, cfg, activeConfigPath(), g, rec.ID, completed)
}
