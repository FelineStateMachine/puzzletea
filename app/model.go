// Package app implements the root TUI model for PuzzleTea, including
// the Elm-architecture update/view loop, game spawning, and debug overlay.
package app

import (
	"context"
	"strconv"
	"time"

	"github.com/FelineStateMachine/puzzletea/config"
	"github.com/FelineStateMachine/puzzletea/game"
	"github.com/FelineStateMachine/puzzletea/registry"
	"github.com/FelineStateMachine/puzzletea/stats"
	"github.com/FelineStateMachine/puzzletea/store"
	"github.com/FelineStateMachine/puzzletea/theme"
	"github.com/FelineStateMachine/puzzletea/ui"
	"github.com/FelineStateMachine/puzzletea/weekly"

	"charm.land/bubbles/v2/list"
	"charm.land/bubbles/v2/spinner"
	"charm.land/bubbles/v2/table"
	"charm.land/bubbles/v2/textinput"
	"charm.land/bubbles/v2/viewport"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/glamour"
)

var (
	gameCategoryItems = buildCategoryItems()
	GameCategories    = gameCategoryItems
)

type viewState int

const (
	mainMenuActionPlay    = "play"
	mainMenuActionExport  = "export"
	mainMenuActionStats   = "stats"
	mainMenuActionOptions = "options"
	mainMenuActionQuit    = "quit"

	playMenuActionCreate   = "create"
	playMenuActionContinue = "continue"
	playMenuActionDaily    = "daily"
	playMenuActionWeekly   = "weekly"
	playMenuActionSeeded   = "seeded"

	optionsMenuActionTheme  = "theme"
	optionsMenuActionGuides = "guides"
)

var mainMenuItems = []ui.MenuItem{
	{Action: mainMenuActionPlay, ItemTitle: "Play", Desc: "start or continue a puzzle"},
	{Action: mainMenuActionExport, ItemTitle: "Export", Desc: "build a printable puzzle pack"},
	{Action: mainMenuActionStats, ItemTitle: "Stats", Desc: "your progress"},
	{Action: mainMenuActionOptions, ItemTitle: "Options", Desc: "configure and learn"},
	{Action: mainMenuActionQuit, ItemTitle: "Quit", Desc: "exit puzzletea"},
}

var optionsMenuItems = []ui.MenuItem{
	{Action: optionsMenuActionTheme, ItemTitle: "Theme", Desc: "change colors"},
	{Action: optionsMenuActionGuides, ItemTitle: "Guides", Desc: "learn the rules"},
}

const (
	mainMenuView viewState = iota
	playMenuView
	optionsMenuView
	seedInputView
	gameSelectView
	modeSelectView
	exportView
	generatingView
	exportRunningView
	gameView
	continueView
	weeklyView
	helpSelectView
	helpDetailView
	statsView
	themeSelectView
)

type navigationState struct {
	mainMenu          ui.MainMenu
	playMenu          ui.MainMenu
	optionsMenu       ui.MainMenu
	gameSelectList    list.Model
	categoryDetail    viewport.Model
	modeSelectList    list.Model
	selectedCategory  registry.Entry
	selectedModeTitle string
}

type continueState struct {
	table table.Model
	games []store.GameRecord
}

type weeklyState struct {
	table  table.Model
	rows   []weeklyRow
	cursor time.Time
}

type seedState struct {
	input       textinput.Model
	modeOptions []seedModeOption
	modeIndex   int
	focus       seedInputFocus
	lastModeKey string
}

type sessionState struct {
	game            game.Gamer
	activeGameID    int64
	completionSaved bool
	returnState     viewState
	weeklyAdvance   *weekly.Info
	generating      bool
	spawnJobID      int64
	spawnCancel     context.CancelFunc
	spawn           *spawnRequest
}

type helpState struct {
	selectList    list.Model
	category      registry.Entry
	viewport      viewport.Model
	renderer      *glamour.TermRenderer
	rendererWidth int
	rendererTheme string
	showFull      bool
}

type statsState struct {
	cards    []stats.Card
	profile  stats.ProfileBanner
	viewport viewport.Model
}

type themeState struct {
	list     list.Model
	previous string
}

