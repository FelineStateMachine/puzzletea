package pdfexport

import (
	"fmt"
	"math"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/go-pdf/fpdf"
)

const (
	a5WidthMM  = 148.0
	a5HeightMM = 210.0
)

func WritePDF(outputPath string, docs []PackDocument, puzzles []Puzzle, cfg RenderConfig) error {
	if strings.TrimSpace(outputPath) == "" {
		return fmt.Errorf("output path is required")
	}
	if len(puzzles) == 0 {
		return fmt.Errorf("no puzzles to render")
	}

	if cfg.GeneratedAt.IsZero() {
		cfg.GeneratedAt = time.Now()
	}
	if strings.TrimSpace(cfg.Title) == "" {
		cfg.Title = defaultTitle(docs)
	}
	if strings.TrimSpace(cfg.AdvertText) == "" {
		cfg.AdvertText = "Find more puzzles at github.com/FelineStateMachine/puzzletea"
	}

	pdf := fpdf.NewCustom(&fpdf.InitType{
		OrientationStr: "P",
		UnitStr:        "mm",
		Size: fpdf.SizeType{
			Wd: a5WidthMM,
			Ht: a5HeightMM,
		},
	})
	pdf.SetAutoPageBreak(false, 0)
	pdf.SetCreator("PuzzleTea", true)
	pdf.SetAuthor("PuzzleTea", true)
	pdf.SetTitle(cfg.Title, true)
	pdf.SetFooterFunc(func() {
		if pdf.PageNo() <= 1 {
			return
		}
		pdf.SetY(-8)
		pdf.SetFont("Helvetica", "", 8)
		pdf.SetTextColor(105, 105, 105)
		pdf.CellFormat(0, 4, strconv.Itoa(pdf.PageNo()), "", 0, "C", false, 0, "")
	})

	renderTitlePage(pdf, docs, puzzles, cfg)
	for _, puzzle := range puzzles {
		renderPuzzlePage(pdf, puzzle)
	}

	dir := filepath.Dir(outputPath)
	if dir != "" && dir != "." {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return fmt.Errorf("create output directory: %w", err)
		}
	}

	if err := pdf.OutputFileAndClose(outputPath); err != nil {
		return fmt.Errorf("write pdf file: %w", err)
	}
	return nil
}

func renderTitlePage(pdf *fpdf.Fpdf, docs []PackDocument, puzzles []Puzzle, cfg RenderConfig) {
	pdf.AddPage()
	pageW, pageH := pdf.GetPageSize()
	margin := 12.0

	pdf.SetTextColor(20, 20, 20)
	pdf.SetFont("Helvetica", "B", 22)
	pdf.SetXY(0, 24)
	pdf.CellFormat(pageW, 10, cfg.Title, "", 0, "C", false, 0, "")

	pdf.SetFont("Helvetica", "", 11)
	pdf.SetTextColor(70, 70, 70)
	pdf.SetXY(0, 36)
	pdf.CellFormat(pageW, 6, "PuzzleTea Puzzle Pack", "", 0, "C", false, 0, "")

	metaY := 50.0
	pdf.SetTextColor(25, 25, 25)
	pdf.SetFont("Helvetica", "", 10)
	pdf.SetXY(margin, metaY)
	pdf.CellFormat(pageW-2*margin, 6, fmt.Sprintf("Generated: %s", cfg.GeneratedAt.Format("January 2, 2006")), "", 0, "L", false, 0, "")
	metaY += 6

	categories := summarizeCategories(puzzles)
	pdf.SetXY(margin, metaY)
	pdf.CellFormat(pageW-2*margin, 6, fmt.Sprintf("Puzzles: %d", len(puzzles)), "", 0, "L", false, 0, "")
	metaY += 6
	pdf.SetXY(margin, metaY)
	pdf.CellFormat(pageW-2*margin, 6, fmt.Sprintf("Categories: %s", strings.Join(categories, ", ")), "", 0, "L", false, 0, "")
	metaY += 8

	pdf.SetFont("Helvetica", "B", 10)
	pdf.SetTextColor(45, 45, 45)
	pdf.SetXY(margin, metaY)
	pdf.CellFormat(pageW-2*margin, 6, "Source Exports", "", 0, "L", false, 0, "")
	metaY += 7

	renderSourceExportsTable(pdf, docs, margin, metaY, pageW-2*margin, pageH-45)

	pdf.SetTextColor(50, 50, 50)
	pdf.SetFont("Helvetica", "B", 12)
	pdf.SetXY(margin, pageH-30)
	pdf.CellFormat(pageW-2*margin, 7, "Made with PuzzleTea", "", 0, "C", false, 0, "")

	pdf.SetFont("Helvetica", "", 10)
	pdf.SetXY(margin, pageH-22)
	pdf.CellFormat(pageW-2*margin, 6, cfg.AdvertText, "", 0, "C", false, 0, "")
}

