// Package wordsearch implements the word-finding grid puzzle.
package wordsearch

import (
	"fmt"
	"strings"

	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"github.com/FelineStateMachine/puzzletea/game"
)

type selectionState int

const (
	noSelection selectionState = iota
	startSelected
)

var _ game.Gamer = Model{}

type Model struct {
	width, height  int
	grid           grid
	words          []Word
	cursor         game.Cursor
	selection      selectionState
	selectionStart game.Cursor
	keys           KeyMap
	modeTitle      string
	solved         bool
	showFullHelp   bool
	foundCells     [][]bool

	// Mouse drag support: true while left-button is held after clicking a cell.
	mouseDragging bool

	// Screen geometry for mouse hit-testing.
	termWidth, termHeight int

	// Cached grid origin for mouse hit-testing (recomputed on resize/solve).
	originX, originY int
	originValid      bool

	// Debug: last mouse event info.
	lastMouseX, lastMouseY int
	lastMouseBtn           string
	lastMouseGridCol       int
	lastMouseGridRow       int
	lastMouseHit           bool
}

func buildFoundCells(width, height int, words []Word) [][]bool {
	fc := make([][]bool, height)
	for y := range fc {
		fc[y] = make([]bool, width)
	}
	for i := range words {
		if words[i].Found {
			for _, pos := range words[i].Positions() {
				fc[pos.Y][pos.X] = true
			}
		}
	}
	return fc
}

func New(mode WordSearchMode, g grid, words []Word) (game.Gamer, error) {
	return &Model{
		width:      mode.Width,
		height:     mode.Height,
		grid:       g,
		words:      words,
		selection:  noSelection,
		keys:       DefaultKeyMap,
		modeTitle:  mode.Title(),
		solved:     false,
		foundCells: buildFoundCells(mode.Width, mode.Height, words),
	}, nil
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
		return m.handleKeyPress(msg)

	case tea.MouseClickMsg:
		m.handleMouseClick(msg)

	case tea.MouseMotionMsg:
		m.handleMouseMotion(msg)

	case tea.MouseReleaseMsg:
		m.handleMouseRelease()
	}
	return m, nil
}

func (m *Model) handleMouseClick(msg tea.MouseClickMsg) {
	m.lastMouseX, m.lastMouseY = msg.X, msg.Y
	m.lastMouseBtn = msg.String()

	col, row, ok := m.screenToGrid(msg.X, msg.Y)
	m.lastMouseGridCol, m.lastMouseGridRow = col, row
	m.lastMouseHit = ok

	if m.solved || !ok {
		return
	}

	switch msg.Button {
	case tea.MouseLeft:
		// Click sets the selection start and begins a drag.
		m.cursor.X, m.cursor.Y = col, row
		m.selectionStart = m.cursor
		m.selection = startSelected
		m.mouseDragging = true

	case tea.MouseRight:
		// Right-click cancels the current selection.
		m.selection = noSelection
		m.mouseDragging = false
	}
}

func (m *Model) handleMouseMotion(msg tea.MouseMotionMsg) {
	if m.solved || !m.mouseDragging || m.selection != startSelected {
		return
	}

	col, row, ok := m.screenToGrid(msg.X, msg.Y)
	if !ok {
		return
	}

	// Move cursor to track the drag endpoint; the existing selection
	// rendering uses selectionStart → cursor.
	m.cursor.X, m.cursor.Y = col, row
}

func (m *Model) handleMouseRelease() {
	if !m.mouseDragging {
		return
	}
	m.mouseDragging = false

	if m.selection != startSelected {
		return
	}

	// Same cell as start: keep selection active (start is set, waiting
	// for a second click/drag to define the end — mirrors keyboard flow).
	if m.cursor.X == m.selectionStart.X && m.cursor.Y == m.selectionStart.Y {
		return
	}

	// Try to validate and submit the selection.
	m.validateSelection()
	m.selection = noSelection
}

func (m Model) handleKeyPress(msg tea.KeyPressMsg) (game.Gamer, tea.Cmd) {
	switch {
	case key.Matches(msg, m.keys.Select):
		m.handleSelect()
	case key.Matches(msg, m.keys.Cancel):
		m.selection = noSelection
	default:
		m.cursor.Move(m.keys.CursorKeyMap, msg, m.width-1, m.height-1)
	}

	return m, nil
}

func (m *Model) handleSelect() {
	if m.selection == noSelection {
		// Mark selection start
		m.selectionStart = m.cursor
		m.selection = startSelected
	} else {
		// Try to validate selection
		m.validateSelection()
		m.selection = noSelection
	}
}

func (m *Model) validateSelection() {
	var letters strings.Builder
	valid := walkLine(m.selectionStart, m.cursor, func(x, y int) {
		letters.WriteRune(m.grid.Get(x, y))
	})
	if !valid {
		return
	}

	word := letters.String()
	wordReverse := reverseString(word)

	for i := range m.words {
		if m.words[i].Found {
			continue
		}
		if m.words[i].Text == word || m.words[i].Text == wordReverse {
			m.words[i].Found = true
			for _, pos := range m.words[i].Positions() {
				m.foundCells[pos.Y][pos.X] = true
			}
			m.checkWin()
			return
		}
	}
}

