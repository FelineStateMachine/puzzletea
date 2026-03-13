package app

import (
	"github.com/FelineStateMachine/puzzletea/game"
	"github.com/FelineStateMachine/puzzletea/theme"
	"github.com/FelineStateMachine/puzzletea/ui"

	tea "charm.land/bubbletea/v2"
)

func (m model) updateActiveState(msg tea.Msg) (model, tea.Cmd) {
	var cmd tea.Cmd

	switch m.state {
	case mainMenuView:
		updateMainMenuCursor(msg, &m.nav.mainMenu)
	case playMenuView:
		updateMainMenuCursor(msg, &m.nav.playMenu)
	case optionsMenuView:
		updateMainMenuCursor(msg, &m.nav.optionsMenu)
	case seedInputView:
		m, cmd = m.handleSeedInputUpdate(msg)
	case generatingView:
		m.spinner, cmd = m.spinner.Update(msg)
	case gameView:
		m.session.game, cmd = m.session.game.Update(msg)
		if m.debug.enabled {
			m.debug.info = m.renderDebugInfo()
		}
		m = m.persistCompletionIfSolved()
	case gameSelectView:
		prev := selectedCategoryName(m.nav.gameSelectList.SelectedItem())
		m.nav.gameSelectList, cmd = m.nav.gameSelectList.Update(msg)
		if selectedCategoryName(m.nav.gameSelectList.SelectedItem()) != prev {
			m = m.updateCategoryDetailViewport()
		}
	case modeSelectView:
		m.nav.modeSelectList, cmd = m.nav.modeSelectList.Update(msg)
	case continueView:
		m.nav.continueTable, cmd = m.nav.continueTable.Update(msg)
	case weeklyView:
		m.nav.weeklyTable, cmd = m.nav.weeklyTable.Update(msg)
	case helpSelectView:
		m.nav.helpSelectList, cmd = m.nav.helpSelectList.Update(msg)
	case helpDetailView:
		m.help.viewport, cmd = m.help.viewport.Update(msg)
	case statsView:
		m.stats.viewport, cmd = m.stats.viewport.Update(msg)
	case themeSelectView:
		prev := m.theme.list.Index()
		m.theme.list, cmd = m.theme.list.Update(msg)
		if m.theme.list.Index() != prev {
			if item, ok := m.theme.list.SelectedItem().(ui.MenuItem); ok {
				_ = theme.Apply(item.ItemTitle)
				ui.UpdateThemeListStyles(&m.theme.list)
			}
		}
	}

	return m, cmd
}

func (m model) resizeActiveState(menuW int) model {
	switch m.state {
	case seedInputView:
		m.nav.seedInput.SetWidth(min(m.width, 48))
	case modeSelectView:
		m.nav.modeSelectList.SetSize(menuW, min(m.height, ui.ListHeight(m.nav.modeSelectList)))
	case continueView:
		m.nav.continueTable.SetWidth(m.width)
		visibleRows := min(len(m.nav.continueGames), ui.MaxTableRows)
		m.nav.continueTable.SetHeight(min(m.height, visibleRows))
	case weeklyView:
		m = m.refreshWeeklyBrowser()
	case helpSelectView:
		listWidth, listHeight := helpSelectListSize(m.width, m.height, m.nav.helpSelectList)
		m.nav.helpSelectList.SetSize(listWidth, listHeight)
	case helpDetailView:
		m = m.updateHelpDetailViewport()
	case themeSelectView:
		const maxVisibleItems = 8
		listW := min(menuW, theme.MaxNameLen+4)
		m.theme.list.SetSize(listW, min(m.height, maxVisibleItems*3))
	case statsView:
		m = m.updateStatsViewport()
	}

	return m
}

func (m model) handleFrameworkMsg(msg tea.Msg) (model, tea.Cmd, bool) {
	switch msg := msg.(type) {
	case spawnCompleteMsg:
		next, cmd := m.handleSpawnComplete(msg.jobID, msg.result)
		return next.(model), cmd, true
	case game.SpawnCompleteMsg:
		next, cmd := m.handleSpawnComplete(m.session.spawnJobID, msg)
		return next.(model), cmd, true
	case tea.WindowSizeMsg:
		return m.handleWindowSize(msg), nil, false
	case tea.KeyPressMsg:
		next, cmd, done := m.handleGlobalKey(msg)
		return next, cmd, done
	default:
		return m, nil, false
	}
}