func renderPuzzlePage(pdf *fpdf.Fpdf, puzzle Puzzle) {
	pdf.AddPage()
	pageW, pageH := pdf.GetPageSize()
	hydratePuzzlePrintData(&puzzle)

	pdf.SetTextColor(20, 20, 20)
	pdf.SetFont("Helvetica", "B", 13)
	pdf.SetXY(0, 10)
	title := fmt.Sprintf("%s %d: %s", puzzle.Category, puzzle.Index, puzzle.Name)
	pdf.CellFormat(pageW, 7, title, "", 0, "C", false, 0, "")

	pdf.SetFont("Helvetica", "", 9)
	pdf.SetTextColor(95, 95, 95)
	pdf.SetXY(0, 17)
	subtitleParts := []string{fmt.Sprintf("Difficulty Score: %d/10", difficultyScoreOutOfTen(puzzle.DifficultyScore))}
	if !isMixedModes(puzzle.ModeSelection) {
		subtitleParts = append([]string{fmt.Sprintf("Mode: %s", puzzle.ModeSelection)}, subtitleParts...)
	}
	subtitle := strings.Join(subtitleParts, " | ")
	pdf.CellFormat(pageW, 5, subtitle, "", 0, "C", false, 0, "")

	if puzzle.Nonogram != nil {
		renderNonogramPage(pdf, puzzle.Nonogram)
		return
	}
	if puzzle.Nurikabe != nil {
		renderNurikabePage(pdf, puzzle.Nurikabe)
		return
	}
	if puzzle.Shikaku != nil {
		renderShikakuPage(pdf, puzzle.Shikaku)
		return
	}
	if puzzle.Hashi != nil {
		renderHashiPage(pdf, puzzle.Hashi)
		return
	}
	if puzzle.Hitori != nil {
		renderHitoriPage(pdf, puzzle.Hitori)
		return
	}
	if puzzle.Takuzu != nil {
		renderTakuzuPage(pdf, puzzle.Takuzu)
		return
	}
	if puzzle.Sudoku != nil {
		renderSudokuPage(pdf, puzzle.Sudoku)
		return
	}
	if puzzle.WordSearch != nil {
		renderWordSearchPage(pdf, puzzle.WordSearch)
		return
	}
	if puzzle.Table != nil {
		renderGridTablePage(pdf, puzzle.Table)
		return
	}
	renderFallbackPage(pdf, puzzle, pageH)
}

