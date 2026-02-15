package main

import (
	"fmt"
	"log"
	"math/rand/v2"
	"strconv"
	"time"

	"github.com/FelineStateMachine/puzzletea/game"
	"github.com/FelineStateMachine/puzzletea/hashiwokakero"
	"github.com/FelineStateMachine/puzzletea/hitori"
	"github.com/FelineStateMachine/puzzletea/lightsout"
	"github.com/FelineStateMachine/puzzletea/namegen"
	"github.com/FelineStateMachine/puzzletea/nonogram"
	"github.com/FelineStateMachine/puzzletea/store"
	"github.com/FelineStateMachine/puzzletea/sudoku"
	"github.com/FelineStateMachine/puzzletea/takuzu"
	"github.com/FelineStateMachine/puzzletea/wordsearch"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/lipgloss"
)

var (
	// Earth-tone palette for menus — ANSI 256 colors with light/dark adaptivity.
	menuAccent      = lipgloss.AdaptiveColor{Light: "130", Dark: "173"}
	menuAccentLight = lipgloss.AdaptiveColor{Light: "137", Dark: "180"}
	menuText        = lipgloss.AdaptiveColor{Light: "235", Dark: "252"}
	menuTextDim     = lipgloss.AdaptiveColor{Light: "243", Dark: "243"}
	menuDim         = lipgloss.AdaptiveColor{Light: "250", Dark: "238"}
	menuTableHeader = lipgloss.AdaptiveColor{Light: "130", Dark: "180"}

	rootStyle  = lipgloss.NewStyle().Margin(1, 2)
	debugStyle = lipgloss.NewStyle().Border(lipgloss.RoundedBorder(), true).BorderForeground(lipgloss.Color("124"))

	GameCategories = []list.Item{
		game.Category{Name: "Hashiwokakero", Desc: "Connect islands with bridges.", Modes: hashiwokakero.Modes},
		game.Category{Name: "Hitori", Desc: "Shade cells to eliminate duplicates.", Modes: hitori.Modes},
		game.Category{Name: "Lights Out", Desc: "Turn off all the lights.", Modes: lightsout.Modes},
		game.Category{Name: "Nonogram", Desc: "Fill cells to match row and column hints.", Modes: nonogram.Modes},
		game.Category{Name: "Sudoku", Desc: "Fill the 9x9 grid following sudoku rules.", Modes: sudoku.Modes},
		game.Category{Name: "Takuzu", Desc: "Fill the grid with ● and ○.", Modes: takuzu.Modes},
		game.Category{Name: "Word Search", Desc: "Find hidden words in a letter grid.", Modes: wordsearch.Modes},
	}
)

const (
	mainMenuView = iota
	gameSelectView
	modeSelectView
	generatingView
	gameView
	continueView
)

type menuItem struct {
	title string
	desc  string
}

func (i menuItem) Title() string       { return i.title }
func (i menuItem) Description() string { return i.desc }
func (i menuItem) FilterValue() string { return i.title }

var mainMenuItems = []list.Item{
	menuItem{title: "Daily Puzzle", desc: time.Now().Format("Jan _2 06")},
	menuItem{title: "Generate", desc: "a new puzzle."},
	menuItem{title: "Continue", desc: "a previous puzzle."},
	menuItem{title: "Quit", desc: "the game."},
}

type model struct {
	state int

	mainMenuList     list.Model
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
}

func initialModel(s *store.Store) model {
	r := initDebugRenderer()
	l := initGameSelectList()
	ml := initMainMenuList()
	sp := spinner.New()
	sp.Spinner = spinner.Dot
	sp.Style = lipgloss.NewStyle().Foreground(menuAccent)
	return model{
		state:          mainMenuView,
		debugRenderer:  r,
		gameSelectList: l,
		mainMenuList:   ml,
		spinner:        sp,
		store:          s,
	}
}

