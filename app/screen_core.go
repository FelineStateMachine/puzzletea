package app

import (
	"time"

	"github.com/FelineStateMachine/puzzletea/config"
	"github.com/FelineStateMachine/puzzletea/registry"
	sessionflow "github.com/FelineStateMachine/puzzletea/session"
	"github.com/FelineStateMachine/puzzletea/store"
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
	return m.initScreen(playMenuView), nil
}

type openStatsAction struct{}

func (a openStatsAction) applyToModel(m model) (model, tea.Cmd) {
	return m.handleStatsEnter()
}

type openOptionsMenuAction struct{}

func (a openOptionsMenuAction) applyToModel(m model) (model, tea.Cmd) {
	m.nav.optionsMenu = ui.NewMainMenu(optionsMenuItems)
	m.state = optionsMenuView
	m = m.clearNotice()
	return m.initScreen(optionsMenuView), nil
}

type quitAction struct{}

func (a quitAction) applyToModel(m model) (model, tea.Cmd) {
	return m, tea.Quit
}

type openCreateAction struct{}

func (a openCreateAction) applyToModel(m model) (model, tea.Cmd) {
	createCfg := config.CreateConfig{}
	if m.cfg != nil {
		createCfg = m.cfg.Create
	}
	m.create = newCreateState(createCfg, m.width)
	m.state = createView
	m = m.clearNotice()
	return m.initScreen(createView), nil
}

type openContinueAction struct{}

func (a openContinueAction) applyToModel(m model) (model, tea.Cmd) {
	m.cont.table, m.cont.games = ui.InitContinueTable(m.store, m.height)
	m.state = continueView
	m = m.clearNotice()
	return m.initScreen(continueView), nil
}

type openDailyAction struct{}

func (a openDailyAction) applyToModel(m model) (model, tea.Cmd) {
	return m.handleDailyPuzzle()
}

type openExportAction struct{}

func (a openExportAction) applyToModel(m model) (model, tea.Cmd) {
	return m.handleExportEnter()
}

type openWeeklyAction struct{}

func (a openWeeklyAction) applyToModel(m model) (model, tea.Cmd) {
	return m.enterWeeklyView()
}

type backAction struct {
	target viewState
}

func (a backAction) applyToModel(m model) (model, tea.Cmd) {
	if m.state == themeSelectView {
		previous := m.theme.previous
		if ts, ok := m.screens[themeSelectView].(themeSelectScreen); ok {
			previous = ts.theme.previous
		}
		_ = theme.Apply(previous)
	}
	m.state = a.target
	return m.resizeActiveScreen(), nil
}

type gameSelectEnterAction struct {
	entry registry.Entry
}

func (a gameSelectEnterAction) applyToModel(m model) (model, tea.Cmd) {
	m.nav.selectedCategory = a.entry
	m.nav.modeSelectList = ui.InitList(buildModeDisplayItems(a.entry), a.entry.Definition.Name+" - Select Mode")
	m.nav.modeSelectList.SetSize(min(m.width, 64), min(m.height, ui.ListHeight(m.nav.modeSelectList)))
	m.state = modeSelectView
	m = m.initScreen(modeSelectView)
	return m, nil
}

type modeSelectEnterAction struct {
	entry registry.Entry
	mode  registry.ModeEntry
}

func (a modeSelectEnterAction) applyToModel(m model) (model, tea.Cmd) {
	m.nav.selectedCategory = a.entry
	m.nav.selectedModeTitle = a.mode.Definition.Title
	name := sessionflow.GenerateUniqueName(m.store)
	request := spawnRequest{
		source:      spawnSourceNormal,
		name:        name,
		gameType:    a.entry.Definition.Name,
		modeTitle:   a.mode.Definition.Title,
		run:         store.NormalRunMetadata(),
		returnState: modeSelectView,
		exitState:   mainMenuView,
	}
	if a.mode.Definition.PresetElo != nil && a.mode.Elo != nil {
		cmd := newSessionController(&m).startEloSpawn(a.mode.Elo, name, *a.mode.Definition.PresetElo, request)
		return m, cmd
	}
	cmd := newSessionController(&m).startSpawn(a.mode.Spawner, request)
	return m, cmd
}

