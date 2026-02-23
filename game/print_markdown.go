package game

import (
	"fmt"
	"strconv"
	"strings"
)

func SplitLines(s string) []string {
	if strings.TrimSpace(s) == "" {
		return []string{}
	}
	return strings.Split(s, "\n")
}

func EscapeMarkdownCell(s string) string {
	return strings.ReplaceAll(s, "|", "\\|")
}

func MakeStringGrid(width, height int, fill string) [][]string {
	if width <= 0 || height <= 0 {
		return [][]string{}
	}

	grid := make([][]string, height)
	for y := range height {
		grid[y] = make([]string, width)
		for x := range width {
			grid[y][x] = fill
		}
	}
	return grid
}

func RenderGridTable(cells [][]string) string {
	if len(cells) == 0 {
		return "_(empty grid)_"
	}

	width := 0
	for _, row := range cells {
		if len(row) > width {
			width = len(row)
		}
	}
	if width <= 0 {
		return "_(empty grid)_"
	}

	var b strings.Builder
	b.WriteString("|   |")
	for x := range width {
		fmt.Fprintf(&b, " %d |", x+1)
	}
	b.WriteString("\n| --- |")
	for range width {
		b.WriteString(" --- |")
	}
	b.WriteString("\n")

	for y := range len(cells) {
		fmt.Fprintf(&b, "| %d |", y+1)
		for x := range width {
			cell := "."
			if x < len(cells[y]) && strings.TrimSpace(cells[y][x]) != "" {
				cell = cells[y][x]
			}
			fmt.Fprintf(&b, " %s |", EscapeMarkdownCell(cell))
		}
		b.WriteString("\n")
	}

	return strings.TrimRight(b.String(), "\n")
}

func ParseCSVIntGrid(raw string, width, height int) [][]int {
	if width <= 0 || height <= 0 {
		return [][]int{}
	}

	clues := make([][]int, height)
	for y := range height {
		clues[y] = make([]int, width)
	}

	rows := SplitLines(raw)
	for y := range min(len(rows), len(clues)) {
		parts := strings.Split(rows[y], ",")
		for x := range min(len(parts), len(clues[y])) {
			val, err := strconv.Atoi(strings.TrimSpace(parts[x]))
			if err != nil {
				continue
			}
			clues[y][x] = val
		}
	}
	return clues
}

func NormalizeNonogramHints(src [][]int, size int) [][]int {
	if size <= 0 {
		return [][]int{}
	}

	hints := make([][]int, size)
	for i := range len(hints) {
		if i >= len(src) || len(src[i]) == 0 {
			hints[i] = []int{0}
			continue
		}
		hints[i] = append([]int(nil), src[i]...)
	}
	return hints
}

func MaxNonogramHintLen(hints [][]int) int {
	maxLen := 0
	for _, hint := range hints {
		if len(hint) > maxLen {
			maxLen = len(hint)
		}
	}
	return maxLen
}

func RenderNonogramTable(
	rowHints, colHints [][]int,
	width, height, rowHintCols, colHintRows int,
) string {
	var b strings.Builder

	b.WriteString("|")
	for i := range rowHintCols {
		fmt.Fprintf(&b, " R%d |", i+1)
	}
	for x := range width {
		fmt.Fprintf(&b, " C%d |", x+1)
	}
	b.WriteString("\n|")
	for range rowHintCols + width {
		b.WriteString(" --- |")
	}
	b.WriteString("\n")

	for hintRow := range colHintRows {
		b.WriteString("|")
		for range rowHintCols {
			b.WriteString(" . |")
		}
		for x := range width {
			b.WriteString(" ")
			b.WriteString(renderColumnHintCell(colHints[x], colHintRows, hintRow))
			b.WriteString(" |")
		}
		b.WriteString("\n")
	}

	for y := range height {
		rowHint := rowHints[y]
		hintStart := rowHintCols - len(rowHint)

		b.WriteString("|")
		for hintCol := range rowHintCols {
			if hintCol < hintStart {
				b.WriteString(" . |")
				continue
			}
			fmt.Fprintf(&b, " %d |", rowHint[hintCol-hintStart])
		}
		for range width {
			b.WriteString(" . |")
		}
		b.WriteString("\n")
	}

	return strings.TrimRight(b.String(), "\n")
}

func renderColumnHintCell(hint []int, depth, row int) string {
	start := depth - len(hint)
	if row < start {
		return "."
	}
	return strconv.Itoa(hint[row-start])
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
