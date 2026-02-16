package shikaku

import (
	"fmt"

	"github.com/FelineStateMachine/puzzletea/game"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// expansion tracks how far the preview rectangle extends from the clue cell.
type expansion struct {
	clueID                int
	left, right, up, down int
}

// rect converts the expansion to a Rectangle.
func (e expansion) rect(clue *Clue) Rectangle {
	return Rectangle{
		ClueID: e.clueID,
		X:      clue.X - e.left,
		Y:      clue.Y - e.up,
		W:      e.left + 1 + e.right,
		H:      e.up + 1 + e.down,
	}
}

var _ game.Gamer = Model{}

// Model implements game.Gamer for Shikaku.
type Model struct {
	puzzle       Puzzle
	cursor       game.Cursor
	selectedClue *int // nil = navigation mode, non-nil = expansion mode
	expansion    expansion
	keys         KeyMap
	modeTitle    string
	showFullHelp bool
}

// New creates a new Shikaku game model.
func New(mode ShikakuMode, puzzle Puzzle) game.Gamer {
	return Model{
		puzzle:    puzzle,
		cursor:    game.Cursor{X: 0, Y: 0},
		keys:      DefaultKeyMap,
		modeTitle: mode.Title(),
	}
}

func (m Model) Init() tea.Cmd { return nil }

func (m Model) Update(msg tea.Msg) (game.Gamer, tea.Cmd) {
	switch msg := msg.(type) {
	case game.HelpToggleMsg:
		m.showFullHelp = msg.Show
	case tea.KeyMsg:
		if m.puzzle.IsSolved() {
			m.cursor.Move(m.keys.CursorKeyMap, msg, m.puzzle.Width-1, m.puzzle.Height-1)
		} else if m.selectedClue != nil {
			m = m.handleExpansionMode(msg)
		} else {
			m = m.handleNavMode(msg)
		}
	}
	return m, nil
}

func (m Model) handleNavMode(msg tea.KeyMsg) Model {
	switch {
	case key.Matches(msg, m.keys.Select):
		clue := m.puzzle.FindClueAt(m.cursor.X, m.cursor.Y)
		if clue != nil {
			m.selectedClue = &clue.ID
			m.expansion = expansion{clueID: clue.ID}
			// Initialize from existing rectangle if present.
			if r := m.puzzle.FindRectangleForClue(clue.ID); r != nil {
				m.expansion.left = clue.X - r.X
				m.expansion.right = r.X + r.W - 1 - clue.X
				m.expansion.up = clue.Y - r.Y
				m.expansion.down = r.Y + r.H - 1 - clue.Y
			}
		}
	case key.Matches(msg, m.keys.Delete):
		// Delete rectangle at cursor position.
		owner := m.puzzle.CellOwner(m.cursor.X, m.cursor.Y)
		if owner >= 0 {
			m.puzzle.RemoveRectangle(owner)
		}
	default:
		m.cursor.Move(m.keys.CursorKeyMap, msg, m.puzzle.Width-1, m.puzzle.Height-1)
	}
	return m
}

func (m Model) handleExpansionMode(msg tea.KeyMsg) Model {
	clue := m.puzzle.FindClueByID(*m.selectedClue)
	if clue == nil {
		m.selectedClue = nil
		return m
	}

	switch {
	// Expand each side independently.
	case key.Matches(msg, m.keys.Right):
		if m.expansion.right < m.puzzle.Width-1-clue.X {
			m.expansion.right++
		}
	case key.Matches(msg, m.keys.Left):
		if m.expansion.left < clue.X {
			m.expansion.left++
		}
	case key.Matches(msg, m.keys.Down):
		if m.expansion.down < m.puzzle.Height-1-clue.Y {
			m.expansion.down++
		}
	case key.Matches(msg, m.keys.Up):
		if m.expansion.up < clue.Y {
			m.expansion.up++
		}
	// Shrink each side independently.
	case key.Matches(msg, m.keys.ShrinkRight):
		if m.expansion.right > 0 {
			m.expansion.right--
		}
	case key.Matches(msg, m.keys.ShrinkLeft):
		if m.expansion.left > 0 {
			m.expansion.left--
		}
	case key.Matches(msg, m.keys.ShrinkDown):
		if m.expansion.down > 0 {
			m.expansion.down--
		}
	case key.Matches(msg, m.keys.ShrinkUp):
		if m.expansion.up > 0 {
			m.expansion.up--
		}
	case key.Matches(msg, m.keys.Select):
		rect := m.expansion.rect(clue)
		if rect.Area() == clue.Value && !m.puzzle.Overlaps(rect, clue.ID) {
			m.puzzle.SetRectangle(rect)
			m.selectedClue = nil
		}
	case key.Matches(msg, m.keys.Cancel):
		m.selectedClue = nil
	case key.Matches(msg, m.keys.Delete):
		m.puzzle.RemoveRectangle(clue.ID)
		m.selectedClue = nil
	}
	return m
}

func (m Model) View() string {
	solved := m.puzzle.IsSolved()
	title := game.TitleBarView("Shikaku", m.modeTitle, solved)
	grid := gridView(m, solved)
	info := infoView(&m.puzzle)
	status := statusBarView(m.selectedClue != nil, m.showFullHelp)
	return lipgloss.JoinVertical(lipgloss.Center, title, grid, info, status)
}

func (m Model) SetTitle(t string) game.Gamer {
	m.modeTitle = t
	return m
}

func (m Model) IsSolved() bool {
	return m.puzzle.IsSolved()
}

func (m Model) Reset() game.Gamer {
	m.puzzle.Rectangles = nil
	m.cursor = game.Cursor{}
	m.selectedClue = nil
	return m
}

func (m Model) GetDebugInfo() string {
	status := "In Progress"
	if m.puzzle.IsSolved() {
		status = "Solved"
	}

	selectedStr := "None"
	if m.selectedClue != nil {
		selectedStr = fmt.Sprintf("%d", *m.selectedClue)
	}

	s := game.DebugHeader("Shikaku", [][2]string{
		{"Status", status},
		{"Grid Size", fmt.Sprintf("%dx%d", m.puzzle.Width, m.puzzle.Height)},
		{"Clues", fmt.Sprintf("%d", len(m.puzzle.Clues))},
		{"Rectangles", fmt.Sprintf("%d", len(m.puzzle.Rectangles))},
		{"Cursor", fmt.Sprintf("(%d, %d)", m.cursor.X, m.cursor.Y)},
		{"Selected Clue", selectedStr},
	})

	var rows [][]string
	for _, c := range m.puzzle.Clues {
		r := m.puzzle.FindRectangleForClue(c.ID)
		rectStr := "-"
		rectStatus := "Empty"
		if r != nil {
			rectStr = fmt.Sprintf("(%d,%d) %dx%d", r.X, r.Y, r.W, r.H)
			if r.Area() == c.Value {
				rectStatus = "Correct"
			} else {
				rectStatus = fmt.Sprintf("Area %d", r.Area())
			}
		}
		rows = append(rows, []string{
			fmt.Sprintf("%d", c.ID),
			fmt.Sprintf("(%d,%d)", c.X, c.Y),
			fmt.Sprintf("%d", c.Value),
			rectStr,
			rectStatus,
		})
	}
	s += game.DebugTable("Clues", []string{"ID", "Pos", "Value", "Rectangle", "Status"}, rows)

	return s
}