// initialModelWithGame creates a model that starts directly in gameView,
// bypassing the menu. Used by CLI flags (--new, --continue).
func initialModelWithGame(s *store.Store, g game.Gamer, activeGameID int64, completionSaved bool) model {
	r := initDebugRenderer()
	l := initGameSelectList()
	ml := initMainMenuList()
	sp := spinner.New()
	sp.Spinner = spinner.Dot
	sp.Style = lipgloss.NewStyle().Foreground(menuAccent)
	return model{
		state:           gameView,
		debugRenderer:   r,
		gameSelectList:  l,
		mainMenuList:    ml,
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

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case game.SpawnCompleteMsg:
		return m.handleSpawnComplete(msg)

	case tea.WindowSizeMsg:
		h, v := rootStyle.GetFrameSize()
		w, ht := msg.Width-h, msg.Height-v
		m.width = w
		m.height = ht
		menuW, menuH := min(w, 64), min(ht, 24)
		m.mainMenuList.SetSize(menuW, menuH)
		m.gameSelectList.SetSize(menuW, menuH)
		if m.state == modeSelectView {
			m.modeSelectList.SetSize(menuW, menuH)
		}
		if m.state == continueView {
			m.continueTable.SetWidth(w)
			m.continueTable.SetHeight(ht)
		}

	case tea.KeyMsg:
		// During generation, only allow Escape (to cancel) and Ctrl+C (to quit).
		if m.state == generatingView {
			switch msg.Type {
			case tea.KeyEscape:
				if m.dailyPending {
					m.dailyPending = false
					m.state = mainMenuView
				} else {
					m.state = modeSelectView
				}
				return m, nil
			case tea.KeyCtrlC:
				return m, tea.Quit
			}
			return m, nil
		}
		switch msg.Type {
		case tea.KeyCtrlN:
			m = saveCurrentGame(m, store.StatusInProgress)
			m.state = mainMenuView
			m.debug = false
		case tea.KeyEnter:
			if m.state != gameView {
				return m.handleEnter()
			}
		case tea.KeyEscape:
			return m.handleEscape()
		case tea.KeyCtrlC:
			m = saveCurrentGame(m, store.StatusAbandoned)
			return m, tea.Quit
		case tea.KeyCtrlE:
			m.debug = !m.debug
		case tea.KeyCtrlH:
			m.showFullHelp = !m.showFullHelp
			if m.state == gameView && m.game != nil {
				m.game, _ = m.game.Update(game.HelpToggleMsg{Show: m.showFullHelp})
			}
		case tea.KeyCtrlR:
			if m.state == gameView && m.game != nil {
				m.game = m.game.Reset()
			}
		}
	}

	switch m.state {
	case mainMenuView:
		m.mainMenuList, cmd = m.mainMenuList.Update(msg)
	case generatingView:
		m.spinner, cmd = m.spinner.Update(msg)
	case gameView:
		m.game, cmd = m.game.Update(msg)
		if m.debug {
			m.debuginfo = m.renderDebugInfo()
		}
		if !m.completionSaved && m.game.IsSolved() {
			m.completionSaved = true
			saveData, err := m.game.GetSave()
			if err == nil {
				_ = m.store.UpdateSaveState(m.activeGameID, string(saveData))
			}
			_ = m.store.UpdateStatus(m.activeGameID, store.StatusCompleted)
		}
	case gameSelectView:
		m.gameSelectList, cmd = m.gameSelectList.Update(msg)
	case modeSelectView:
		m.modeSelectList, cmd = m.modeSelectList.Update(msg)
	case continueView:
		m.continueTable, cmd = m.continueTable.Update(msg)
	}

	return m, cmd
}

func (m model) handleEnter() (tea.Model, tea.Cmd) {
	switch m.state {
	case mainMenuView:
		return m.handleMainMenuEnter()
	case gameSelectView:
		return m.handleGameSelectEnter()
	case modeSelectView:
		return m.handleModeSelectEnter()
	case continueView:
		return m.handleContinueEnter()
	}
	return m, nil
}

func (m model) handleMainMenuEnter() (tea.Model, tea.Cmd) {
	item, ok := m.mainMenuList.SelectedItem().(menuItem)
	if !ok {
		return m, nil
	}
	switch item.title {
	case "Daily Puzzle":
		return m.handleDailyPuzzle()
	case "Generate":
		m.state = gameSelectView
	case "Continue":
		m.initContinueTable()
		m.state = continueView
	case "Quit":
		return m, tea.Quit
	}
	return m, nil
}

