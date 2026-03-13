package fillomino

import (
	"strconv"

	"codeberg.org/go-pdf/fpdf"
	"github.com/FelineStateMachine/puzzletea/pdfexport"
)

type printAdapter struct{}

var PDFPrintAdapter = printAdapter{}

func (printAdapter) CanonicalGameType() string { return "Fillomino" }
func (printAdapter) Aliases() []string         { return []string{"fillomino", "polyomino", "regions"} }

func (printAdapter) BuildPDFPayload(save []byte) (any, error) {
	return pdfexport.ParseFillominoPrintData(save)
}

func (printAdapter) RenderPDFBody(pdf *fpdf.Fpdf, payload any) error {
	switch data := payload.(type) {
	case *pdfexport.FillominoData:
		renderFillominoPage(pdf, data)
	}
	return nil
}

func renderFillominoPage(pdf *fpdf.Fpdf, data *pdfexport.FillominoData) {
	if data == nil || data.Width <= 0 || data.Height <= 0 {
		return
	}

	pageW, pageH := pdf.GetPageSize()
	pageNo := pdf.PageNo()
	area := pdfexport.PuzzleBoardRect(pageW, pageH, pageNo, 1)
	cellSize := pdfexport.FitCompactCellSize(data.Width, data.Height, area)
	if cellSize <= 0 {
		return
	}

	blockW := float64(data.Width) * cellSize
	blockH := float64(data.Height) * cellSize
	startX, startY := pdfexport.CenteredOrigin(area, data.Width, data.Height, cellSize)

	pdf.SetDrawColor(60, 60, 60)
	pdf.SetLineWidth(pdfexport.ThinGridLineMM)
	for y := 0; y < data.Height; y++ {
		for x := 0; x < data.Width; x++ {
			cellX := startX + float64(x)*cellSize
			cellY := startY + float64(y)*cellSize
			pdf.Rect(cellX, cellY, cellSize, cellSize, "D")
			if data.Givens[y][x] == 0 {
				continue
			}
			drawFillominoGiven(pdf, cellX, cellY, cellSize, strconv.Itoa(data.Givens[y][x]))
		}
	}

	pdf.SetDrawColor(35, 35, 35)
	pdf.SetLineWidth(pdfexport.OuterBorderLineMM)
	pdf.Rect(startX, startY, blockW, blockH, "D")

	ruleY := pdfexport.InstructionY(startY+blockH, pageH, 1)
	pdfexport.SetInstructionStyle(pdf)
	pdf.SetXY(area.X, ruleY)
	pdf.CellFormat(
		area.W,
		pdfexport.InstructionLineHMM,
		"Each connected region must contain exactly the number shown in its cells.",
		"",
		0,
		"C",
		false,
		0,
		"",
	)
}

func drawFillominoGiven(pdf *fpdf.Fpdf, x, y, cellSize float64, text string) {
	fontSize := pdfexport.ClampStandardCellFontSize(pdfexport.StandardCellFontSize(cellSize, 0.58))
	pdf.SetTextColor(pdfexport.PrimaryTextGray, pdfexport.PrimaryTextGray, pdfexport.PrimaryTextGray)
	pdf.SetFont(pdfexport.SansFontFamily, "B", fontSize)
	lineH := fontSize * 0.9
	pdf.SetXY(x, y+(cellSize-lineH)/2)
	pdf.CellFormat(cellSize, lineH, text, "", 0, "C", false, 0, "")
}
