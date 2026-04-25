package takuzu

import (
	"image/color"
	"strconv"
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/FelineStateMachine/puzzletea/game"
	"github.com/FelineStateMachine/puzzletea/theme"
)

const cellWidth = game.DynamicGridCellWidth

var renderRuneMap = map[rune]string{
	zeroCell:  " ● ",
	oneCell:   " ○ ",
	emptyCell: " · ",
}

func zeroStyle() lipgloss.Style {
	p := theme.Current()
	return lipgloss.NewStyle().Foreground(p.Accent).Background(p.BG)
}

func oneStyle() lipgloss.Style {
	p := theme.Current()
	return lipgloss.NewStyle().Foreground(p.Secondary).Background(p.BG)
}

func emptyStyle() lipgloss.Style {
	p := theme.Current()
	return lipgloss.NewStyle().Foreground(p.TextDim).Background(p.BG)
}

func renderStyleMap() map[rune]lipgloss.Style {
	return map[rune]lipgloss.Style{
		zeroCell:  zeroStyle(),
		oneCell:   oneStyle(),
		emptyCell: emptyStyle(),
	}
}

type countPair struct {
	zeros int
	ones  int
}

type countContext struct {
	row    countPair
	col    countPair
	target int
}

func cellView(val rune, isProvided, isCursor, inCursorRow, inCursorCol, solved, inDuplicateRow, inDuplicateCol bool) string {
	p := theme.Current()
	styles := renderStyleMap()
	style, ok := styles[val]
	if !ok {
		style = emptyStyle()
	}

	text, ok := renderRuneMap[val]
	if !ok {
		text = renderRuneMap[emptyCell]
	}

	switch {
	case solved:
		style = style.Foreground(p.SolvedFG).Background(p.SuccessBG)
	}

	if isProvided && val != emptyCell && !solved {
		style = style.Bold(true).Background(theme.GivenTint(p.BG))
	}
	if (inDuplicateRow || inDuplicateCol) && !solved {
		style = style.Foreground(game.ConflictFG()).Background(game.ConflictBG())
		text = conflictText(text)
	}
	if isCursor {
		text = cursorText(text)
	}

	return style.Width(cellWidth).AlignHorizontal(lipgloss.Center).Render(text)
}

func cursorText(text string) string {
	runes := []rune(text)
	if len(runes) != cellWidth {
		return text
	}

	switch {
	case runes[0] == ' ' && runes[cellWidth-1] == ' ':
		return game.CursorLeft + string(runes[1]) + game.CursorRight
	case runes[cellWidth-1] == ' ':
		return game.CursorLeft + string(runes[:cellWidth-1])
	case runes[0] == ' ':
		return string(runes[1:]) + game.CursorRight
	default:
		return text
	}
}

func conflictText(text string) string {
	runes := []rune(text)
	if len(runes) != cellWidth {
		return text
	}
	return "!" + string(runes[1]) + "!"
}

func lineComplete(row []rune) bool {
	for _, r := range row {
		if r == emptyCell {
			return false
		}
	}
	return true
}

func colComplete(g grid, size, col int) bool {
	for y := range size {
		if col >= len(g[y]) || g[y][col] == emptyCell {
			return false
		}
	}
	return true
}

func duplicateRowSet(g grid, size int) map[int]bool {
	dup := map[int]bool{}
	for i := range size {
		if !lineComplete(g[i]) {
			continue
		}
		for j := i + 1; j < size; j++ {
			if lineComplete(g[j]) && rowEqual(g[i], g[j]) {
				dup[i] = true
				dup[j] = true
			}
		}
	}
	return dup
}

func duplicateColSet(g grid, size int) map[int]bool {
	dup := map[int]bool{}
	for i := range size {
		if !colComplete(g, size, i) {
			continue
		}
		for j := i + 1; j < size; j++ {
			if colComplete(g, size, j) && colEqual(g, size, i, j) {
				dup[i] = true
				dup[j] = true
			}
		}
	}
	return dup
}

func gridView(m Model) string {
	dupRows := duplicateRowSet(m.grid, m.size)
	dupCols := duplicateColSet(m.grid, m.size)
	return game.RenderDynamicGrid(game.DynamicGridSpec{
		Width:  m.size,
		Height: m.size,
		Solved: m.solved,
		Cell: func(x, y int) string {
			return cellView(
				m.grid[y][x],
				m.provided[y][x],
				x == m.cursor.X && y == m.cursor.Y,
				y == m.cursor.Y,
				x == m.cursor.X,
				m.solved,
				dupRows[y],
				dupCols[x],
			)
		},
		ZoneAt: func(_, _ int) int {
			return 0
		},
		BridgeFill: func(bridge game.DynamicGridBridge) color.Color {
			return bridgeFill(m, bridge)
		},
	})
}

