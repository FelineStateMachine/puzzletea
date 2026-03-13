package app

import (
	"github.com/FelineStateMachine/puzzletea/game"
	"github.com/FelineStateMachine/puzzletea/registry"
	"github.com/FelineStateMachine/puzzletea/resolve"
	sessionflow "github.com/FelineStateMachine/puzzletea/session"

	tea "charm.land/bubbletea/v2"
)

func (m model) handleSeedConfirm() (tea.Model, tea.Cmd) {
	seed := sessionflow.NormalizeSeed(m.nav.seedInput.Value())
	if seed == "" {
		return m, nil
	}

	selectedMode := m.currentSeedMode()
	name := sessionflow.SeededName(seed)
	if selectedMode.key != "" {
		name = sessionflow.SeededNameForGame(seed, selectedMode.gameType)
	}

	rec, err := m.store.GetDailyGame(name)
	if err != nil {
		return m.setErrorf("Could not check saved seeded puzzle: %v", err), nil
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

	var spawner game.SeededSpawner
	var gameType string
	modeTitle := ""
	if selectedMode.key == "" {
		spawner, gameType, modeTitle, err = resolve.SeededMode(seed, registry.Entries())
		if err != nil {
			return m.setErrorf("Could not choose a seeded mode: %v", err), nil
		}
	} else {
		spawner, gameType, modeTitle, err = resolve.SeededModeForGame(seed, selectedMode.gameType, registry.Entries())
		if err != nil {
			return m.setErrorf("Could not choose a seeded mode for %s: %v", selectedMode.gameType, err), nil
		}
	}

	rng := resolve.RNGFromString(seed)
	ctx, jobID := m.beginSpawnContext()
	m.session.spawn = &spawnRequest{
		source:      spawnSourceSeed,
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
