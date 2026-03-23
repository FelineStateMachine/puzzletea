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
	m = m.initScreen(modeSelectView)
	return m, nil
}

func (m model) handleModeSelectEnter() (tea.Model, tea.Cmd) {
	item := unwrapModeDisplayItem(m.nav.modeSelectList.SelectedItem())
	mode, ok := item.(registry.ModeEntry)
	if !ok {
		return m, nil
	}
	m.nav.selectedModeTitle = mode.Definition.Title
	cmd := newSessionController(&m).startSpawn(mode.Spawner, spawnRequest{
		source:      spawnSourceNormal,
		name:        sessionflow.GenerateUniqueName(m.store),
		gameType:    m.nav.selectedCategory.Definition.Name,
		modeTitle:   m.nav.selectedModeTitle,
		run:         store.NormalRunMetadata(),
		returnState: modeSelectView,
		exitState:   mainMenuView,
	})
	return m, cmd
}

func (m model) handleContinueEnter() (tea.Model, tea.Cmd) {
	idx := m.cont.table.Cursor()
	if idx < 0 || idx >= len(m.cont.games) {
		return m, nil
	}
	rec := m.cont.games[idx]
	m, _ = m.importAndActivateRecord(rec)
	return m, nil
}

func (m model) persistCompletionIfSolved() model {
	newSessionController(&m).persistCompletionIfSolved()
	return m
}
