package wordsearch

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

var (
	normalStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#1a1a1a")).
			Background(lipgloss.Color("#ffffff"))

	cursorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#ffffff")).
			Background(lipgloss.Color("#ff00ff"))

	selectionStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#1a1a1a")).
			Background(lipgloss.Color("#ffff00"))

	foundStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#ffffff")).
			Background(lipgloss.Color("#00ff00"))

	wordListHeaderStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#00ffff")).
				Bold(true).
				Underline(true)

	foundWordStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#00ff00")).
			Strikethrough(true)

	unfoundWordStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#888888"))

	winMessageStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#00ff00")).
			Bold(true).
			Padding(1, 2).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#00ff00"))
)

func renderView(m Model) string {
	gridView := renderGrid(m)
	wordListView := renderWordList(m)

	// Join grid and word list horizontally with spacing
	spacer := strings.Repeat(" ", 4)
	mainView := lipgloss.JoinHorizontal(lipgloss.Top, gridView, spacer, wordListView)

	// Add win message if won
	if m.won {
		winMsg := winMessageStyle.Render("🎉 Congratulations! You found all the words! 🎉")
		return lipgloss.JoinVertical(lipgloss.Left, mainView, "\n", winMsg)
	}

	return mainView
}

func renderGrid(m Model) string {
	var sb strings.Builder

	for y := 0; y < m.height; y++ {
		for x := 0; x < m.width; x++ {
			letter := m.grid.Get(x, y)
			style := getCellStyle(m, x, y)
			sb.WriteString(style.Render(fmt.Sprintf(" %c ", letter)))
		}
		sb.WriteString("\n")
	}

	return sb.String()
}

func getCellStyle(m Model, x, y int) lipgloss.Style {
	// Priority order: cursor > selection > found word > normal

	// Check if cursor position
	if m.cursor.x == x && m.cursor.y == y {
		return cursorStyle
	}

	// Check if in current selection
	if m.selection == startSelected && isInSelection(m, x, y) {
		return selectionStyle
	}

	// Check if part of found word
	for _, word := range m.words {
		if word.Found && word.Contains(Position{X: x, Y: y}) {
			return foundStyle
		}
	}

	return normalStyle
}

func isInSelection(m Model, x, y int) bool {
	// Check if (x, y) is on the line from selectionStart to cursor
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

	// Verify it's a valid straight line before walking
	distX := abs(end.x - start.x)
	distY := abs(end.y - start.y)
	if dx != 0 && dy != 0 && distX != distY {
		return false // Not a valid diagonal, skip to avoid infinite loop
	}

	// Check all positions along the line
	cx, cy := start.x, start.y
	for {
		if cx == x && cy == y {
			return true
		}

		if cx == end.x && cy == end.y {
			break
		}

		cx += dx
		cy += dy
	}

	return false
}

func renderWordList(m Model) string {
	var sb strings.Builder

	sb.WriteString(wordListHeaderStyle.Render("Words to Find:"))
	sb.WriteString("\n\n")

	for _, word := range m.words {
		if word.Found {
			sb.WriteString(foundWordStyle.Render(fmt.Sprintf("✓ %s", word.Text)))
		} else {
			sb.WriteString(unfoundWordStyle.Render(fmt.Sprintf("○ %s", word.Text)))
		}
		sb.WriteString("\n")
	}

	sb.WriteString("\n")
	sb.WriteString(fmt.Sprintf("Found: %d/%d", m.countFoundWords(), len(m.words)))

	return sb.String()
}
