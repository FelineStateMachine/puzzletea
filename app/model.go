// Package app implements the root TUI model for PuzzleTea, including
// the Elm-architecture update/view loop, game spawning, and debug overlay.
package app

import (
	"context"
	"time"

	"github.com/FelineStateMachine/puzzletea/catalog"
	"github.com/FelineStateMachine/puzzletea/config"
	"github.com/FelineStateMachine/puzzletea/game"
	"github.com/FelineStateMachine/puzzletea/stats"
	"github.com/FelineStateMachine/puzzletea/store"
	"github.com/FelineStateMachine/puzzletea/theme"
	"github.com/FelineStateMachine/puzzletea/ui"

	"charm.land/bubbles/v2/list"
	"charm.land/bubbles/v2/spinner"
	"charm.land/bubbles/v2/table"
	"charm.land/bubbles/v2/textinput"
	"charm.land/bubbles/v2/viewport"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/glamour"
)

// Categories is the typed catalog view used by stats and print-export helpers.
var Categories = catalog.Categories()

// GameCategories is the list.Item view of Categories for Bubble Tea lists.
var GameCategories = catalog.CategoryItems()

type viewState int

var mainMenuItems = []ui.MenuItem{
	{ItemTitle: "Play", Desc: "start or continue a puzzle"},
	{ItemTitle: "Stats", Desc: "your progress"},
	{ItemTitle: "Options", Desc: "configure and learn"},
	{ItemTitle: "Quit", Desc: "exit puzzletea"},
}

var playMenuItems = []ui.MenuItem{
	{ItemTitle: "Daily Puzzle", Desc: time.Now().Format("Jan _2 06")},
	{ItemTitle: "Generate", Desc: "a new puzzle"},
	{ItemTitle: "Continue", Desc: "a previously played puzzle"},
	{ItemTitle: "Enter Seed", Desc: "for a set puzzle"},
}

var optionsMenuItems = []ui.MenuItem{
	{ItemTitle: "Theme", Desc: "change colors"},
	{ItemTitle: "Guides", Desc: "learn the rules"},
}

const (
	mainMenuView viewState = iota
	playMenuView
	optionsMenuView
	seedInputView
	gameSelectView
	modeSelectView
	generatingView
	gameView
	continueView
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
	modeSelectList    list.Model
	selectedCategory  game.Category
	selectedModeTitle string
	continueTable     table.Model
	continueGames     []store.GameRecord
	seedInput         textinput.Model
	helpSelectList    list.Model
}

type sessionState struct {
	game            game.Gamer
	activeGameID    int64
	completionSaved bool
	generating      bool
	spawnJobID      int64
	spawnCancel     context.CancelFunc
	spawn           *spawnRequest
}

type helpState struct {
	category      game.Category
	viewport      viewport.Model
	renderer      *glamour.TermRenderer
	rendererWidth int
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
)

type spawnRequest struct {
	source      spawnSource
	name        string
	gameType    string
	modeTitle   string
	returnState viewState
}

type model struct {
	state viewState

	nav     navigationState
	session sessionState
	help    helpState
	stats   statsState
	theme   themeState
	debug   debugState

	spinner spinner.Model

	width  int // available content width (terminal - rootStyle frame)
	height int // available content height (terminal - rootStyle frame)

	store *store.Store

	// Config
	cfg *config.Config
}

func newSpinner() spinner.Model {
	p := theme.Current()
	sp := spinner.New()
	sp.Spinner = spinner.Dot
	sp.Style = lipgloss.NewStyle().Foreground(p.Accent)
	return sp
}

// InitialModel creates the root TUI model for the main menu.
func InitialModel(s *store.Store, cfg *config.Config) model {
	r := initDebugRenderer()
	l := ui.InitCategoryList(GameCategories, "Select Category")
	mm := ui.NewMainMenu(mainMenuItems)
	return model{
		state:   mainMenuView,
		debug:   debugState{renderer: r},
		nav:     navigationState{gameSelectList: l, mainMenu: mm},
		spinner: newSpinner(),
		store:   s,
		cfg:     cfg,
	}
}

// InitialModelWithGame creates a model that starts directly in gameView,
// bypassing the menu. Used by CLI flags (--new, --continue).
func InitialModelWithGame(s *store.Store, cfg *config.Config, g game.Gamer, activeGameID int64, completionSaved bool) model {
	r := initDebugRenderer()
	l := ui.InitCategoryList(GameCategories, "Select Category")
	mm := ui.NewMainMenu(mainMenuItems)
	return model{
		state:   gameView,
		debug:   debugState{renderer: r},
		nav:     navigationState{gameSelectList: l, mainMenu: mm},
		spinner: newSpinner(),
		store:   s,
		cfg:     cfg,
		session: sessionState{
			game:            g,
			activeGameID:    activeGameID,
			completionSaved: completionSaved,
		},
	}
}

func (m model) Init() tea.Cmd {
	return nil
}
