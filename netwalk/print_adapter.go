package netwalk

import (
	"math"

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

	drawNetwalkGrid(pdf, startX, startY, blockW, blockH, data.Size, cellSize)

	for y := range data.Size {
		for x := range data.Size {
			cellX := startX + float64(x)*cellSize
			cellY := startY + float64(y)*cellSize
			drawNetwalkTile(pdf, data, x, y, cellX, cellY, cellSize)
		}
	}

	pdf.SetDrawColor(55, 55, 55)
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

func drawNetwalkGrid(pdf *fpdf.Fpdf, startX, startY, blockW, blockH float64, size int, cellSize float64) {
	if size <= 1 {
		return
	}

	pdf.SetDrawColor(115, 115, 115)
	pdf.SetLineWidth(math.Max(pdfexport.ThinGridLineMM*0.72, 0.14))

	for i := 1; i < size; i++ {
		offset := float64(i) * cellSize
		pdf.Line(startX+offset, startY, startX+offset, startY+blockH)
		pdf.Line(startX, startY+offset, startX+blockW, startY+offset)
	}
}

func drawNetwalkTile(pdf *fpdf.Fpdf, data *pdfexport.NetwalkData, x, y int, cellX, cellY, cellSize float64) {
	mask := directionMask(data.Masks[y][x])
	if mask == 0 {
		return
	}
	mask = rotateMask(mask, data.Rotations[y][x])

	centerX := cellX + cellSize/2
	centerY := cellY + cellSize/2
	isRoot := x == data.RootX && y == data.RootY
	isLeaf := degree(mask) == 1
	pad := cellSize * 0.16
	if isRoot || isLeaf {
		pad = math.Max(pad, netwalkMarkerRadius(cellSize, isRoot))
	}
	pdf.SetLineWidth(math.Max(cellSize*0.055, 0.26))
	pdf.SetDrawColor(65, 65, 65)

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

	switch {
	case isRoot:
		drawNetwalkSourceMarker(pdf, centerX, centerY, cellSize)
		return
	case isLeaf:
		drawNetwalkSinkMarker(pdf, centerX, centerY, cellSize)
	}
}

func netwalkMarkerRadius(cellSize float64, root bool) float64 {
	scale := 0.16
	if root {
		scale = 0.20
	}
	return math.Max(1.2, math.Min(2.6, cellSize*scale))
}

func drawNetwalkSinkMarker(pdf *fpdf.Fpdf, centerX, centerY, cellSize float64) {
	radius := netwalkMarkerRadius(cellSize, false)
	pdf.SetDrawColor(65, 65, 65)
	pdf.SetFillColor(255, 255, 255)
	pdf.SetLineWidth(math.Max(cellSize*0.045, 0.22))
	pdf.Circle(centerX, centerY, radius, "DF")
}

func drawNetwalkSourceMarker(pdf *fpdf.Fpdf, centerX, centerY, cellSize float64) {
	radius := netwalkMarkerRadius(cellSize, true)
	pdf.SetDrawColor(65, 65, 65)
	pdf.SetFillColor(255, 255, 255)
	pdf.SetLineWidth(math.Max(cellSize*0.05, 0.24))
	pdf.Circle(centerX, centerY, radius, "DF")

	pdf.SetFillColor(65, 65, 65)
	pdf.Circle(centerX, centerY, radius*0.36, "F")
}
