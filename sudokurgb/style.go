package sudokurgb

import (
	"image/color"
	"strconv"
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/FelineStateMachine/puzzletea/game"
	"github.com/FelineStateMachine/puzzletea/theme"
)

const (
	cellWidth     = game.DynamicGridCellWidth
	rowHintWidth  = valueCount*2 - 1
	colHintHeight = valueCount
)

func emptyCellStyle() lipgloss.Style {
	p := theme.Current()
	return lipgloss.NewStyle().
		Foreground(p.TextDim).
		Background(game.DefaultBorderColors().BackgroundBG)
}

func providedCellStyle(value int, bg color.Color) lipgloss.Style {
	return lipgloss.NewStyle().
		Bold(true).
		Foreground(symbolColor(value)).
		Background(theme.GivenTint(bg))
}

func userCellStyle(value int) lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(symbolColor(value))
}

func boxConflictCellStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(theme.TextOnBG(game.ConflictBG())).
		Background(game.ConflictBG())
}

func valueCursorStyle(value int) lipgloss.Style {
	if value == 0 {
		return game.CursorStyle()
	}

	bg := symbolColor(value)
	return lipgloss.NewStyle().
		Bold(true).
		Foreground(theme.TextOnBG(bg)).
		Background(bg)
}

func hintStyle(value int) lipgloss.Style {
	return lipgloss.NewStyle().Foreground(symbolColor(value))
}

func hintSolvedStyle(value int) lipgloss.Style {
	bg := symbolColor(value)
	return lipgloss.NewStyle().
		Bold(true).
		Foreground(theme.TextOnBG(bg)).
		Background(bg)
}

func hintErrorStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Bold(true).
		Foreground(theme.TextOnBG(game.ConflictBG())).
		Background(game.ConflictBG())
}

func solvedBoardStyle() lipgloss.Style {
	p := theme.Current()
	return lipgloss.NewStyle().
		Foreground(p.SolvedFG).
		Background(p.SuccessBG)
}

type boardBlockLayout struct {
	Block      string
	Grid       string
	HintWidth  int
	HintHeight int
}

func buildBoardBlock(m Model, solved bool) boardBlockLayout {
	grid := renderGrid(m, solved)
	rowHints := rowHintView(m.analysis, solved)
	colHints := colHintView(m.analysis, solved)
	spacerStyle := lipgloss.NewStyle()
	if solved {
		spacerStyle = solvedBoardStyle()
	}
	spacer := spacerStyle.Width(lipgloss.Width(rowHints)).Height(colHintHeight).Render("")
	topBand := lipgloss.JoinHorizontal(lipgloss.Top, spacer, colHints)
	block := lipgloss.JoinVertical(
		lipgloss.Left,
		topBand,
		lipgloss.JoinHorizontal(lipgloss.Top, rowHints, grid),
	)

	return boardBlockLayout{
		Block:      block,
		Grid:       grid,
		HintWidth:  lipgloss.Width(rowHints),
		HintHeight: lipgloss.Height(topBand),
	}
}

func renderGrid(m Model, solved bool) string {
	return game.RenderDynamicGrid(game.DynamicGridSpec{
		Width:  gridSize,
		Height: gridSize,
		Solved: solved,
		Cell: func(x, y int) string {
			return cellView(m, x, y, solved)
		},
		ZoneAt: func(x, y int) int {
			return sudokuBoxIndex(x, y)
		},
	})
}

func cellView(m Model, x, y int, solved bool) string {
	c := m.grid[y][x]
	style := cellStyle(m, c, x, y, solved)
	text := cellContent(c)
	if x == m.cursor.X && y == m.cursor.Y {
		if c.v == 0 {
			text = game.CursorLeft + "·" + game.CursorRight
		} else {
			text = game.CursorLeft + cellContentValue(c.v) + game.CursorRight
		}
	}

	return style.Width(cellWidth).AlignHorizontal(lipgloss.Center).Render(text)
}

