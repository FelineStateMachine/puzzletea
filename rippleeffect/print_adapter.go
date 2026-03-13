package rippleeffect

import (
	"strconv"

	"codeberg.org/go-pdf/fpdf"
	"github.com/FelineStateMachine/puzzletea/pdfexport"
)

type printAdapter struct{}

var PDFPrintAdapter = printAdapter{}

func (printAdapter) CanonicalGameType() string { return "Ripple Effect" }
func (printAdapter) Aliases() []string         { return []string{"ripple effect", "rippleeffect", "ripple"} }

func (printAdapter) BuildPDFPayload(save []byte) (any, error) {
	return pdfexport.ParseRippleEffectPrintData(save)
}

func (printAdapter) RenderPDFBody(pdf *fpdf.Fpdf, payload any) error {
	switch data := payload.(type) {
	case *pdfexport.RippleEffectData:
		renderRippleEffectPage(pdf, data)
	}
	return nil
}

func renderRippleEffectPage(pdf *fpdf.Fpdf, data *pdfexport.RippleEffectData) {
	if data == nil || data.Width <= 0 || data.Height <= 0 {
		return
	}

	pageW, pageH := pdf.GetPageSize()
	pageNo := pdf.PageNo()
	area := pdfexport.PuzzleBoardRect(pageW, pageH, pageNo, 2)
	cellSize := pdfexport.FitCompactCellSize(data.Width, data.Height, area)
	if cellSize <= 0 {
		return
	}

	blockW := float64(data.Width) * cellSize
	blockH := float64(data.Height) * cellSize
	startX, startY := pdfexport.CenteredOrigin(area, data.Width, data.Height, cellSize)

	cageGrid := make([][]int, data.Height)
	for y := 0; y < data.Height; y++ {
		cageGrid[y] = make([]int, data.Width)
		for x := 0; x < data.Width; x++ {
			cageGrid[y][x] = -1
		}
	}
	for cageIdx, cage := range data.Cages {
		for _, cell := range cage.Cells {
			if cell.X >= 0 && cell.X < data.Width && cell.Y >= 0 && cell.Y < data.Height {
				cageGrid[cell.Y][cell.X] = cageIdx
			}
		}
	}
	masks := buildBoundaryMasks(cageGrid)

	pdf.SetDrawColor(150, 150, 150)
	pdf.SetLineWidth(pdfexport.ThinGridLineMM)
	for y := 0; y <= data.Height; y++ {
		lineY := startY + float64(y)*cellSize
		pdf.Line(startX, lineY, startX+blockW, lineY)
	}
	for x := 0; x <= data.Width; x++ {
		lineX := startX + float64(x)*cellSize
		pdf.Line(lineX, startY, lineX, startY+blockH)
	}

	pdf.SetDrawColor(35, 35, 35)
	pdf.SetLineWidth(pdfexport.MajorGridLineMM)
	for y := 0; y < data.Height; y++ {
		for x := 0; x < data.Width; x++ {
			cellX := startX + float64(x)*cellSize
			cellY := startY + float64(y)*cellSize
			mask := masks[y][x]
			if mask.has(boundaryTop) {
				pdf.Line(cellX, cellY, cellX+cellSize, cellY)
			}
			if mask.has(boundaryLeft) {
				pdf.Line(cellX, cellY, cellX, cellY+cellSize)
			}
			if mask.has(boundaryRight) {
				pdf.Line(cellX+cellSize, cellY, cellX+cellSize, cellY+cellSize)
			}
			if mask.has(boundaryBottom) {
				pdf.Line(cellX, cellY+cellSize, cellX+cellSize, cellY+cellSize)
			}

			if data.Givens[y][x] == 0 {
				continue
			}
			drawRippleEffectGiven(pdf, cellX, cellY, cellSize, strconv.Itoa(data.Givens[y][x]))
		}
	}

	ruleY := pdfexport.InstructionY(startY+blockH, pageH, 2)
	pdfexport.SetInstructionStyle(pdf)
	pdf.SetXY(area.X, ruleY)
	pdf.CellFormat(
		area.W,
		pdfexport.InstructionLineHMM,
		"Each cage uses 1..n once; equal digits must be at least their value apart in rows and columns.",
		"",
		0,
		"C",
		false,
		0,
		"",
	)
}

func drawRippleEffectGiven(pdf *fpdf.Fpdf, x, y, cellSize float64, text string) {
	fontSize := pdfexport.ClampStandardCellFontSize(pdfexport.StandardCellFontSize(cellSize, 0.58))
	pdf.SetTextColor(pdfexport.PrimaryTextGray, pdfexport.PrimaryTextGray, pdfexport.PrimaryTextGray)
	pdf.SetFont(pdfexport.SansFontFamily, "B", fontSize)
	lineH := fontSize * 0.9
	pdf.SetXY(x, y+(cellSize-lineH)/2)
	pdf.CellFormat(cellSize, lineH, text, "", 0, "C", false, 0, "")
}