func (m *Model) checkWin() {
	allFound := true
	for _, word := range m.words {
		if !word.Found {
			allFound = false
			break
		}
	}
	if allFound != m.solved {
		m.originValid = false
	}
	m.solved = allFound
}

func (m Model) View() string {
	return renderView(m)
}

func (m Model) GetDebugInfo() string {
	ox, oy := m.gridOrigin()
	hitStr := "miss"
	if m.lastMouseHit {
		hitStr = fmt.Sprintf("(%d, %d)", m.lastMouseGridCol, m.lastMouseGridRow)
	}

	rows := [][2]string{
		{"Grid Size", fmt.Sprintf("%dx%d", m.width, m.height)},
		{"Cursor", fmt.Sprintf("(%d, %d)", m.cursor.X, m.cursor.Y)},
		{"Selection State", fmt.Sprintf("%v", m.selection)},
	}
	if m.selection == startSelected {
		rows = append(rows, [2]string{"Selection Start", fmt.Sprintf("(%d, %d)", m.selectionStart.X, m.selectionStart.Y)})
	}
	rows = append(rows,
		[2]string{"Words Found", fmt.Sprintf("%d/%d", m.countFoundWords(), len(m.words))},
		[2]string{"Won", fmt.Sprintf("%v", m.solved)},
		[2]string{"Term Size", fmt.Sprintf("%d x %d", m.termWidth, m.termHeight)},
		[2]string{"Grid Origin", fmt.Sprintf("(%d, %d)", ox, oy)},
		[2]string{"Last Mouse", fmt.Sprintf("screen=(%d, %d) btn=%s grid=%s", m.lastMouseX, m.lastMouseY, m.lastMouseBtn, hitStr)},
		[2]string{"Mouse Drag", fmt.Sprintf("%v", m.mouseDragging)},
	)

	s := game.DebugHeader("Word Search", rows)

	var tableRows [][]string
	for _, word := range m.words {
		found := "No"
		if word.Found {
			found = "Yes"
		}
		tableRows = append(tableRows, []string{
			word.Text, found,
			fmt.Sprintf("(%d,%d)", word.Start.X, word.Start.Y),
			fmt.Sprintf("(%d,%d)", word.End.X, word.End.Y),
			fmt.Sprintf("%d", word.Direction),
		})
	}
	s += game.DebugTable("Words", []string{"Word", "Found", "Start", "End", "Direction"}, tableRows)

	return s
}

func (m Model) SetTitle(t string) game.Gamer {
	m.modeTitle = t
	return m
}

func (m Model) IsSolved() bool {
	return m.solved
}

func (m Model) Reset() game.Gamer {
	words := make([]Word, len(m.words))
	for i, w := range m.words {
		words[i] = Word{
			Text:      w.Text,
			Start:     w.Start,
			End:       w.End,
			Direction: w.Direction,
			Found:     false,
		}
	}
	m.words = words
	m.foundCells = buildFoundCells(m.width, m.height, m.words)
	m.solved = false
	m.selection = noSelection
	m.selectionStart = game.Cursor{}
	m.cursor = game.Cursor{}
	m.mouseDragging = false
	m.originValid = false
	return m
}

func (m Model) countFoundWords() int {
	count := 0
	for _, word := range m.words {
		if word.Found {
			count++
		}
	}
	return count
}

// lineDirection returns the unit direction vector from start to end
// and whether the line is valid (horizontal, vertical, or diagonal).
func lineDirection(start, end game.Cursor) (dx, dy int, valid bool) {
	if start.X == end.X && start.Y == end.Y {
		return 0, 0, false
	}

	if end.X > start.X {
		dx = 1
	} else if end.X < start.X {
		dx = -1
	}

	if end.Y > start.Y {
		dy = 1
	} else if end.Y < start.Y {
		dy = -1
	}

	distX := abs(end.X - start.X)
	distY := abs(end.Y - start.Y)
	if dx != 0 && dy != 0 && distX != distY {
		return 0, 0, false
	}

	return dx, dy, true
}

// walkLine calls fn for each cell on the line from start to end.
// Returns false if the line is not valid.
func walkLine(start, end game.Cursor, fn func(x, y int)) bool {
	dx, dy, valid := lineDirection(start, end)
	if !valid {
		return false
	}

	x, y := start.X, start.Y
	for {
		fn(x, y)
		if x == end.X && y == end.Y {
			break
		}
		x += dx
		y += dy
	}
	return true
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

func reverseString(s string) string {
	runes := []rune(s)
	for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
		runes[i], runes[j] = runes[j], runes[i]
	}
	return string(runes)
}
