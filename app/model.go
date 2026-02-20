// Package app implements the root TUI model for PuzzleTea, including
// the Elm-architecture update/view loop, game spawning, and debug overlay.
package app

import (
	"context"
	"time"

	"github.com/FelineStateMachine/puzzletea/config"
	"github.com/FelineStateMachine/puzzletea/game"
	"github.com/FelineStateMachine/puzzletea/hashiwokakero"
	"github.com/FelineStateMachine/puzzletea/hitori"
	"github.com/FelineStateMachine/puzzletea/lightsout"
	"github.com/FelineStateMachine/puzzletea/nonogram"
	"github.com/FelineStateMachine/puzzletea/nurikabe"
	"github.com/FelineStateMachine/puzzletea/shikaku"
	"github.com/FelineStateMachine/puzzletea/stats"
	"github.com/FelineStateMachine/puzzletea/store"
	"github.com/FelineStateMachine/puzzletea/sudoku"
	"github.com/FelineStateMachine/puzzletea/takuzu"
	"github.com/FelineStateMachine/puzzletea/theme"
	"github.com/FelineStateMachine/puzzletea/ui"
	"github.com/FelineStateMachine/puzzletea/wordsearch"

	"charm.land/bubbles/v2/list"
	"charm.land/bubbles/v2/spinner"
	"charm.land/bubbles/v2/table"
	"charm.land/bubbles/v2/textinput"
	"charm.land/bubbles/v2/viewport"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/glamour"
)

// Categories is the master typed registry of all puzzle types.
var Categories = []game.Category{
	{Name: "Hashiwokakero", Desc: "Connect islands with bridges.", Modes: hashiwokakero.Modes, Help: hashiwokakero.HelpContent},
	{Name: "Hitori", Desc: "Shade cells to eliminate duplicates.", Modes: hitori.Modes, Help: hitori.HelpContent},
	{Name: "Lights Out", Desc: "Turn off all the lights.", Modes: lightsout.Modes, Help: lightsout.HelpContent},
	{Name: "Nonogram", Desc: "Fill cells to match row and column hints.", Modes: nonogram.Modes, Help: nonogram.HelpContent},
	{Name: "Nurikabe", Desc: "Build islands while keeping one connected sea.", Modes: nurikabe.Modes, Help: nurikabe.HelpContent},
	{Name: "Shikaku", Desc: "Divide the grid into rectangles.", Modes: shikaku.Modes, Help: shikaku.HelpContent},
	{Name: "Sudoku", Desc: "Fill the 9x9 grid following sudoku rules.", Modes: sudoku.Modes, Help: sudoku.HelpContent},
	{Name: "Takuzu", Desc: "Fill the grid with ● and ○.", Modes: takuzu.Modes, Help: takuzu.HelpContent},
	{Name: "Word Search", Desc: "Find hidden words in a letter grid.", Modes: wordsearch.Modes, Help: wordsearch.HelpContent},
}

// GameCategories is the list.Item view of Categories for bubbles lists.
var GameCategories = categoriesAsItems(Categories)

func categoriesAsItems(categories []game.Category) []list.Item {
	items := make([]list.Item, len(categories))
	for i, cat := range categories {
		items[i] = cat
	}
	return items
}

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
	mainMenuView = iota
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

type model struct {
	state int

	mainMenu         ui.MainMenu
	playMenu         ui.MainMenu
	optionsMenu      ui.MainMenu
	gameSelectList   list.Model
	modeSelectList   list.Model
	selectedCategory game.Category

	continueTable table.Model
	continueGames []store.GameRecord

	selectedModeTitle string
	game              game.Gamer

	spinner     spinner.Model
	generating  bool // true while an async Spawn is in flight
	spawnJobID  int64
	spawnCancel context.CancelFunc

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

	// Seed input state
	seedInput     textinput.Model
	seedPending   bool   // true while generating a seeded puzzle
	seedName      string // pre-computed seed-derived name
	seedGameType  string // game type from rendezvous hashing
	seedModeTitle string // mode title from rendezvous hashing

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

	// Theme picker state
	themeList     list.Model
	previousTheme string // for revert on Esc

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
	l := ui.InitList(GameCategories, "Select Category")
	mm := ui.NewMainMenu(mainMenuItems)
	return model{
		state:          mainMenuView,
		debugRenderer:  r,
		gameSelectList: l,
		mainMenu:       mm,
		spinner:        newSpinner(),
		store:          s,
		cfg:            cfg,
	}
}

// InitialModelWithGame creates a model that starts directly in gameView,
// bypassing the menu. Used by CLI flags (--new, --continue).
func InitialModelWithGame(s *store.Store, cfg *config.Config, g game.Gamer, activeGameID int64, completionSaved bool) model {
	r := initDebugRenderer()
	l := ui.InitList(GameCategories, "Select Category")
	mm := ui.NewMainMenu(mainMenuItems)
	return model{
		state:           gameView,
		debugRenderer:   r,
		gameSelectList:  l,
		mainMenu:        mm,
		spinner:         newSpinner(),
		store:           s,
		cfg:             cfg,
		game:            g,
		activeGameID:    activeGameID,
		completionSaved: completionSaved,
	}
}

func (m model) Init() tea.Cmd {
	return nil
}
