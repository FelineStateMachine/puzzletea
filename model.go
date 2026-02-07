package main

import (
	"fmt"

	"github.com/FelineStateMachine/puzzletea/game"
	"github.com/FelineStateMachine/puzzletea/nonogram"
	"github.com/FelineStateMachine/puzzletea/wordsearch"
	// "github.com/FelineStateMachine/puzzletea/sudoku"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/lipgloss"
)

var (
	rootStyle  = lipgloss.NewStyle().Margin(1, 2)
	debugStyle = lipgloss.NewStyle().Border(lipgloss.RoundedBorder(), true).BorderForeground(lipgloss.Color("124"))

	GameModes = []list.Item{
		nonogram.NewMode("Easy - 5x5", "A random nonogram on a five by five board.", 5, 5),
		nonogram.NewMode("Medium - 10x10", "A random nonogram on a ten by ten board.", 10, 10),
		nonogram.NewMode("Hard - 15x15", "A random nonogram on a fifteen by fifteen board.", 15, 15),
		nonogram.NewMode("Extra - 5x10", "A random nonogram on a five by ten board.", 5, 10),

		wordsearch.NewMode("Word Search Easy - 10x10", "Find 6 words in a 10x10 grid.", 10, 10, 6, 3, 5, []wordsearch.Direction{wordsearch.Right, wordsearch.Down, wordsearch.DownRight}),
		wordsearch.NewMode("Word Search Medium - 15x15", "Find 10 words in a 15x15 grid.", 15, 15, 10, 4, 7, []wordsearch.Direction{wordsearch.Right, wordsearch.Down, wordsearch.DownRight, wordsearch.DownLeft, wordsearch.Left, wordsearch.Up}),
		wordsearch.NewMode("Word Search Hard - 20x20", "Find 15 words in a 20x20 grid.", 20, 20, 15, 5, 10, []wordsearch.Direction{wordsearch.Right, wordsearch.Down, wordsearch.DownRight, wordsearch.DownLeft, wordsearch.Left, wordsearch.Up, wordsearch.UpRight, wordsearch.UpLeft}),

		// sudoku.NewMode("Easy - 38 Provided Cells", "A random sudoku with at least 38 cells provided to start.", 38),
		// sudoku.NewMode("Hard - 26 Provided Cells", "A random sudoku with at least 26 cells provided to start.", 26)
	}
)

const (
	menuView = iota
	gameView
)

type model struct {
	state int

	gamemodesList list.Model

	mode game.Mode
	game game.Gamer

	debug         bool
	debugRenderer *glamour.TermRenderer
	debuginfo     string
}

func initialModel() model {
	r := initDebugRenderer()
	l := initGameModeList()
	return model{
		debugRenderer: r,
		gamemodesList: l,
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
		m.gamemodesList.SetSize(msg.Width-h, msg.Height-v)

	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlN:
			m.state = menuView
			m.debug = false
		case tea.KeyEnter:
			if m.state == menuView {
				m.mode, _ = m.gamemodesList.SelectedItem().(game.Mode)
				m.game, _ = m.SpawnGame(m.mode)
				m.state = gameView
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
	case menuView:
		m.gamemodesList, cmd = m.gamemodesList.Update(msg)
	}

	return m, cmd
}

func (m model) View() string {
	var s string

	switch m.state {
	case menuView:
		s = rootStyle.Render(m.gamemodesList.View())
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

func initGameModeList() list.Model {
	gamemodes := list.New(GameModes, list.NewDefaultDelegate(), 64, 64)
	gamemodes.Title = "Select Gamemode"
	gamemodes.DisableQuitKeybindings()
	gamemodes.SetFilteringEnabled(false)
	gamemodes.SetShowHelp(false)
	return gamemodes
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
