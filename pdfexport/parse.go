package pdfexport

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"
)

var (
	puzzleHeadingPattern   = regexp.MustCompile(`^##\s+(.+?)\s+-\s+(\d+)\s*$`)
	nonogramRowHeaderRegex = regexp.MustCompile(`^R\d+$`)
	nonogramColHeaderRegex = regexp.MustCompile(`^C\d+$`)
	tableSepCellRegex      = regexp.MustCompile(`^:?-{3,}:?$`)
)

func ParseFiles(paths []string) ([]PackDocument, error) {
	docs := make([]PackDocument, 0, len(paths))
	for _, path := range paths {
		doc, err := ParseFile(path)
		if err != nil {
			return nil, err
		}
		docs = append(docs, doc)
	}
	return docs, nil
}

func ParseFile(path string) (PackDocument, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return PackDocument{}, fmt.Errorf("read input markdown: %w", err)
	}
	return ParseMarkdown(path, string(data))
}

func ParseMarkdown(path, content string) (PackDocument, error) {
	content = strings.ReplaceAll(content, "\r\n", "\n")
	content = strings.ReplaceAll(content, "\r", "\n")
	lines := strings.Split(content, "\n")

	firstContentLine := firstNonEmptyLine(lines)
	if firstContentLine < 0 {
		return PackDocument{}, parseError(path, 1, "input markdown is empty")
	}
	if strings.TrimSpace(lines[firstContentLine]) != "# PuzzleTea Export" {
		return PackDocument{}, parseError(path, firstContentLine+1, "expected markdown title '# PuzzleTea Export'")
	}

	headings := findHeadingLines(lines)
	if len(headings) == 0 {
		return PackDocument{}, parseError(path, firstContentLine+1, "expected at least one puzzle section heading")
	}

	meta, err := parseMetadata(lines[firstContentLine+1:headings[0]], path, firstContentLine+2)
	if err != nil {
		return PackDocument{}, err
	}
	meta.SourceFileName = filepath.Base(path)

	puzzles := make([]Puzzle, 0, len(headings))
	for i, start := range headings {
		end := len(lines)
		if i+1 < len(headings) {
			end = headings[i+1]
		}

		puzzle, err := parsePuzzleSection(lines[start:end], path, start+1, meta)
		if err != nil {
			return PackDocument{}, err
		}
		puzzles = append(puzzles, puzzle)
	}

	if len(puzzles) == 0 {
		return PackDocument{}, parseError(path, firstContentLine+1, "no puzzle sections were parsed")
	}

	return PackDocument{
		SourcePath: path,
		Metadata:   meta,
		Puzzles:    puzzles,
	}, nil
}

func parseMetadata(lines []string, path string, startLine int) (PackMetadata, error) {
	meta := PackMetadata{}

	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}
		if !strings.HasPrefix(trimmed, "- ") {
			continue
		}

		entry := strings.TrimSpace(strings.TrimPrefix(trimmed, "- "))
		key, value, ok := strings.Cut(entry, ":")
		if !ok {
			continue
		}

		lineNo := startLine + i
		key = strings.ToLower(strings.TrimSpace(key))
		value = strings.TrimSpace(value)

		switch key {
		case "generated":
			meta.GeneratedRaw = value
			if ts, err := time.Parse(time.RFC3339, value); err == nil {
				meta.GeneratedAt = ts
			}
		case "version":
			meta.Version = value
		case "category":
			meta.Category = value
		case "mode selection":
			meta.ModeSelection = value
		case "count":
			count, err := strconv.Atoi(value)
			if err != nil {
				return PackMetadata{}, parseError(path, lineNo, "invalid Count value %q", value)
			}
			meta.Count = count
		case "seed":
			meta.Seed = value
		case "export format":
			meta.Format = value
		}
	}

	if strings.TrimSpace(meta.Category) == "" {
		return PackMetadata{}, parseError(path, startLine, "missing required metadata field: Category")
	}
	if strings.TrimSpace(meta.ModeSelection) == "" {
		return PackMetadata{}, parseError(path, startLine, "missing required metadata field: Mode Selection")
	}

	return meta, nil
}

