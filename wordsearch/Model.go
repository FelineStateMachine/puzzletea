package wordsearch

import (
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
}

// New creates a new word search game
func New(mode WordSearchMode, g grid, words []Word) (game.Gamer, error) {
	return &Model{
		width:     mode.Width,
		height:    mode.Height,
		grid:      g,
		words:     words,
		selection: noSelection,
		keys:      DefaultKeyMap,
		modeTitle: mode.Title(),
		solved:    false,
	}, nil
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (game.Gamer, tea.Cmd) {
	switch msg := msg.(type) {
	case game.HelpToggleMsg:
		m.showFullHelp = msg.Show
	case tea.KeyMsg:
		return m.handleKeyPress(msg)
	}
	return m, nil
}

func (m Model) handleKeyPress(msg tea.KeyMsg) (game.Gamer, tea.Cmd) {
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
	start := m.selectionStart
	end := m.cursor

	// Calculate direction
	dx := 0
	dy := 0

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

	// Must be a valid line (horizontal, vertical, or diagonal)
	if dx == 0 && dy == 0 {
		return // Same position
	}

	// Verify it's a straight line
	distX := abs(end.X - start.X)
	distY := abs(end.Y - start.Y)

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

func (m *Model) extractLetters(start, end game.Cursor, dx, dy int) string {
	var letters strings.Builder
	x, y := start.X, start.Y

	for {
		letters.WriteRune(m.grid.Get(x, y))

		if x == end.X && y == end.Y {
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
	m.solved = allFound
}

func (m Model) View() string {
	return renderView(m)
}

func (m Model) GetDebugInfo() string {
	var sb strings.Builder

	sb.WriteString("# Word Search Debug\n\n")
	sb.WriteString(fmt.Sprintf("**Grid Size:** %dx%d\n\n", m.width, m.height))
	sb.WriteString(fmt.Sprintf("**Cursor:** (%d, %d)\n\n", m.cursor.X, m.cursor.Y))
	sb.WriteString(fmt.Sprintf("**Selection State:** %v\n\n", m.selection))

	if m.selection == startSelected {
		sb.WriteString(fmt.Sprintf("**Selection Start:** (%d, %d)\n\n", m.selectionStart.X, m.selectionStart.Y))
	}

	sb.WriteString(fmt.Sprintf("**Words Found:** %d/%d\n\n", m.countFoundWords(), len(m.words)))
	sb.WriteString(fmt.Sprintf("**Won:** %v\n\n", m.solved))

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

func (m Model) SetTitle(t string) game.Gamer {
	m.modeTitle = t
	return m
}

func (m Model) IsSolved() bool {
	return m.solved
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
