package netwalk

import (
	"math"
	"strings"

	"codeberg.org/go-pdf/fpdf"
	"github.com/FelineStateMachine/puzzletea/pdfexport"
)

type printAdapter struct{}

var PDFPrintAdapter = printAdapter{}

func (printAdapter) CanonicalGameType() string { return "Netwalk" }
func (printAdapter) Aliases() []string         { return []string{"netwalk", "network"} }

func (printAdapter) BuildPDFPayload(save []byte) (any, error) {
	return pdfexport.ParseNetwalkPrintData(save)
}

func (printAdapter) RenderPDFBody(pdf *fpdf.Fpdf, payload any) error {
	switch data := payload.(type) {
	case *pdfexport.NetwalkData:
		renderNetwalkPage(pdf, data)
	}
	return nil
}

func renderNetwalkPage(pdf *fpdf.Fpdf, data *pdfexport.NetwalkData) {
	if data == nil || data.Size <= 0 {
		return
	}

	pageW, pageH := pdf.GetPageSize()
	pageNo := pdf.PageNo()
	area := pdfexport.PuzzleBoardRect(pageW, pageH, pageNo, 1)
	cellSize := pdfexport.FitCompactCellSize(data.Size, data.Size, area)
	if cellSize <= 0 {
		return
	}

	blockW := float64(data.Size) * cellSize
	blockH := float64(data.Size) * cellSize
	startX, startY := pdfexport.CenteredOrigin(area, data.Size, data.Size, cellSize)

	pdf.SetDrawColor(60, 60, 60)
	pdf.SetLineWidth(pdfexport.ThinGridLineMM)
	for y := range data.Size {
		for x := range data.Size {
			cellX := startX + float64(x)*cellSize
			cellY := startY + float64(y)*cellSize
			pdf.Rect(cellX, cellY, cellSize, cellSize, "D")
			drawNetwalkTile(pdf, data, x, y, cellX, cellY, cellSize)
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
		"Rotate tiles so every connector matches and the full network reaches the server without loops.",
		"",
		0,
		"C",
		false,
		0,
		"",
	)
}

func drawNetwalkTile(pdf *fpdf.Fpdf, data *pdfexport.NetwalkData, x, y int, cellX, cellY, cellSize float64) {
	mask := directionMask(data.Masks[y][x])
	if mask == 0 {
		return
	}
	mask = rotateMask(mask, data.Rotations[y][x])

	centerX := cellX + cellSize/2
	centerY := cellY + cellSize/2
	pad := cellSize * 0.16
	pdf.SetLineWidth(math.Max(cellSize*0.08, 0.35))
	pdf.SetDrawColor(50, 50, 50)

	if mask&north != 0 {
		pdf.Line(centerX, centerY, centerX, cellY+pad)
	}
	if mask&east != 0 {
		pdf.Line(centerX, centerY, cellX+cellSize-pad, centerY)
	}
	if mask&south != 0 {
		pdf.Line(centerX, centerY, centerX, cellY+cellSize-pad)
	}
	if mask&west != 0 {
		pdf.Line(centerX, centerY, cellX+pad, centerY)
	}

	if x == data.RootX && y == data.RootY {
		drawNetwalkCenteredText(pdf, centerX, centerY, cellSize, "★", 0.62)
		return
	}

	if degree(mask) == 1 {
		drawNetwalkCenteredText(pdf, centerX, centerY, cellSize, "•", 0.68)
	}
}

func drawNetwalkCenteredText(pdf *fpdf.Fpdf, centerX, centerY, cellSize float64, text string, scale float64) {
	fontSize := pdfexport.ClampStandardCellFontSize(pdfexport.StandardCellFontSize(cellSize, scale))
	pdf.SetTextColor(50, 50, 50)
	pdf.SetFont(pdfexport.SansFontFamily, "B", fontSize)
	lineH := fontSize * 0.9
	width := pdf.GetStringWidth(text)
	pdf.SetXY(centerX-width/2, centerY-lineH/2)
	pdf.CellFormat(width, lineH, strings.TrimSpace(text), "", 0, "C", false, 0, "")
}
