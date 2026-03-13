package app

import (
	"github.com/FelineStateMachine/puzzletea/registry"
	sessionflow "github.com/FelineStateMachine/puzzletea/session"
	"github.com/FelineStateMachine/puzzletea/store"
	"github.com/FelineStateMachine/puzzletea/ui"

	tea "charm.land/bubbletea/v2"
)

func (m model) handleGameSelectEnter() (tea.Model, tea.Cmd) {
	entry, ok := selectedCategoryEntry(m.nav.gameSelectList.SelectedItem())
	if !ok {
		return m, nil
	}
	m.nav.selectedCategory = entry
	m.nav.modeSelectList = ui.InitList(buildModeDisplayItems(entry), entry.Definition.Name+" - Select Mode")
	m.nav.modeSelectList.SetSize(min(m.width, 64), min(m.height, ui.ListHeight(m.nav.modeSelectList)))
	m.state = modeSelectView
	return m, nil
}

func (m model) handleModeSelectEnter() (tea.Model, tea.Cmd) {
	item := unwrapModeDisplayItem(m.nav.modeSelectList.SelectedItem())
	mode, ok := item.(registry.ModeEntry)
	if !ok {
		return m, nil
	}
	m.nav.selectedModeTitle = mode.Definition.Title
	ctx, jobID := m.beginSpawnContext()
	m.session.spawn = &spawnRequest{
		source:      spawnSourceNormal,
		name:        sessionflow.GenerateUniqueName(m.store),
		gameType:    m.nav.selectedCategory.Definition.Name,
		modeTitle:   m.nav.selectedModeTitle,
		returnState: modeSelectView,
		exitState:   mainMenuView,
	}
	m.state = generatingView
	return m, tea.Batch(m.spinner.Tick, spawnCmd(mode.Spawner, ctx, jobID))
}

func (m model) handleContinueEnter() (tea.Model, tea.Cmd) {
	idx := m.nav.continueTable.Cursor()
	if idx < 0 || idx >= len(m.nav.continueGames) {
		return m, nil
	}
	rec := m.nav.continueGames[idx]
	m, _ = m.importAndActivateRecord(rec)
	return m, nil
}

func (m model) persistCompletionIfSolved() model {
	if m.session.game == nil || m.session.activeGameID == 0 ||
		m.session.completionSaved || !m.session.game.IsSolved() {
		return m
	}

	m.session.completionSaved = true
	saveData, err := m.session.game.GetSave()
	if err == nil {
		if err := m.store.UpdateSaveState(m.session.activeGameID, string(saveData)); err != nil {
			m = m.setErrorf("Puzzle completed, but saving the final state failed: %v", err)
		}
	} else {
		m = m.setErrorf("Puzzle completed, but reading the final state failed: %v", err)
	}
	if err := m.store.UpdateStatus(m.session.activeGameID, store.StatusCompleted); err != nil {
		m = m.setErrorf("Puzzle completed, but updating completion status failed: %v", err)
	}
	return m
}
