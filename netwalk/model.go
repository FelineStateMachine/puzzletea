package netwalk

import (
	"fmt"

	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"github.com/FelineStateMachine/puzzletea/game"
)

type Model struct {
	puzzle       Puzzle
	cursor       game.Cursor
	keys         KeyMap
	modeTitle    string
	showFullHelp bool
	termWidth    int
	termHeight   int
	originX      int
	originY      int
	originValid  bool
	state        boardState
}

var _ game.Gamer = Model{}

func New(mode NetwalkMode, puzzle Puzzle) (game.Gamer, error) {
	cursor := puzzle.firstActive()
	m := Model{
		puzzle:    puzzle,
		cursor:    game.Cursor{X: cursor.X, Y: cursor.Y},
		keys:      DefaultKeyMap,
		modeTitle: mode.Title(),
	}
	m.recompute()
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
		m = m.handleMouseClick(msg)
	case tea.KeyPressMsg:
		switch {
		case key.Matches(msg, m.keys.Rotate):
			m.rotateCurrent(1)
		case key.Matches(msg, m.keys.RotateBack):
			m.rotateCurrent(3)
		case key.Matches(msg, m.keys.Lock):
			m.toggleCurrentLock()
		default:
			m.cursor.Move(m.keys.CursorKeyMap, msg, m.puzzle.Size-1, m.puzzle.Size-1)
		}
	}
	m.updateKeyBindings()
	return m, nil
}

func (m Model) View() string {
	title := game.TitleBarView("Netwalk", m.modeTitle, m.state.solved)
	grid := gridView(m)
	if m.state.solved {
		return game.ComposeGameView(title, grid)
	}
	return game.ComposeGameViewRows(
		title,
		grid,
		game.StableRow(statusBarView(m, m.showFullHelp), statusBarView(m, false), statusBarView(m, true)),
	)
}

func (m Model) SetTitle(t string) game.Gamer {
	m.modeTitle = t
	return m
}

func (m Model) IsSolved() bool {
	return m.state.solved
}

func (m Model) Reset() game.Gamer {
	for y := range m.puzzle.Size {
		for x := range m.puzzle.Size {
			m.puzzle.Tiles[y][x].Rotation = m.puzzle.Tiles[y][x].InitialRotation
			m.puzzle.Tiles[y][x].Locked = false
		}
	}
	cursor := m.puzzle.firstActive()
	m.cursor = game.Cursor{X: cursor.X, Y: cursor.Y}
	m.originValid = false
	m.recompute()
	return m
}

func (m Model) GetDebugInfo() string {
	return game.DebugHeader("Netwalk", [][2]string{
		{"Status", stateLabel(m.state.solved)},
		{"Cursor", fmt.Sprintf("(%d, %d)", m.cursor.X, m.cursor.Y)},
		{"Grid Size", fmt.Sprintf("%d×%d", m.puzzle.Size, m.puzzle.Size)},
		{"Root", fmt.Sprintf("(%d, %d)", m.puzzle.Root.X, m.puzzle.Root.Y)},
		{"Connected", fmt.Sprintf("%d / %d", m.state.connected, m.state.nonEmpty)},
		{"Dangling", fmt.Sprintf("%d", m.state.dangling)},
		{"Locked", fmt.Sprintf("%d", m.state.locked)},
	})
}

func stateLabel(solved bool) string {
	if solved {
		return "Solved"
	}
	return "In Progress"
}

func (m *Model) rotateCurrent(delta uint8) {
	if m.state.solved {
		return
	}
	t := &m.puzzle.Tiles[m.cursor.Y][m.cursor.X]
	if !isActive(*t) || t.Locked {
		return
	}
	t.Rotation = (t.Rotation + delta) % 4
	m.recompute()
}

func (m *Model) toggleCurrentLock() {
	if m.state.solved {
		return
	}
	t := &m.puzzle.Tiles[m.cursor.Y][m.cursor.X]
	if !isActive(*t) {
		return
	}
	t.Locked = !t.Locked
	m.recompute()
}

func (m *Model) recompute() {
	m.state = analyzePuzzle(m.puzzle)
}

func (m Model) handleMouseClick(msg tea.MouseClickMsg) Model {
	col, row, ok := m.screenToGrid(msg.X, msg.Y)
	if !ok {
		return m
	}
	m.cursor.X = col
	m.cursor.Y = row

	switch msg.Button {
	case tea.MouseLeft:
		m.rotateCurrent(1)
	case tea.MouseRight:
		m.toggleCurrentLock()
	}
	return m
}

func (m *Model) screenToGrid(screenX, screenY int) (col, row int, ok bool) {
	ox, oy := m.cachedGridOrigin()
	return game.DynamicGridScreenToCell(
		game.DynamicGridMetrics{
			Width:     m.puzzle.Size,
			Height:    m.puzzle.Size,
			CellWidth: cellWidth,
		},
		ox,
		oy,
		screenX,
		screenY,
		false,
	)
}

func (m *Model) cachedGridOrigin() (x, y int) {
	if m.originValid {
		return m.originX, m.originY
	}
	x, y = m.gridOrigin()
	m.originX, m.originY = x, y
	m.originValid = true
	return x, y
}

func (m *Model) gridOrigin() (x, y int) {
	title := game.TitleBarView("Netwalk", m.modeTitle, m.state.solved)
	grid := gridView(*m)
	return game.DynamicGridOrigin(m.termWidth, m.termHeight, m.View(), title, grid)
}
