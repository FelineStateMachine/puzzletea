// Package takuzuplus implements the Takuzu+ puzzle game.
package takuzuplus

import (
	"fmt"

	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"github.com/FelineStateMachine/puzzletea/game"
	"github.com/FelineStateMachine/puzzletea/games/takuzu"
)

type Model struct {
	size         int
	grid         grid
	initialGrid  grid
	provided     [][]bool
	relations    relations
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

func New(mode TakuzuPlusMode, puzzle grid, provided [][]bool, rels relations) (game.Gamer, error) {
	m := Model{
		size:        mode.Size,
		grid:        puzzle,
		initialGrid: puzzle.clone(),
		provided:    provided,
		relations:   rels,
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
	title := game.TitleBarView("Takuzu+", m.modeTitle, m.solved)
	grid := gridView(m)
	if m.solved {
		return game.ComposeGameView(title, grid)
	}
	info := countContextView(m)
	status := statusBarView(m.showFullHelp)
	return game.ComposeGameViewRows(
		title,
		grid,
		game.StaticRow(info),
		game.StableRow(status, statusBarView(false), statusBarView(true)),
	)
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
	for y := range m.size {
		for x := range m.size {
			if m.grid[y][x] == emptyCell {
				return false
			}
		}
	}
	return takuzu.CheckConstraintsGrid(m.grid, m.size) &&
		takuzu.HasUniqueLinesGrid(m.grid, m.size) &&
		checkRelations(m.grid, m.relations)
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

	return game.DebugHeader("Takuzu+", [][2]string{
		{"Status", status},
		{"Cursor", fmt.Sprintf("(%d, %d)", m.cursor.X, m.cursor.Y)},
		{"Grid Size", fmt.Sprintf("%d×%d", m.size, m.size)},
		{"Cells Filled", fmt.Sprintf("%d / %d", filled, m.size*m.size)},
		{"Row Counts", countPairString(countLine(m.grid[m.cursor.Y]))},
		{"Col Counts", countPairString(countColumn(m.grid, m.cursor.X, m.size))},
		{"Relation Clues", fmt.Sprintf("%d", countRelations(m.relations))},
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
