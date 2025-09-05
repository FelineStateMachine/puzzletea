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
	maxWidth, maxHeight := m.rowHints.RequiredLen()*cellWidth, m.colHints.RequiredLen()
	g := gridView(m.grid, m.cursor)
	r := rowHintView(m.rowHints, maxWidth)
	c := colHintView(m.colHints, maxHeight)
	spacer := baseStyle.Width(maxWidth).Height(maxHeight).Render("")

	s1 := lipgloss.JoinHorizontal(lipgloss.Bottom, spacer, c)
	s2 := lipgloss.JoinHorizontal(lipgloss.Top, r, g)

	s := lipgloss.JoinVertical(lipgloss.Left, s1, s2)

	return s
}

func (m Model) GetDebugInfo() string {
	curr := generateTomography(m.grid)
	solved := curr.rows.equal(m.rowHints) && curr.cols.equal(m.colHints)
	s := fmt.Sprintf(
		"# Nonogram - Dev Info\n"+
			"| Property          | Value   |\n"+
			"| :----- | :---- |\n"+
			"| Solved | %v |\n"+
			"| Cursor | (%d,%d) |\n"+
			"| Dimensions | [%dx%d] |\n"+
			"| Hint Render Width | %dr %dc |\n"+
			"\n"+
			"Row Tomography\n```%v```\n"+
			"Row Hints\n```%v```\n\n"+
			"Col Tomography\n```%v```\n"+
			"Col Hints\n```%v```\n",
		solved,
		m.cursor.x,
		m.cursor.y,
		m.width,
		m.height,
		m.rowHints.RequiredLen()*cellWidth,
		m.colHints.RequiredLen(),
		curr.rows,
		m.rowHints,
		curr.cols,
		m.colHints,
	)

	return s
}

func (m *Model) updateTile(r rune) {
	m.grid[m.cursor.y][m.cursor.x] = r
}