type continueEnterAction struct {
	record store.GameRecord
}

func (a continueEnterAction) applyToModel(m model) (model, tea.Cmd) {
	m, _ = m.importAndActivateRecord(a.record)
	return m, nil
}

type weeklyShiftAction struct {
	delta  int
	weekly weeklyState
}

func (a weeklyShiftAction) applyToModel(m model) (model, tea.Cmd) {
	m.weekly = a.weekly
	m = m.moveWeeklyWeek(a.delta)
	m = m.initScreen(weeklyView)
	return m, nil
}

type weeklyEnterAction struct {
	weekly weeklyState
}

func (a weeklyEnterAction) applyToModel(m model) (model, tea.Cmd) {
	m.weekly = a.weekly
	return m.handleWeeklyEnter()
}

type helpSelectEnterAction struct {
	entry registry.Entry
}

func (a helpSelectEnterAction) applyToModel(m model) (model, tea.Cmd) {
	m.help.category = a.entry
	return m.handleHelpSelectEnter()
}

type openThemeSelectAction struct{}

func (a openThemeSelectAction) applyToModel(m model) (model, tea.Cmd) {
	return m.handleThemeEnter()
}

type openHelpSelectAction struct{}

func (a openHelpSelectAction) applyToModel(m model) (model, tea.Cmd) {
	m.help.selectList = ui.InitList(gameCategoryItems, "How to Play")
	m.state = helpSelectView
	m = m.clearNotice()
	return m.initScreen(helpSelectView), nil
}

type previewThemeAction struct {
	name string
}

func (a previewThemeAction) applyToModel(m model) (model, tea.Cmd) {
	_ = theme.Apply(a.name)
	if ts, ok := m.screens[themeSelectView].(themeSelectScreen); ok {
		ui.UpdateThemeListStyles(&ts.theme.list)
		m.screens[themeSelectView] = ts
	}
	return m, nil
}

type confirmThemeAction struct {
	theme themeState
}

func (a confirmThemeAction) applyToModel(m model) (model, tea.Cmd) {
	m.theme = a.theme
	return m.handleThemeConfirm()
}

type seedConfirmAction struct {
	seed seedState
}

func (a seedConfirmAction) applyToModel(m model) (model, tea.Cmd) {
	m.seed = a.seed
	return m.handleSeedConfirm()
}

// exportSubmitAction is a pure tea.Msg emitted by exportSubmitCmd.
// It is handled in the root Update() and does not implement screenAction.
type exportSubmitAction struct{}

type screenModel interface {
	Resize(width, height int) screenModel
	Update(msg tea.Msg) (screenModel, tea.Cmd, screenAction)
	View(notice noticeState) string
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
	createView: func(m model) screenModel {
		return createScreen{width: m.width, height: m.height, create: m.create}
	},
	gameSelectView: func(m model) screenModel {
		return gameSelectScreen{width: m.width, height: m.height, list: m.nav.gameSelectList, detail: m.nav.categoryDetail}
	},
	modeSelectView: func(m model) screenModel {
		return modeSelectScreen{width: m.width, height: m.height, entry: m.nav.selectedCategory, list: m.nav.modeSelectList}
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
	return m.screens[m.state]
}

// initScreen creates a fresh screen for the given state (using the registry
// factory) and stores it in m.screens. It also calls Resize so dimensions
// are set correctly.
func (m model) initScreen(state viewState) model {
	factory, ok := screenRegistry[state]
	if !ok {
		return m
	}
	if m.screens == nil {
		m.screens = make(map[viewState]screenModel)
	}
	m.screens[state] = factory(m).Resize(m.width, m.height)
	return m
}
