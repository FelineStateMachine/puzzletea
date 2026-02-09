// Package nonogram implements the grid-based picture logic puzzle.
package nonogram

import (
	"errors"
	"fmt"

	"github.com/FelineStateMachine/puzzletea/game"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const (
	filledTile = '.'
	markedTile = '-'
	emptyTile  = ' '
)

type Model struct {
	width, height int
	rowHints      TomographyDefinition
	colHints      TomographyDefinition
	cursor        game.Cursor
	grid          grid
	keys          KeyMap
	modeTitle     string
	showFullHelp  bool
	currentHints  Hints // cached tomography of the current grid
	solved        bool  // cached solved state
}

func New(mode NonogramMode, hints Hints) (game.Gamer, error) {
	h, w := mode.Height, mode.Width
	r, c := hints.rows, hints.cols
	if w < r.RequiredLen() {
		return Model{}, errors.New("puzzle width does not support row tomography definition")
	}
	if h < c.RequiredLen() {
		return Model{}, errors.New("puzzle height does not support column tomography definition")
	}
	g := newGrid(createEmptyState(h, w))
	curr := generateTomography(g)
	m := Model{
		width:        w,
		height:       h,
		rowHints:     r,
		colHints:     c,
		grid:         g,
		keys:         DefaultKeyMap,
		modeTitle:    mode.Title(),
		currentHints: curr,
		solved:       curr.rows.equal(r) && curr.cols.equal(c),
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
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keys.FillTile):
			m.updateTile(filledTile)
		case key.Matches(msg, m.keys.MarkTile):
			m.updateTile(markedTile)
		case key.Matches(msg, m.keys.ClearTile):
			m.updateTile(emptyTile)
		default:
			m.cursor.Move(m.keys.CursorKeyMap, msg, m.width-1, m.height-1)
		}
	}
	m.updateKeyBindings()
	return m, nil
}

func (m Model) View() string {
	maxWidth, maxHeight := m.rowHints.RequiredLen()*cellWidth, m.colHints.RequiredLen()

	title := game.TitleBarView("Nonogram", m.modeTitle, m.solved)
	g := gridView(m.grid, m.cursor, m.solved)
	r := rowHintView(m.rowHints, maxWidth, m.currentHints.rows)
	c := colHintView(m.colHints, maxHeight, m.currentHints.cols)
	spacer := baseStyle.Width(maxWidth).Height(maxHeight).Render("")
	status := nonoStatusBarView(m.showFullHelp)

	s1 := lipgloss.JoinHorizontal(lipgloss.Bottom, spacer, c)
	s2 := lipgloss.JoinHorizontal(lipgloss.Top, r, g)

	grid := lipgloss.JoinVertical(lipgloss.Left, s1, s2)

	return lipgloss.JoinVertical(lipgloss.Left, title, grid, status)
}

func (m Model) GetDebugInfo() string {
	status := "In Progress"
	if m.solved {
		status = "Solved"
	}

	s := game.DebugHeader("Nonogram", [][2]string{
		{"Status", status},
		{"Cursor", fmt.Sprintf("(%d, %d)", m.cursor.X, m.cursor.Y)},
		{"Grid Size", fmt.Sprintf("%d x %d", m.width, m.height)},
		{"Hint Widths", fmt.Sprintf("row: %d, col: %d", m.rowHints.RequiredLen()*cellWidth, m.colHints.RequiredLen())},
	})

	s += tomoDebugTable("Row Tomography", "Row", m.rowHints, m.currentHints.rows)
	s += tomoDebugTable("Column Tomography", "Col", m.colHints, m.currentHints.cols)

	return s
}

func tomoDebugTable(heading, label string, hints, current TomographyDefinition) string {
	var rows [][]string
	for i, hint := range hints {
		var currStr string
		match := false
		if i < len(current) {
			currStr = intSliceStr(current[i])
			match = intSliceEqual(hint, current[i])
		}
		matchStr := "No"
		if match {
			matchStr = "Yes"
		}
		rows = append(rows, []string{fmt.Sprintf("%d", i), intSliceStr(hint), currStr, matchStr})
	}
	return game.DebugTable(heading, []string{label, "Hint", "Current", "Match"}, rows)
}

func intSliceStr(s []int) string {
	if len(s) == 0 {
		return "[]"
	}
	result := ""
	for i, v := range s {
		if i > 0 {
			result += ", "
		}
		result += fmt.Sprintf("%d", v)
	}
	return result
}

func (m Model) SetTitle(t string) game.Gamer {
	m.modeTitle = t
	return m
}

func (m Model) IsSolved() bool {
	return m.solved
}

func (m *Model) updateTile(r rune) {
	m.grid[m.cursor.Y][m.cursor.X] = r
	m.currentHints = generateTomography(m.grid)
	m.solved = m.currentHints.rows.equal(m.rowHints) && m.currentHints.cols.equal(m.colHints)
}
