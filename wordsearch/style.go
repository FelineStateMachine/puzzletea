package wordsearch

import (
	"fmt"
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/FelineStateMachine/puzzletea/game"
	"github.com/FelineStateMachine/puzzletea/theme"
)

func normalStyle() lipgloss.Style {
	p := theme.Current()
	return lipgloss.NewStyle().
		Foreground(p.FG).
		Background(p.BG)
}

func foundStyle() lipgloss.Style {
	p := theme.Current()
	return lipgloss.NewStyle().
		Foreground(p.SolvedFG).
		Background(p.SuccessBG)
}

func wordListHeaderStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(theme.Current().Accent).
		Bold(true).
		Underline(true)
}

func foundWordStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(theme.Current().SuccessBorder).
		Strikethrough(true)
}

func unfoundWordStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(theme.Current().Info)
}

func gridBorderStyle() lipgloss.Style {
	p := theme.Current()
	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(p.Border).
		BorderBackground(p.BG)
}

func gridBorderSolvedStyle() lipgloss.Style {
	p := theme.Current()
	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(p.SuccessBorder).
		BorderBackground(p.BG)
}

func wordListBorderStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(theme.Current().Border).
		Padding(0, 1)
}

func wordListBorderSolvedStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(theme.Current().SuccessBorder).
		Padding(0, 1)
}

// selectionStyle returns the style for cells in the active drag selection,
// using the current theme's SelectionBG color for the background.
func selectionStyle() lipgloss.Style {
	bg := theme.Current().SelectionBG
	return lipgloss.NewStyle().
		Foreground(theme.TextOnBG(bg)).
		Background(bg)
}

func renderView(m Model) string {
	title := game.TitleBarView("Word Search", m.modeTitle, m.solved)
	gridContent := renderGrid(m)
	wordListContent := renderWordList(m)

	gBorder := gridBorderStyle()
	wBorder := wordListBorderStyle()
	if m.solved {
		gBorder = gridBorderSolvedStyle()
		wBorder = wordListBorderSolvedStyle()
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
		return game.StatusBarStyle().Render("arrows/wasd: move  enter/space: select  esc: cancel  mouse: click & drag  ctrl+n: menu  ctrl+r: reset  ctrl+h: help")
	}
	return game.StatusBarStyle().Render("enter/space: select  esc: cancel  mouse: click & drag")
}

func renderGrid(m Model) string {
	var rows []string

	for y := 0; y < m.height; y++ {
		var cells []string
		for x := 0; x < m.width; x++ {
			letter := m.grid.Get(x, y)
			style := getCellStyle(m, x, y)
			display := fmt.Sprintf(" %c ", letter)
			if m.cursor.X == x && m.cursor.Y == y {
				display = game.CursorLeft + fmt.Sprintf("%c", letter) + game.CursorRight
			}
			cells = append(cells, style.Render(display))
		}
		rows = append(rows, lipgloss.JoinHorizontal(lipgloss.Top, cells...))
	}

	return lipgloss.JoinVertical(lipgloss.Left, rows...)
}

func getCellStyle(m Model, x, y int) lipgloss.Style {
	// Priority order: cursor > selection > found word > normal

	if m.cursor.X == x && m.cursor.Y == y {
		return game.CursorStyle()
	}

	if m.selection == startSelected && isInSelection(m, x, y) {
		return selectionStyle()
	}

	if m.foundCells[y][x] {
		return foundStyle()
	}

	return normalStyle()
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

	sb.WriteString(wordListHeaderStyle().Render("Words to Find:"))
	sb.WriteString("\n\n")

	for _, word := range m.words {
		if word.Found {
			sb.WriteString(foundWordStyle().Render(fmt.Sprintf("\u2713 %s", word.Text)))
		} else {
			sb.WriteString(unfoundWordStyle().Render(fmt.Sprintf("\u25cb %s", word.Text)))
		}
		sb.WriteString("\n")
	}

	sb.WriteString("\n")
	fmt.Fprintf(&sb, "Found: %d/%d", m.countFoundWords(), len(m.words))

	return sb.String()
}