func (m model) handleDailyPuzzle() (tea.Model, tea.Cmd) {
	today := time.Now()
	rng := dailyRNG(today)
	name := dailyName(today, rng)

	// Check if a daily game already exists for today.
	rec, err := m.store.GetDailyGame(name)
	if err != nil {
		log.Printf("failed to check daily game: %v", err)
		return m, nil
	}

	if rec != nil {
		// Daily exists - resume or review it.
		importFn, ok := game.Registry[rec.GameType]
		if !ok {
			log.Printf("unknown game type for daily: %s", rec.GameType)
			return m, nil
		}
		g, err := importFn([]byte(rec.SaveState))
		if err != nil {
			log.Printf("failed to import daily game: %v", err)
			return m, nil
		}
		m.game = g.SetTitle(rec.Name)
		m.game, _ = m.game.Update(game.HelpToggleMsg{Show: m.showFullHelp})
		m.activeGameID = rec.ID
		m.state = gameView
		m.completionSaved = rec.Status == store.StatusCompleted

		// Resume abandoned dailies by resetting their status.
		if rec.Status == store.StatusAbandoned {
			_ = m.store.UpdateStatus(rec.ID, store.StatusInProgress)
		}
		return m, nil
	}

	// No existing daily - generate a new one.
	spawner, gameType, modeTitle := dailyMode(rng)
	m.dailyPending = true
	m.dailyName = name
	m.dailyGameType = gameType
	m.dailyModeTitle = modeTitle
	m.state = generatingView
	m.generating = true
	return m, tea.Batch(m.spinner.Tick, spawnSeededCmd(spawner, rng))
}

func (m model) handleGameSelectEnter() (tea.Model, tea.Cmd) {
	cat, ok := m.gameSelectList.SelectedItem().(game.Category)
	if !ok {
		return m, nil
	}
	m.selectedCategory = cat
	m.modeSelectList = initModeSelectList(cat)
	m.modeSelectList.SetSize(m.gameSelectList.Width(), m.gameSelectList.Height())
	m.state = modeSelectView
	return m, nil
}

func (m model) handleModeSelectEnter() (tea.Model, tea.Cmd) {
	m.mode, _ = m.modeSelectList.SelectedItem().(game.Mode)
	spawner, ok := m.mode.(game.Spawner)
	if !ok {
		return m, nil
	}
	m.state = generatingView
	m.generating = true
	return m, tea.Batch(m.spinner.Tick, spawnCmd(spawner))
}

// spawnCmd returns a tea.Cmd that runs Spawn() off the main goroutine.
func spawnCmd(spawner game.Spawner) tea.Cmd {
	return func() tea.Msg {
		g, err := spawner.Spawn()
		return game.SpawnCompleteMsg{Game: g, Err: err}
	}
}

// spawnSeededCmd returns a tea.Cmd that runs SpawnSeeded() off the main goroutine.
// The caller must not use rng after this call — *rand.Rand is not goroutine-safe
// and ownership is transferred to the spawned goroutine.
func spawnSeededCmd(spawner game.SeededSpawner, rng *rand.Rand) tea.Cmd {
	return func() tea.Msg {
		g, err := spawner.SpawnSeeded(rng)
		return game.SpawnCompleteMsg{Game: g, Err: err}
	}
}

