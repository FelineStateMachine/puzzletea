package nurikabe

import (
	"fmt"

	"charm.land/lipgloss/v2"
	"github.com/FelineStateMachine/puzzletea/game"
	"github.com/FelineStateMachine/puzzletea/theme"
)

const cellWidth = 3

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
		BorderBackground(p.SuccessBG)
}

func cellStyle(c cellState, clue int, isCursor, solved, conflict, inSeaSquare bool) (lipgloss.Style, string) {
	p := theme.Current()
	display := "   "
	landBg := theme.Blend(p.BG, p.Success, 0.45)
	style := lipgloss.NewStyle().Background(landBg).Foreground(theme.TextOnBG(landBg))

	switch {
	case clue > 0:
		display = fmt.Sprintf("%2d ", clue)
		if clue < 10 {
			display = fmt.Sprintf(" %d ", clue)
		}

		style = lipgloss.NewStyle().
			Foreground(p.Info).
			Background(landBg).
			Bold(true)
	case c == seaCell:
		display = " ~ "
		if inSeaSquare {
			display = " @ "
		}
		seaBg := theme.Blend(p.BG, p.Secondary, 0.24)
		style = style.Background(seaBg).Foreground(theme.TextOnBG(seaBg))
	case c == islandCell:
		display = " \u00b7 "
		style = style.Background(landBg).Foreground(theme.TextOnBG(landBg))
	}

	if solved {
		style = style.Background(p.SuccessBG).Foreground(theme.TextOnBG(p.SuccessBG))
	}
	if conflict {
		style = style.Foreground(game.ConflictFG()).Background(game.ConflictBG())
	}
	if isCursor && solved {
		style = game.CursorSolvedStyle()
	} else if isCursor {
		style = game.CursorStyle()
	}

	return style.Width(cellWidth).AlignHorizontal(lipgloss.Center), display
}

func gridView(m Model) string {
	rows := make([]string, m.height)
	for y := range m.height {
		cells := make([]string, m.width)
		for x := range m.width {
			isCursor := x == m.cursor.X && y == m.cursor.Y
			inSeaSquare := isSeaSquareCell(m.marks, x, y)
			style, display := cellStyle(m.marks[y][x], m.clues[y][x], isCursor, m.solved, m.conflicts[y][x], inSeaSquare)
			cells[x] = style.Render(display)
		}
		rows[y] = lipgloss.JoinHorizontal(lipgloss.Top, cells...)
	}

	content := lipgloss.JoinVertical(lipgloss.Left, rows...)
	if m.solved {
		return gridBorderSolvedStyle().Render(content)
	}
	return gridBorderStyle().Render(content)
}

// isSeaSquareCell reports whether (x,y) is part of any 2x2 sea block.
func isSeaSquareCell(marks grid, x, y int) bool {
	if len(marks) == 0 || len(marks[0]) == 0 {
		return false
	}
	if y < 0 || y >= len(marks) || x < 0 || x >= len(marks[0]) {
		return false
	}
	if marks[y][x] != seaCell {
		return false
	}

	startX := max(0, x-1)
	endX := min(len(marks[0])-2, x)
	startY := max(0, y-1)
	endY := min(len(marks)-2, y)

	for yy := startY; yy <= endY; yy++ {
		for xx := startX; xx <= endX; xx++ {
			if marks[yy][xx] == seaCell &&
				marks[yy][xx+1] == seaCell &&
				marks[yy+1][xx] == seaCell &&
				marks[yy+1][xx+1] == seaCell {
				return true
			}
		}
	}

	return false
}

func statusBarView(showFullHelp bool) string {
	if showFullHelp {
		return game.StatusBarStyle().Render("arrows/wasd: move x/LMB: sea  z/RMB: island  bkspc: clear  ctrl+n: menu  ctrl+r: reset  ctrl+h: help")
	}
	return game.StatusBarStyle().Render("x/LMB: sea  z/RMB: island  bkspc: clear")
}