func renderNonogramPage(pdf *fpdf.Fpdf, data *NonogramData) {
	if data == nil || data.Width <= 0 || data.Height <= 0 {
		return
	}

	pageW, pageH := pdf.GetPageSize()
	marginX := 10.0
	padding := 3.0
	top := 28.0
	bottom := 12.0

	rowHints := normalizeNonogramHintsForRender(data.RowHints, data.Height)
	colHints := normalizeNonogramHintsForRender(data.ColHints, data.Width)

	rowHintCols := maxHintDepth(rowHints)
	colHintRows := maxHintDepth(colHints)
	if rowHintCols < 1 {
		rowHintCols = 1
	}
	if colHintRows < 1 {
		colHintRows = 1
	}

	totalCols := rowHintCols + data.Width
	totalRows := colHintRows + data.Height

	availW := pageW - 2*(marginX+padding)
	availH := pageH - top - bottom - 2*padding
	cellSize := math.Min(availW/float64(totalCols), availH/float64(totalRows))
	if cellSize > 8.8 {
		cellSize = 8.8
	}
	if cellSize < 3.2 {
		cellSize = 3.2
	}

	blockW := float64(totalCols) * cellSize
	blockH := float64(totalRows) * cellSize
	gridW := float64(data.Width) * cellSize
	gridH := float64(data.Height) * cellSize
	startX := marginX + padding + (availW-blockW)/2
	startY := top + padding + (availH-blockH)/2

	xSep := startX + float64(rowHintCols)*cellSize
	ySep := startY + float64(colHintRows)*cellSize

	for row := 0; row < colHintRows; row++ {
		for col := 0; col < data.Width; col++ {
			cellX := xSep + float64(col)*cellSize
			cellY := startY + float64(row)*cellSize
			if text := colHintText(colHints[col], colHintRows, row); text != "" {
				drawNonogramHintText(pdf, cellX, cellY, cellSize, cellSize, text)
			}
		}
	}

	for row := 0; row < data.Height; row++ {
		for col := 0; col < rowHintCols; col++ {
			cellX := startX + float64(col)*cellSize
			cellY := ySep + float64(row)*cellSize
			if text := rowHintText(rowHints[row], rowHintCols, col); text != "" {
				drawNonogramHintText(pdf, cellX, cellY, cellSize, cellSize, text)
			}
		}
	}

	drawNonogramPuzzleGrid(pdf, xSep, ySep, data.Width, data.Height, cellSize)

	drawNonogramMajorLines(pdf, xSep, ySep, cellSize, data.Width, data.Height, 5)

	pdf.SetLineWidth(0.60)
	pdf.Rect(xSep, ySep, gridW, gridH, "D")
}

func drawNonogramPuzzleGrid(
	pdf *fpdf.Fpdf,
	startX,
	startY float64,
	width,
	height int,
	cellSize float64,
) {
	if width <= 0 || height <= 0 || cellSize <= 0 {
		return
	}

	gridW := float64(width) * cellSize
	gridH := float64(height) * cellSize

	pdf.SetDrawColor(45, 45, 45)
	pdf.SetLineWidth(0.12)
	for col := 0; col <= width; col++ {
		x := startX + float64(col)*cellSize
		pdf.Line(x, startY, x, startY+gridH)
	}
	for row := 0; row <= height; row++ {
		y := startY + float64(row)*cellSize
		pdf.Line(startX, y, startX+gridW, y)
	}
}

func drawNonogramHintText(pdf *fpdf.Fpdf, x, y, w, h float64, text string) {
	if strings.TrimSpace(text) == "" {
		return
	}

	// Darker, compact hint typography for dense nonograms.
	pdf.SetTextColor(55, 55, 55)
	fontSize := standardCellFontSize(h, 0.70)
	pdf.SetFont("Helvetica", "", fontSize)
	lineH := fontSize * 0.86
	pdf.SetXY(x, y+(h-lineH)/2)
	pdf.CellFormat(w, lineH, text, "", 0, "C", false, 0, "")
}

func renderSudokuPage(pdf *fpdf.Fpdf, data *SudokuData) {
	if data == nil {
		return
	}

	pageW, pageH := pdf.GetPageSize()
	marginX := 10.0
	top := 28.0
	bottom := 16.0
	availW := pageW - 2*marginX
	availH := pageH - top - bottom

	cellSize := math.Min(availW/9.0, availH/9.0)
	if cellSize > 12.5 {
		cellSize = 12.5
	}
	if cellSize < 10.5 {
		cellSize = 10.5
	}

	boardW := 9.0 * cellSize
	boardH := 9.0 * cellSize
	startX := (pageW - boardW) / 2
	startY := top + (availH-boardH)/2

	drawSudokuGridLines(pdf, startX, startY, cellSize)
	drawSudokuGivens(pdf, startX, startY, cellSize, data.Givens)

	pdf.SetFont("Helvetica", "", 8)
	pdf.SetTextColor(85, 85, 85)
	pdf.SetXY(marginX, pageH-11)
	pdf.CellFormat(pageW-2*marginX, 5, "Fill rows, columns, and 3x3 boxes with 1-9", "", 0, "C", false, 0, "")
}

