package sudoku

import (
	"fmt"
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/FelineStateMachine/puzzletea/game"
	"github.com/FelineStateMachine/puzzletea/theme"
)

var sudokuCellWidth = 2

func emptyCellStyle() lipgloss.Style {
	p := theme.Current()
	return lipgloss.NewStyle().
		Foreground(p.TextDim).
		Background(p.BG)
}

func providedCellStyle() lipgloss.Style {
	p := theme.Current()
	return lipgloss.NewStyle().
		Bold(true).
		Foreground(p.Given).
		Background(p.BG)
}

func userCellStyle() lipgloss.Style {
	p := theme.Current()
	return lipgloss.NewStyle().
		Foreground(p.AccentSoft).
		Background(p.BG)
}

func conflictCellStyle() lipgloss.Style {
	p := theme.Current()
	return lipgloss.NewStyle().
		Foreground(p.Error).
		Background(p.ErrorBG)
}

func sameNumberStyle() lipgloss.Style {
	p := theme.Current()
	return lipgloss.NewStyle().
		Bold(true).
		Foreground(p.Accent)
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

func renderGrid(m Model, solved bool, conflicts [gridSize][gridSize]bool) string {
	p := theme.Current()

	var rows []string

	for y := range gridSize {
		var cells []string
		for x := range gridSize {
			c := m.grid[y][x]
			style := getCellStyle(m, c, x, y, conflicts[y][x], solved)
			content := cellContent(c)
			rendered := style.Width(sudokuCellWidth).Align(lipgloss.Center).Render(content)
			cells = append(cells, rendered)

			if x == 2 || x == 5 {
				bg := p.BG
				if m.cursor.Y == y {
					bg = p.Surface
				}
				sep := lipgloss.NewStyle().Foreground(p.Border).Background(bg).Render("\u2502")
				cells = append(cells, sep)
			}
		}
		row := lipgloss.JoinHorizontal(lipgloss.Top, cells...)
		rows = append(rows, row)

		if y == 2 || y == 5 {
			sepLine := strings.Repeat("\u2500", sudokuCellWidth)
			var renderedParts []string
			for x := range gridSize {
				bg := p.BG
				if m.cursor.X == x {
					bg = p.Surface
				}
				renderedParts = append(renderedParts, lipgloss.NewStyle().Foreground(p.Border).Background(bg).Render(sepLine))
				if x == 2 || x == 5 {
					renderedParts = append(renderedParts, lipgloss.NewStyle().Foreground(p.Border).Background(p.BG).Render("\u253c"))
				}
			}
			rows = append(rows, lipgloss.JoinHorizontal(lipgloss.Top, renderedParts...))
		}
	}

	grid := lipgloss.JoinVertical(lipgloss.Left, rows...)

	if solved {
		return gridBorderSolvedStyle().Render(grid)
	}
	return gridBorderStyle().Render(grid)
}

func getCellStyle(m Model, c cell, x, y int, conflict, solved bool) lipgloss.Style {
	p := theme.Current()
	isCursor := m.cursor.X == x && m.cursor.Y == y
	cursorVal := m.grid[m.cursor.Y][m.cursor.X].v

	// Priority: cursor+solved > cursor > conflict > same number > provided > user > empty
	if isCursor && solved {
		return game.CursorSolvedStyle()
	}

	if isCursor {
		return game.CursorStyle()
	}

	if solved {
		return lipgloss.NewStyle().Foreground(p.SolvedFG).Background(p.SuccessBG)
	}

	if conflict {
		return conflictCellStyle()
	}

	if cursorVal != 0 && c.v == cursorVal {
		return sameNumberStyle()
	}

	isProvided := m.providedGrid[y][x]
	inCursorRow := m.cursor.Y == y
	inCursorCol := m.cursor.X == x
	inCursorBox := (x/3 == m.cursor.X/3) && (y/3 == m.cursor.Y/3)
	inCrosshair := inCursorRow || inCursorCol || inCursorBox

	if isProvided {
		s := providedCellStyle()
		if inCrosshair {
			s = s.Background(p.Surface)
		}
		return s
	}

	if c.v != 0 {
		s := userCellStyle()
		if inCrosshair {
			s = s.Background(p.Surface)
		}
		return s
	}

	s := emptyCellStyle()
	if inCrosshair {
		s = s.Background(p.Surface)
	}
	return s
}

func cellContent(c cell) string {
	if c.v == 0 {
		return "\u00b7"
	}
	return fmt.Sprintf("%d", c.v)
}

// computeConflicts returns a grid of booleans indicating which cells have conflicts.
// A cell has a conflict if its value appears more than once in its row, column, or 3x3 box.
func computeConflicts(g grid) [gridSize][gridSize]bool {
	var conflicts [gridSize][gridSize]bool

	for y := range gridSize {
		var seen [10][]int // value -> list of x positions
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
		return game.StatusBarStyle().Render("arrows/wasd: move  1-9: fill  bkspc: clear  ctrl+n: menu  ctrl+r: reset  ctrl+h: help")
	}
	return game.StatusBarStyle().Render("1-9: fill  bkspc: clear")
}
