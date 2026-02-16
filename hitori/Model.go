// Package hitori implements the Hitori number puzzle game.
package hitori

import (
	"fmt"

	"github.com/FelineStateMachine/puzzletea/game"

	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

// Model implements game.Gamer for Hitori.
type Model struct {
	size         int
	numbers      grid
	marks        [][]cellMark
	initialMarks [][]cellMark
	conflicts    [][]bool
	cursor       game.Cursor
	solved       bool
	keys         KeyMap
	modeTitle    string
	showFullHelp bool
}

var _ game.Gamer = Model{}

// New creates a new Hitori game model.
func New(mode HitoriMode, numbers grid) (game.Gamer, error) {
	if mode.Size <= 0 {
		return Model{}, fmt.Errorf("invalid grid size: %d", mode.Size)
	}
	if len(numbers) != mode.Size {
		return Model{}, fmt.Errorf("numbers grid has %d rows, expected %d", len(numbers), mode.Size)
	}
	for y, row := range numbers {
		if len(row) != mode.Size {
			return Model{}, fmt.Errorf("numbers row %d has %d columns, expected %d", y, len(row), mode.Size)
		}
	}

	marks := newMarks(mode.Size)
	m := Model{
		size:         mode.Size,
		numbers:      numbers,
		marks:        marks,
		initialMarks: cloneMarks(marks),
		cursor:       game.Cursor{X: 0, Y: 0},
		keys:         DefaultKeyMap,
		modeTitle:    mode.Title(),
	}
	m.solved = m.checkSolved()
	m.conflicts = computeConflicts(m.numbers, m.marks, m.size)
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
		case key.Matches(msg, m.keys.ShadeCell):
			if !m.solved {
				m.toggleMark(shaded)
			}
		case key.Matches(msg, m.keys.CircleCell):
			if !m.solved {
				m.toggleMark(circled)
			}
		case key.Matches(msg, m.keys.ClearCell):
			if !m.solved {
				m.marks[m.cursor.Y][m.cursor.X] = unmarked
				m.solved = m.checkSolved()
				m.conflicts = computeConflicts(m.numbers, m.marks, m.size)
			}
		default:
			m.cursor.Move(m.keys.CursorKeyMap, msg, m.size-1, m.size-1)
		}
	}
	m.updateKeyBindings()
	return m, nil
}

// toggleMark sets the cell to the given mark if it's currently different,
// otherwise clears it to unmarked.
func (m *Model) toggleMark(mark cellMark) {
	if m.marks[m.cursor.Y][m.cursor.X] == mark {
		m.marks[m.cursor.Y][m.cursor.X] = unmarked
	} else {
		m.marks[m.cursor.Y][m.cursor.X] = mark
	}
	m.solved = m.checkSolved()
	m.conflicts = computeConflicts(m.numbers, m.marks, m.size)
}

func (m Model) View() string {
	title := game.TitleBarView("Hitori", m.modeTitle, m.solved)
	grid := gridView(m.numbers, m.marks, m.cursor, m.solved, m.conflicts)
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

func (m Model) Reset() game.Gamer {
	m.marks = cloneMarks(m.initialMarks)
	m.solved = m.checkSolved()
	m.conflicts = computeConflicts(m.numbers, m.marks, m.size)
	m.cursor = game.Cursor{}
	return m
}

func (m Model) checkSolved() bool {
	// At least one cell must be shaded for a valid solution.
	hasShaded := false
	for y := range m.size {
		for x := range m.size {
			if m.marks[y][x] == shaded {
				hasShaded = true
				break
			}
		}
		if hasShaded {
			break
		}
	}
	if !hasShaded {
		return false
	}

	return isValidSolution(m.numbers, m.marks, m.size)
}

func (m Model) GetDebugInfo() string {
	status := "In Progress"
	if m.solved {
		status = "Solved"
	}

	shadedCount := 0
	circledCount := 0
	for y := range m.size {
		for x := range m.size {
			switch m.marks[y][x] {
			case shaded:
				shadedCount++
			case circled:
				circledCount++
			}
		}
	}

	return game.DebugHeader("Hitori", [][2]string{
		{"Status", status},
		{"Cursor", fmt.Sprintf("(%d, %d)", m.cursor.X, m.cursor.Y)},
		{"Grid Size", fmt.Sprintf("%d\u00d7%d", m.size, m.size)},
		{"Shaded Cells", fmt.Sprintf("%d", shadedCount)},
		{"Circled Cells", fmt.Sprintf("%d", circledCount)},
	})
}
