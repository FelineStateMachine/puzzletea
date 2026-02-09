package main

import (
	"encoding/json"
	"fmt"
	"log"
	"strconv"

	"github.com/FelineStateMachine/puzzletea/game"
	"github.com/FelineStateMachine/puzzletea/hashiwokakero"
	"github.com/FelineStateMachine/puzzletea/namegen"
	"github.com/FelineStateMachine/puzzletea/nonogram"
	"github.com/FelineStateMachine/puzzletea/store"
	"github.com/FelineStateMachine/puzzletea/sudoku"
	"github.com/FelineStateMachine/puzzletea/wordsearch"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/lipgloss"
)

var (
	rootStyle  = lipgloss.NewStyle().Margin(1, 2)
	debugStyle = lipgloss.NewStyle().Border(lipgloss.RoundedBorder(), true).BorderForeground(lipgloss.Color("124"))

	GameCategories = []list.Item{
		game.Category{Name: "Hashiwokakero", Desc: "Connect islands with bridges.", Modes: hashiwokakero.Modes},
		game.Category{Name: "Nonogram", Desc: "Fill cells to match row and column hints.", Modes: nonogram.Modes},
		game.Category{Name: "Sudoku", Desc: "Fill the 9x9 grid following sudoku rules.", Modes: sudoku.Modes},
		game.Category{Name: "Word Search", Desc: "Find hidden words in a letter grid.", Modes: wordsearch.Modes},
	}
)

const (
	mainMenuView = iota
	gameSelectView
	modeSelectView
	gameView
	continueView
)

// menuItem satisfies list.Item for the main menu.
type menuItem struct {
	title string
	desc  string
}

func (i menuItem) Title() string       { return i.title }
func (i menuItem) Description() string { return i.desc }
func (i menuItem) FilterValue() string { return i.title }

var mainMenuItems = []list.Item{
	menuItem{title: "Daily Puzzle", desc: "A new puzzle every day."},
	menuItem{title: "New Puzzle", desc: "Start a new puzzle game."},
	menuItem{title: "Continue", desc: "Resume a saved game."},
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

	debug         bool
	debugRenderer *glamour.TermRenderer
	debuginfo     string

	store           *store.Store
	activeGameID    int64
	completionSaved bool
}

