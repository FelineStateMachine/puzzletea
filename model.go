package main

import (
	"fmt"

	"github.com/FelineStateMachine/puzzletea/game"
	"github.com/FelineStateMachine/puzzletea/hashiwokakero"
	"github.com/FelineStateMachine/puzzletea/nonogram"
	"github.com/FelineStateMachine/puzzletea/sudoku"
	"github.com/FelineStateMachine/puzzletea/wordsearch"
	"github.com/charmbracelet/bubbles/list"
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
	gameSelectView = iota
	modeSelectView
	gameView
)

type model struct {
	state int

	gameSelectList   list.Model
	modeSelectList   list.Model
	selectedCategory game.Category

	mode game.Mode
	game game.Gamer

	debug         bool
	debugRenderer *glamour.TermRenderer
	debuginfo     string
}

func initialModel() model {
	r := initDebugRenderer()
	l := initGameSelectList()
	return model{
		debugRenderer:  r,
		gameSelectList: l,
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
		m.gameSelectList.SetSize(w, ht)
		if m.state == modeSelectView {
			m.modeSelectList.SetSize(w, ht)
		}

	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlN:
			m.state = gameSelectView
			m.debug = false
		case tea.KeyEnter:
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
				m.game, _ = m.mode.(game.Spawner).Spawn()
				m.state = gameView
				return m, nil
			}
		case tea.KeyEscape:
			if m.state == modeSelectView {
				m.state = gameSelectView
				return m, nil
			}
		case tea.KeyCtrlC:
			return m, tea.Quit
		case tea.KeyCtrlE:
			m.debug = !m.debug
		}
	}

	switch m.state {
	case gameView:
		m.game, cmd = m.game.Update(msg)
		if m.debug {
			m.debuginfo = m.renderDebugInfo()
		}
	case gameSelectView:
		m.gameSelectList, cmd = m.gameSelectList.Update(msg)
	case modeSelectView:
		m.modeSelectList, cmd = m.modeSelectList.Update(msg)
	}

	return m, cmd
}

func (m model) View() string {
	var s string

	switch m.state {
	case gameSelectView:
		s = rootStyle.Render(m.gameSelectList.View())
		return s
	case modeSelectView:
		s = rootStyle.Render(m.modeSelectList.View())
		return s
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

func initGameSelectList() list.Model {
	return initList(GameCategories, "Select Game")
}

func initModeSelectList(cat game.Category) list.Model {
	return initList(cat.Modes, cat.Name+" - Select Mode")
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
