package app

import (
	"math/rand/v2"

	"github.com/FelineStateMachine/puzzletea/config"
	sessionflow "github.com/FelineStateMachine/puzzletea/session"
	"github.com/FelineStateMachine/puzzletea/store"

	tea "charm.land/bubbletea/v2"
)

func (m model) handleCreateGenerate() (model, tea.Cmd) {
	leaves := m.create.selectedLeaves()
	if len(leaves) == 0 {
		return m.setErrorf("Select at least one puzzle."), nil
	}

	elo, err := m.create.elo()
	if err != nil {
		return m.setErrorf("Invalid Elo: %v", err), nil
	}

	leaf := leaves[m.createChoiceIndex(len(leaves))]
	spawner, modeTitle, err := m.create.resolveLeafVariant(leaf, elo)
	if err != nil {
		return m.setErrorf("Could not prepare puzzle: %v", err), nil
	}

	if err := m.persistCreatePreset(); err != nil {
		return m.setErrorf("Could not save create settings: %v", err), nil
	}

	seed := ""
	name := sessionflow.GenerateUniqueName(m.store)
	run := store.NormalRunMetadata()
	if len(leaves) == 1 {
		seed = normalizedCreateSeed(m.create.seedInput.Value())
	}
	if seed != "" {
		name = sessionflow.SeededNameForCreateLeaf(seed, leaf.gameType, leaf.id, elo)
		run = store.SeededRunMetadata(seed)
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
	} else {
		seed = name
	}

	cmd := newSessionController(&m).startEloSpawn(spawner, seed, elo, spawnRequest{
		source:      spawnSourceNormal,
		name:        name,
		gameType:    leaf.gameType,
		modeTitle:   modeTitle,
		run:         run,
		returnState: createView,
		exitState:   mainMenuView,
	})
	return m, cmd
}

func (m model) createChoiceIndex(n int) int {
	if n <= 1 {
		return 0
	}
	if m.createRandIndex != nil {
		idx := m.createRandIndex(n)
		if idx < 0 || idx >= n {
			return 0
		}
		return idx
	}
	return rand.IntN(n)
}

func (m model) persistCreatePresetIfValid() model {
	if len(m.create.selectedLeaves()) == 0 {
		return m
	}
	if _, err := m.create.elo(); err != nil {
		return m
	}
	if err := m.persistCreatePreset(); err != nil {
		return m.setErrorf("Could not save create settings: %v", err)
	}
	return m
}

func (m model) persistCreatePreset() error {
	if m.cfg == nil {
		m.cfg = config.Default()
	}
	elo, err := m.create.elo()
	if err != nil {
		return err
	}
	m.cfg.Create.SelectedLeafIDs = m.create.selectedLeafIDs()
	m.cfg.Create.Elo = int(elo)
	if m.configPath == "" {
		return nil
	}
	return m.cfg.Save(m.configPath)
}
