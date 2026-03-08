// Package takuzu implements the binary (Binairo) puzzle game.
package takuzu

import (
	"fmt"

	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"github.com/FelineStateMachine/puzzletea/game"
)

// Model implements game.Gamer for Takuzu.
type Model struct {
	size         int
	grid         grid
	initialGrid  grid
	provided     [][]bool
	cursor       game.Cursor
	solved       bool
	keys         KeyMap
	modeTitle    string
	showFullHelp bool
	termWidth    int
	termHeight   int
	originX      int
	originY      int
	originValid  bool
}

var _ game.Gamer = Model{}

// New creates a new Takuzu game model.
func New(mode TakuzuMode, puzzle grid, provided [][]bool) (game.Gamer, error) {
	m := Model{
		size:        mode.Size,
		grid:        puzzle,
		initialGrid: puzzle.clone(),
		provided:    provided,
		cursor:      game.Cursor{X: 0, Y: 0},
		keys:        DefaultKeyMap,
		modeTitle:   mode.Title(),
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
		m.originValid = false
	case tea.WindowSizeMsg:
		m.termWidth = msg.Width
		m.termHeight = msg.Height
		m.originValid = false
	case tea.MouseClickMsg:
		if msg.Button == tea.MouseLeft {
			m = m.handleMouseClick(msg)
		}
	case tea.KeyPressMsg:
		switch {
		case key.Matches(msg, m.keys.PlaceZero):
			m.setCurrentCell(zeroCell)
		case key.Matches(msg, m.keys.PlaceOne):
			m.setCurrentCell(oneCell)
		case key.Matches(msg, m.keys.Clear):
			m.setCurrentCell(emptyCell)
		default:
			m.cursor.Move(m.keys.CursorKeyMap, msg, m.size-1, m.size-1)
		}
	}
	m.updateKeyBindings()
	return m, nil
}

func (m Model) View() string {
	title := game.TitleBarView("Takuzu", m.modeTitle, m.solved)
	grid := gridView(m)
	if m.solved {
		return game.ComposeGameView(title, grid)
	}
	status := statusBarView(m.showFullHelp)
	return game.ComposeGameViewRows(title, grid, game.StableRow(status, statusBarView(false), statusBarView(true)))
}

func (m Model) SetTitle(t string) game.Gamer {
	m.modeTitle = t
	return m
}

func (m Model) IsSolved() bool {
	return m.solved
}

func (m Model) Reset() game.Gamer {
	m.grid = m.initialGrid.clone()
	m.solved = m.checkSolved()
	m.cursor = game.Cursor{}
	m.originValid = false
	return m
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

func (m *Model) setCurrentCell(val rune) {
	if m.solved || m.provided[m.cursor.Y][m.cursor.X] {
		return
	}

	wasSolved := m.solved
	m.grid[m.cursor.Y][m.cursor.X] = val
	m.solved = m.checkSolved()
	if m.solved != wasSolved {
		m.originValid = false
	}
}

func (m *Model) cycleCurrentCell() {
	if m.solved || m.provided[m.cursor.Y][m.cursor.X] {
		return
	}

	switch m.grid[m.cursor.Y][m.cursor.X] {
	case zeroCell:
		m.setCurrentCell(oneCell)
	case oneCell:
		m.setCurrentCell(emptyCell)
	default:
		m.setCurrentCell(zeroCell)
	}
}

func (m Model) handleMouseClick(msg tea.MouseClickMsg) Model {
	col, row, ok := m.screenToGrid(msg.X, msg.Y)
	if !ok {
		return m
	}

	sameCell := col == m.cursor.X && row == m.cursor.Y
	m.cursor.X = col
	m.cursor.Y = row
	if sameCell {
		m.cycleCurrentCell()
	}
	return m
}
