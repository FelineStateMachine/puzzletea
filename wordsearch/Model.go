package wordsearch

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/FelineStateMachine/puzzletea/game"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
)

type selectionState int

const (
	noSelection selectionState = iota
	startSelected
)

type cursor struct {
	x, y int
}

type Model struct {
	width, height  int
	grid           grid
	words          []Word
	cursor         cursor
	selection      selectionState
	selectionStart cursor
	keys           KeyMap
	won            bool
}

// New creates a new word search game
func New(mode WordSearchMode, g grid, words []Word) *Model {
	return &Model{
		width:     mode.Width,
		height:    mode.Height,
		grid:      g,
		words:     words,
		cursor:    cursor{x: 0, y: 0},
		selection: noSelection,
		keys:      DefaultKeyMap,
		won:       false,
	}
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (game.Gamer, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		return m.handleKeyPress(msg)
	}
	return m, nil
}

func (m Model) handleKeyPress(msg tea.KeyMsg) (game.Gamer, tea.Cmd) {
	switch {
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
	case key.Matches(msg, m.keys.Select):
		m.handleSelect()
	case key.Matches(msg, m.keys.Cancel):
		m.selection = noSelection
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
	start := m.selectionStart
	end := m.cursor

	// Calculate direction
	dx := 0
	dy := 0

	if end.x > start.x {
		dx = 1
	} else if end.x < start.x {
		dx = -1
	}

	if end.y > start.y {
		dy = 1
	} else if end.y < start.y {
		dy = -1
	}

	// Must be a valid line (horizontal, vertical, or diagonal)
	if dx == 0 && dy == 0 {
		return // Same position
	}

	// Verify it's a straight line
	distX := abs(end.x - start.x)
	distY := abs(end.y - start.y)

	if dx != 0 && dy != 0 && distX != distY {
		return // Not a valid diagonal
	}

	// Extract letters along the selection
	letters := m.extractLetters(start, end, dx, dy)
	lettersReverse := reverseString(letters)

	// Check against all unfound words
	for i := range m.words {
		if m.words[i].Found {
			continue
		}

		if m.words[i].Text == letters || m.words[i].Text == lettersReverse {
			m.words[i].Found = true
			m.checkWin()
			return
		}
	}
}

func (m *Model) extractLetters(start, end cursor, dx, dy int) string {
	var letters strings.Builder
	x, y := start.x, start.y

	for {
		letters.WriteRune(m.grid.Get(x, y))

		if x == end.x && y == end.y {
			break
		}

		x += dx
		y += dy
	}

	return letters.String()
}

func (m *Model) checkWin() {
	allFound := true
	for _, word := range m.words {
		if !word.Found {
			allFound = false
			break
		}
	}
	m.won = allFound
}

func (m Model) View() string {
	return renderView(m)
}

func (m Model) GetDebugInfo() string {
	var sb strings.Builder

	sb.WriteString("# Word Search Debug\n\n")
	sb.WriteString(fmt.Sprintf("**Grid Size:** %dx%d\n\n", m.width, m.height))
	sb.WriteString(fmt.Sprintf("**Cursor:** (%d, %d)\n\n", m.cursor.x, m.cursor.y))
	sb.WriteString(fmt.Sprintf("**Selection State:** %v\n\n", m.selection))

	if m.selection == startSelected {
		sb.WriteString(fmt.Sprintf("**Selection Start:** (%d, %d)\n\n", m.selectionStart.x, m.selectionStart.y))
	}

	sb.WriteString(fmt.Sprintf("**Words Found:** %d/%d\n\n", m.countFoundWords(), len(m.words)))
	sb.WriteString(fmt.Sprintf("**Won:** %v\n\n", m.won))

	sb.WriteString("## Words\n\n")
	sb.WriteString("| Word | Found | Start | End | Direction |\n")
	sb.WriteString("|------|-------|-------|-----|----------|\n")

	for _, word := range m.words {
		found := "❌"
		if word.Found {
			found = "✓"
		}
		sb.WriteString(fmt.Sprintf("| %s | %s | (%d,%d) | (%d,%d) | %d |\n",
			word.Text, found, word.Start.X, word.Start.Y, word.End.X, word.End.Y, word.Direction))
	}

	return sb.String()
}

func (m Model) GetFullHelp() [][]key.Binding {
	return m.keys.FullHelp()
}

func (m Model) GetSave() ([]byte, error) {
	data := Save{
		Width:      m.width,
		Height:     m.height,
		Grid:       m.grid.String(),
		Words:      m.words,
		CursorX:    m.cursor.x,
		CursorY:    m.cursor.y,
		Selection:  int(m.selection),
		SelectionX: m.selectionStart.x,
		SelectionY: m.selectionStart.y,
		Won:        m.won,
	}
	return json.Marshal(data)
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