func initialModel(s *store.Store) model {
	r := initDebugRenderer()
	l := initGameSelectList()
	ml := initMainMenuList()
	return model{
		state:          mainMenuView,
		debugRenderer:  r,
		gameSelectList: l,
		mainMenuList:   ml,
		store:          s,
	}
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		h, v := rootStyle.GetFrameSize()
		w, ht := msg.Width-h, msg.Height-v
		m.mainMenuList.SetSize(w, ht)
		m.gameSelectList.SetSize(w, ht)
		if m.state == modeSelectView {
			m.modeSelectList.SetSize(w, ht)
		}
		if m.state == continueView {
			m.continueTable.SetWidth(w)
			m.continueTable.SetHeight(ht)
		}

	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlN:
			m = saveCurrentGame(m, store.StatusInProgress)
			m.state = mainMenuView
			m.debug = false
		case tea.KeyEnter:
			if m.state == mainMenuView {
				item, ok := m.mainMenuList.SelectedItem().(menuItem)
				if !ok {
					return m, nil
				}
				switch item.title {
				case "New Puzzle":
					m.state = gameSelectView
				case "Continue":
					m.initContinueTable()
					m.state = continueView
				}
				return m, nil
			}
			if m.state == gameSelectView {
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
			if m.state == modeSelectView {
				m.mode, _ = m.modeSelectList.SelectedItem().(game.Mode)
				g, err := m.mode.(game.Spawner).Spawn()
				if err != nil {
					return m, nil
				}
				m.game = g
				m.state = gameView
				m.completionSaved = false

				// Capture initial state and create DB record.
				initialState, err := m.game.GetSave()
				if err != nil {
					log.Printf("failed to get initial save: %v", err)
					return m, nil
				}
				name := generateUniqueName(m.store)
				rec := &store.GameRecord{
					Name:         name,
					GameType:     m.selectedCategory.Name,
					Mode:         m.mode.Title(),
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
			if m.state == continueView {
				idx := m.continueTable.Cursor()
				if idx < 0 || idx >= len(m.continueGames) {
					return m, nil
				}
				rec := m.continueGames[idx]
				var g game.Gamer
				var err error
				switch rec.GameType {
				case "Nonogram":
					g, err = nonogram.ImportModel([]byte(rec.SaveState))
				case "Word Search":
					g, err = wordsearch.ImportModel([]byte(rec.SaveState))
				case "Sudoku":
					g, err = sudoku.ImportModel([]byte(rec.SaveState))
				case "Hashiwokakero":
					g, err = hashiwokakero.ImportModel([]byte(rec.SaveState))
				default:
					return m, nil
				}
				if err != nil {
					log.Printf("failed to import game: %v", err)
					return m, nil
				}
				m.game = g
				m.activeGameID = rec.ID
				m.state = gameView
				m.completionSaved = rec.Status == store.StatusCompleted
				return m, nil
			}
		case tea.KeyEscape:
			if m.state == modeSelectView {
				m.state = gameSelectView
				return m, nil
			}
			if m.state == gameSelectView || m.state == continueView {
				m.state = mainMenuView
				return m, nil
			}
		case tea.KeyCtrlC:
			m = saveCurrentGame(m, store.StatusAbandoned)
			return m, tea.Quit
		case tea.KeyCtrlE:
			m.debug = !m.debug
		}
	}

	switch m.state {
	case mainMenuView:
		m.mainMenuList, cmd = m.mainMenuList.Update(msg)
	case gameView:
		m.game, cmd = m.game.Update(msg)
		if m.debug {
			m.debuginfo = m.renderDebugInfo()
		}
		if !m.completionSaved && isSolvedFromSave(m.game) {
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

func (m model) View() string {
	var s string

	switch m.state {
	case mainMenuView:
		return rootStyle.Render(m.mainMenuList.View())
	case gameSelectView:
		return rootStyle.Render(m.gameSelectList.View())
	case modeSelectView:
		return rootStyle.Render(m.modeSelectList.View())
	case continueView:
		if len(m.continueGames) == 0 {
			return rootStyle.Render("No saved games yet.\n\nPress Escape to return.")
		}
		return rootStyle.Render(m.continueTable.View())
	case gameView:
		var debugInfo string
		if m.debug {
			debugInfo = debugStyle.Render(m.debuginfo)
		}
		s = lipgloss.JoinVertical(lipgloss.Left,
			m.game.View(),
			debugInfo,
		)
	default:
		s = fmt.Sprintf("unknown state: %d", m.state)
		panic(s)
	}

	return rootStyle.Render(s)
}

func initList(items []list.Item, title string) list.Model {
	l := list.New(items, list.NewDefaultDelegate(), 64, 64)
	l.Title = title
	l.DisableQuitKeybindings()
	l.SetFilteringEnabled(false)
	l.SetShowHelp(false)
	return l
}

func initMainMenuList() list.Model {
	return initList(mainMenuItems, "PuzzleTea")
}

func initGameSelectList() list.Model {
	return initList(GameCategories, "Select Game")
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
		{Title: "Name", Width: 20},
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
		table.WithHeight(10),
	)
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

// isSolvedFromSave checks whether the game's save data indicates a solved/won state.
func isSolvedFromSave(g game.Gamer) bool {
	data, err := g.GetSave()
	if err != nil {
		return false
	}
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(data, &fields); err != nil {
		return false
	}
	for _, key := range []string{"solved", "won"} {
		raw, ok := fields[key]
		if !ok {
			continue
		}
		var val bool
		if json.Unmarshal(raw, &val) == nil && val {
			return true
		}
	}
	return false
}