func drawSudokuGridLines(pdf *fpdf.Fpdf, startX, startY, cellSize float64) {
	pdf.SetDrawColor(25, 25, 25)

	for i := range 10 {
		x := startX + float64(i)*cellSize
		pdf.SetLineWidth(sudokuLineWidthFor(i))
		pdf.Line(x, startY, x, startY+9.0*cellSize)
	}

	for i := range 10 {
		y := startY + float64(i)*cellSize
		pdf.SetLineWidth(sudokuLineWidthFor(i))
		pdf.Line(startX, y, startX+9.0*cellSize, y)
	}
}

func sudokuLineWidthFor(index int) float64 {
	switch {
	case index == 0 || index == 9:
		return 0.70
	case index%3 == 0:
		return 0.55
	default:
		return 0.15
	}
}

func drawSudokuGivens(pdf *fpdf.Fpdf, startX, startY, cellSize float64, givens [9][9]int) {
	fontSize := standardCellFontSize(cellSize, 0.62)
	lineH := fontSize * 0.85
	pdf.SetFont("Helvetica", "B", fontSize)
	pdf.SetTextColor(20, 20, 20)

	for y := range 9 {
		for x := range 9 {
			value := givens[y][x]
			if value < 1 || value > 9 {
				continue
			}
			cellX := startX + float64(x)*cellSize
			cellY := startY + float64(y)*cellSize
			pdf.SetXY(cellX, cellY+(cellSize-lineH)/2)
			pdf.CellFormat(cellSize, lineH, strconv.Itoa(value), "", 0, "C", false, 0, "")
		}
	}
}

func renderWordSearchPage(pdf *fpdf.Fpdf, data *WordSearchData) {
	if data == nil || data.Width <= 0 || data.Height <= 0 {
		return
	}

	pageW, pageH := pdf.GetPageSize()
	marginX := 10.0
	top := 28.0
	bottom := 12.0
	availW := pageW - 2*marginX
	availH := pageH - top - bottom

	columnCount := wordSearchColumnCount(data.Width, len(data.Words))
	wordFontSize := 8.3
	wordLineHeight := 4.2
	gridListGap := 4.0

	estimatedWordLines := estimateWordBankLineCount(pdf, data.Words, columnCount, availW, wordFontSize)
	wordBankHeight := 7.0 + float64(estimatedWordLines)*wordLineHeight
	maxWordBankHeight := availH * 0.42
	if wordBankHeight > maxWordBankHeight {
		wordBankHeight = maxWordBankHeight
	}
	if wordBankHeight < 16 {
		wordBankHeight = 16
	}

	gridAreaH := availH - wordBankHeight - gridListGap
	if gridAreaH < availH*0.5 {
		gridAreaH = availH * 0.5
	}

	cellSize := math.Min(availW/float64(data.Width), gridAreaH/float64(data.Height))
	if cellSize > 10.5 {
		cellSize = 10.5
	}
	if cellSize < 4.2 {
		cellSize = 4.2
	}

	gridW := float64(data.Width) * cellSize
	gridH := float64(data.Height) * cellSize
	gridX := (pageW - gridW) / 2
	gridY := top + (gridAreaH-gridH)/2

	drawWordSearchGrid(pdf, data, gridX, gridY, cellSize)
	drawWordBank(pdf, data.Words, marginX, gridY+gridH+gridListGap, availW, pageH-bottom-(gridY+gridH+gridListGap), columnCount)
}

