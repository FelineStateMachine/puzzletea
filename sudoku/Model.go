package sudoku

import (
	"fmt"
	"slices"
	"strconv"

	"github.com/FelineStateMachine/puzzletea/game"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const GRIDSIZE = 9

const (
	cellBGColor     = lipgloss.Color("249")
	providedFGColor = lipgloss.Color("229")
	cursorBGColor   = lipgloss.Color("10")
)

type cursor struct {
	x, y int
}

type Model struct {
	cursor   cursor
	grid     grid
	provided []cell
	keys     KeyMap
}

// New creates a new sudoku game using the provided cell values.
func New(mode SudokuMode, provided []cell, save ...string) (game.Gamer, error) {
	g := loadSave(newGrid(provided), save...)
	m := Model{
		grid:     g,
		provided: provided,
		keys:     DefaultKeyMap,
	}
	return m, nil
}

// Init implements game.Gamer.
func (m Model) Init() tea.Cmd {
	return nil
}

// Update implements game.Gamer.
func (m Model) Update(msg tea.Msg) (game.Gamer, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keys.FillValue):
			val, _ := strconv.Atoi(msg.String())
			m.updateCell(val)
		case key.Matches(msg, m.keys.Up):
			if m.cursor.y > 0 {
				m.cursor.y--
			}
		case key.Matches(msg, m.keys.Down):
			if m.cursor.y < GRIDSIZE-1 {
				m.cursor.y++
			}
		case key.Matches(msg, m.keys.Left):
			if m.cursor.x > 0 {
				m.cursor.x--
			}
		case key.Matches(msg, m.keys.Right):
			if m.cursor.x < GRIDSIZE-1 {
				m.cursor.x++
			}
		}

	}
	m.updateKeyBindinds()
	return m, nil
}

func (m *Model) updateCell(v int) {
	c := &m.grid[m.cursor.y][m.cursor.x]
	if slices.Contains(m.provided, *c) {

	}
	c.v = v
}

// View implements game.Gamer.
func (m Model) View() string {
	var rows []string
	for _, row := range m.grid {
		var cells []string
		for _, cell := range row {

			cellStyle := lipgloss.NewStyle().Background(cellBGColor)
			if slices.Contains(m.provided, cell) {
				cellStyle = cellStyle.Foreground(providedFGColor)
			}
			if m.cursor.x == cell.x && m.cursor.y == cell.y {
				cellStyle = cellStyle.Background(cursorBGColor)
			}

			c := cellStyle.Render(fmt.Sprintf("%d", cell.v))
			cells = append(cells, c)
		}
		r := lipgloss.JoinHorizontal(lipgloss.Right, cells...)
		rows = append(rows, r)
	}
	return lipgloss.JoinVertical(lipgloss.Left, rows...)
}

// GetDebugInfo implements game.Gamer.
func (m Model) GetDebugInfo() string {
	return fmt.Sprintf(`
# Sudoko
Cursor (%dx%d)
Provided Cells: '%v'
`,
		m.cursor.x, m.cursor.y,
		m.provided,
	)
}
