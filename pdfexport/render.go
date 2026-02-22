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
	pdf.CellFormat(pageW, 6, "A5 Printable Puzzle Pack", "", 0, "C", false, 0, "")

	metaY := 50.0
	pdf.SetTextColor(25, 25, 25)
	pdf.SetFont("Helvetica", "", 10)
	pdf.SetXY(margin, metaY)
	pdf.CellFormat(pageW-2*margin, 6, fmt.Sprintf("Generated: %s", cfg.GeneratedAt.Format(time.RFC3339)), "", 0, "L", false, 0, "")
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

	pdf.SetFont("Helvetica", "", 8.5)
	pdf.SetTextColor(85, 85, 85)
	for _, doc := range docs {
		line := fmt.Sprintf("%s | Category: %s | Mode: %s | Count: %d | Seed: %s",
			doc.Metadata.SourceFileName,
			doc.Metadata.Category,
			doc.Metadata.ModeSelection,
			doc.Metadata.Count,
			emptyAs(doc.Metadata.Seed, "none"),
		)
		wrapped := pdf.SplitLines([]byte(line), pageW-2*margin)
		for _, raw := range wrapped {
			if metaY > pageH-45 {
				break
			}
			pdf.SetXY(margin, metaY)
			pdf.CellFormat(pageW-2*margin, 5, string(raw), "", 0, "L", false, 0, "")
			metaY += 5
		}
		if metaY > pageH-45 {
			break
		}
	}

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

	pdf.SetTextColor(20, 20, 20)
	pdf.SetFont("Helvetica", "B", 13)
	pdf.SetXY(0, 10)
	title := fmt.Sprintf("%s %d: %s", puzzle.Category, puzzle.Index, puzzle.Name)
	pdf.CellFormat(pageW, 7, title, "", 0, "C", false, 0, "")

	pdf.SetFont("Helvetica", "", 9)
	pdf.SetTextColor(95, 95, 95)
	pdf.SetXY(0, 17)
	subtitle := fmt.Sprintf("Mode: %s | Source: %s | Difficulty score: %.2f",
		emptyAs(puzzle.ModeSelection, "mixed modes"),
		puzzle.SourceFileName,
		puzzle.DifficultyScore,
	)
	pdf.CellFormat(pageW, 5, subtitle, "", 0, "C", false, 0, "")

	if puzzle.Nonogram != nil {
		renderNonogramPage(pdf, puzzle.Nonogram)
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
	top := 28.0
	bottom := 12.0

	rowHintCols := maxHintDepth(data.RowHints)
	colHintRows := maxHintDepth(data.ColHints)
	if rowHintCols < 1 {
		rowHintCols = 1
	}
	if colHintRows < 1 {
		colHintRows = 1
	}

	totalCols := rowHintCols + data.Width
	totalRows := colHintRows + data.Height

	availW := pageW - 2*marginX
	availH := pageH - top - bottom
	cellSize := math.Min(availW/float64(totalCols), availH/float64(totalRows))
	if cellSize > 8.6 {
		cellSize = 8.6
	}

	blockW := float64(totalCols) * cellSize
	blockH := float64(totalRows) * cellSize
	startX := (pageW - blockW) / 2
	startY := top + (availH-blockH)/2

	pdf.SetDrawColor(45, 45, 45)
	pdf.SetLineWidth(0.12)
	for r := 0; r < totalRows; r++ {
		for c := 0; c < totalCols; c++ {
			x := startX + float64(c)*cellSize
			y := startY + float64(r)*cellSize
			pdf.Rect(x, y, cellSize, cellSize, "D")

			switch {
			case r < colHintRows && c >= rowHintCols:
				col := c - rowHintCols
				if text := colHintText(data.ColHints[col], colHintRows, r); text != "" {
					drawCellText(pdf, x, y, cellSize, cellSize, text, true)
				}
			case r >= colHintRows && c < rowHintCols:
				row := r - colHintRows
				if text := rowHintText(data.RowHints[row], rowHintCols, c); text != "" {
					drawCellText(pdf, x, y, cellSize, cellSize, text, true)
				}
			case r >= colHintRows && c >= rowHintCols:
				row := r - colHintRows
				col := c - rowHintCols
				if row < len(data.Grid) && col < len(data.Grid[row]) {
					cellText := strings.TrimSpace(data.Grid[row][col])
					if cellText != "" && cellText != "." {
						drawCellText(pdf, x, y, cellSize, cellSize, cellText, false)
					}
				}
			}
		}
	}

	xSep := startX + float64(rowHintCols)*cellSize
	ySep := startY + float64(colHintRows)*cellSize
	pdf.SetLineWidth(0.4)
	pdf.Line(xSep, startY, xSep, startY+blockH)
	pdf.Line(startX, ySep, startX+blockW, ySep)

	pdf.SetLineWidth(0.55)
	pdf.Rect(startX, startY, blockW, blockH, "D")
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

	fontSize := math.Max(3.1, math.Min(6.3, h*0.63))
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

func emptyAs(v, fallback string) string {
	if strings.TrimSpace(v) == "" {
		return fallback
	}
	return v
}
