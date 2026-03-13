package hashiwokakero

import (
	"math"
	"strconv"
	"unicode/utf8"

	"codeberg.org/go-pdf/fpdf"
	"github.com/FelineStateMachine/puzzletea/pdfexport"
)

type printAdapter struct{}

var PDFPrintAdapter = printAdapter{}

func (printAdapter) CanonicalGameType() string { return "Hashiwokakero" }
func (printAdapter) Aliases() []string {
	return []string{"hashi", "hashiwokakero", "hashi wokakero"}
}

func (printAdapter) BuildPDFPayload(save []byte) (any, error) {
	return pdfexport.ParseHashiPrintData(save)
}

func (printAdapter) RenderPDFBody(pdf *fpdf.Fpdf, payload any) error {
	switch data := payload.(type) {
	case *pdfexport.HashiData:
		renderHashiPage(pdf, data)
	}
	return nil
}

func renderHashiPage(pdf *fpdf.Fpdf, data *pdfexport.HashiData) {
	if data == nil || data.Width <= 0 || data.Height <= 0 {
		return
	}

	pageW, pageH := pdf.GetPageSize()
	pageNo := pdf.PageNo()

	spanX := max(data.Width-1, 1)
	spanY := max(data.Height-1, 1)
	area := pdfexport.PuzzleBoardRect(pageW, pageH, pageNo, 1)
	step := pdfexport.FitHashiCellSize(spanX, spanY, area)
	if step <= 0 {
		return
	}

	boardW := float64(spanX) * step
	boardH := float64(spanY) * step
	originX, originY := pdfexport.CenteredOrigin(area, spanX, spanY, step)
	islandRadius := hashiIslandRadius(step)

	drawHashiGuideDots(pdf, originX, originY, data.Width, data.Height, step)
	drawHashiBoardBorder(pdf, originX, originY, boardW, boardH, islandRadius)
	drawHashiIslands(pdf, originX, originY, step, islandRadius, data.Islands)

	ruleY := pdfexport.InstructionY(originY+boardH+pdfexport.InstructionLineHMM, pageH, 1)
	pdfexport.SetInstructionStyle(pdf)
	pdf.SetXY(area.X, ruleY)
	pdf.CellFormat(
		area.W,
		pdfexport.InstructionLineHMM,
		"Connect islands horizontally/vertically with up to two bridges and no crossings.",
		"",
		0,
		"C",
		false,
		0,
		"",
	)
}

func drawHashiGuideDots(pdf *fpdf.Fpdf, originX, originY float64, width, height int, step float64) {
	pdf.SetFillColor(230, 230, 230)
	r := math.Max(0.20, math.Min(0.55, step*0.035))
	for y := range height {
		for x := range width {
			cx := originX + float64(x)*step
			cy := originY + float64(y)*step
			pdf.Circle(cx, cy, r, "F")
		}
	}
}

func drawHashiBoardBorder(pdf *fpdf.Fpdf, originX, originY, boardW, boardH, islandRadius float64) {
	if boardW <= 0 || boardH <= 0 {
		return
	}
	borderPad := islandRadius + 1.2
	pdf.SetDrawColor(35, 35, 35)
	pdf.SetLineWidth(pdfexport.OuterBorderLineMM)
	pdf.Rect(originX-borderPad, originY-borderPad, boardW+2*borderPad, boardH+2*borderPad, "D")
}

func drawHashiIslands(
	pdf *fpdf.Fpdf,
	originX,
	originY,
	step,
	radius float64,
	islands []pdfexport.HashiIsland,
) {
	pdf.SetDrawColor(20, 20, 20)
	pdf.SetFillColor(255, 255, 255)
	pdf.SetLineWidth(pdfexport.MajorGridLineMM)

	for _, island := range islands {
		cx := originX + float64(island.X)*step
		cy := originY + float64(island.Y)*step
		pdf.Circle(cx, cy, radius, "DF")
		drawHashiIslandNumber(pdf, cx, cy, radius, island.Required)
	}
}

func hashiIslandRadius(step float64) float64 {
	return math.Max(1.4, math.Min(3.2, step*0.23))
}

func drawHashiIslandNumber(pdf *fpdf.Fpdf, cx, cy, radius float64, required int) {
	text := strconv.Itoa(required)
	fontSize := pdfexport.StandardCellFontSize(radius*2.0, 0.95)
	runeCount := utf8.RuneCountInString(text)
	switch {
	case runeCount >= 3:
		fontSize *= 0.70
	case runeCount == 2:
		fontSize *= 0.82
	}
	fontSize = pdfexport.ClampStandardCellFontSize(fontSize)

	pdf.SetTextColor(pdfexport.PrimaryTextGray, pdfexport.PrimaryTextGray, pdfexport.PrimaryTextGray)
	pdf.SetFont(pdfexport.SansFontFamily, "B", fontSize)
	lineH := fontSize * 0.88
	pdf.SetXY(cx-radius, cy-lineH/2)
	pdf.CellFormat(radius*2, lineH, text, "", 0, "C", false, 0, "")
}
