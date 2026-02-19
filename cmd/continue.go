package cmd

import (
	"fmt"

	"github.com/FelineStateMachine/puzzletea/app"
	"github.com/FelineStateMachine/puzzletea/config"
	"github.com/FelineStateMachine/puzzletea/game"
	"github.com/FelineStateMachine/puzzletea/stats"
	"github.com/FelineStateMachine/puzzletea/store"

	tea "charm.land/bubbletea/v2"
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

	stats.InitModeXP(app.GameCategories)

	rec, err := s.GetGameByName(name)
	if err != nil {
		return err
	}
	if rec == nil {
		return fmt.Errorf("no saved game found with name %q\nRun 'puzzletea list' to see available games.", name)
	}

	importFn, ok := game.Registry[rec.GameType]
	if !ok {
		return fmt.Errorf("unknown game type %q in save data", rec.GameType)
	}

	g, err := importFn([]byte(rec.SaveState))
	if err != nil {
		return fmt.Errorf("failed to import game: %w", err)
	}
	g = g.SetTitle(rec.Name)

	completed := rec.Status == store.StatusCompleted
	m := app.InitialModelWithGame(s, cfg, g, rec.ID, completed)
	p := tea.NewProgram(m)
	_, err = p.Run()
	return err
}