func drawWordSearchGrid(pdf *fpdf.Fpdf, data *WordSearchData, startX, startY, cellSize float64) {
	pdf.SetDrawColor(45, 45, 45)
	pdf.SetLineWidth(0.12)
	for y := range data.Height {
		for x := range data.Width {
			cellX := startX + float64(x)*cellSize
			cellY := startY + float64(y)*cellSize
			pdf.Rect(cellX, cellY, cellSize, cellSize, "D")

			cellText := " "
			if y < len(data.Grid) && x < len(data.Grid[y]) {
				cellText = strings.TrimSpace(strings.ToUpper(data.Grid[y][x]))
			}
			if cellText == "" || cellText == "." {
				continue
			}

			fontSize := standardCellFontSize(cellSize, 0.74)
			lineH := fontSize * 0.86
			pdf.SetFont("Helvetica", "B", fontSize)
			pdf.SetTextColor(18, 18, 18)
			pdf.SetXY(cellX, cellY+(cellSize-lineH)/2)
			pdf.CellFormat(cellSize, lineH, cellText, "", 0, "C", false, 0, "")
		}
	}

	pdf.SetLineWidth(0.55)
	pdf.Rect(startX, startY, float64(data.Width)*cellSize, float64(data.Height)*cellSize, "D")
}

func drawWordBank(pdf *fpdf.Fpdf, words []string, x, y, width, height float64, columns int) {
	if height <= 0 {
		return
	}

	pdf.SetTextColor(40, 40, 40)
	pdf.SetFont("Helvetica", "B", 9.2)
	pdf.SetXY(x, y)
	pdf.CellFormat(width, 4.8, "Word Bank", "", 0, "L", false, 0, "")

	pdf.SetFont("Helvetica", "", 8.3)
	pdf.SetTextColor(70, 70, 70)
	pdf.SetXY(x, y+4.8)
	pdf.CellFormat(width, 4.2, "Words may run in all 8 directions", "", 0, "L", false, 0, "")

	listY := y + 9.0
	if len(words) == 0 {
		pdf.SetFont("Helvetica", "", 8.8)
		pdf.SetTextColor(95, 95, 95)
		pdf.SetXY(x, listY)
		pdf.CellFormat(width, 4.6, "(word list unavailable)", "", 0, "L", false, 0, "")
		return
	}

	columnGap := 4.0
	if columns < 1 {
		columns = 1
	}
	colWidth := (width - float64(columns-1)*columnGap) / float64(columns)
	if colWidth <= 0 {
		colWidth = width
		columns = 1
	}

	colLines := layoutWordBankColumns(pdf, words, columns, colWidth)
	lineHeight := 4.1
	maxLines := int(height / lineHeight)
	if maxLines <= 0 {
		return
	}

	pdf.SetTextColor(25, 25, 25)
	pdf.SetFont("Helvetica", "", 8.8)
	for c := range columns {
		colX := x + float64(c)*(colWidth+columnGap)
		curY := listY
		lines := colLines[c]
		for _, line := range lines {
			if int((curY-listY)/lineHeight) >= maxLines {
				break
			}
			pdf.SetXY(colX, curY)
			pdf.CellFormat(colWidth, lineHeight, line, "", 0, "L", false, 0, "")
			curY += lineHeight
		}
	}
}

func wordSearchColumnCount(width, wordCount int) int {
	if width >= 20 || wordCount > 18 {
		return 3
	}
	if wordCount <= 0 {
		return 1
	}
	return 2
}

func estimateWordBankLineCount(pdf *fpdf.Fpdf, words []string, columns int, availW, fontSize float64) int {
	if len(words) == 0 {
		return 1
	}
	if columns < 1 {
		columns = 1
	}

	colWidth := (availW - float64(columns-1)*4.0) / float64(columns)
	if colWidth <= 0 {
		colWidth = availW
	}

	pdf.SetFont("Helvetica", "", fontSize)
	lineCounts := make([]int, columns)
	for _, word := range words {
		text := strings.ToUpper(strings.TrimSpace(word))
		if text == "" {
			continue
		}
		lines := pdf.SplitLines([]byte(text), colWidth)
		if len(lines) == 0 {
			lines = [][]byte{[]byte(text)}
		}
		idx := minLineCountColumn(lineCounts)
		lineCounts[idx] += len(lines)
	}

	maxLines := 0
	for _, lineCount := range lineCounts {
		if lineCount > maxLines {
			maxLines = lineCount
		}
	}
	if maxLines == 0 {
		return 1
	}
	return maxLines
}