func (m model) handleSpawnComplete(msg game.SpawnCompleteMsg) (tea.Model, tea.Cmd) {
	m.generating = false
	isDaily := m.dailyPending
	m.dailyPending = false

	// If the user navigated away while generating, discard the result.
	if m.state != generatingView {
		return m, nil
	}

	if msg.Err != nil {
		log.Printf("failed to spawn game: %v", msg.Err)
		if isDaily {
			m.state = mainMenuView
		} else {
			m.state = modeSelectView
		}
		return m, nil
	}

	var name, gameType, modeTitle string
	if isDaily {
		name = m.dailyName
		gameType = m.dailyGameType
		modeTitle = m.dailyModeTitle
		m.game = msg.Game.SetTitle(name)
	} else {
		name = generateUniqueName(m.store)
		gameType = m.selectedCategory.Name
		modeTitle = m.mode.Title()
		m.game = msg.Game.SetTitle(name)
	}
	m.game, _ = m.game.Update(game.HelpToggleMsg{Show: m.showFullHelp})
	m.state = gameView
	m.completionSaved = false

	// Capture initial state and create DB record.
	initialState, err := m.game.GetSave()
	if err != nil {
		log.Printf("failed to get initial save: %v", err)
		return m, nil
	}
	rec := &store.GameRecord{
		Name:         name,
		GameType:     gameType,
		Mode:         modeTitle,
		InitialState: string(initialState),
		SaveState:    string(initialState),
		Status:       store.StatusNew,
	}
	if err := m.store.CreateGame(rec); err != nil {
		log.Printf("failed to create game record: %v", err)
	} else {
		m.activeGameID = rec.ID
	}
	return m, nil
}

func (m model) handleContinueEnter() (tea.Model, tea.Cmd) {
	idx := m.continueTable.Cursor()
	if idx < 0 || idx >= len(m.continueGames) {
		return m, nil
	}
	rec := m.continueGames[idx]
	importFn, ok := game.Registry[rec.GameType]
	if !ok {
		return m, nil
	}
	g, err := importFn([]byte(rec.SaveState))
	if err != nil {
		log.Printf("failed to import game: %v", err)
		return m, nil
	}
	m.game = g.SetTitle(rec.Name)
	m.game, _ = m.game.Update(game.HelpToggleMsg{Show: m.showFullHelp})
	m.activeGameID = rec.ID
	m.state = gameView
	m.completionSaved = rec.Status == store.StatusCompleted
	return m, nil
}

func (m model) handleEscape() (tea.Model, tea.Cmd) {
	switch m.state {
	case generatingView:
		m.state = modeSelectView
	case modeSelectView:
		m.state = gameSelectView
	case gameSelectView, continueView:
		m.state = mainMenuView
	}
	return m, nil
}

func (m model) View() string {
	switch m.state {
	case mainMenuView:
		return rootStyle.Render(lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, m.mainMenuList.View()))
	case gameSelectView:
		return rootStyle.Render(lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, m.gameSelectList.View()))
	case modeSelectView:
		return rootStyle.Render(lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, m.modeSelectList.View()))
	case generatingView:
		s := m.spinner.View() + " Generating puzzle..."
		return rootStyle.Render(lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, s))
	case continueView:
		var s string
		if len(m.continueGames) == 0 {
			s = "No saved games yet.\n\nPress Escape to return."
		} else {
			title := lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.AdaptiveColor{Light: "255", Dark: "255"}).
				Background(menuAccent).
				Padding(0, 1).
				Render("Saved Games")
			s = lipgloss.JoinVertical(lipgloss.Left, title, "", m.continueTable.View())
		}
		return rootStyle.Render(lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, s))
	case gameView:
		if m.game == nil {
			return ""
		}
		var debugInfo string
		if m.debug {
			debugInfo = debugStyle.Render(m.debuginfo)
		}
		s := lipgloss.JoinVertical(lipgloss.Center,
			m.game.View(),
			debugInfo,
		)
		return rootStyle.Render(lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, s))
	default:
		s := fmt.Sprintf("unknown state: %d", m.state)
		panic(s)
	}
}

func initList(items []list.Item, title string) list.Model {
	d := list.NewDefaultDelegate()
	d.Styles.SelectedTitle = d.Styles.SelectedTitle.
		Foreground(menuAccent).
		BorderLeftForeground(menuAccent)
	d.Styles.SelectedDesc = d.Styles.SelectedDesc.
		Foreground(menuAccentLight).
		BorderLeftForeground(menuAccent)
	d.Styles.NormalTitle = d.Styles.NormalTitle.
		Foreground(menuText)
	d.Styles.NormalDesc = d.Styles.NormalDesc.
		Foreground(menuTextDim)

	l := list.New(items, d, 64, 64)
	l.Title = title
	l.Styles.Title = lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.AdaptiveColor{Light: "255", Dark: "255"}).
		Background(menuAccent).
		Padding(0, 1)
	l.DisableQuitKeybindings()
	l.SetFilteringEnabled(false)
	l.SetShowHelp(false)
	return l
}