func parsePuzzleSection(section []string, path string, startLine int, meta PackMetadata) (Puzzle, error) {
	if len(section) == 0 {
		return Puzzle{}, parseError(path, startLine, "empty puzzle section")
	}

	heading := strings.TrimSpace(section[0])
	matches := puzzleHeadingPattern.FindStringSubmatch(heading)
	if len(matches) != 3 {
		return Puzzle{}, parseError(path, startLine, "invalid puzzle heading %q", heading)
	}

	index, err := strconv.Atoi(matches[2])
	if err != nil {
		return Puzzle{}, parseError(path, startLine, "invalid puzzle index %q", matches[2])
	}

	bodyLines := append([]string(nil), section[1:]...)
	trimSectionBody(&bodyLines)
	body := strings.Join(bodyLines, "\n")

	p := Puzzle{
		SourcePath:     path,
		SourceFileName: meta.SourceFileName,
		Category:       meta.Category,
		ModeSelection:  meta.ModeSelection,
		Name:           matches[1],
		Index:          index,
		Body:           body,
	}

	if strings.EqualFold(strings.TrimSpace(meta.Category), "nonogram") {
		nonogram, err := parseNonogramBody(bodyLines, path, startLine+1)
		if err != nil {
			return Puzzle{}, err
		}
		p.PrintPayload = nonogram
		return p, nil
	}

	table, err := parseGridTableBody(bodyLines, path, startLine+1)
	if err != nil {
		return Puzzle{}, err
	}
	p.PrintPayload = table

	return p, nil
}

func parseNonogramBody(bodyLines []string, path string, bodyStartLine int) (*NonogramData, error) {
	tableLines, tableLineNumbers := findFirstMarkdownTable(bodyLines, bodyStartLine)
	if len(tableLines) < 3 {
		lineNo := bodyStartLine
		if len(tableLineNumbers) > 0 {
			lineNo = tableLineNumbers[0]
		}
		return nil, parseError(path, lineNo, "expected nonogram markdown table with header, separator, and data rows")
	}

	header := parseTableRow(tableLines[0])
	if len(header) == 0 {
		return nil, parseError(path, tableLineNumbers[0], "nonogram header row is empty")
	}

	rowHintCols := 0
	colCount := 0
	for _, cell := range header {
		switch {
		case nonogramRowHeaderRegex.MatchString(cell):
			rowHintCols++
		case nonogramColHeaderRegex.MatchString(cell):
			colCount++
		}
	}
	if rowHintCols < 1 || colCount < 1 {
		return nil, parseError(path, tableLineNumbers[0], "expected nonogram header cells like R1.. and C1..")
	}

	expectedCols := rowHintCols + colCount
	dataRows := make([][]string, 0, len(tableLines)-2)
	for i := 2; i < len(tableLines); i++ {
		cells := parseTableRow(tableLines[i])
		if len(cells) < expectedCols {
			return nil, parseError(path, tableLineNumbers[i], "expected %d columns, found %d", expectedCols, len(cells))
		}
		if len(cells) > expectedCols {
			cells = cells[:expectedCols]
		}
		dataRows = append(dataRows, cells)
	}

	if len(dataRows) == 0 {
		return nil, parseError(path, tableLineNumbers[0], "nonogram table has no data rows")
	}

	colHintRows := 0
	for _, row := range dataRows {
		if rowHintPlaceholderRow(row[:rowHintCols]) {
			colHintRows++
			continue
		}
		break
	}

	height := len(dataRows) - colHintRows
	if height <= 0 {
		return nil, parseError(path, tableLineNumbers[len(tableLineNumbers)-1], "nonogram table does not contain puzzle rows")
	}

	rowHints := make([][]int, height)
	grid := make([][]string, height)
	for y := range height {
		row := dataRows[colHintRows+y]
		hints := parseHintCells(row[:rowHintCols])
		if len(hints) == 0 {
			hints = []int{0}
		}
		rowHints[y] = hints

		gridRow := make([]string, colCount)
		for x := 0; x < colCount; x++ {
			cell := strings.TrimSpace(row[rowHintCols+x])
			switch cell {
			case "", ".":
				gridRow[x] = " "
			default:
				gridRow[x] = cell
			}
		}
		grid[y] = gridRow
	}

	colHints := make([][]int, colCount)
	for x := 0; x < colCount; x++ {
		hints := make([]int, 0, colHintRows)
		for r := 0; r < colHintRows; r++ {
			if v, ok := parseHintValue(dataRows[r][rowHintCols+x]); ok {
				hints = append(hints, v)
			}
		}
		if len(hints) == 0 {
			hints = []int{0}
		}
		colHints[x] = hints
	}

	return &NonogramData{
		Width:    colCount,
		Height:   height,
		RowHints: rowHints,
		ColHints: colHints,
		Grid:     grid,
	}, nil
}