func layoutWordBankColumns(pdf *fpdf.Fpdf, words []string, columns int, colWidth float64) [][]string {
	colLines := make([][]string, columns)
	if len(words) == 0 || columns <= 0 {
		return colLines
	}

	lineCounts := make([]int, columns)
	for _, word := range words {
		text := strings.ToUpper(strings.TrimSpace(word))
		if text == "" {
			continue
		}
		wrapped := pdf.SplitLines([]byte(text), colWidth)
		if len(wrapped) == 0 {
			wrapped = [][]byte{[]byte(text)}
		}

		idx := minLineCountColumn(lineCounts)
		for _, line := range wrapped {
			colLines[idx] = append(colLines[idx], string(line))
		}
		lineCounts[idx] += len(wrapped)
	}

	return colLines
}

func minLineCountColumn(lineCounts []int) int {
	idx := 0
	for i := 1; i < len(lineCounts); i++ {
		if lineCounts[i] < lineCounts[idx] {
			idx = i
		}
	}
	return idx
}

func renderGridTablePage(pdf *fpdf.Fpdf, table *GridTable) {
	if table == nil || len(table.Rows) == 0 {
		return
	}

	pageW, pageH := pdf.GetPageSize()
	marginX := 10.0
	top := 28.0
	bottom := 12.0

	rows := len(table.Rows)
	cols := 0
	for _, row := range table.Rows {
		if len(row) > cols {
			cols = len(row)
		}
	}
	if cols == 0 {
		return
	}

	availW := pageW - 2*marginX
	availH := pageH - top - bottom
	cellSize := math.Min(availW/float64(cols), availH/float64(rows))
	if cellSize > 11.2 {
		cellSize = 11.2
	}
	if cellSize < 3.3 {
		cellSize = 3.3
	}

	blockW := float64(cols) * cellSize
	blockH := float64(rows) * cellSize
	startX := (pageW - blockW) / 2
	startY := top + (availH-blockH)/2

	pdf.SetDrawColor(55, 55, 55)
	pdf.SetLineWidth(0.16)
	for r := 0; r < rows; r++ {
		for c := 0; c < cols; c++ {
			x := startX + float64(c)*cellSize
			y := startY + float64(r)*cellSize
			pdf.Rect(x, y, cellSize, cellSize, "D")

			var text string
			if c < len(table.Rows[r]) {
				text = strings.TrimSpace(table.Rows[r][c])
			}
			if text == "." {
				text = " "
			}
			if text == "" {
				continue
			}

			dim := (table.HasHeaderRow && r == 0) || (table.HasHeaderCol && c == 0)
			drawCellText(pdf, x, y, cellSize, cellSize, text, dim)
		}
	}

	if table.HasHeaderRow {
		ySep := startY + cellSize
		pdf.SetLineWidth(0.42)
		pdf.Line(startX, ySep, startX+blockW, ySep)
	}
	if table.HasHeaderCol {
		xSep := startX + cellSize
		pdf.SetLineWidth(0.42)
		pdf.Line(xSep, startY, xSep, startY+blockH)
	}

	pdf.SetLineWidth(0.6)
	pdf.Rect(startX, startY, blockW, blockH, "D")
}

func renderFallbackPage(pdf *fpdf.Fpdf, puzzle Puzzle, pageH float64) {
	pageW, _ := pdf.GetPageSize()
	marginX := 10.0
	top := 28.0
	bottom := 12.0
	availW := pageW - 2*marginX
	availH := pageH - top - bottom

	lines := sanitizeBody(puzzle.Body)
	fontSize := 9.2
	lineHeight := 4.8

	pdf.SetFont("Courier", "", fontSize)
	wrapped := make([]string, 0, len(lines))
	for _, line := range lines {
		chunks := pdf.SplitLines([]byte(line), availW)
		if len(chunks) == 0 {
			wrapped = append(wrapped, "")
			continue
		}
		for _, raw := range chunks {
			wrapped = append(wrapped, string(raw))
		}
	}

	if total := float64(len(wrapped)) * lineHeight; total > availH && len(wrapped) > 0 {
		maxLines := int(availH / lineHeight)
		if maxLines < len(wrapped) {
			wrapped = append(wrapped[:max(0, maxLines-1)], "...")
		}
	}

	blockH := float64(len(wrapped)) * lineHeight
	startY := top + (availH-blockH)/2

	pdf.SetTextColor(50, 50, 50)
	y := startY
	for _, line := range wrapped {
		w := pdf.GetStringWidth(line)
		x := (pageW - w) / 2
		if x < marginX {
			x = marginX
		}
		pdf.SetXY(x, y)
		pdf.CellFormat(availW, lineHeight, line, "", 0, "L", false, 0, "")
		y += lineHeight
	}
}

