package nurikabe

import (
	"fmt"

	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"github.com/FelineStateMachine/puzzletea/game"
)

type Model struct {
	width, height int
	clues         clueGrid
	marks         grid
	initialMarks  grid
	cursor        game.Cursor
	solved        bool
	conflicts     [][]bool
	keys          KeyMap
	modeTitle     string
	showFullHelp  bool

	dragging   bool
	dragTarget cellState

	termWidth, termHeight int
	originX, originY      int
	originValid           bool

	lastMouseX, lastMouseY int
	lastMouseBtn           string
	lastMouseGridCol       int
	lastMouseGridRow       int
	lastMouseHit           bool
}

var _ game.Gamer = Model{}

func New(mode NurikabeMode, puzzle Puzzle) (game.Gamer, error) {
	if puzzle.Width != mode.Width || puzzle.Height != mode.Height {
		return Model{}, fmt.Errorf("puzzle size %dx%d does not match mode %dx%d", puzzle.Width, puzzle.Height, mode.Width, mode.Height)
	}
	if err := validateClues(puzzle.Clues, puzzle.Width, puzzle.Height); err != nil {
		return Model{}, err
	}

	marks := newGrid(puzzle.Width, puzzle.Height, unknownCell)
	for y := range puzzle.Height {
		for x := range puzzle.Width {
			if puzzle.Clues[y][x] > 0 {
				marks[y][x] = islandCell
			}
		}
	}

	m := Model{
		width:        puzzle.Width,
		height:       puzzle.Height,
		clues:        cloneClues(puzzle.Clues),
		marks:        marks,
		initialMarks: marks.clone(),
		cursor:       game.Cursor{X: 0, Y: 0},
		keys:         DefaultKeyMap,
		modeTitle:    mode.Title(),
		dragTarget:   unknownCell,
	}
	m.recomputeDerivedState()
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

	case tea.KeyPressMsg:
		switch {
		case key.Matches(msg, m.keys.SetSea):
			if !m.solved {
				m.setCellAtCursor(seaCell)
			}
		case key.Matches(msg, m.keys.SetIsland):
			if !m.solved {
				current := m.marks[m.cursor.Y][m.cursor.X]
				m.setCellAtCursor(m.islandTarget(current))
			}
		case key.Matches(msg, m.keys.Clear):
			if !m.solved {
				m.setCellAtCursor(unknownCell)
			}
		default:
			m.cursor.Move(m.keys.CursorKeyMap, msg, m.width-1, m.height-1)
		}

	case tea.MouseClickMsg:
		m.lastMouseX, m.lastMouseY = msg.X, msg.Y
		m.lastMouseBtn = msg.String()
		col, row, ok := m.screenToGrid(msg.X, msg.Y)
		m.lastMouseGridCol, m.lastMouseGridRow = col, row
		m.lastMouseHit = ok
		if m.solved || !ok {
			break
		}
		m.cursor.X, m.cursor.Y = col, row
		switch msg.Button {
		case tea.MouseLeft:
			target := m.leftClickTarget()
			m.dragging = true
			m.dragTarget = target
			m.setCellAtCursor(target)
		case tea.MouseRight:
			target := m.rightClickTarget(m.marks[row][col])
			m.dragging = true
			m.dragTarget = target
			m.setCellAtCursor(target)
		}

	case tea.MouseMotionMsg:
		if m.solved || !m.dragging {
			break
		}
		col, row, ok := m.screenToGrid(msg.X, msg.Y)
		if !ok {
			break
		}
		if col == m.cursor.X && row == m.cursor.Y {
			break
		}
		m.cursor.X, m.cursor.Y = col, row
		m.setCellAtCursor(m.dragTarget)

	case tea.MouseReleaseMsg:
		m.dragging = false
		m.dragTarget = unknownCell
	}

	m.updateKeyBindings()
	return m, nil
}

func (m *Model) setCellAtCursor(state cellState) {
	x, y := m.cursor.X, m.cursor.Y
	if isClueCell(m.clues, x, y) {
		// Clue cells are immutable island cells.
		m.marks[y][x] = islandCell
		return
	}
	m.marks[y][x] = state
	m.afterBoardChange()
}

func (m *Model) afterBoardChange() {
	wasSolved := m.solved
	m.recomputeDerivedState()
	if wasSolved != m.solved {
		m.originValid = false
	}
}

func (m *Model) recomputeDerivedState() {
	m.solved = isSolvedGrid(m.marks, m.clues)
	m.conflicts = computeConflicts(m.marks, m.clues)
}

func (m Model) leftClickTarget() cellState {
	return seaCell
}

func (m Model) rightClickTarget(current cellState) cellState {
	return m.islandTarget(current)
}

func (m Model) islandTarget(current cellState) cellState {
	if current == islandCell {
		return unknownCell
	}
	return islandCell
}

func (m Model) View() string {
	title := game.TitleBarView("Nurikabe", m.modeTitle, m.solved)
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
	m.marks = m.initialMarks.clone()
	m.cursor = game.Cursor{}
	m.dragging = false
	m.dragTarget = unknownCell
	m.recomputeDerivedState()
	m.originValid = false
	return m
}

func (m Model) GetDebugInfo() string {
	status := "In Progress"
	if m.solved {
		status = "Solved"
	}

	unknownCount := 0
	seaCount := 0
	islandCount := 0
	conflictCount := 0
	for y := range m.height {
		for x := range m.width {
			switch m.marks[y][x] {
			case unknownCell:
				unknownCount++
			case seaCell:
				seaCount++
			case islandCell:
				islandCount++
			}
			if m.conflicts[y][x] {
				conflictCount++
			}
		}
	}

	ox, oy := m.gridOrigin()
	hit := "miss"
	if m.lastMouseHit {
		hit = fmt.Sprintf("(%d, %d)", m.lastMouseGridCol, m.lastMouseGridRow)
	}

	return game.DebugHeader("Nurikabe", [][2]string{
		{"Status", status},
		{"Cursor", fmt.Sprintf("(%d, %d)", m.cursor.X, m.cursor.Y)},
		{"Grid Size", fmt.Sprintf("%d x %d", m.width, m.height)},
		{"Clues", fmt.Sprintf("%d", countClues(m.clues))},
		{"Cells", fmt.Sprintf("unknown=%d sea=%d island=%d", unknownCount, seaCount, islandCount)},
		{"Conflicts", fmt.Sprintf("%d", conflictCount)},
		{"Term Size", fmt.Sprintf("%d x %d", m.termWidth, m.termHeight)},
		{"Grid Origin", fmt.Sprintf("(%d, %d)", ox, oy)},
		{"Last Mouse", fmt.Sprintf("screen=(%d, %d) btn=%s grid=%s", m.lastMouseX, m.lastMouseY, m.lastMouseBtn, hit)},
		{"Drag", fmt.Sprintf("active=%v target=%q", m.dragging, rune(m.dragTarget))},
	})
}
