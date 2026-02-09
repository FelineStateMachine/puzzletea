package sudoku

import (
	"fmt"
	"slices"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

var (
	emptyCellStyle = lipgloss.NewStyle().
			Foreground(lipgloss.AdaptiveColor{Light: "250", Dark: "240"}).
			Background(lipgloss.AdaptiveColor{Light: "254", Dark: "235"})

	providedCellStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.AdaptiveColor{Light: "130", Dark: "179"}).
				Background(lipgloss.AdaptiveColor{Light: "254", Dark: "235"})

	userCellStyle = lipgloss.NewStyle().
			Foreground(lipgloss.AdaptiveColor{Light: "236", Dark: "187"}).
			Background(lipgloss.AdaptiveColor{Light: "254", Dark: "235"})

	cursorCellStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.AdaptiveColor{Light: "255", Dark: "235"}).
			Background(lipgloss.AdaptiveColor{Light: "130", Dark: "173"})

	conflictCellStyle = lipgloss.NewStyle().
				Foreground(lipgloss.AdaptiveColor{Light: "160", Dark: "167"}).
				Background(lipgloss.AdaptiveColor{Light: "224", Dark: "52"})

	crosshairBG = lipgloss.AdaptiveColor{Light: "254", Dark: "237"}

	boxBorderFG = lipgloss.AdaptiveColor{Light: "250", Dark: "240"}

	gridBorderStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.AdaptiveColor{Light: "250", Dark: "240"})

	gridBorderSolvedStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.AdaptiveColor{Light: "22", Dark: "149"})

	statusBarStyle = lipgloss.NewStyle().
			Foreground(lipgloss.AdaptiveColor{Light: "137", Dark: "137"}).
			MarginTop(1)

	sudokuCellWidth = 2
)

func renderGrid(m Model) string {
	var rows []string

	for y := range gridSize {
		var cells []string
		for x := range gridSize {
			c := m.grid[y][x]
			style := getCellStyle(m, c, x, y)
			content := cellContent(c)
			rendered := style.Width(sudokuCellWidth).Align(lipgloss.Center).Render(content)
			cells = append(cells, rendered)

			// Insert vertical box separator after columns 3 and 6
			if x == 2 || x == 5 {
				sep := lipgloss.NewStyle().Foreground(boxBorderFG).Render("│")
				cells = append(cells, sep)
			}
		}
		row := lipgloss.JoinHorizontal(lipgloss.Top, cells...)
		rows = append(rows, row)

		// Insert horizontal box separator after rows 3 and 6
		if y == 2 || y == 5 {
			sepLine := strings.Repeat("─", sudokuCellWidth)
			var sepParts []string
			for x := range gridSize {
				sepParts = append(sepParts, sepLine)
				if x == 2 || x == 5 {
					sepParts = append(sepParts, "┼")
				}
			}
			sep := lipgloss.NewStyle().Foreground(boxBorderFG).Render(strings.Join(sepParts, ""))
			rows = append(rows, sep)
		}
	}

	grid := lipgloss.JoinVertical(lipgloss.Left, rows...)

	if m.isSolved() {
		return gridBorderSolvedStyle.Render(grid)
	}
	return gridBorderStyle.Render(grid)
}

func getCellStyle(m Model, c cell, x, y int) lipgloss.Style {
	// Priority: cursor > conflict > provided > user > empty
	if m.cursor.X == x && m.cursor.Y == y {
		return cursorCellStyle
	}

	if c.v != 0 && hasConflict(m, c, x, y) {
		return conflictCellStyle
	}

	isProvided := slices.Contains(m.provided, c)
	inCursorRow := m.cursor.Y == y
	inCursorCol := m.cursor.X == x
	inCursorBox := (x/3 == m.cursor.X/3) && (y/3 == m.cursor.Y/3)
	inCrosshair := inCursorRow || inCursorCol || inCursorBox

	if isProvided {
		s := providedCellStyle
		if inCrosshair {
			s = s.Background(crosshairBG)
		}
		return s
	}

	if c.v != 0 {
		s := userCellStyle
		if inCrosshair {
			s = s.Background(crosshairBG)
		}
		return s
	}

	s := emptyCellStyle
	if inCrosshair {
		s = s.Background(crosshairBG)
	}
	return s
}

func cellContent(c cell) string {
	if c.v == 0 {
		return "·"
	}
	return fmt.Sprintf("%d", c.v)
}

func hasConflict(m Model, c cell, x, y int) bool {
	if c.v == 0 {
		return false
	}

	// Check row
	for cx := range gridSize {
		if cx != x && m.grid[y][cx].v == c.v {
			return true
		}
	}

	// Check column
	for cy := range gridSize {
		if cy != y && m.grid[cy][x].v == c.v {
			return true
		}
	}

	// Check 3x3 box
	boxStartX := (x / 3) * 3
	boxStartY := (y / 3) * 3
	for by := boxStartY; by < boxStartY+3; by++ {
		for bx := boxStartX; bx < boxStartX+3; bx++ {
			if (bx != x || by != y) && m.grid[by][bx].v == c.v {
				return true
			}
		}
	}

	return false
}

func statusBarView(showFullHelp bool) string {
	if showFullHelp {
		return statusBarStyle.Render("arrows/wasd: move  1-9: fill  bkspc: clear  ctrl+n: menu  ctrl+h: help")
	}
	return statusBarStyle.Render("1-9: fill  bkspc: clear")
}