func bridgeFill(_ Model, _ game.DynamicGridBridge) color.Color {
	return nil
}

func countContextView(m Model) string {
	ctx := buildCountContext(m.grid, m.cursor, m.size)
	width := countValueWidth(ctx.target)
	label := lipgloss.NewStyle().Foreground(theme.Current().Info)

	var b strings.Builder
	b.WriteString(label.Render("row  "))
	b.WriteString(countLabelStyle(zeroCell).Render(string([]rune(renderRuneMap[zeroCell])[1]) + ":"))
	b.WriteString(renderCountValue(ctx.row.zeros, ctx.target, zeroCell, width))
	b.WriteString(label.Render("  "))
	b.WriteString(countLabelStyle(oneCell).Render(string([]rune(renderRuneMap[oneCell])[1]) + ":"))
	b.WriteString(renderCountValue(ctx.row.ones, ctx.target, oneCell, width))
	b.WriteString(label.Render("  col  "))
	b.WriteString(countLabelStyle(zeroCell).Render(string([]rune(renderRuneMap[zeroCell])[1]) + ":"))
	b.WriteString(renderCountValue(ctx.col.zeros, ctx.target, zeroCell, width))
	b.WriteString(label.Render("  "))
	b.WriteString(countLabelStyle(oneCell).Render(string([]rune(renderRuneMap[oneCell])[1]) + ":"))
	b.WriteString(renderCountValue(ctx.col.ones, ctx.target, oneCell, width))
	return b.String()
}

func buildCountContext(g grid, cursor game.Cursor, size int) countContext {
	if size <= 0 || len(g) == 0 {
		return countContext{}
	}
	row := 0
	if cursor.Y >= 0 && cursor.Y < len(g) {
		row = cursor.Y
	}
	col := 0
	if cursor.X >= 0 && len(g[row]) > 0 && cursor.X < len(g[row]) {
		col = cursor.X
	}
	return countContext{
		row:    countLine(g[row]),
		col:    countColumn(g, col, size),
		target: size / 2,
	}
}

func countLine(row []rune) countPair {
	var counts countPair
	for _, value := range row {
		switch value {
		case zeroCell:
			counts.zeros++
		case oneCell:
			counts.ones++
		}
	}
	return counts
}

func countColumn(g grid, col, size int) countPair {
	var counts countPair
	for y := 0; y < size && y < len(g); y++ {
		if col < 0 || col >= len(g[y]) {
			continue
		}
		switch g[y][col] {
		case zeroCell:
			counts.zeros++
		case oneCell:
			counts.ones++
		}
	}
	return counts
}

func countPairString(counts countPair) string {
	return strconv.Itoa(counts.zeros) + "/" + strconv.Itoa(counts.ones)
}

func renderCountValue(count, target int, value rune, width int) string {
	text := strconv.Itoa(count) + "/" + strconv.Itoa(target)
	style := countValueStyle(value)
	if count == target {
		style = metGoalCountStyle(value)
	}
	if count > target {
		style = overGoalCountStyle()
	}
	return style.Width(width).AlignHorizontal(lipgloss.Right).Render(text)
}

func countValueWidth(target int) int {
	return len(strconv.Itoa(target))*2 + 1
}

func countValueStyle(value rune) lipgloss.Style {
	return lipgloss.NewStyle().Foreground(binaryValueColor(value))
}

func countLabelStyle(value rune) lipgloss.Style {
	return lipgloss.NewStyle().
		Bold(true).
		Foreground(binaryValueColor(value))
}

func metGoalCountStyle(value rune) lipgloss.Style {
	bg := binaryValueColor(value)
	return lipgloss.NewStyle().
		Bold(true).
		Foreground(theme.TextOnBG(bg)).
		Background(bg)
}

func overGoalCountStyle() lipgloss.Style {
	bg := game.ConflictBG()
	return lipgloss.NewStyle().
		Bold(true).
		Foreground(theme.TextOnBG(bg)).
		Background(bg)
}

func binaryValueColor(value rune) color.Color {
	p := theme.Current()
	switch value {
	case zeroCell:
		return p.Accent
	case oneCell:
		return p.Secondary
	default:
		return p.TextDim
	}
}

func statusBarView(showFullHelp bool) string {
	if showFullHelp {
		return game.StatusBarStyle().Render("arrows/wasd: move  mouse: click/cycle  z/0: ●  x/1: ○  bkspc: clear  esc: menu  ctrl+r: reset  ctrl+h: help")
	}
	return game.StatusBarStyle().Render("mouse: click/cycle  z/0: ●  x/1: ○")
}