func parseGridTableBody(bodyLines []string, path string, bodyStartLine int) (*GridTable, error) {
	tableLines, tableLineNumbers := findFirstMarkdownTable(bodyLines, bodyStartLine)
	if len(tableLines) == 0 {
		return nil, nil
	}
	if len(tableLines) < 2 {
		return nil, parseError(path, tableLineNumbers[0], "expected markdown table to include data rows")
	}

	rows := make([][]string, 0, len(tableLines))
	for i, line := range tableLines {
		cells := parseTableRow(line)
		if len(cells) == 0 {
			return nil, parseError(path, tableLineNumbers[i], "empty table row")
		}
		rows = append(rows, cells)
	}

	hasHeaderRow := false
	if len(rows) > 1 && isMarkdownSeparatorRow(rows[1]) {
		hasHeaderRow = true
		rows = append(rows[:1], rows[2:]...)
	}
	if len(rows) < 2 {
		return nil, parseError(path, tableLineNumbers[0], "table must include header and data rows")
	}

	width := 0
	for _, row := range rows {
		if len(row) > width {
			width = len(row)
		}
	}
	for i := range rows {
		if len(rows[i]) >= width {
			continue
		}
		padded := make([]string, width)
		copy(padded, rows[i])
		rows[i] = padded
	}

	return &GridTable{
		Rows:         rows,
		HasHeaderRow: hasHeaderRow,
		HasHeaderCol: detectHeaderColumn(rows, hasHeaderRow),
	}, nil
}

func findFirstMarkdownTable(lines []string, startLine int) ([]string, []int) {
	table := []string{}
	lineNumbers := []int{}
	started := false

	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "|") {
			started = true
			table = append(table, line)
			lineNumbers = append(lineNumbers, startLine+i)
			continue
		}
		if started {
			break
		}
	}

	return table, lineNumbers
}

func parseTableRow(line string) []string {
	trimmed := strings.TrimSpace(line)
	trimmed = strings.TrimPrefix(trimmed, "|")
	trimmed = strings.TrimSuffix(trimmed, "|")
	if trimmed == "" {
		return []string{}
	}

	parts := strings.Split(trimmed, "|")
	cells := make([]string, 0, len(parts))
	for _, part := range parts {
		cells = append(cells, strings.TrimSpace(part))
	}
	return cells
}

func isMarkdownSeparatorRow(cells []string) bool {
	if len(cells) == 0 {
		return false
	}
	for _, cell := range cells {
		if !tableSepCellRegex.MatchString(strings.TrimSpace(cell)) {
			return false
		}
	}
	return true
}

func rowHintPlaceholderRow(cells []string) bool {
	for _, cell := range cells {
		trimmed := strings.TrimSpace(cell)
		if trimmed != "" && trimmed != "." {
			return false
		}
	}
	return true
}

func parseHintCells(cells []string) []int {
	hints := make([]int, 0, len(cells))
	for _, cell := range cells {
		if v, ok := parseHintValue(cell); ok {
			hints = append(hints, v)
		}
	}
	return hints
}

func detectHeaderColumn(rows [][]string, hasHeaderRow bool) bool {
	if len(rows) < 2 {
		return false
	}

	start := 1
	if !hasHeaderRow {
		start = 0
	}

	total := 0
	numeric := 0
	for i := start; i < len(rows); i++ {
		if len(rows[i]) == 0 {
			continue
		}
		total++
		if _, err := strconv.Atoi(strings.TrimSpace(rows[i][0])); err == nil {
			numeric++
		}
	}

	return total > 0 && numeric*100/total >= 70
}

func parseHintValue(cell string) (int, bool) {
	trimmed := strings.TrimSpace(cell)
	if trimmed == "" || trimmed == "." {
		return 0, false
	}
	v, err := strconv.Atoi(trimmed)
	if err != nil {
		return 0, false
	}
	return v, true
}

func trimSectionBody(lines *[]string) {
	for len(*lines) > 0 {
		trimmed := strings.TrimSpace((*lines)[0])
		if trimmed != "" && trimmed != "---" {
			break
		}
		*lines = (*lines)[1:]
	}
	for len(*lines) > 0 {
		trimmed := strings.TrimSpace((*lines)[len(*lines)-1])
		if trimmed != "" && trimmed != "---" {
			break
		}
		*lines = (*lines)[:len(*lines)-1]
	}
}

func firstNonEmptyLine(lines []string) int {
	for i, line := range lines {
		if strings.TrimSpace(line) != "" {
			return i
		}
	}
	return -1
}

func findHeadingLines(lines []string) []int {
	indexes := []int{}
	for i, line := range lines {
		if strings.HasPrefix(strings.TrimSpace(line), "## ") {
			indexes = append(indexes, i)
		}
	}
	return indexes
}

func parseError(path string, line int, format string, args ...any) error {
	if line < 1 {
		line = 1
	}
	return fmt.Errorf("%s:%d: %s", path, line, fmt.Sprintf(format, args...))
}
