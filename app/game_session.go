package app

import (
	"log"

	"github.com/FelineStateMachine/puzzletea/game"
	sessionflow "github.com/FelineStateMachine/puzzletea/session"
	"github.com/FelineStateMachine/puzzletea/store"

	tea "charm.land/bubbletea/v2"
)

// activateGame prepares the game with global UI state and moves to game view.
func (m model) activateGame(g game.Gamer, activeGameID int64, completionSaved bool) model {
	g, _ = g.Update(game.HelpToggleMsg{Show: m.help.showFull})
	g, _ = g.Update(tea.WindowSizeMsg{Width: m.width, Height: m.height})

	m.session.game = g
	m.session.activeGameID = activeGameID
	m.session.completionSaved = completionSaved
	m.state = gameView
	return m
}

func (m model) importAndActivateRecord(rec store.GameRecord) (model, bool) {
	g, err := sessionflow.ImportRecord(&rec)
	if err != nil {
		log.Printf("failed to import game: %v", err)
		return m, false
	}

	return m.activateGame(g, rec.ID, rec.Status == store.StatusCompleted), true
}
