package cmd

import (
	"fmt"

	"github.com/FelineStateMachine/puzzletea/app"
	"github.com/FelineStateMachine/puzzletea/config"
	"github.com/FelineStateMachine/puzzletea/stats"
	"github.com/FelineStateMachine/puzzletea/store"

	"github.com/spf13/cobra"
)

var continueCmd = &cobra.Command{
	Use:   "continue <name>",
	Short: "Resume a saved game by name",
	Long:  "Resume a previously saved game using its unique name.\nUse 'puzzletea list' to see available saved games.",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg := loadConfig()
		return continueGame(args[0], cfg)
	},
}

// continueGame looks up a saved game by name and launches the TUI.
func continueGame(name string, cfg *config.Config) error {
	s, err := store.Open(store.DefaultDBPath())
	if err != nil {
		return err
	}
	defer s.Close()

	stats.InitModeXP(app.Categories)

	rec, err := s.GetGameByName(name)
	if err != nil {
		return err
	}
	if rec == nil {
		return fmt.Errorf("no saved game found with name %q\nRun 'puzzletea list' to see available games.", name)
	}

	g, err := importSavedGame(rec)
	if err != nil {
		return err
	}

	completed := rec.Status == store.StatusCompleted
	return runGameProgram(s, cfg, g, rec.ID, completed)
}
