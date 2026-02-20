package cmd

import (
	"fmt"

	"github.com/FelineStateMachine/puzzletea/app"
	"github.com/FelineStateMachine/puzzletea/config"
	"github.com/FelineStateMachine/puzzletea/game"
	"github.com/FelineStateMachine/puzzletea/store"

	tea "charm.land/bubbletea/v2"
)

func importSavedGame(rec *store.GameRecord) (game.Gamer, error) {
	importFn, ok := game.Registry[rec.GameType]
	if !ok {
		return nil, fmt.Errorf("unknown game type %q in save data", rec.GameType)
	}

	g, err := importFn([]byte(rec.SaveState))
	if err != nil {
		return nil, fmt.Errorf("failed to import game: %w", err)
	}
	return g.SetTitle(rec.Name), nil
}

func createGameRecord(
	s *store.Store,
	g game.Gamer,
	name string,
	gameType string,
	modeTitle string,
) (*store.GameRecord, error) {
	initialState, err := g.GetSave()
	if err != nil {
		return nil, fmt.Errorf("failed to get initial save: %w", err)
	}

	rec := &store.GameRecord{
		Name:         name,
		GameType:     gameType,
		Mode:         modeTitle,
		InitialState: string(initialState),
		SaveState:    string(initialState),
		Status:       store.StatusNew,
	}
	if err := s.CreateGame(rec); err != nil {
		return nil, fmt.Errorf("failed to create game record: %w", err)
	}

	return rec, nil
}

func runGameProgram(
	s *store.Store,
	cfg *config.Config,
	g game.Gamer,
	activeGameID int64,
	completionSaved bool,
) error {
	m := app.InitialModelWithGame(s, cfg, g, activeGameID, completionSaved)
	p := tea.NewProgram(m)
	_, err := p.Run()
	return err
}
