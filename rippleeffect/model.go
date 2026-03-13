package rippleeffect

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
	geo          *geometry
	grid         grid
	initialGrid  grid
	givens       grid
	cursor       game.Cursor
	conflicts    [][]bool
	solved       bool
	keys         KeyMap
	modeTitle    string
	showFullHelp bool
}

func New(mode Mode, puzzle Puzzle) (game.Gamer, error) {
	if puzzle.Width != mode.Size || puzzle.Height != mode.Size {
		return nil, fmt.Errorf("ripple effect puzzle size %dx%d does not match mode %dx%d", puzzle.Width, puzzle.Height, mode.Size, mode.Size)
	}

	geo, err := buildGeometry(puzzle.Width, puzzle.Height, puzzle.Cages)
	if err != nil {
		return nil, err
	}
	if !gridsMatchPuzzle(puzzle.Givens, puzzle.Width, puzzle.Height) {
		return nil, fmt.Errorf("ripple effect givens grid does not match puzzle size")
	}

	m := Model{
		width:       puzzle.Width,
		height:      puzzle.Height,
		geo:         geo,
		grid:        cloneGrid(puzzle.Givens),
		initialGrid: cloneGrid(puzzle.Givens),
		givens:      cloneGrid(puzzle.Givens),
		cursor:      game.Cursor{},
		keys:        DefaultKeyMap,
		modeTitle:   mode.Title(),
	}
	m.recompute()
	return m, nil
}

func gridsMatchPuzzle(g grid, width, height int) bool {
	if len(g) != height {
		return false
	}
	for y := range height {
		if len(g[y]) != width {
			return false
		}
	}
	return true
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (game.Gamer, tea.Cmd) {
	switch msg := msg.(type) {
	case game.HelpToggleMsg:
		m.showFullHelp = msg.Show
	case tea.KeyPressMsg:
		switch {
		case key.Matches(msg, m.keys.Clear):
			m.setCell(0)
		case key.Matches(msg, m.keys.FillValue):
			value, _ := strconv.Atoi(msg.String())
			m.setCell(value)
		default:
			m.cursor.Move(m.keys.CursorKeyMap, msg, m.width-1, m.height-1)
		}
	}

	m.updateKeyBindings()
	return m, nil
}

func (m *Model) setCell(value int) {
	if m.solved || m.givens[m.cursor.Y][m.cursor.X] != 0 {
		return
	}
	if value != 0 && value > m.geo.cageSizes[m.geo.cageGrid[m.cursor.Y][m.cursor.X]] {
		return
	}
	m.grid[m.cursor.Y][m.cursor.X] = value
	m.recompute()
}

func (m *Model) recompute() {
	result := validateGridState(m.grid, m.geo)
	m.conflicts = result.conflicts
	m.solved = result.solved
}

func (m Model) View() string {
	title := game.TitleBarView("Ripple Effect", m.modeTitle, m.solved)
	body := gridView(m)
	if m.solved {
		return game.ComposeGameView(title, body)
	}
	status := statusBarView(m.showFullHelp)
	return game.ComposeGameViewRows(title, body, game.StableRow(status, statusBarView(false), statusBarView(true)))
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
	conflictCount := 0
	for y := range m.height {
		for x := range m.width {
			if m.grid[y][x] != 0 {
				filled++
			}
			if m.conflicts[y][x] {
				conflictCount++
			}
		}
	}

	cageIdx := m.geo.cageGrid[m.cursor.Y][m.cursor.X]
	return game.DebugHeader("Ripple Effect", [][2]string{
		{"Status", status},
		{"Cursor", fmt.Sprintf("(%d, %d)", m.cursor.X, m.cursor.Y)},
		{"Filled Cells", fmt.Sprintf("%d / %d", filled, m.width*m.height)},
		{"Conflict Count", fmt.Sprintf("%d", conflictCount)},
		{"Current Cage", fmt.Sprintf("%d (size %d)", m.geo.cages[cageIdx].ID, m.geo.cageSizes[cageIdx])},
	})
}
