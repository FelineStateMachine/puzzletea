package markdownexport

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/FelineStateMachine/puzzletea/hashiwokakero"
	"github.com/FelineStateMachine/puzzletea/hitori"
	"github.com/FelineStateMachine/puzzletea/nonogram"
	"github.com/FelineStateMachine/puzzletea/nurikabe"
	"github.com/FelineStateMachine/puzzletea/shikaku"
	"github.com/FelineStateMachine/puzzletea/sudoku"
	"github.com/FelineStateMachine/puzzletea/takuzu"
	"github.com/FelineStateMachine/puzzletea/wordsearch"
)

func RenderPuzzleSnippet(gameType, _ string, save []byte) (string, error) {
	switch normalizeGameType(gameType) {
	case "hashiwokakero":
		return renderHashi(save)
	case "hitori":
		return renderHitori(save)
	case "nonogram":
		return renderNonogram(save)
	case "nurikabe":
		return renderNurikabe(save)
	case "shikaku":
		return renderShikaku(save)
	case "sudoku":
		return renderSudoku(save)
	case "takuzu":
		return renderTakuzu(save)
	case "word search", "wordsearch":
		return renderWordSearch(save)
	case "lights out", "lightsout":
		return "", ErrUnsupportedGame
	default:
		return "", ErrUnsupportedGame
	}
}

func renderHashi(data []byte) (string, error) {
	var save hashiwokakero.Save
	if err := json.Unmarshal(data, &save); err != nil {
		return "", fmt.Errorf("decode hashiwokakero save: %w", err)
	}

	cells := makeGrid(save.Width, save.Height, ".")
	for _, island := range save.Islands {
		if island.Y >= 0 && island.Y < len(cells) && island.X >= 0 && island.X < len(cells[island.Y]) {
			cells[island.Y][island.X] = strconv.Itoa(island.Required)
		}
	}

	var b strings.Builder
	b.WriteString("### Grid\n\n")
	b.WriteString(renderGridTable(cells))
	b.WriteString("\n\n")
	b.WriteString("Rules: connect numbered islands with horizontal/vertical bridges. ")
	b.WriteString("Use up to two bridges per connection and never cross bridges.")
	return b.String(), nil
}

func renderHitori(data []byte) (string, error) {
	var save hitori.Save
	if err := json.Unmarshal(data, &save); err != nil {
		return "", fmt.Errorf("decode hitori save: %w", err)
	}

	rows := splitLines(save.Numbers)
	cells := make([][]string, 0, max(len(rows), save.Size))
	targetHeight := max(len(rows), save.Size)
	for y := range targetHeight {
		row := []rune{}
		if y < len(rows) {
			row = []rune(rows[y])
		}
		cellsRow := make([]string, max(len(row), save.Size))
		if len(cellsRow) == 0 {
			cellsRow = []string{"."}
		}
		for x := range len(cellsRow) {
			if x < len(row) {
				cellsRow[x] = string(row[x])
			} else {
				cellsRow[x] = "."
			}
		}
		cells = append(cells, cellsRow)
	}
	if len(cells) == 0 {
		cells = [][]string{{"."}}
	}

	var b strings.Builder
	b.WriteString("### Grid\n\n")
	b.WriteString(renderGridTable(cells))
	b.WriteString("\n\n")
	b.WriteString("Goal: shade cells so no row or column has duplicate unshaded values, ")
	b.WriteString("shaded cells do not touch orthogonally, and all unshaded cells stay connected.")
	return b.String(), nil
}

func renderNonogram(data []byte) (string, error) {
	var save nonogram.Save
	if err := json.Unmarshal(data, &save); err != nil {
		return "", fmt.Errorf("decode nonogram save: %w", err)
	}

	width := save.Width
	height := save.Height
	if width <= 0 {
		width = len(save.ColHints)
	}
	if height <= 0 {
		height = len(save.RowHints)
	}
	if width <= 0 || height <= 0 {
		return "### Puzzle Grid with Integrated Hints\n\n_(empty grid)_", nil
	}

	rowHints := normalizeNonogramHints(save.RowHints, height)
	colHints := normalizeNonogramHints(save.ColHints, width)

	rowHintCols := maxNonogramHintLen(rowHints)
	colHintRows := maxNonogramHintLen(colHints)
	if rowHintCols < 1 {
		rowHintCols = 1
	}
	if colHintRows < 1 {
		colHintRows = 1
	}

	var b strings.Builder
	b.WriteString("### Puzzle Grid with Integrated Hints\n\n")
	b.WriteString(renderNonogramTable(rowHints, colHints, width, height, rowHintCols, colHintRows))
	b.WriteString("\n\n")
	b.WriteString("Row hints are right-aligned beside each row. ")
	b.WriteString("Column hints are stacked above each column and bottom-aligned to the grid.")
	return b.String(), nil
}

func renderNurikabe(data []byte) (string, error) {
	var save nurikabe.Save
	if err := json.Unmarshal(data, &save); err != nil {
		return "", fmt.Errorf("decode nurikabe save: %w", err)
	}

	clues := parseNurikabeClues(save.Clues, save.Width, save.Height)
	cells := makeGrid(save.Width, save.Height, ".")
	for y := range len(cells) {
		for x := range len(cells[y]) {
			if y < len(clues) && x < len(clues[y]) && clues[y][x] > 0 {
				cells[y][x] = strconv.Itoa(clues[y][x])
			}
		}
	}

	var b strings.Builder
	b.WriteString("### Clue Grid\n\n")
	b.WriteString(renderGridTable(cells))
	b.WriteString("\n\n")
	b.WriteString("Goal: build one connected sea while each numbered island has the exact size of its clue.")
	return b.String(), nil
}

