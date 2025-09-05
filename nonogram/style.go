package nonogram

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

var (
	baseStyle = lipgloss.NewStyle()

	filledStyle = baseStyle.
			Background(lipgloss.Color("255")).
			Foreground(lipgloss.Color("255"))

	markedStyle = baseStyle.
			Foreground(lipgloss.Color("245"))

	emptyStyle = baseStyle.
			Foreground(lipgloss.Color("250"))

	highlightStyle = baseStyle.
			Background(lipgloss.Color("205")).
			Foreground(lipgloss.Color("255"))

	cellWidth = 3

	hintCellStyle = baseStyle.Width(cellWidth)

	renderStyleMap = map[rune]lipgloss.Style{
		filledTile: filledStyle,
		markedTile: markedStyle,
		emptyTile:  emptyStyle,
	}

	renderRuneMap = map[rune]string{
		filledTile: "⬤",
		markedTile: "⊗",
		emptyTile:  "◯",
	}
)

func colHintView(c TomographyDefinition, height int) string {
	var renderedCols []string
	for _, hints := range c {
		var colHints []string
		for range height - len(hints) {
			spacer := hintCellStyle.Render(" ")
			colHints = append(colHints, spacer)
		}
		for _, hint := range hints {
			hintCell := hintCellStyle.
				Align(lipgloss.Center).
				Render(fmt.Sprintf("%d", hint))
			colHints = append(colHints, hintCell)
		}
		renderedCol := lipgloss.JoinVertical(lipgloss.Left, colHints...)
		renderedCols = append(renderedCols, renderedCol)
	}
	return lipgloss.JoinHorizontal(lipgloss.Top, renderedCols...)
}
func rowHintView(r TomographyDefinition, width int) string {
	var renderedRows []string
	for _, hints := range r {
		var rowHints []string
		for _, hint := range hints {
			hintCell := fmt.Sprintf("%2d", hint)
			rowHints = append(rowHints, hintCell)
		}
		renderedRow := baseStyle.Width(width).
			Align(lipgloss.Right).
			Render(strings.Join(rowHints, " "))
		renderedRows = append(renderedRows, renderedRow)
	}
	s := lipgloss.JoinVertical(lipgloss.Left, renderedRows...)
	return s
}

func gridView(g grid, c cursor) string {
	var rows []string
	for y, row := range g {
		var rowBuider []string
		for x, cell := range row {
			highlighted := x == c.x && y == c.y
			cell := tileView(cell, highlighted)
			rowBuider = append(rowBuider, cell)

		}
		row := lipgloss.JoinHorizontal(lipgloss.Top, rowBuider...)
		rows = append(rows, row)
	}
	grid := lipgloss.JoinVertical(lipgloss.Left, rows...)
	return grid
}

func tileView(val rune, h bool) string {
	s, ok := renderStyleMap[val]
	if !ok {
		s = renderStyleMap[emptyTile]
	}
	if h {
		s = highlightStyle
	}
	r, ok := renderRuneMap[val]
	if !ok {
		r = renderRuneMap[emptyTile]
	}
	return s.Width(cellWidth).AlignHorizontal(lipgloss.Center).Render(r)
}
