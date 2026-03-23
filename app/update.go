package app

import (
	"time"

	"github.com/FelineStateMachine/puzzletea/game"
	"github.com/FelineStateMachine/puzzletea/store"
	"github.com/FelineStateMachine/puzzletea/theme"
	"github.com/FelineStateMachine/puzzletea/ui"

	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
)

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case spawnCompleteMsg:
		next, cmd := m.handleSpawnComplete(msg.jobID, msg.result)
		return next, cmd
	case game.SpawnCompleteMsg:
		next, cmd := m.handleSpawnComplete(m.session.spawnJobID, msg)
		return next, cmd
	case exportCompleteMsg:
		return m.handleExportComplete(msg)
	case tea.WindowSizeMsg:
		m = m.handleWindowSize(msg)
		if m.state != gameView {
			return m, nil
		}
	}

	next, cmd, handled := m.handleGlobalKey(msg)
	if handled {
		return next, cmd
	}
	m = next

	if m.state == gameView {
		gameCmd := newSessionController(&m).updateActiveGame(msg)
		return m, gameCmd
	}

	screen := m.activeScreen()
	if screen == nil {
		return m, nil
	}

	nextScreen, screenCmd, action := screen.Update(msg)
	m = nextScreen.Apply(m)
	if action == nil {
		return m, screenCmd
	}

	next, actionCmd := m.handleScreenAction(action)
	return next, tea.Batch(screenCmd, actionCmd)
}

func (m model) resizeActiveScreen() model {
	screen := m.activeScreen()
	if screen == nil {
		return m
	}
	return screen.Resize(m.width, m.height).Apply(m)
}

func (m model) handleWindowSize(msg tea.WindowSizeMsg) model {
	m.width = msg.Width
	m.height = msg.Height
	if m.state == gameView {
		return m
	}
	return m.resizeActiveScreen()
}

func (m model) handleGlobalKey(msg tea.Msg) (model, tea.Cmd, bool) {
	keyMsg, ok := msg.(tea.KeyPressMsg)
	if !ok {
		return m, nil, false
	}

	if m.state == generatingView {
		switch {
		case key.Matches(keyMsg, rootKeys.Escape):
			returnState := m.activeSpawnReturnState()
			m.cancelActiveSpawn()
			m.state = returnState
			return m.resizeActiveScreen(), nil, true
		case key.Matches(keyMsg, rootKeys.Quit):
			return m, tea.Quit, true
		}
		return m, nil, true
	}

	if m.state == exportRunningView {
		switch {
		case key.Matches(keyMsg, rootKeys.Escape):
			m.cancelActiveExport()
			m.state = exportView
			return m.resizeActiveScreen(), nil, true
		case key.Matches(keyMsg, rootKeys.Quit):
			return m, tea.Quit, true
		}
		return m, nil, true
	}

	if m.state == gameView {
		switch {
		case key.Matches(keyMsg, rootKeys.Enter):
			next, cmd, handled := m.advanceSolvedWeekly()
			if handled {
				return next, cmd, true
			}
		case key.Matches(keyMsg, rootKeys.Escape):
			returnState := m.session.returnState
			m = saveCurrentGame(m, store.StatusInProgress)
			m.state = returnState
			if returnState == weeklyView {
				m = m.refreshWeeklyBrowser()
			}
			m.debug.enabled = false
			return m, nil, true
		case key.Matches(keyMsg, rootKeys.Quit):
			m = saveCurrentGame(m, store.StatusAbandoned)
			return m, tea.Quit, true
		case key.Matches(keyMsg, rootKeys.Debug):
			m.debug.enabled = !m.debug.enabled
			return m, nil, true
		case key.Matches(keyMsg, rootKeys.FullHelp):
			m.help.showFull = !m.help.showFull
			if m.session.game != nil {
				m.session.game, _ = m.session.game.Update(game.HelpToggleMsg{Show: m.help.showFull})
			}
			return m, nil, true
		case key.Matches(keyMsg, rootKeys.ResetGame):
			if m.session.game != nil {
				m.session.game = m.session.game.Reset()
			}
			return m, nil, true
		}
		return m, nil, false
	}

	switch {
	case key.Matches(keyMsg, rootKeys.Quit):
		return m, tea.Quit, true
	case key.Matches(keyMsg, rootKeys.Debug):
		m.debug.enabled = !m.debug.enabled
		return m, nil, true
	case key.Matches(keyMsg, rootKeys.FullHelp):
		m.help.showFull = !m.help.showFull
		return m, nil, true
	default:
		return m, nil, false
	}
}