func initMainMenuList() list.Model {
	return initList(mainMenuItems, "puzzletea")
}

func initGameSelectList() list.Model {
	return initList(GameCategories, "Select Category")
}

func initModeSelectList(cat game.Category) list.Model {
	return initList(cat.Modes, cat.Name+" - Select Mode")
}

func (m *model) initContinueTable() {
	games, err := m.store.ListGames()
	if err != nil {
		log.Printf("failed to list games: %v", err)
		games = nil
	}
	m.continueGames = games

	columns := []table.Column{
		{Title: "Name", Width: 34},
		{Title: "Game", Width: 15},
		{Title: "Mode", Width: 15},
		{Title: "Status", Width: 12},
		{Title: "Last Updated", Width: 20},
	}

	rows := make([]table.Row, len(games))
	for i, g := range games {
		rows[i] = table.Row{
			g.Name,
			g.GameType,
			g.Mode,
			formatStatus(g.Status),
			g.UpdatedAt.Local().Format("Jan 02 15:04"),
		}
	}

	t := table.New(
		table.WithColumns(columns),
		table.WithRows(rows),
		table.WithFocused(true),
		table.WithHeight(max(m.height-2, 1)),
	)

	s := table.DefaultStyles()
	s.Header = s.Header.
		Bold(true).
		Foreground(menuTableHeader).
		BorderStyle(lipgloss.NormalBorder()).
		BorderBottom(true).
		BorderForeground(menuDim)
	s.Selected = s.Selected.
		Foreground(lipgloss.AdaptiveColor{Light: "255", Dark: "255"}).
		Background(menuAccent).
		Bold(true)
	s.Cell = s.Cell.
		Foreground(menuText)
	t.SetStyles(s)

	m.continueTable = t
}

func formatStatus(s store.GameStatus) string {
	switch s {
	case store.StatusNew:
		return "New"
	case store.StatusInProgress:
		return "In Progress"
	case store.StatusCompleted:
		return "Completed"
	default:
		return string(s)
	}
}

func initDebugRenderer() *glamour.TermRenderer {
	renderer, err := glamour.NewTermRenderer(
		glamour.WithAutoStyle(),
		glamour.WithPreservedNewLines(),
		glamour.WithChromaFormatter("terminal16m"),
	)
	if err != nil {
		panic(err)
	}
	return renderer
}

func (m model) renderDebugInfo() string {
	s, err := m.debugRenderer.Render(m.game.GetDebugInfo())
	if err != nil {
		panic(err)
	}

	return s
}

// saveCurrentGame saves the current game state to the DB if a game is active.
func saveCurrentGame(m model, status store.GameStatus) model {
	if m.game == nil || m.activeGameID == 0 {
		return m
	}
	saveData, err := m.game.GetSave()
	if err != nil {
		log.Printf("failed to get save data: %v", err)
		return m
	}
	_ = m.store.UpdateSaveState(m.activeGameID, string(saveData))
	// Don't overwrite a completed status when navigating away.
	if !(m.completionSaved && status != store.StatusCompleted) {
		_ = m.store.UpdateStatus(m.activeGameID, status)
	}
	m.activeGameID = 0
	m.game = nil
	return m
}

// generateUniqueName generates a unique adjective-noun name, retrying if needed.
func generateUniqueName(s *store.Store) string {
	for range 100 {
		name := namegen.Generate()
		exists, err := s.NameExists(name)
		if err != nil || !exists {
			return name
		}
	}
	// Fallback: append a number
	for i := 1; ; i++ {
		name := namegen.Generate() + "-" + strconv.Itoa(i)
		exists, err := s.NameExists(name)
		if err != nil || !exists {
			return name
		}
	}
}
