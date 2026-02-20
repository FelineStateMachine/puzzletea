package app

import (
	"log"

	"github.com/FelineStateMachine/puzzletea/game"
	"github.com/FelineStateMachine/puzzletea/store"

	tea "charm.land/bubbletea/v2"
)

// activateGame prepares the game with global UI state and moves to game view.
func (m model) activateGame(g game.Gamer, activeGameID int64, completionSaved bool) model {
	g, _ = g.Update(game.HelpToggleMsg{Show: m.showFullHelp})
	g, _ = g.Update(tea.WindowSizeMsg{Width: m.width, Height: m.height})

	m.game = g
	m.activeGameID = activeGameID
	m.completionSaved = completionSaved
	m.state = gameView
	return m
}

func (m model) importAndActivateRecord(rec store.GameRecord) model {
	importFn, ok := game.Registry[rec.GameType]
	if !ok {
		log.Printf("unknown game type in save data: %s", rec.GameType)
		return m
	}

	g, err := importFn([]byte(rec.SaveState))
	if err != nil {
		log.Printf("failed to import game: %v", err)
		return m
	}

	return m.activateGame(g.SetTitle(rec.Name), rec.ID, rec.Status == store.StatusCompleted)
}