func (m model) activeSpawnReturnState() viewState {
	if m.session.spawn == nil {
		return modeSelectView
	}
	return m.session.spawn.returnState
}

func (m model) handleScreenAction(action screenAction) (model, tea.Cmd) {
	switch action := action.(type) {
	case openPlayMenuAction:
		m.nav.playMenu = ui.NewMainMenu(buildPlayMenuItems(time.Now(), m.currentWeeklyMenuIndex()))
		m.state = playMenuView
		m = m.clearNotice()
		return m.resizeActiveScreen(), nil
	case openStatsAction:
		return asModel(m.handleStatsEnter())
	case openOptionsMenuAction:
		m.nav.optionsMenu = ui.NewMainMenu(optionsMenuItems)
		m.state = optionsMenuView
		m = m.clearNotice()
		return m.resizeActiveScreen(), nil
	case quitAction:
		return m, tea.Quit
	case openGameSelectAction:
		m.state = gameSelectView
		m = m.updateCategoryDetailViewport()
		m = m.clearNotice()
		return m.resizeActiveScreen(), nil
	case openContinueAction:
		m.cont.table, m.cont.games = ui.InitContinueTable(m.store, m.height)
		m.state = continueView
		m = m.clearNotice()
		return m.resizeActiveScreen(), nil
	case openDailyAction:
		return asModel(m.handleDailyPuzzle())
	case openExportAction:
		return asModel(m.handleExportEnter())
	case openWeeklyAction:
		return asModel(m.enterWeeklyView())
	case openSeedInputAction:
		return m.enterSeedInputView()
	case backAction:
		if m.state == themeSelectView {
			_ = theme.Apply(m.theme.previous)
		}
		m.state = action.target
		return m.resizeActiveScreen(), nil
	case gameSelectEnterAction:
		return asModel(m.handleGameSelectEnter())
	case modeSelectEnterAction:
		return asModel(m.handleModeSelectEnter())
	case continueEnterAction:
		return asModel(m.handleContinueEnter())
	case weeklyShiftAction:
		m = m.moveWeeklyWeek(action.delta)
		return m, nil
	case weeklyEnterAction:
		return asModel(m.handleWeeklyEnter())
	case helpSelectEnterAction:
		return asModel(m.handleHelpSelectEnter())
	case openThemeSelectAction:
		return asModel(m.handleThemeEnter())
	case openHelpSelectAction:
		m.help.selectList = ui.InitList(gameCategoryItems, "How to Play")
		listWidth, listHeight := helpSelectListSize(m.width, m.height, m.help.selectList)
		m.help.selectList.SetSize(listWidth, listHeight)
		m.state = helpSelectView
		m = m.clearNotice()
		return m.resizeActiveScreen(), nil
	case previewThemeAction:
		_ = theme.Apply(action.name)
		ui.UpdateThemeListStyles(&m.theme.list)
		return m, nil
	case confirmThemeAction:
		return asModel(m.handleThemeConfirm())
	case seedConfirmAction:
		return asModel(m.handleSeedConfirm())
	case exportSubmitAction:
		return asModel(m.handleExportSubmit())
	default:
		return m, nil
	}
}

func asModel(next tea.Model, cmd tea.Cmd) (model, tea.Cmd) {
	return next.(model), cmd
}
