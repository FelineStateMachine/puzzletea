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

const gridSize = 9

type Model struct {
	cursor       game.Cursor
	grid         grid
	provided     []cell
	keys         KeyMap
	modeTitle    string
	showFullHelp bool
}

// New creates a new sudoku game using the provided cell values.
func New(mode SudokuMode, provided []cell, save ...string) (game.Gamer, error) {
	g := loadSave(newGrid(provided), save...)
	m := Model{
		grid:      g,
		provided:  provided,
		keys:      DefaultKeyMap,
		modeTitle: mode.Title(),
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
	case game.HelpToggleMsg:
		m.showFullHelp = msg.Show
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keys.FillValue):
			val, _ := strconv.Atoi(msg.String())
			m.updateCell(val)
		case key.Matches(msg, m.keys.ClearCell):
			m.updateCell(0)
		default:
			m.cursor.Move(m.keys.CursorKeyMap, msg, gridSize-1, gridSize-1)
		}
	}
	m.updateKeyBindings()
	return m, nil
}

func (m Model) SetTitle(t string) game.Gamer {
	m.modeTitle = t
	return m
}

// IsSolved implements game.Gamer.
func (m Model) IsSolved() bool {
	return m.isSolved()
}

func (m *Model) updateCell(v int) {
	c := &m.grid[m.cursor.Y][m.cursor.X]
	if slices.Contains(m.provided, *c) {
		return
	}
	c.v = v
}

// View implements game.Gamer.
func (m Model) View() string {
	title := titleBarView(m.modeTitle, m.isSolved())
	grid := renderGrid(m)
	status := statusBarView(m.showFullHelp)

	return lipgloss.JoinVertical(lipgloss.Left, title, grid, status)
}

// GetDebugInfo implements game.Gamer.
func (m Model) GetDebugInfo() string {
	cursorCell := m.grid[m.cursor.Y][m.cursor.X]
	isProvided := slices.Contains(m.provided, cursorCell)
	conflict := hasConflict(m, cursorCell, m.cursor.X, m.cursor.Y)
	solved := m.isSolved()

	filledCount := 0
	conflictCount := 0
	for y := range gridSize {
		for x := range gridSize {
			c := m.grid[y][x]
			if c.v != 0 {
				filledCount++
			}
			if c.v != 0 && hasConflict(m, c, x, y) {
				conflictCount++
			}
		}
	}

	status := "In Progress"
	if solved {
		status = "Solved"
	}

	s := fmt.Sprintf(
		"# Sudoku\n\n"+
			"## Game State\n\n"+
			"| Property | Value |\n"+
			"| :--- | :--- |\n"+
			"| Status | %s |\n"+
			"| Cursor | (%d, %d) |\n"+
			"| Cell Value | %s |\n"+
			"| Is Provided | %v |\n"+
			"| Has Conflict | %v |\n"+
			"| Cells Filled | %d / 81 |\n"+
			"| Conflict Count | %d |\n"+
			"| Provided Count | %d |\n",
		status,
		m.cursor.X, m.cursor.Y,
		cellContent(cursorCell),
		isProvided,
		conflict,
		filledCount,
		conflictCount,
		len(m.provided),
	)

	if len(m.provided) > 0 {
		s += "\n## Provided Cells\n\n"
		s += "| Row | Col | Value |\n"
		s += "| :--- | :--- | :--- |\n"
		for _, p := range m.provided {
			s += fmt.Sprintf("| %d | %d | %d |\n", p.y, p.x, p.v)
		}
	}

	return s
}
