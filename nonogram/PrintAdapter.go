package nonogram

import (
	"strconv"
	"strings"

	"codeberg.org/go-pdf/fpdf"
	"github.com/FelineStateMachine/puzzletea/pdfexport"
)

type printAdapter struct{}

func (printAdapter) CanonicalGameType() string { return "Nonogram" }
func (printAdapter) Aliases() []string         { return []string{"nonogram"} }

func (printAdapter) BuildPDFPayload(save []byte) (any, error) {
	return pdfexport.ParseNonogramPrintData(save)
}

func (printAdapter) RenderPDFBody(pdf *fpdf.Fpdf, payload any) error {
	switch data := payload.(type) {
	case *pdfexport.NonogramData:
		renderNonogramPage(pdf, data)
	}
	return nil
}

func renderNonogramPage(pdf *fpdf.Fpdf, data *pdfexport.NonogramData) {
	if data == nil || data.Width <= 0 || data.Height <= 0 {
		return
	}

	pageW, pageH := pdf.GetPageSize()
	pageNo := pdf.PageNo()

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

	layout := layoutNonogram(
		pageW,
		pageH,
		pageNo,
		data.Width,
		data.Height,
		rowHintCols,
		colHintRows,
	)
	cellSize := layout.cellSize
	if cellSize <= 0 {
		return
	}

	gridW := float64(data.Width) * cellSize
	gridH := float64(data.Height) * cellSize
	startX := layout.hintStartX
	startY := layout.hintStartY
	xSep := layout.gridX
	ySep := layout.gridY

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

	pdf.SetLineWidth(pdfexport.OuterBorderLineMM)
	pdf.Rect(xSep, ySep, gridW, gridH, "D")

	ruleY := ySep + gridH + 3.5
	ruleY = pdfexport.InstructionY(ruleY-3.5, pageH, 1)
	body := pdfexport.PuzzleBodyRect(pageW, pageH, pageNo)
	pdfexport.SetInstructionStyle(pdf)
	pdf.SetXY(body.X, ruleY)
	pdf.CellFormat(
		body.W,
		pdfexport.InstructionLineHMM,
		"Use row/column hints to fill blocks in order; groups are separated by at least one blank cell.",
		"",
		0,
		"C",
		false,
		0,
		"",
	)
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
	pdf.SetLineWidth(pdfexport.ThinGridLineMM)
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

	pdf.SetTextColor(pdfexport.PrimaryTextGray, pdfexport.PrimaryTextGray, pdfexport.PrimaryTextGray)
	fontSize := pdfexport.StandardCellFontSize(h, 0.70)
	pdf.SetFont(pdfexport.SansFontFamily, "B", fontSize)
	lineH := fontSize * 0.86
	pdf.SetXY(x, y+(h-lineH)/2)
	pdf.CellFormat(w, lineH, text, "", 0, "C", false, 0, "")
}

type nonogramLayout struct {
	cellSize   float64
	hintStartX float64
	hintStartY float64
	gridX      float64
	gridY      float64
}

func layoutNonogram(
	pageW,
	pageH float64,
	pageNo,
	gridCols,
	gridRows,
	rowHintCols,
	colHintRows int,
) nonogramLayout {
	totalCols := rowHintCols + gridCols
	totalRows := colHintRows + gridRows
	area := pdfexport.PuzzleBoardRect(pageW, pageH, pageNo, 1)
	cellSize := pdfexport.FitNonogramCellSize(totalCols, totalRows, area)
	if cellSize <= 0 {
		return nonogramLayout{}
	}

	if rowHintCols > 0 {
		centeredCapW := area.W / float64(gridCols+2*rowHintCols)
		if centeredCapW > 0 && centeredCapW < cellSize {
			cellSize = centeredCapW
		}
	}
	if colHintRows > 0 {
		centeredCapH := area.H / float64(gridRows+2*colHintRows)
		if centeredCapH > 0 && centeredCapH < cellSize {
			cellSize = centeredCapH
		}
	}

	gridW := float64(gridCols) * cellSize
	gridH := float64(gridRows) * cellSize
	gridX := area.X + (area.W-gridW)/2
	gridY := area.Y + (area.H-gridH)/2
	hintStartX := gridX - float64(rowHintCols)*cellSize
	hintStartY := gridY - float64(colHintRows)*cellSize

	return nonogramLayout{
		cellSize:   cellSize,
		hintStartX: hintStartX,
		hintStartY: hintStartY,
		gridX:      gridX,
		gridY:      gridY,
	}
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
	pdf.SetLineWidth(pdfexport.MajorGridLineMM)

	for col := step; col < width; col += step {
		x := puzzleStartX + float64(col)*cellSize
		pdf.Line(x, puzzleStartY, x, puzzleStartY+float64(height)*cellSize)
	}
	for row := step; row < height; row += step {
		y := puzzleStartY + float64(row)*cellSize
		pdf.Line(puzzleStartX, y, puzzleStartX+float64(width)*cellSize, y)
	}
}

func init() {
	pdfexport.RegisterPrintAdapter(printAdapter{})
}
