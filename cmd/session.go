package cmd

import (
	"github.com/FelineStateMachine/puzzletea/app"
	"github.com/FelineStateMachine/puzzletea/config"
	"github.com/FelineStateMachine/puzzletea/game"
	"github.com/FelineStateMachine/puzzletea/store"

	tea "charm.land/bubbletea/v2"
)

var runGameProgramFn = runGameProgram

func runGameProgram(
	s *store.Store,
	cfg *config.Config,
	configPath string,
	g game.Gamer,
	activeGameID int64,
	completionSaved bool,
) error {
	m := app.InitialModelWithGame(s, cfg, configPath, g, activeGameID, completionSaved)
	p := tea.NewProgram(m)
	_, err := p.Run()
	return err
}
