// Package takuzu implements the binary (Binairo) puzzle game.
package takuzu

import (
	"fmt"

	"github.com/FelineStateMachine/puzzletea/game"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Model implements game.Gamer for Takuzu.
type Model struct {
	size         int
	grid         grid
	provided     [][]bool
	cursor       game.Cursor
	solved       bool
	keys         KeyMap
	modeTitle    string
	showFullHelp bool
}

var _ game.Gamer = Model{}

// New creates a new Takuzu game model.
func New(mode TakuzuMode, puzzle grid, provided [][]bool) (game.Gamer, error) {
	m := Model{
		size:      mode.Size,
		grid:      puzzle,
		provided:  provided,
		cursor:    game.Cursor{X: 0, Y: 0},
		keys:      DefaultKeyMap,
		modeTitle: mode.Title(),
	}
	m.solved = m.checkSolved()
	return m, nil
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (game.Gamer, tea.Cmd) {
	switch msg := msg.(type) {
	case game.HelpToggleMsg:
		m.showFullHelp = msg.Show
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keys.PlaceZero):
			if !m.provided[m.cursor.Y][m.cursor.X] && !m.solved {
				m.grid[m.cursor.Y][m.cursor.X] = zeroCell
				m.solved = m.checkSolved()
			}
		case key.Matches(msg, m.keys.PlaceOne):
			if !m.provided[m.cursor.Y][m.cursor.X] && !m.solved {
				m.grid[m.cursor.Y][m.cursor.X] = oneCell
				m.solved = m.checkSolved()
			}
		case key.Matches(msg, m.keys.Clear):
			if !m.provided[m.cursor.Y][m.cursor.X] && !m.solved {
				m.grid[m.cursor.Y][m.cursor.X] = emptyCell
				m.solved = m.checkSolved()
			}
		default:
			m.cursor.Move(m.keys.CursorKeyMap, msg, m.size-1, m.size-1)
		}
	}
	m.updateKeyBindings()
	return m, nil
}

func (m Model) View() string {
	title := game.TitleBarView("Takuzu", m.modeTitle, m.solved)
	grid := gridView(m.grid, m.provided, m.cursor, m.solved)
	status := statusBarView(m.showFullHelp)
	return lipgloss.JoinVertical(lipgloss.Center, title, grid, status)
}

func (m Model) SetTitle(t string) game.Gamer {
	m.modeTitle = t
	return m
}

func (m Model) IsSolved() bool {
	return m.solved
}

func (m Model) checkSolved() bool {
	// All cells must be filled.
	for y := range m.size {
		for x := range m.size {
			if m.grid[y][x] == emptyCell {
				return false
			}
		}
	}
	return checkConstraints(m.grid, m.size) && hasUniqueLines(m.grid, m.size)
}

func (m Model) GetDebugInfo() string {
	status := "In Progress"
	if m.solved {
		status = "Solved"
	}

	filled := 0
	for y := range m.size {
		for x := range m.size {
			if m.grid[y][x] != emptyCell {
				filled++
			}
		}
	}

	return game.DebugHeader("Takuzu", [][2]string{
		{"Status", status},
		{"Cursor", fmt.Sprintf("(%d, %d)", m.cursor.X, m.cursor.Y)},
		{"Grid Size", fmt.Sprintf("%d×%d", m.size, m.size)},
		{"Cells Filled", fmt.Sprintf("%d / %d", filled, m.size*m.size)},
	})
}
