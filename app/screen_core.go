package app

import (
	"time"

	"github.com/FelineStateMachine/puzzletea/theme"
	"github.com/FelineStateMachine/puzzletea/ui"

	tea "charm.land/bubbletea/v2"
)

type screenAction interface {
	applyToModel(m model) (model, tea.Cmd)
}

type openPlayMenuAction struct{}

func (a openPlayMenuAction) applyToModel(m model) (model, tea.Cmd) {
	m.nav.playMenu = ui.NewMainMenu(buildPlayMenuItems(time.Now(), m.currentWeeklyMenuIndex()))
	m.state = playMenuView
	m = m.clearNotice()
	return m.resizeActiveScreen(), nil
}

type openStatsAction struct{}

func (a openStatsAction) applyToModel(m model) (model, tea.Cmd) {
	return asModel(m.handleStatsEnter())
}

type openOptionsMenuAction struct{}

func (a openOptionsMenuAction) applyToModel(m model) (model, tea.Cmd) {
	m.nav.optionsMenu = ui.NewMainMenu(optionsMenuItems)
	m.state = optionsMenuView
	m = m.clearNotice()
	return m.resizeActiveScreen(), nil
}

type quitAction struct{}

func (a quitAction) applyToModel(m model) (model, tea.Cmd) {
	return m, tea.Quit
}

type openGameSelectAction struct{}

func (a openGameSelectAction) applyToModel(m model) (model, tea.Cmd) {
	m.state = gameSelectView
	m = m.updateCategoryDetailViewport()
	m = m.clearNotice()
	return m.resizeActiveScreen(), nil
}

type openContinueAction struct{}

func (a openContinueAction) applyToModel(m model) (model, tea.Cmd) {
	m.cont.table, m.cont.games = ui.InitContinueTable(m.store, m.height)
	m.state = continueView
	m = m.clearNotice()
	return m.resizeActiveScreen(), nil
}

type openDailyAction struct{}

func (a openDailyAction) applyToModel(m model) (model, tea.Cmd) {
	return asModel(m.handleDailyPuzzle())
}

type openExportAction struct{}

func (a openExportAction) applyToModel(m model) (model, tea.Cmd) {
	return asModel(m.handleExportEnter())
}

type openWeeklyAction struct{}

func (a openWeeklyAction) applyToModel(m model) (model, tea.Cmd) {
	return asModel(m.enterWeeklyView())
}

type openSeedInputAction struct{}

func (a openSeedInputAction) applyToModel(m model) (model, tea.Cmd) {
	return m.enterSeedInputView()
}

type backAction struct {
	target viewState
}

func (a backAction) applyToModel(m model) (model, tea.Cmd) {
	if m.state == themeSelectView {
		_ = theme.Apply(m.theme.previous)
	}
	m.state = a.target
	return m.resizeActiveScreen(), nil
}

type gameSelectEnterAction struct{}

func (a gameSelectEnterAction) applyToModel(m model) (model, tea.Cmd) {
	return asModel(m.handleGameSelectEnter())
}

type modeSelectEnterAction struct{}

func (a modeSelectEnterAction) applyToModel(m model) (model, tea.Cmd) {
	return asModel(m.handleModeSelectEnter())
}

type continueEnterAction struct{}

func (a continueEnterAction) applyToModel(m model) (model, tea.Cmd) {
	return asModel(m.handleContinueEnter())
}

type weeklyShiftAction struct {
	delta int
}

func (a weeklyShiftAction) applyToModel(m model) (model, tea.Cmd) {
	m = m.moveWeeklyWeek(a.delta)
	return m, nil
}

type weeklyEnterAction struct{}

func (a weeklyEnterAction) applyToModel(m model) (model, tea.Cmd) {
	return asModel(m.handleWeeklyEnter())
}

type helpSelectEnterAction struct{}

func (a helpSelectEnterAction) applyToModel(m model) (model, tea.Cmd) {
	return asModel(m.handleHelpSelectEnter())
}

type openThemeSelectAction struct{}