func cellStyle(m Model, c cell, x, y int, solved bool) lipgloss.Style {
	isCursor := m.cursor.X == x && m.cursor.Y == y
	var base lipgloss.Style

	switch c.v {
	case 0:
		base = emptyCellStyle()
	default:
		base = userCellStyle(c.v)
	}

	if m.providedGrid[y][x] {
		base = providedCellStyle(c.v, game.DefaultBorderColors().BackgroundBG)
	}

	switch {
	case isCursor:
		return valueCursorStyle(c.v)
	case m.analysis.boxConflictCells[y][x]:
		base = base.
			Foreground(theme.TextOnBG(game.ConflictBG())).
			Background(game.ConflictBG())
	case solved:
		base = base.
			Bold(true).
			Foreground(theme.Current().SolvedFG).
			Background(theme.Current().SuccessBG)
	}

	return base
}

func sudokuBoxIndex(x, y int) int {
	return (y / 3 * 3) + (x / 3)
}

func cellContent(c cell) string {
	if c.v == 0 {
		return "·"
	}
	return cellContentValue(c.v)
}

func cellContentValue(value int) string {
	switch value {
	case 1:
		return "▲"
	case 2:
		return "■"
	case 3:
		return "●"
	default:
		return "·"
	}
}

func symbolColor(value int) color.Color {
	p := theme.Current()
	switch value {
	case 1:
		return p.Error
	case 2:
		return p.Success
	case 3:
		return p.Secondary
	default:
		return p.TextDim
	}
}

func rowHintView(analysis boardAnalysis, solved bool) string {
	lines := make([]string, 0, gridSize*2+1)
	blankStyle := lipgloss.NewStyle()
	if solved {
		blankStyle = solvedBoardStyle()
	}
	blank := blankStyle.Width(rowHintWidth).Render("")
	lines = append(lines, blank)
	for row := range gridSize {
		lines = append(lines, renderRowHintCounts(analysis, row, solved))
		lines = append(lines, blank)
	}
	return strings.Join(lines, "\n")
}

func renderRowHintCounts(analysis boardAnalysis, row int, solved bool) string {
	var b strings.Builder
	spacerStyle := lipgloss.NewStyle()
	if solved {
		spacerStyle = solvedBoardStyle()
	}

	for value := 1; value <= valueCount; value++ {
		if value > 1 {
			b.WriteString(spacerStyle.Render(" "))
		}
		b.WriteString(renderHintCount(analysis.rowCounts[row][value], value, analysis.rowOverQuota[row][value], solved))
	}

	rowStyle := lipgloss.NewStyle()
	if solved {
		rowStyle = solvedBoardStyle()
	}
	return rowStyle.Width(rowHintWidth).Render(b.String())
}

func colHintView(analysis boardAnalysis, solved bool) string {
	lines := make([]string, 0, colHintHeight)
	for value := 1; value <= valueCount; value++ {
		var b strings.Builder
		lineStyle := lipgloss.NewStyle()
		if solved {
			lineStyle = solvedBoardStyle()
		}
		b.WriteString(lineStyle.Render(" "))
		for col := range gridSize {
			cellStyle := lipgloss.NewStyle().
				Width(cellWidth).
				AlignHorizontal(lipgloss.Center)
			if solved {
				cellStyle = cellStyle.Inherit(solvedBoardStyle())
			}
			b.WriteString(
				cellStyle.Render(renderHintCount(analysis.colCounts[col][value], value, analysis.colOverQuota[col][value], solved)),
			)
			if col < gridSize-1 {
				b.WriteString(lineStyle.Render(" "))
			}
		}
		b.WriteString(lineStyle.Render(" "))
		lines = append(lines, b.String())
	}
	return strings.Join(lines, "\n")
}

func renderHintCount(count, value int, overQuota, solved bool) string {
	if solved {
		return solvedBoardStyle().Render(strconv.Itoa(count))
	}

	style := hintStyle(value)
	if count == houseQuota {
		style = hintSolvedStyle(value)
	}
	if overQuota {
		style = hintErrorStyle()
	}
	return style.Render(strconv.Itoa(count))
}

func statusBarView(showFullHelp bool) string {
	if showFullHelp {
		return game.StatusBarStyle().Render("mouse: click focus  arrows/wasd: move  1/2/3: ▲■●  bkspc: clear  esc: menu  ctrl+r: reset  ctrl+h: help")
	}
	return game.StatusBarStyle().Render("mouse: click focus  1/2/3: ▲■●  bkspc: clear")
}
