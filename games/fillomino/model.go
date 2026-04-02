package fillomino

import (
	"fmt"
	"strconv"

	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"github.com/FelineStateMachine/puzzletea/game"
)

var _ game.Gamer = Model{}

type Model struct {
	width        int
	height       int
	grid         grid
	initialGrid  grid
	provided     [][]bool
	cursor       game.Cursor
	conflicts    [][]bool
	solved       bool
	keys         KeyMap
	modeTitle    string
	showFullHelp bool
	maxCellValue int
	termWidth    int
	termHeight   int
	originX      int
	originY      int
	originValid  bool
}

func New(mode Mode, puzzle Puzzle) (game.Gamer, error) {
	if puzzle.Width != mode.Size || puzzle.Height != mode.Size {
		return nil, fmt.Errorf("fillomino puzzle size %dx%d does not match mode %dx%d", puzzle.Width, puzzle.Height, mode.Size, mode.Size)
	}

	m := Model{
		width:        puzzle.Width,
		height:       puzzle.Height,
		grid:         cloneGrid(puzzle.Givens),
		initialGrid:  cloneGrid(puzzle.Givens),
		provided:     newProvidedMask(puzzle.Givens),
		cursor:       game.Cursor{},
		keys:         DefaultKeyMap,
		modeTitle:    mode.Title(),
		maxCellValue: mode.MaxRegion,
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
		if msg.Button == tea.MouseLeft {
			if col, row, ok := m.screenToGrid(msg.X, msg.Y); ok {
				m.cursor.X = col
				m.cursor.Y = row
			}
		}
	case tea.KeyPressMsg:
		switch {
		case key.Matches(msg, m.keys.Clear):
			m.setCell(0)
		case isDigitKey(msg.String()):
			value, _ := strconv.Atoi(msg.String())
			if value > 0 && value <= m.maxCellValue {
				m.setCell(value)
			}
		default:
			m.cursor.Move(m.keys.CursorKeyMap, msg, m.width-1, m.height-1)
		}
	}

	m.updateKeyBindings()
	return m, nil
}

func isDigitKey(s string) bool {
	return len(s) == 1 && s[0] >= '1' && s[0] <= '9'
}

func (m *Model) setCell(value int) {
	if m.solved || m.provided[m.cursor.Y][m.cursor.X] {
		return
	}
	m.grid[m.cursor.Y][m.cursor.X] = value
	m.recompute()
}

func (m *Model) recompute() {
	result := validateGridState(m.grid)
	m.conflicts = result.conflicts
	m.solved = result.solved
}

func (m Model) View() string {
	title := game.TitleBarView("Fillomino", m.modeTitle, m.solved)
	body := gridView(m)
	if m.solved {
		return game.ComposeGameView(title, body)
	}
	info := cursorRegionInfoView(m)
	status := statusBarView(m.showFullHelp)
	return game.ComposeGameViewRows(
		title,
		body,
		game.StableRow(info, cursorRegionInfoVariants(m)...),
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
	m.grid = cloneGrid(m.initialGrid)
	m.cursor = game.Cursor{}
	m.recompute()
	return m
}

func (m Model) GetDebugInfo() string {
	status := "In Progress"
	if m.solved {
		status = "Solved"
	}

	filled := 0
	conflicts := 0
	for y := range m.height {
		for x := range m.width {
			if m.grid[y][x] != 0 {
				filled++
			}
			if m.conflicts[y][x] {
				conflicts++
			}
		}
	}

	return game.DebugHeader("Fillomino", [][2]string{
		{"Status", status},
		{"Cursor", fmt.Sprintf("(%d, %d)", m.cursor.X, m.cursor.Y)},
		{"Cells Filled", fmt.Sprintf("%d / %d", filled, m.width*m.height)},
		{"Conflict Count", fmt.Sprintf("%d", conflicts)},
		{"Max Region", fmt.Sprintf("%d", m.maxCellValue)},
	})
}
