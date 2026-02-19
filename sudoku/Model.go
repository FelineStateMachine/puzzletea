// Package sudoku implements the classic number-placement puzzle.
package sudoku

import (
	"fmt"
	"strconv"

	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/FelineStateMachine/puzzletea/game"
)

const gridSize = 9

var _ game.Gamer = Model{}

type Model struct {
	cursor       game.Cursor
	grid         grid
	provided     []cell
	providedGrid [gridSize][gridSize]bool
	conflicts    [gridSize][gridSize]bool
	keys         KeyMap
	modeTitle    string
	showFullHelp bool
}

func buildProvidedGrid(provided []cell) [gridSize][gridSize]bool {
	var pg [gridSize][gridSize]bool
	for _, c := range provided {
		pg[c.y][c.x] = true
	}
	return pg
}

func New(mode SudokuMode, provided []cell) (game.Gamer, error) {
	g := newGrid(provided)
	m := Model{
		grid:         g,
		provided:     provided,
		providedGrid: buildProvidedGrid(provided),
		conflicts:    computeConflicts(g),
		keys:         DefaultKeyMap,
		modeTitle:    mode.Title(),
	}
	return m, nil
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (game.Gamer, tea.Cmd) {
	switch msg := msg.(type) {
	case game.HelpToggleMsg:
		m.showFullHelp = msg.Show
	case tea.KeyPressMsg:
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

func (m Model) IsSolved() bool {
	return m.isSolved()
}

func (m Model) Reset() game.Gamer {
	m.grid = newGrid(m.provided)
	m.conflicts = computeConflicts(m.grid)
	return m
}

func (m *Model) updateCell(v int) {
	if m.providedGrid[m.cursor.Y][m.cursor.X] {
		return
	}
	m.grid[m.cursor.Y][m.cursor.X].v = v
	m.conflicts = computeConflicts(m.grid)
}

func (m Model) View() string {
	solved := isSolvedWith(m.grid, m.conflicts)
	title := game.TitleBarView("Sudoku", m.modeTitle, solved)
	grid := renderGrid(m, solved, m.conflicts)
	status := statusBarView(m.showFullHelp)

	return lipgloss.JoinVertical(lipgloss.Center, title, grid, status)
}

func (m Model) GetDebugInfo() string {
	cursorCell := m.grid[m.cursor.Y][m.cursor.X]
	isProvided := m.providedGrid[m.cursor.Y][m.cursor.X]
	conflict := m.conflicts[m.cursor.Y][m.cursor.X]
	solved := m.isSolved()

	filledCount := 0
	conflictCount := 0
	for y := range gridSize {
		for x := range gridSize {
			if m.grid[y][x].v != 0 {
				filledCount++
			}
			if m.conflicts[y][x] {
				conflictCount++
			}
		}
	}

	status := "In Progress"
	if solved {
		status = "Solved"
	}

	s := game.DebugHeader("Sudoku", [][2]string{
		{"Status", status},
		{"Cursor", fmt.Sprintf("(%d, %d)", m.cursor.X, m.cursor.Y)},
		{"Cell Value", cellContent(cursorCell)},
		{"Is Provided", fmt.Sprintf("%v", isProvided)},
		{"Has Conflict", fmt.Sprintf("%v", conflict)},
		{"Cells Filled", fmt.Sprintf("%d / 81", filledCount)},
		{"Conflict Count", fmt.Sprintf("%d", conflictCount)},
		{"Provided Count", fmt.Sprintf("%d", len(m.provided))},
	})

	if len(m.provided) > 0 {
		var rows [][]string
		for _, p := range m.provided {
			rows = append(rows, []string{
				fmt.Sprintf("%d", p.y),
				fmt.Sprintf("%d", p.x),
				fmt.Sprintf("%d", p.v),
			})
		}
		s += game.DebugTable("Provided Cells", []string{"Row", "Col", "Value"}, rows)
	}

	return s
}