func (a openThemeSelectAction) applyToModel(m model) (model, tea.Cmd) {
	return asModel(m.handleThemeEnter())
}

type openHelpSelectAction struct{}

func (a openHelpSelectAction) applyToModel(m model) (model, tea.Cmd) {
	m.help.selectList = ui.InitList(gameCategoryItems, "How to Play")
	listWidth, listHeight := helpSelectListSize(m.width, m.height, m.help.selectList)
	m.help.selectList.SetSize(listWidth, listHeight)
	m.state = helpSelectView
	m = m.clearNotice()
	return m.resizeActiveScreen(), nil
}

type previewThemeAction struct {
	name string
}

func (a previewThemeAction) applyToModel(m model) (model, tea.Cmd) {
	_ = theme.Apply(a.name)
	ui.UpdateThemeListStyles(&m.theme.list)
	return m, nil
}

type confirmThemeAction struct{}

func (a confirmThemeAction) applyToModel(m model) (model, tea.Cmd) {
	return asModel(m.handleThemeConfirm())
}

type seedConfirmAction struct{}

func (a seedConfirmAction) applyToModel(m model) (model, tea.Cmd) {
	return asModel(m.handleSeedConfirm())
}

// exportSubmitAction is a pure tea.Msg emitted by exportSubmitCmd.
// It is handled in the root Update() and does not implement screenAction.
type exportSubmitAction struct{}

type screenModel interface {
	State() viewState
	Resize(width, height int) screenModel
	Update(msg tea.Msg) (screenModel, tea.Cmd, screenAction)
	View(notice noticeState) string
	Apply(m model) model
}

type screenFactory func(m model) screenModel

var screenRegistry = map[viewState]screenFactory{
	mainMenuView: func(m model) screenModel {
		return mainMenuScreen{width: m.width, height: m.height, menu: m.nav.mainMenu}
	},
	playMenuView: func(m model) screenModel {
		return playMenuScreen{width: m.width, height: m.height, menu: m.nav.playMenu}
	},
	optionsMenuView: func(m model) screenModel {
		return optionsMenuScreen{width: m.width, height: m.height, menu: m.nav.optionsMenu}
	},
	seedInputView: func(m model) screenModel {
		return seedInputScreen{width: m.width, height: m.height, seed: m.seed}
	},
	gameSelectView: func(m model) screenModel {
		return gameSelectScreen{width: m.width, height: m.height, list: m.nav.gameSelectList, detail: m.nav.categoryDetail}
	},
	modeSelectView: func(m model) screenModel {
		return modeSelectScreen{width: m.width, height: m.height, entry: m.nav.selectedCategory, list: m.nav.modeSelectList}
	},
	exportView: func(m model) screenModel {
		return exportScreen{width: m.width, height: m.height, export: m.export}
	},
	continueView: func(m model) screenModel {
		return continueScreen{width: m.width, height: m.height, cont: m.cont}
	},
	weeklyView: func(m model) screenModel {
		return weeklyScreen{width: m.width, height: m.height, weekly: m.weekly}
	},
	helpSelectView: func(m model) screenModel {
		return helpSelectScreen{width: m.width, height: m.height, help: m.help}
	},
	helpDetailView: func(m model) screenModel {
		return helpDetailScreen{width: m.width, height: m.height, help: m.help}
	},
	statsView: func(m model) screenModel {
		return statsScreen{width: m.width, height: m.height, stats: m.stats}
	},
	themeSelectView: func(m model) screenModel {
		return themeSelectScreen{width: m.width, height: m.height, theme: m.theme}
	},
	generatingView: func(m model) screenModel {
		return generatingScreen{width: m.width, height: m.height, spinner: m.spinner}
	},
	exportRunningView: func(m model) screenModel {
		return exportRunningScreen{width: m.width, height: m.height, spinner: m.spinner}
	},
}

func (m model) activeScreen() screenModel {
	factory, ok := screenRegistry[m.state]
	if !ok {
		return nil
	}
	return factory(m)
}
