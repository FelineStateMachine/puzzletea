package cmd

import (
	"fmt"
	"log"
	"strings"

	"github.com/FelineStateMachine/puzzletea/app"
	"github.com/FelineStateMachine/puzzletea/resolve"
	"github.com/FelineStateMachine/puzzletea/stats"
	"github.com/FelineStateMachine/puzzletea/store"

	tea "charm.land/bubbletea/v2"
	"github.com/spf13/cobra"
)

var newCmd = &cobra.Command{
	Use:   "new <game> [mode]",
	Short: "Start a new puzzle game",
	Long: fmt.Sprintf("Start a new puzzle game, optionally specifying the difficulty mode.\n\nAvailable games:\n  %s",
		strings.Join(resolve.CategoryNames(app.GameCategories), "\n  ")),
	Args: cobra.RangeArgs(1, 2),
	RunE: func(cmd *cobra.Command, args []string) error {
		gameArg := args[0]
		var modeArg string
		if len(args) > 1 {
			modeArg = args[1]
		}
		return launchNewGame(gameArg, modeArg)
	},
}

// launchNewGame resolves the game/mode, spawns a new game, and launches the TUI.
func launchNewGame(gameArg, modeArg string) error {
	cat, err := resolve.Category(gameArg, app.GameCategories)
	if err != nil {
		return err
	}

	spawner, modeTitle, err := resolve.Mode(cat, modeArg)
	if err != nil {
		return err
	}

	g, err := spawner.Spawn()
	if err != nil {
		return fmt.Errorf("failed to spawn game: %w", err)
	}

	s, err := store.Open(store.DefaultDBPath())
	if err != nil {
		return err
	}
	defer s.Close()

	stats.InitModeXP(app.GameCategories)

	name := app.GenerateUniqueName(s)
	g = g.SetTitle(name)

	initialState, err := g.GetSave()
	if err != nil {
		return fmt.Errorf("failed to get initial save: %w", err)
	}

	rec := &store.GameRecord{
		Name:         name,
		GameType:     cat.Name,
		Mode:         modeTitle,
		InitialState: string(initialState),
		SaveState:    string(initialState),
		Status:       store.StatusNew,
	}
	if err := s.CreateGame(rec); err != nil {
		log.Printf("failed to create game record: %v", err)
	}

	m := app.InitialModelWithGame(s, g, rec.ID, false)
	p := tea.NewProgram(m)
	_, err = p.Run()
	return err
}
