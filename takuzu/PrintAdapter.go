package takuzu

import (
	"strings"

	"codeberg.org/go-pdf/fpdf"
	"github.com/FelineStateMachine/puzzletea/game"
	"github.com/FelineStateMachine/puzzletea/pdfexport"
)

type printAdapter struct{}

func (printAdapter) CanonicalGameType() string { return "Takuzu" }
func (printAdapter) Aliases() []string         { return []string{"takuzu"} }

func (printAdapter) BuildPDFPayload(save []byte) (any, error) {
	return pdfexport.ParseTakuzuPrintData(save)
}

func (printAdapter) RenderPDFBody(pdf *fpdf.Fpdf, payload any) error {
	switch data := payload.(type) {
	case *pdfexport.TakuzuData:
		renderTakuzuPage(pdf, data)
	}
	return nil
}

func renderTakuzuPage(pdf *fpdf.Fpdf, data *pdfexport.TakuzuData) {
	if data == nil || data.Size <= 0 {
		return
	}

	size := data.Size
	pageW, pageH := pdf.GetPageSize()
	pageNo := pdf.PageNo()
	area := pdfexport.PuzzleBoardRect(pageW, pageH, pageNo, 2)
	cellSize := pdfexport.FitCompactCellSize(size, size, area)
	if cellSize <= 0 {
		return
	}

	blockW := float64(size) * cellSize
	blockH := float64(size) * cellSize
	startX, startY := pdfexport.CenteredOrigin(area, size, size, cellSize)

	pdf.SetDrawColor(60, 60, 60)
	pdf.SetLineWidth(pdfexport.ThinGridLineMM)
	for y := 0; y < size; y++ {
		for x := 0; x < size; x++ {
			cellX := startX + float64(x)*cellSize
			cellY := startY + float64(y)*cellSize
			pdf.Rect(cellX, cellY, cellSize, cellSize, "D")
		}
	}

	if data.GroupEveryTwo {
		pdf.SetDrawColor(130, 130, 130)
		pdf.SetLineWidth(pdfexport.MajorGridLineMM)
		for i := 2; i < size; i += 2 {
			x := startX + float64(i)*cellSize
			y := startY + float64(i)*cellSize
			pdf.Line(x, startY, x, startY+blockH)
			pdf.Line(startX, y, startX+blockW, y)
		}
	}

	for y := 0; y < size; y++ {
		for x := 0; x < size; x++ {
			text := ""
			if y < len(data.Givens) && x < len(data.Givens[y]) {
				text = strings.TrimSpace(data.Givens[y][x])
			}
			if text == "" {
				continue
			}

			cellX := startX + float64(x)*cellSize
			cellY := startY + float64(y)*cellSize
			drawTakuzuGiven(pdf, cellX, cellY, cellSize, size, text)
		}
	}

	pdf.SetDrawColor(35, 35, 35)
	pdf.SetLineWidth(pdfexport.OuterBorderLineMM)
	pdf.Rect(startX, startY, blockW, blockH, "D")

	ruleY := pdfexport.InstructionY(startY+blockH, pageH, 2)
	pdfexport.SetInstructionStyle(pdf)
	pdf.SetXY(area.X, ruleY)
	pdf.CellFormat(
		area.W,
		pdfexport.InstructionLineHMM,
		"No three equal adjacent in any row or column.",
		"",
		0,
		"C",
		false,
		0,
		"",
	)
	pdf.SetXY(area.X, ruleY+pdfexport.InstructionLineHMM)
	pdf.CellFormat(
		area.W,
		pdfexport.InstructionLineHMM,
		"Each row/column has equal 0 and 1 counts, and rows/columns are unique.",
		"",
		0,
		"C",
		false,
		0,
		"",
	)
}

func drawTakuzuGiven(pdf *fpdf.Fpdf, x, y, cellSize float64, size int, text string) {
	fontSize := takuzuGivenFontSize(cellSize, size)

	pdf.SetTextColor(pdfexport.PrimaryTextGray, pdfexport.PrimaryTextGray, pdfexport.PrimaryTextGray)
	pdf.SetFont(pdfexport.SansFontFamily, "B", fontSize)
	lineH := fontSize * 0.9
	pdf.SetXY(x, y+(cellSize-lineH)/2)
	pdf.CellFormat(cellSize, lineH, text, "", 0, "C", false, 0, "")
}

func takuzuGivenFontSize(cellSize float64, size int) float64 {
	fontSize := pdfexport.StandardCellFontSize(cellSize, 0.68)
	switch {
	case size >= 14:
		fontSize *= 0.94
	case size >= 12:
		fontSize *= 0.97
	}
	return pdfexport.ClampStandardCellFontSize(fontSize)
}

func init() {
	game.RegisterPrintAdapter(printAdapter{})
}
