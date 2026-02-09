package wordsearch

import (
	"fmt"
	"strings"

	"github.com/FelineStateMachine/puzzletea/game"
	"github.com/charmbracelet/lipgloss"
)

var (
	normalStyle = lipgloss.NewStyle().
			Foreground(lipgloss.AdaptiveColor{Light: "236", Dark: "252"}).
			Background(lipgloss.AdaptiveColor{Light: "254", Dark: "236"})

	cursorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.AdaptiveColor{Light: "255", Dark: "235"}).
			Background(lipgloss.AdaptiveColor{Light: "130", Dark: "173"})

	selectionStyle = lipgloss.NewStyle().
			Foreground(lipgloss.AdaptiveColor{Light: "235", Dark: "235"}).
			Background(lipgloss.AdaptiveColor{Light: "172", Dark: "180"})

	foundStyle = lipgloss.NewStyle().
			Foreground(lipgloss.AdaptiveColor{Light: "22", Dark: "151"}).
			Background(lipgloss.AdaptiveColor{Light: "194", Dark: "237"})

	wordListHeaderStyle = lipgloss.NewStyle().
				Foreground(lipgloss.AdaptiveColor{Light: "130", Dark: "173"}).
				Bold(true).
				Underline(true)

	foundWordStyle = lipgloss.NewStyle().
			Foreground(lipgloss.AdaptiveColor{Light: "28", Dark: "115"}).
			Strikethrough(true)

	unfoundWordStyle = lipgloss.NewStyle().
				Foreground(lipgloss.AdaptiveColor{Light: "137", Dark: "137"})

	statusBarStyle = lipgloss.NewStyle().
			Foreground(lipgloss.AdaptiveColor{Light: "137", Dark: "137"}).
			MarginTop(1)
)

func renderView(m Model) string {
	title := game.TitleBarView("Word Search", m.modeTitle, m.solved)
	gridView := renderGrid(m)
	wordListView := renderWordList(m)

	// Join grid and word list horizontally with spacing
	spacer := strings.Repeat(" ", 4)
	mainView := lipgloss.JoinHorizontal(lipgloss.Top, gridView, spacer, wordListView)

	status := statusBarView(m.showFullHelp)

	return lipgloss.JoinVertical(lipgloss.Left, title, mainView, status)
}

func statusBarView(showFullHelp bool) string {
	if showFullHelp {
		return statusBarStyle.Render("arrows/wasd: move  enter/space: select  esc: cancel  ctrl+n: menu  ctrl+h: help")
	}
	return statusBarStyle.Render("enter/space: select  esc: cancel")
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
	if m.cursor.X == x && m.cursor.Y == y {
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

	// Verify it's a valid straight line before walking
	distX := abs(end.X - start.X)
	distY := abs(end.Y - start.Y)
	if dx != 0 && dy != 0 && distX != distY {
		return false // Not a valid diagonal, skip to avoid infinite loop
	}

	// Check all positions along the line
	cx, cy := start.X, start.Y
	for {
		if cx == x && cy == y {
			return true
		}

		if cx == end.X && cy == end.Y {
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
