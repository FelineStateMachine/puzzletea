// Package app implements the root TUI model for PuzzleTea, including
// the Elm-architecture update/view loop, game spawning, and debug overlay.
package app

import (
	"time"

	"github.com/FelineStateMachine/puzzletea/game"
	"github.com/FelineStateMachine/puzzletea/game/hashiwokakero"
	"github.com/FelineStateMachine/puzzletea/game/hitori"
	"github.com/FelineStateMachine/puzzletea/game/lightsout"
	"github.com/FelineStateMachine/puzzletea/game/nonogram"
	"github.com/FelineStateMachine/puzzletea/game/shikaku"
	"github.com/FelineStateMachine/puzzletea/game/sudoku"
	"github.com/FelineStateMachine/puzzletea/game/takuzu"
	"github.com/FelineStateMachine/puzzletea/game/wordsearch"
	"github.com/FelineStateMachine/puzzletea/stats"
	"github.com/FelineStateMachine/puzzletea/store"
	"github.com/FelineStateMachine/puzzletea/ui"

	"charm.land/bubbles/v2/list"
	"charm.land/bubbles/v2/spinner"
	"charm.land/bubbles/v2/table"
	"charm.land/bubbles/v2/viewport"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/glamour"
)

// GameCategories is the master registry of all puzzle types.
var GameCategories = []list.Item{
	game.Category{Name: "Hashiwokakero", Desc: "Connect islands with bridges.", Modes: hashiwokakero.Modes, Help: hashiwokakero.HelpContent},
	game.Category{Name: "Hitori", Desc: "Shade cells to eliminate duplicates.", Modes: hitori.Modes, Help: hitori.HelpContent},
	game.Category{Name: "Lights Out", Desc: "Turn off all the lights.", Modes: lightsout.Modes, Help: lightsout.HelpContent},
	game.Category{Name: "Nonogram", Desc: "Fill cells to match row and column hints.", Modes: nonogram.Modes, Help: nonogram.HelpContent},
	game.Category{Name: "Shikaku", Desc: "Divide the grid into rectangles.", Modes: shikaku.Modes, Help: shikaku.HelpContent},
	game.Category{Name: "Sudoku", Desc: "Fill the 9x9 grid following sudoku rules.", Modes: sudoku.Modes, Help: sudoku.HelpContent},
	game.Category{Name: "Takuzu", Desc: "Fill the grid with ● and ○.", Modes: takuzu.Modes, Help: takuzu.HelpContent},
	game.Category{Name: "Word Search", Desc: "Find hidden words in a letter grid.", Modes: wordsearch.Modes, Help: wordsearch.HelpContent},
}

var mainMenuItems = []ui.MenuItem{
	{ItemTitle: "Daily Puzzle", Desc: time.Now().Format("Jan _2 06")},
	{ItemTitle: "Generate", Desc: "a new puzzle"},
	{ItemTitle: "Continue", Desc: "a saved puzzle"},
	{ItemTitle: "Stats", Desc: "your progress"},
	{ItemTitle: "Guides", Desc: "learn the rules"},
	{ItemTitle: "Quit", Desc: "exit puzzletea"},
}

const (
	mainMenuView = iota
	gameSelectView
	modeSelectView
	generatingView
	gameView
	continueView
	helpSelectView
	helpDetailView
	statsView
)

type model struct {
	state int

	mainMenu         ui.MainMenu
	gameSelectList   list.Model
	modeSelectList   list.Model
	selectedCategory game.Category

	continueTable table.Model
	continueGames []store.GameRecord

	mode game.Mode
	game game.Gamer

	spinner    spinner.Model
	generating bool // true while an async Spawn is in flight

	width  int // available content width (terminal - rootStyle frame)
	height int // available content height (terminal - rootStyle frame)

	debug         bool
	debugRenderer *glamour.TermRenderer
	debuginfo     string
	showFullHelp  bool

	store           *store.Store
	activeGameID    int64
	completionSaved bool

	// Daily puzzle state
	dailyPending   bool   // true while generating a daily puzzle
	dailyName      string // pre-computed daily name
	dailyGameType  string // e.g. "Nonogram"
	dailyModeTitle string // e.g. "Standard"

	// Help page state
	helpSelectList    list.Model
	helpCategory      game.Category
	helpViewport      viewport.Model
	helpRenderer      *glamour.TermRenderer
	helpRendererWidth int

	// Stats page state
	statsCards    []stats.Card
	statsProfile  stats.ProfileBanner
	statsViewport viewport.Model
}

// InitialModel creates the root TUI model for the main menu.
func InitialModel(s *store.Store) model {
	r := initDebugRenderer()
	l := ui.InitList(GameCategories, "Select Category")
	mm := ui.NewMainMenu(mainMenuItems)
	sp := spinner.New()
	sp.Spinner = spinner.Dot
	sp.Style = lipgloss.NewStyle().Foreground(ui.MenuAccent)
	return model{
		state:          mainMenuView,
		debugRenderer:  r,
		gameSelectList: l,
		mainMenu:       mm,
		spinner:        sp,
		store:          s,
	}
}

// InitialModelWithGame creates a model that starts directly in gameView,
// bypassing the menu. Used by CLI flags (--new, --continue).
func InitialModelWithGame(s *store.Store, g game.Gamer, activeGameID int64, completionSaved bool) model {
	r := initDebugRenderer()
	l := ui.InitList(GameCategories, "Select Category")
	mm := ui.NewMainMenu(mainMenuItems)
	sp := spinner.New()
	sp.Spinner = spinner.Dot
	sp.Style = lipgloss.NewStyle().Foreground(ui.MenuAccent)
	return model{
		state:           gameView,
		debugRenderer:   r,
		gameSelectList:  l,
		mainMenu:        mm,
		spinner:         sp,
		store:           s,
		game:            g,
		activeGameID:    activeGameID,
		completionSaved: completionSaved,
	}
}

func (m model) Init() tea.Cmd {
	return nil
}
