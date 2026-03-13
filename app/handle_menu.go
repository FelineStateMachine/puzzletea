package app

import (
	"time"

	"github.com/FelineStateMachine/puzzletea/theme"
	"github.com/FelineStateMachine/puzzletea/ui"

	tea "charm.land/bubbletea/v2"
)

func (m model) handleMainMenuEnter() (tea.Model, tea.Cmd) {
	item := m.nav.mainMenu.Selected()
	switch item.Title() {
	case "Play":
		m.nav.playMenu = ui.NewMainMenu(buildPlayMenuItems(time.Now(), m.currentWeeklyMenuIndex()))
		m.state = playMenuView
	case "Stats":
		return m.handleStatsEnter()
	case "Options":
		m.nav.optionsMenu = ui.NewMainMenu(optionsMenuItems)
		m.state = optionsMenuView
	case "Quit":
		return m, tea.Quit
	}
	return m, nil
}

func (m model) currentWeeklyMenuIndex() int {
	if m.store == nil {
		return 1
	}

	year, week := time.Now().ISOWeek()
	highestCompleted, err := m.store.GetCurrentWeeklyHighestCompletedIndex(year, week)
	if err != nil {
		return 1
	}
	if highestCompleted >= weeklyEntryCount {
		return weeklyEntryCount
	}
	if highestCompleted < 1 {
		return 1
	}
	return highestCompleted + 1
}

func (m model) handlePlayMenuEnter() (tea.Model, tea.Cmd) {
	item := m.nav.playMenu.Selected()
	switch item.Title() {
	case "Create":
		m.state = gameSelectView
		m = m.updateCategoryDetailViewport()
	case "Continue":
		m.nav.continueTable, m.nav.continueGames = ui.InitContinueTable(m.store, m.height)
		m.state = continueView
	case "Daily":
		return m.handleDailyPuzzle()
	case "Weekly":
		return m.enterWeeklyView()
	case "Seeded":
		return m.enterSeedInputView()
	}
	return m, nil
}

func (m model) handleOptionsMenuEnter() (tea.Model, tea.Cmd) {
	item := m.nav.optionsMenu.Selected()
	switch item.Title() {
	case "Theme":
		return m.handleThemeEnter()
	case "Guides":
		m.nav.helpSelectList = ui.InitList(gameCategoryItems, "How to Play")
		listWidth, listHeight := helpSelectListSize(m.width, m.height, m.nav.helpSelectList)
		m.nav.helpSelectList.SetSize(listWidth, listHeight)
		m.state = helpSelectView
	}
	return m, nil
}

func (m model) handleEscape() (tea.Model, tea.Cmd) {
	switch m.state {
	case playMenuView, optionsMenuView, statsView:
		m.state = mainMenuView
	case seedInputView, gameSelectView, continueView, weeklyView:
		m.state = playMenuView
	case generatingView:
		returnState := m.activeSpawnReturnState()
		m.cancelActiveSpawn()
		m.state = returnState
	case modeSelectView:
		m.state = gameSelectView
		m = m.updateCategoryDetailViewport()
	case helpDetailView:
		m.state = helpSelectView
	case helpSelectView, themeSelectView:
		if m.state == themeSelectView {
			_ = theme.Apply(m.theme.previous)
		}
		m.state = optionsMenuView
	}
	return m, nil
}

func updateMainMenuCursor(msg tea.Msg, menu *ui.MainMenu) {
	keyMsg, ok := msg.(tea.KeyPressMsg)
	if !ok {
		return
	}
	switch keyMsg.String() {
	case "up", "k":
		menu.CursorUp()
	case "down", "j":
		menu.CursorDown()
	}
}
