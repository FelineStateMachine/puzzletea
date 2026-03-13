package app

import (
	"time"

	"github.com/FelineStateMachine/puzzletea/daily"
	sessionflow "github.com/FelineStateMachine/puzzletea/session"

	tea "charm.land/bubbletea/v2"
)

func (m model) handleDailyPuzzle() (tea.Model, tea.Cmd) {
	today := time.Now()
	name := daily.Name(today)

	rec, err := m.store.GetDailyGame(name)
	if err != nil {
		return m.setErrorf("Could not check today’s daily puzzle: %v", err), nil
	}
	if rec != nil {
		var resumed bool
		m, resumed = m.importAndActivateRecord(*rec)
		if resumed {
			if err := sessionflow.ResumeAbandonedDeterministicRecord(m.store, rec); err != nil {
				m = m.setErrorf("%v", err)
			}
		}
		return m, nil
	}

	spawner, gameType, modeTitle := daily.Mode(today)
	if spawner == nil {
		return m.setErrorf("No daily puzzle is configured for %s", today.Format("2006-01-02")), nil
	}

	rng := daily.RNG(today)
	ctx, jobID := m.beginSpawnContext()
	m.session.spawn = &spawnRequest{
		source:      spawnSourceDaily,
		name:        name,
		gameType:    gameType,
		modeTitle:   modeTitle,
		returnState: playMenuView,
		exitState:   mainMenuView,
	}
	m.state = generatingView
	m = m.clearNotice()
	return m, tea.Batch(m.spinner.Tick, spawnSeededCmd(spawner, rng, ctx, jobID))
}
