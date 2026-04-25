package sudokurgb

import (
	"fmt"
	"strconv"

	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"github.com/FelineStateMachine/puzzletea/game"
)

const (
	gridSize   = 9
	valueCount = 3
	houseQuota = 3
)

var _ game.Gamer = Model{}

type Model struct {
	cursor       game.Cursor
	grid         grid
	provided     []cell
	providedGrid [gridSize][gridSize]bool
	analysis     boardAnalysis
	keys         KeyMap
	modeTitle    string
	showFullHelp bool
	termWidth    int
	termHeight   int
	originX      int
	originY      int
	originValid  bool
}

func buildProvidedGrid(provided []cell) [gridSize][gridSize]bool {
	var pg [gridSize][gridSize]bool
	for _, c := range provided {
		pg[c.y][c.x] = true
	}
	return pg
}

func New(mode SudokuRGBMode, provided []cell) (game.Gamer, error) {
	g := newGrid(provided)
	m := Model{
		grid:         g,
		provided:     provided,
		providedGrid: buildProvidedGrid(provided),
		analysis:     analyzeGrid(g),
		keys:         DefaultKeyMap,
		modeTitle:    mode.Title(),
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
		case key.Matches(msg, m.keys.FillValue):
			value, _ := strconv.Atoi(msg.String())
			m.updateCell(value)
		case key.Matches(msg, m.keys.ClearCell):
			m.updateCell(0)
		default:
			m.cursor.Move(m.keys.CursorKeyMap, msg, gridSize-1, gridSize-1)
		}
	}

	m.updateKeyBindings()
	return m, nil
}

func (m Model) SetTitle(t string) game.Gamer {
	m.modeTitle = t
	return m
}

func (m Model) IsSolved() bool {
	return m.isSolved()
}

func (m Model) Reset() game.Gamer {
	m.grid = newGrid(m.provided)
	m.analysis = analyzeGrid(m.grid)
	m.originValid = false
	return m
}

func (m *Model) updateCell(value int) {
	if m.providedGrid[m.cursor.Y][m.cursor.X] {
		return
	}

	wasSolved := m.isSolved()
	m.grid[m.cursor.Y][m.cursor.X].v = value
	m.analysis = analyzeGrid(m.grid)
	if m.isSolved() != wasSolved {
		m.originValid = false
	}
}

func (m Model) View() string {
	solved := isSolvedWith(m.grid, m.analysis)
	title := game.TitleBarView("Sudoku RGB", m.modeTitle, solved)
	board := buildBoardBlock(m, solved)
	if solved {
		return game.ComposeGameView(title, board.Block)
	}

	status := statusBarView(m.showFullHelp)
	return game.ComposeGameViewRows(title, board.Block, game.StableRow(status, statusBarView(false), statusBarView(true)))
}

func (m Model) GetDebugInfo() string {
	cursorCell := m.grid[m.cursor.Y][m.cursor.X]
	isProvided := m.providedGrid[m.cursor.Y][m.cursor.X]
	boxConflict := m.analysis.boxConflictCells[m.cursor.Y][m.cursor.X]
	solved := m.isSolved()

	filledCount := 0
	conflictCount := 0
	for y := range gridSize {
		for x := range gridSize {
			if m.grid[y][x].v != 0 {
				filledCount++
			}
			if m.analysis.boxConflictCells[y][x] {
				conflictCount++
			}
		}
	}

	status := "In Progress"
	if solved {
		status = "Solved"
	}

	return game.DebugHeader("Sudoku RGB", [][2]string{
		{"Status", status},
		{"Cursor", fmt.Sprintf("(%d, %d)", m.cursor.X, m.cursor.Y)},
		{"Cell Value", cellDebugValue(cursorCell.v)},
		{"Is Provided", fmt.Sprintf("%v", isProvided)},
		{"Box Conflict", fmt.Sprintf("%v", boxConflict)},
		{"Row Counts", m.analysis.rowCountsString(m.cursor.Y)},
		{"Col Counts", m.analysis.colCountsString(m.cursor.X)},
		{"Cells Filled", fmt.Sprintf("%d / 81", filledCount)},
		{"Box Conflict Count", fmt.Sprintf("%d", conflictCount)},
		{"Provided Count", fmt.Sprintf("%d", len(m.provided))},
	})
}

func cellDebugValue(value int) string {
	if value == 0 {
		return "empty"
	}
	return fmt.Sprintf("%d (%s)", value, cellContentValue(value))
}
