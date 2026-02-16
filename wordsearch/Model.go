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

// New creates a new word search game
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
	case tea.KeyPressMsg:
		return m.handleKeyPress(msg)
	}
	return m, nil
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
	m.solved = allFound
}

func (m Model) View() string {
	return renderView(m)
}

func (m Model) GetDebugInfo() string {
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
