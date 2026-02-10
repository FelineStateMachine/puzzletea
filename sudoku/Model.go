// Package sudoku implements the classic number-placement puzzle.
package sudoku

import (
	"fmt"
	"strconv"

	"github.com/FelineStateMachine/puzzletea/game"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const gridSize = 9

var _ game.Gamer = Model{}

type Model struct {
	cursor       game.Cursor
	grid         grid
	provided     []cell
	providedGrid [gridSize][gridSize]bool
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
		keys:         DefaultKeyMap,
		modeTitle:    mode.Title(),
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
	if m.providedGrid[m.cursor.Y][m.cursor.X] {
		return
	}
	m.grid[m.cursor.Y][m.cursor.X].v = v
}

// View implements game.Gamer.
func (m Model) View() string {
	conflicts := computeConflicts(m.grid)
	solved := isSolvedWith(m.grid, conflicts)
	title := game.TitleBarView("Sudoku", m.modeTitle, solved)
	grid := renderGrid(m, solved, conflicts)
	status := statusBarView(m.showFullHelp)

	return lipgloss.JoinVertical(lipgloss.Left, title, grid, status)
}

// GetDebugInfo implements game.Gamer.
func (m Model) GetDebugInfo() string {
	conflicts := computeConflicts(m.grid)
	cursorCell := m.grid[m.cursor.Y][m.cursor.X]
	isProvided := m.providedGrid[m.cursor.Y][m.cursor.X]
	conflict := conflicts[m.cursor.Y][m.cursor.X]
	solved := m.isSolved()

	filledCount := 0
	conflictCount := 0
	for y := range gridSize {
		for x := range gridSize {
			if m.grid[y][x].v != 0 {
				filledCount++
			}
			if conflicts[y][x] {
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
