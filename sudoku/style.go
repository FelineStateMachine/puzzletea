package sudoku

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

var (
	backgroundColor = lipgloss.AdaptiveColor{Light: "254", Dark: "235"}

	emptyCellStyle = lipgloss.NewStyle().
			Foreground(lipgloss.AdaptiveColor{Light: "250", Dark: "240"}).
			Background(backgroundColor)

	providedCellStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.AdaptiveColor{Light: "130", Dark: "179"}).
				Background(backgroundColor)

	userCellStyle = lipgloss.NewStyle().
			Foreground(lipgloss.AdaptiveColor{Light: "236", Dark: "187"}).
			Background(backgroundColor)

	cursorCellStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.AdaptiveColor{Light: "255", Dark: "235"}).
			Background(lipgloss.AdaptiveColor{Light: "130", Dark: "173"})

	conflictCellStyle = lipgloss.NewStyle().
				Foreground(lipgloss.AdaptiveColor{Light: "160", Dark: "167"}).
				Background(lipgloss.AdaptiveColor{Light: "224", Dark: "52"})

	sameNumberStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.AdaptiveColor{Light: "130", Dark: "187"})

	crosshairBG = lipgloss.AdaptiveColor{Light: "254", Dark: "237"}

	boxBorderFG = lipgloss.AdaptiveColor{Light: "250", Dark: "240"}

	gridBorderStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.AdaptiveColor{Light: "250", Dark: "240"}).
			BorderBackground(backgroundColor)

	gridBorderSolvedStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.AdaptiveColor{Light: "22", Dark: "149"}).
				BorderBackground(backgroundColor)

	statusBarStyle = lipgloss.NewStyle().
			Foreground(lipgloss.AdaptiveColor{Light: "137", Dark: "137"}).
			MarginTop(1)

	sudokuCellWidth = 2
)

func renderGrid(m Model, solved bool, conflicts [gridSize][gridSize]bool) string {
	var rows []string

	for y := range gridSize {
		var cells []string
		for x := range gridSize {
			c := m.grid[y][x]
			style := getCellStyle(m, c, x, y, conflicts[y][x])
			content := cellContent(c)
			rendered := style.Width(sudokuCellWidth).Align(lipgloss.Center).Render(content)
			cells = append(cells, rendered)

			// Insert vertical box separator after columns 3 and 6
			if x == 2 || x == 5 {
				bg := backgroundColor
				if m.cursor.Y == y {
					bg = crosshairBG
				}
				sep := lipgloss.NewStyle().Foreground(boxBorderFG).Background(bg).Render("│")
				cells = append(cells, sep)
			}
		}
		row := lipgloss.JoinHorizontal(lipgloss.Top, cells...)
		rows = append(rows, row)

		// Insert horizontal box separator after rows 3 and 6
		if y == 2 || y == 5 {
			sepLine := strings.Repeat("─", sudokuCellWidth)
			var renderedParts []string
			for x := range gridSize {
				bg := backgroundColor
				if m.cursor.X == x {
					bg = crosshairBG
				}
				renderedParts = append(renderedParts, lipgloss.NewStyle().Foreground(boxBorderFG).Background(bg).Render(sepLine))
				if x == 2 || x == 5 {
					renderedParts = append(renderedParts, lipgloss.NewStyle().Foreground(boxBorderFG).Background(backgroundColor).Render("┼"))
				}
			}
			rows = append(rows, lipgloss.JoinHorizontal(lipgloss.Top, renderedParts...))
		}
	}

	grid := lipgloss.JoinVertical(lipgloss.Left, rows...)

	if solved {
		return gridBorderSolvedStyle.Render(grid)
	}
	return gridBorderStyle.Render(grid)
}

func getCellStyle(m Model, c cell, x, y int, conflict bool) lipgloss.Style {
	cursorVal := m.grid[m.cursor.Y][m.cursor.X].v

	// Priority: cursor > conflict > same number > provided > user > empty
	if m.cursor.X == x && m.cursor.Y == y {
		return cursorCellStyle
	}

	if conflict {
		return conflictCellStyle
	}

	if cursorVal != 0 && c.v == cursorVal {
		return sameNumberStyle
	}

	isProvided := m.providedGrid[y][x]
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

// computeConflicts returns a grid of booleans indicating which cells have conflicts.
// A cell has a conflict if its value appears more than once in its row, column, or 3x3 box.
func computeConflicts(g grid) [gridSize][gridSize]bool {
	var conflicts [gridSize][gridSize]bool

	// Check rows
	for y := range gridSize {
		var seen [10][]int // value → list of x positions
		for x := range gridSize {
			v := g[y][x].v
			if v != 0 {
				seen[v] = append(seen[v], x)
			}
		}
		for v := 1; v <= 9; v++ {
			if len(seen[v]) > 1 {
				for _, x := range seen[v] {
					conflicts[y][x] = true
				}
			}
		}
	}

	// Check columns
	for x := range gridSize {
		var seen [10][]int
		for y := range gridSize {
			v := g[y][x].v
			if v != 0 {
				seen[v] = append(seen[v], y)
			}
		}
		for v := 1; v <= 9; v++ {
			if len(seen[v]) > 1 {
				for _, y := range seen[v] {
					conflicts[y][x] = true
				}
			}
		}
	}

	// Check 3x3 boxes
	for boxY := range 3 {
		for boxX := range 3 {
			type pos struct{ y, x int }
			var seen [10][]pos
			for dy := range 3 {
				for dx := range 3 {
					y, x := boxY*3+dy, boxX*3+dx
					v := g[y][x].v
					if v != 0 {
						seen[v] = append(seen[v], pos{y, x})
					}
				}
			}
			for v := 1; v <= 9; v++ {
				if len(seen[v]) > 1 {
					for _, p := range seen[v] {
						conflicts[p.y][p.x] = true
					}
				}
			}
		}
	}

	return conflicts
}

func statusBarView(showFullHelp bool) string {
	if showFullHelp {
		return statusBarStyle.Render("arrows/wasd: move  1-9: fill  bkspc: clear  ctrl+n: menu  ctrl+h: help")
	}
	return statusBarStyle.Render("1-9: fill  bkspc: clear")
}