func sanitizeBody(body string) []string {
	lines := strings.Split(body, "\n")
	cleaned := make([]string, 0, len(lines))
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || trimmed == "---" {
			cleaned = append(cleaned, "")
			continue
		}
		if strings.HasPrefix(trimmed, "### ") {
			cleaned = append(cleaned, strings.TrimSpace(strings.TrimPrefix(trimmed, "### ")))
			continue
		}
		if strings.HasPrefix(trimmed, "|") {
			line = strings.ReplaceAll(line, ".", " ")
		}
		cleaned = append(cleaned, line)
	}
	return cleaned
}

func drawCellText(pdf *fpdf.Fpdf, x, y, w, h float64, text string, dim bool) {
	if strings.TrimSpace(text) == "" {
		return
	}
	if dim {
		pdf.SetTextColor(130, 130, 130)
	} else {
		pdf.SetTextColor(25, 25, 25)
	}

	fontSize := standardCellFontSize(h, 0.63)
	pdf.SetFont("Helvetica", "", fontSize)
	lineH := fontSize * 0.9
	pdf.SetXY(x, y+(h-lineH)/2)
	pdf.CellFormat(w, lineH, text, "", 0, "C", false, 0, "")
}

func colHintText(hints []int, depth, row int) string {
	if len(hints) == 0 {
		return ""
	}
	start := depth - len(hints)
	if row < start {
		return ""
	}
	return strconv.Itoa(hints[row-start])
}

func rowHintText(hints []int, depth, col int) string {
	if len(hints) == 0 {
		return ""
	}
	start := depth - len(hints)
	if col < start {
		return ""
	}
	return strconv.Itoa(hints[col-start])
}

func maxHintDepth(hints [][]int) int {
	maxDepth := 0
	for _, h := range hints {
		if len(h) > maxDepth {
			maxDepth = len(h)
		}
	}
	return maxDepth
}

func normalizeNonogramHintsForRender(hints [][]int, size int) [][]int {
	if size <= 0 {
		return nil
	}

	normalized := make([][]int, size)
	for i := 0; i < size; i++ {
		if i >= len(hints) || len(hints[i]) == 0 {
			normalized[i] = []int{0}
			continue
		}
		normalized[i] = append([]int(nil), hints[i]...)
	}
	return normalized
}

func drawNonogramMajorLines(
	pdf *fpdf.Fpdf,
	puzzleStartX,
	puzzleStartY,
	cellSize float64,
	width,
	height,
	step int,
) {
	if step <= 0 || width <= 0 || height <= 0 {
		return
	}

	pdf.SetDrawColor(45, 45, 45)
	pdf.SetLineWidth(0.30)

	for col := step; col < width; col += step {
		x := puzzleStartX + float64(col)*cellSize
		pdf.Line(x, puzzleStartY, x, puzzleStartY+float64(height)*cellSize)
	}
	for row := step; row < height; row += step {
		y := puzzleStartY + float64(row)*cellSize
		pdf.Line(puzzleStartX, y, puzzleStartX+float64(width)*cellSize, y)
	}
}

func summarizeCategories(puzzles []Puzzle) []string {
	set := map[string]struct{}{}
	for _, p := range puzzles {
		category := strings.TrimSpace(p.Category)
		if category == "" {
			continue
		}
		set[category] = struct{}{}
	}

	categories := make([]string, 0, len(set))
	for category := range set {
		categories = append(categories, category)
	}
	sort.Strings(categories)
	if len(categories) == 0 {
		return []string{"Unknown"}
	}
	return categories
}

