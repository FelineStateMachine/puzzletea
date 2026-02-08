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

type cursor struct {
	x, y int
}

type Model struct {
	width, height int
	rowHints      TomographyDefinition
	colHints      TomographyDefinition
	cursor        cursor
	grid          grid
	keys          KeyMap
	modeName      string
}

func New(mode NonogramMode, hints Hints, save ...string) (game.Gamer, error) {
	h, w := mode.Height, mode.Width
	r, c := hints.rows, hints.cols
	if w < r.RequiredLen() {
		return Model{}, errors.New("Puzzle width does not support row tomography definition.")
	}
	if h < c.RequiredLen() {
		return Model{}, errors.New("Puzzle height does not support column tomography definition.")
	}
	s := loadSave(createEmptyState(h, w), save...)
	m := Model{
		width:    w,
		height:   h,
		rowHints: r,
		colHints: c,
		grid:     newGrid(s),
		keys:     DefaultKeyMap,
		modeName: mode.title,
	}

	return m, nil
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (game.Gamer, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keys.FillTile):
			m.updateTile(filledTile)
		case key.Matches(msg, m.keys.MarkTile):
			m.updateTile(markedTile)
		case key.Matches(msg, m.keys.ClearTile):
			m.updateTile(emptyTile)
		case key.Matches(msg, m.keys.Up):
			if m.cursor.y > 0 {
				m.cursor.y--
			}
		case key.Matches(msg, m.keys.Down):
			if m.cursor.y < m.height-1 {
				m.cursor.y++
			}
		case key.Matches(msg, m.keys.Left):
			if m.cursor.x > 0 {
				m.cursor.x--
			}
		case key.Matches(msg, m.keys.Right):
			if m.cursor.x < m.width-1 {
				m.cursor.x++
			}
		}

	}
	m.updateKeyBindinds()
	return m, nil
}

func (m Model) View() string {
	curr := generateTomography(m.grid)
	solved := curr.rows.equal(m.rowHints) && curr.cols.equal(m.colHints)

	maxWidth, maxHeight := m.rowHints.RequiredLen()*cellWidth, m.colHints.RequiredLen()

	title := nonoTitleBarView(m.modeName, solved)
	g := gridView(m.grid, m.cursor, solved)
	r := rowHintView(m.rowHints, maxWidth, curr.rows)
	c := colHintView(m.colHints, maxHeight, curr.cols)
	spacer := baseStyle.Width(maxWidth).Height(maxHeight).Render("")
	status := nonoStatusBarView(m.keys)

	s1 := lipgloss.JoinHorizontal(lipgloss.Bottom, spacer, c)
	s2 := lipgloss.JoinHorizontal(lipgloss.Top, r, g)

	grid := lipgloss.JoinVertical(lipgloss.Left, s1, s2)

	return lipgloss.JoinVertical(lipgloss.Left, title, grid, status)
}

func (m Model) GetDebugInfo() string {
	curr := generateTomography(m.grid)
	solved := curr.rows.equal(m.rowHints) && curr.cols.equal(m.colHints)

	status := "In Progress"
	if solved {
		status = "Solved"
	}

	s := fmt.Sprintf(
		"# Nonogram\n\n"+
			"## Game State\n\n"+
			"| Property | Value |\n"+
			"| :--- | :--- |\n"+
			"| Status | %s |\n"+
			"| Cursor | (%d, %d) |\n"+
			"| Grid Size | %d x %d |\n"+
			"| Hint Widths | row: %d, col: %d |\n",
		status,
		m.cursor.x, m.cursor.y,
		m.width, m.height,
		m.rowHints.RequiredLen()*cellWidth, m.colHints.RequiredLen(),
	)

	s += "\n## Row Tomography\n\n"
	s += "| Row | Hint | Current | Match |\n"
	s += "| :--- | :--- | :--- | :--- |\n"
	for i, hint := range m.rowHints {
		var currStr string
		match := false
		if i < len(curr.rows) {
			currStr = intSliceStr(curr.rows[i])
			match = intSliceEqual(hint, curr.rows[i])
		}
		matchStr := "No"
		if match {
			matchStr = "Yes"
		}
		s += fmt.Sprintf("| %d | %s | %s | %s |\n", i, intSliceStr(hint), currStr, matchStr)
	}

	s += "\n## Column Tomography\n\n"
	s += "| Col | Hint | Current | Match |\n"
	s += "| :--- | :--- | :--- | :--- |\n"
	for i, hint := range m.colHints {
		var currStr string
		match := false
		if i < len(curr.cols) {
			currStr = intSliceStr(curr.cols[i])
			match = intSliceEqual(hint, curr.cols[i])
		}
		matchStr := "No"
		if match {
			matchStr = "Yes"
		}
		s += fmt.Sprintf("| %d | %s | %s | %s |\n", i, intSliceStr(hint), currStr, matchStr)
	}

	return s
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

func (m *Model) updateTile(r rune) {
	m.grid[m.cursor.y][m.cursor.x] = r
}