func renderShikaku(data []byte) (string, error) {
	var save shikaku.Save
	if err := json.Unmarshal(data, &save); err != nil {
		return "", fmt.Errorf("decode shikaku save: %w", err)
	}

	cells := makeGrid(save.Width, save.Height, ".")
	for _, clue := range save.Clues {
		if clue.Y >= 0 && clue.Y < len(cells) && clue.X >= 0 && clue.X < len(cells[clue.Y]) {
			cells[clue.Y][clue.X] = strconv.Itoa(clue.Value)
		}
	}

	var b strings.Builder
	b.WriteString("### Clue Grid\n\n")
	b.WriteString(renderGridTable(cells))
	b.WriteString("\n\n")
	b.WriteString("Goal: partition the grid into rectangles so each rectangle contains one clue and its area matches that clue.")
	return b.String(), nil
}

func renderSudoku(data []byte) (string, error) {
	var save sudoku.Save
	if err := json.Unmarshal(data, &save); err != nil {
		return "", fmt.Errorf("decode sudoku save: %w", err)
	}

	cells := makeGrid(9, 9, ".")
	for _, provided := range save.Provided {
		if provided.Y >= 0 && provided.Y < len(cells) && provided.X >= 0 && provided.X < len(cells[provided.Y]) {
			cells[provided.Y][provided.X] = strconv.Itoa(provided.V)
		}
	}

	var b strings.Builder
	b.WriteString("### Given Grid\n\n")
	b.WriteString(renderGridTable(cells))
	b.WriteString("\n\n")
	b.WriteString("Goal: fill each row, column, and 3x3 box with digits 1-9 exactly once.")
	return b.String(), nil
}

func renderTakuzu(data []byte) (string, error) {
	var save takuzu.Save
	if err := json.Unmarshal(data, &save); err != nil {
		return "", fmt.Errorf("decode takuzu save: %w", err)
	}

	cells := makeGrid(save.Size, save.Size, ".")
	stateRows := splitLines(save.State)
	providedRows := splitLines(save.Provided)

	for y := range len(cells) {
		for x := range len(cells[y]) {
			provided := y < len(providedRows) && x < len(providedRows[y]) && providedRows[y][x] == '#'
			if !provided {
				continue
			}

			if y < len(stateRows) && x < len(stateRows[y]) {
				switch stateRows[y][x] {
				case '0', '1':
					cells[y][x] = string(stateRows[y][x])
				default:
					cells[y][x] = "."
				}
			}
		}
	}

	var b strings.Builder
	b.WriteString("### Given Grid\n\n")
	b.WriteString(renderGridTable(cells))
	b.WriteString("\n\n")
	b.WriteString("Goal: fill with 0/1 so no three equal adjacent cells appear, each row/column has equal 0 and 1 counts, and rows/columns are unique.")
	return b.String(), nil
}

func renderWordSearch(data []byte) (string, error) {
	var save wordsearch.Save
	if err := json.Unmarshal(data, &save); err != nil {
		return "", fmt.Errorf("decode word search save: %w", err)
	}

	rows := splitLines(save.Grid)
	height := max(save.Height, len(rows))
	width := save.Width
	if width <= 0 {
		for _, row := range rows {
			if len([]rune(row)) > width {
				width = len([]rune(row))
			}
		}
	}

	cells := makeGrid(width, height, ".")
	for y := range len(cells) {
		if y >= len(rows) {
			continue
		}
		row := []rune(rows[y])
		for x := range len(cells[y]) {
			if x < len(row) {
				cells[y][x] = string(row[x])
			}
		}
	}

	var b strings.Builder
	b.WriteString("### Grid\n\n")
	b.WriteString(renderGridTable(cells))
	b.WriteString("\n\n")

	b.WriteString("### Word List\n\n")
	b.WriteString("| # | Word |\n")
	b.WriteString("| --- | --- |\n")
	for i, word := range save.Words {
		fmt.Fprintf(&b, "| %d | %s |\n", i+1, escapeCell(word.Text))
	}
	if len(save.Words) == 0 {
		b.WriteString("| 1 | (none) |\n")
	}

	b.WriteString("\nGoal: find all listed words in the grid.")
	return b.String(), nil
}

func makeGrid(width, height int, fill string) [][]string {
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

func renderGridTable(cells [][]string) string {
	if len(cells) == 0 {
		return "_(empty grid)_"
	}
	width := 0
	for _, row := range cells {
		if len(row) > width {
			width = len(row)
		}
	}
	if width == 0 {
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
			fmt.Fprintf(&b, " %s |", escapeCell(cell))
		}
		b.WriteString("\n")
	}

	return strings.TrimRight(b.String(), "\n")
}

func parseNurikabeClues(raw string, width, height int) [][]int {
	clues := make([][]int, max(height, 0))
	for y := range len(clues) {
		clues[y] = make([]int, max(width, 0))
	}

	rows := splitLines(raw)
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

func splitLines(s string) []string {
	if strings.TrimSpace(s) == "" {
		return []string{}
	}
	return strings.Split(s, "\n")
}

func escapeCell(s string) string {
	return strings.ReplaceAll(s, "|", "\\|")
}

func normalizeNonogramHints(src nonogram.TomographyDefinition, size int) [][]int {
	hints := make([][]int, max(size, 0))
	for i := range len(hints) {
		if i >= len(src) {
			hints[i] = []int{0}
			continue
		}
		if len(src[i]) == 0 {
			hints[i] = []int{0}
			continue
		}
		hints[i] = append([]int(nil), src[i]...)
	}
	return hints
}

func maxNonogramHintLen(hints [][]int) int {
	maxLen := 0
	for _, hint := range hints {
		if len(hint) > maxLen {
			maxLen = len(hint)
		}
	}
	return maxLen
}

func renderNonogramTable(
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