func defaultTitle(docs []PackDocument) string {
	if len(docs) == 1 {
		category := strings.TrimSpace(docs[0].Metadata.Category)
		if category != "" {
			return fmt.Sprintf("%s Puzzle Pack", category)
		}
	}
	return "PuzzleTea Mixed Puzzle Pack"
}

func renderSourceExportsTable(
	pdf *fpdf.Fpdf,
	docs []PackDocument,
	x, y, width, maxY float64,
) float64 {
	if width <= 0 || y >= maxY {
		return y
	}

	headers := []string{"Source", "Category", "Mode", "Count", "Seed"}
	columnRatios := []float64{0.33, 0.20, 0.22, 0.10, 0.15}
	columnWidths := make([]float64, len(columnRatios))
	usedWidth := 0.0
	for i := 0; i < len(columnRatios)-1; i++ {
		columnWidths[i] = width * columnRatios[i]
		usedWidth += columnWidths[i]
	}
	columnWidths[len(columnWidths)-1] = width - usedWidth

	headerHeight := 5.2
	rowHeight := 4.8
	availableRowsHeight := maxY - y - headerHeight
	if availableRowsHeight < rowHeight {
		return y
	}
	maxRows := int(math.Floor(availableRowsHeight / rowHeight))
	if maxRows < 1 {
		return y
	}

	rowCount := len(docs)
	if rowCount > maxRows {
		rowCount = maxRows
	}

	pdf.SetDrawColor(125, 125, 125)
	pdf.SetLineWidth(0.14)
	pdf.SetFillColor(245, 245, 245)
	pdf.SetTextColor(45, 45, 45)
	pdf.SetFont("Helvetica", "B", 8.4)

	curX := x
	for i, header := range headers {
		pdf.SetXY(curX, y)
		pdf.CellFormat(columnWidths[i], headerHeight, header, "1", 0, "C", true, 0, "")
		curX += columnWidths[i]
	}

	pdf.SetFont("Helvetica", "", 8.1)
	pdf.SetTextColor(70, 70, 70)
	for i := 0; i < rowCount; i++ {
		rowY := y + headerHeight + float64(i)*rowHeight
		mode := ""
		if !isMixedModes(docs[i].Metadata.ModeSelection) {
			mode = docs[i].Metadata.ModeSelection
		}

		values := []string{
			docs[i].Metadata.SourceFileName,
			docs[i].Metadata.Category,
			mode,
			strconv.Itoa(docs[i].Metadata.Count),
			emptyAs(docs[i].Metadata.Seed, "none"),
		}

		curX = x
		for col := range values {
			cellText := fitTableCellText(pdf, values[col], columnWidths[col]-1.6)
			align := "L"
			if col == 3 {
				align = "C"
			}
			pdf.SetXY(curX, rowY)
			pdf.CellFormat(columnWidths[col], rowHeight, cellText, "1", 0, align, false, 0, "")
			curX += columnWidths[col]
		}
	}

	return y + headerHeight + float64(rowCount)*rowHeight
}

func fitTableCellText(pdf *fpdf.Fpdf, text string, maxWidth float64) string {
	text = strings.TrimSpace(text)
	if text == "" {
		return ""
	}
	if maxWidth <= 0 {
		return ""
	}
	if pdf.GetStringWidth(text) <= maxWidth {
		return text
	}

	ellipsis := "..."
	if pdf.GetStringWidth(ellipsis) > maxWidth {
		return ellipsis
	}

	runes := []rune(text)
	for len(runes) > 0 {
		candidate := string(runes) + ellipsis
		if pdf.GetStringWidth(candidate) <= maxWidth {
			return candidate
		}
		runes = runes[:len(runes)-1]
	}
	return ellipsis
}

func emptyAs(v, fallback string) string {
	if strings.TrimSpace(v) == "" {
		return fallback
	}
	return v
}

func difficultyScoreOutOfTen(score float64) int {
	if score < 0 {
		score = 0
	}
	if score > 1 {
		score = 1
	}
	return int(math.Round(score * 10))
}

func isMixedModes(mode string) bool {
	return normalizeToken(mode) == "mixed modes"
}