type debugState struct {
	enabled  bool
	renderer *glamour.TermRenderer
	info     string
}

type spawnSource string

const (
	spawnSourceNormal spawnSource = "normal"
	spawnSourceDaily  spawnSource = "daily"
	spawnSourceSeed   spawnSource = "seed"
	spawnSourceWeekly spawnSource = "weekly"
)

type spawnRequest struct {
	source      spawnSource
	name        string
	gameType    string
	modeTitle   string
	run         store.RunMetadata
	returnState viewState
	exitState   viewState
	weeklyInfo  *weekly.Info
}

type seedInputFocus int

const (
	seedFocusText seedInputFocus = iota
	seedFocusMode
)

type seedModeOption struct {
	key      string
	label    string
	gameType string
}

type model struct {
	state   viewState
	screens map[viewState]screenModel // persistent screen instances

	nav     navigationState
	cont    continueState
	session sessionState
	seed    seedState
	weekly  weeklyState
	help    helpState
	stats   statsState
	theme   themeState
	export  exportModel
	debug   debugState
	notice  noticeState

	spinner spinner.Model

	width  int // available content width (terminal - rootStyle frame)
	height int // available content height (terminal - rootStyle frame)

	store *store.Store

	// Config
	cfg        *config.Config
	configPath string
}

func newSpinner() spinner.Model {
	p := theme.Current()
	sp := spinner.New()
	sp.Spinner = spinner.Dot
	sp.Style = lipgloss.NewStyle().Foreground(p.Accent)
	return sp
}

// InitialModel creates the root TUI model for the main menu.
func InitialModel(s *store.Store, cfg *config.Config, configPath string) model {
	r := initDebugRenderer()
	l := ui.InitCategoryList(gameCategoryItems, "Select Category")
	mm := ui.NewMainMenu(mainMenuItems)
	m := model{
		state:      mainMenuView,
		screens:    make(map[viewState]screenModel),
		debug:      debugState{renderer: r},
		nav:        navigationState{gameSelectList: l, mainMenu: mm},
		spinner:    newSpinner(),
		store:      s,
		cfg:        cfg,
		configPath: configPath,
	}
	return m.initScreen(mainMenuView)
}

// InitialModelWithGame creates a model that starts directly in gameView,
// bypassing the menu. Used by CLI flags (--new, --continue).
func InitialModelWithGame(s *store.Store, cfg *config.Config, configPath string, g game.Gamer, activeGameID int64, completionSaved bool) model {
	r := initDebugRenderer()
	l := ui.InitCategoryList(gameCategoryItems, "Select Category")
	mm := ui.NewMainMenu(mainMenuItems)
	m := model{
		state:      gameView,
		screens:    make(map[viewState]screenModel),
		debug:      debugState{renderer: r},
		nav:        navigationState{gameSelectList: l, mainMenu: mm},
		spinner:    newSpinner(),
		store:      s,
		cfg:        cfg,
		configPath: configPath,
		session: sessionState{
			game:            g,
			activeGameID:    activeGameID,
			completionSaved: completionSaved,
			returnState:     mainMenuView,
		},
	}
	// Pre-cache the main menu screen so back-navigation from game works.
	return m.initScreen(mainMenuView)
}

func (m model) Init() tea.Cmd {
	return nil
}

func buildPlayMenuItems(now time.Time, currentWeeklyIndex int) []ui.MenuItem {
	year, week := now.ISOWeek()
	return []ui.MenuItem{
		{Action: playMenuActionCreate, ItemTitle: "Create", Desc: "a new puzzle"},
		{Action: playMenuActionContinue, ItemTitle: "Continue", Desc: "a previously played puzzle"},
		{Action: playMenuActionDaily, ItemTitle: "Daily", Desc: now.Format("Jan _2 06")},
		{Action: playMenuActionWeekly, ItemTitle: "Weekly", Desc: "Week " + formatTwoDigits(week) + "-" + strconv.Itoa(year) + " #" + formatWeeklyMenuIndex(currentWeeklyIndex)},
		{Action: playMenuActionSeeded, ItemTitle: "Seeded", Desc: "enter a specific seed"},
	}
}
