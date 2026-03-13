package app

import (
	"github.com/FelineStateMachine/puzzletea/game"
	sessionflow "github.com/FelineStateMachine/puzzletea/session"
	"github.com/FelineStateMachine/puzzletea/store"
	"github.com/FelineStateMachine/puzzletea/weekly"

	tea "charm.land/bubbletea/v2"
)

type gameOpenOptions struct {
	readOnly    bool
	returnState viewState
	weeklyInfo  *weekly.Info
}

// activateGame prepares the game with global UI state and moves to game view.
func (m model) activateGame(g game.Gamer, activeGameID int64, completionSaved bool, options gameOpenOptions) model {
	g, _ = g.Update(game.HelpToggleMsg{Show: m.help.showFull})
	g, _ = g.Update(tea.WindowSizeMsg{Width: m.width, Height: m.height})

	m.session.game = g
	m.session.activeGameID = activeGameID
	m.session.completionSaved = completionSaved
	m.session.returnState = options.returnState
	m.session.weeklyAdvance = options.weeklyInfo
	m.state = gameView
	return m
}

func (m model) importAndActivateRecord(rec store.GameRecord) (model, bool) {
	return m.importAndActivateRecordWithOptions(rec, gameOpenOptions{returnState: mainMenuView})
}

func (m model) importAndActivateRecordWithOptions(rec store.GameRecord, options gameOpenOptions) (model, bool) {
	g, err := sessionflow.ImportRecord(&rec)
	if err != nil {
		return m.setErrorf("Could not load saved puzzle %q: %v", rec.Name, err), false
	}

	activeGameID := rec.ID
	completionSaved := rec.Status == store.StatusCompleted
	if options.readOnly {
		activeGameID = 0
		completionSaved = true
	}

	return m.activateGame(g, activeGameID, completionSaved, options).clearNotice(), true
}
