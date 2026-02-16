package wordsearch

import (
	"fmt"
	"strings"

	"charm.land/lipgloss/v2"
	"charm.land/lipgloss/v2/compat"
	"github.com/FelineStateMachine/puzzletea/game"
)

var (
	normalStyle = lipgloss.NewStyle().
			Foreground(compat.AdaptiveColor{Light: lipgloss.Color("236"), Dark: lipgloss.Color("252")}).
			Background(compat.AdaptiveColor{Light: lipgloss.Color("254"), Dark: lipgloss.Color("236")})

	cursorStyle = game.CursorStyle

	selectionStyle = lipgloss.NewStyle().
			Foreground(compat.AdaptiveColor{Light: lipgloss.Color("235"), Dark: lipgloss.Color("235")}).
			Background(compat.AdaptiveColor{Light: lipgloss.Color("172"), Dark: lipgloss.Color("180")})

	foundStyle = lipgloss.NewStyle().
			Foreground(compat.AdaptiveColor{Light: lipgloss.Color("22"), Dark: lipgloss.Color("151")}).
			Background(compat.AdaptiveColor{Light: lipgloss.Color("194"), Dark: lipgloss.Color("237")})

	wordListHeaderStyle = lipgloss.NewStyle().
				Foreground(compat.AdaptiveColor{Light: lipgloss.Color("130"), Dark: lipgloss.Color("173")}).
				Bold(true).
				Underline(true)

	foundWordStyle = lipgloss.NewStyle().
			Foreground(compat.AdaptiveColor{Light: lipgloss.Color("28"), Dark: lipgloss.Color("115")}).
			Strikethrough(true)

	unfoundWordStyle = lipgloss.NewStyle().
				Foreground(compat.AdaptiveColor{Light: lipgloss.Color("137"), Dark: lipgloss.Color("137")})

	borderFG    = compat.AdaptiveColor{Light: lipgloss.Color("250"), Dark: lipgloss.Color("240")}
	gridBG      = compat.AdaptiveColor{Light: lipgloss.Color("254"), Dark: lipgloss.Color("236")}
	solvedBdrFG = compat.AdaptiveColor{Light: lipgloss.Color("22"), Dark: lipgloss.Color("149")}

	gridBorderStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(borderFG).
			BorderBackground(gridBG)

	gridBorderSolvedStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(solvedBdrFG).
				BorderBackground(gridBG)

	wordListBorderStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(borderFG).
				Padding(0, 1)

	wordListBorderSolvedStyle = lipgloss.NewStyle().
					Border(lipgloss.RoundedBorder()).
					BorderForeground(solvedBdrFG).
					Padding(0, 1)
)

func renderView(m Model) string {
	title := game.TitleBarView("Word Search", m.modeTitle, m.solved)
	gridContent := renderGrid(m)
	wordListContent := renderWordList(m)

	gBorder := gridBorderStyle
	wBorder := wordListBorderStyle
	if m.solved {
		gBorder = gridBorderSolvedStyle
		wBorder = wordListBorderSolvedStyle
	}

	gridView := gBorder.Render(gridContent)
	gridHeight := lipgloss.Height(gridView)
	// Match word list height to grid height, subtracting border lines (top+bottom).
	wBorder = wBorder.Height(gridHeight - 2)
	wordListView := wBorder.Render(wordListContent)

	spacer := strings.Repeat(" ", 2)
	mainView := lipgloss.JoinHorizontal(lipgloss.Top, gridView, spacer, wordListView)

	status := statusBarView(m.showFullHelp)

	return lipgloss.JoinVertical(lipgloss.Center, title, mainView, status)
}

func statusBarView(showFullHelp bool) string {
	if showFullHelp {
		return game.StatusBarStyle.Render("arrows/wasd: move  enter/space: select  esc: cancel  mouse: click & drag  ctrl+n: menu  ctrl+r: reset  ctrl+h: help")
	}
	return game.StatusBarStyle.Render("enter/space: select  esc: cancel  mouse: click & drag")
}

func renderGrid(m Model) string {
	var rows []string

	for y := 0; y < m.height; y++ {
		var cells []string
		for x := 0; x < m.width; x++ {
			letter := m.grid.Get(x, y)
			style := getCellStyle(m, x, y)
			cells = append(cells, style.Render(fmt.Sprintf(" %c ", letter)))
		}
		rows = append(rows, lipgloss.JoinHorizontal(lipgloss.Top, cells...))
	}

	return lipgloss.JoinVertical(lipgloss.Left, rows...)
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
	if m.foundCells[y][x] {
		return foundStyle
	}

	return normalStyle
}

func isInSelection(m Model, x, y int) bool {
	found := false
	walkLine(m.selectionStart, m.cursor, func(cx, cy int) {
		if cx == x && cy == y {
			found = true
		}
	})
	return found
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
	fmt.Fprintf(&sb, "Found: %d/%d", m.countFoundWords(), len(m.words))

	return sb.String()
}
