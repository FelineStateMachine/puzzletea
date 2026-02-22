package pdfexport

import (
	"math"
	"strconv"
	"unicode/utf8"

	"github.com/go-pdf/fpdf"
)

func renderHashiPage(pdf *fpdf.Fpdf, data *HashiData) {
	if data == nil || data.Width <= 0 || data.Height <= 0 {
		return
	}

	pageW, pageH := pdf.GetPageSize()
	pageNo := pdf.PageNo()

	spanX := max(data.Width-1, 1)
	spanY := max(data.Height-1, 1)
	area := puzzleBoardRect(pageW, pageH, pageNo, 1)
	step := fitBoardCellSize(spanX, spanY, area, boardFamilyHashi)
	if step <= 0 {
		return
	}

	boardW := float64(spanX) * step
	boardH := float64(spanY) * step
	originX, originY := centeredOrigin(area, spanX, spanY, step)
	islandRadius := hashiIslandRadius(step)

	drawHashiGuideDots(pdf, originX, originY, data.Width, data.Height, step)
	drawHashiBoardBorder(pdf, originX, originY, boardW, boardH, islandRadius)
	drawHashiIslands(pdf, originX, originY, step, islandRadius, data.Islands)

	// Add an explicit blank line before the Hashi hint text.
	ruleY := instructionY(originY+boardH+instructionLineHMM, pageH, 1)
	setInstructionStyle(pdf)
	pdf.SetXY(area.x, ruleY)
	pdf.CellFormat(
		area.w,
		instructionLineHMM,
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
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
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
	pdf.SetLineWidth(outerBorderLineMM)
	pdf.Rect(originX-borderPad, originY-borderPad, boardW+2*borderPad, boardH+2*borderPad, "D")
}

func drawHashiIslands(pdf *fpdf.Fpdf, originX, originY, step, radius float64, islands []HashiIsland) {
	pdf.SetDrawColor(20, 20, 20)
	pdf.SetFillColor(255, 255, 255)
	pdf.SetLineWidth(majorGridLineMM)

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
	fontSize := standardCellFontSize(radius*2.0, 0.95)
	runeCount := utf8.RuneCountInString(text)
	switch {
	case runeCount >= 3:
		fontSize *= 0.70
	case runeCount == 2:
		fontSize *= 0.82
	}
	fontSize = clampStandardCellFontSize(fontSize)

	pdf.SetTextColor(primaryTextGray, primaryTextGray, primaryTextGray)
	pdf.SetFont(sansFontFamily, "B", fontSize)
	lineH := fontSize * 0.88
	pdf.SetXY(cx-radius, cy-lineH/2)
	pdf.CellFormat(radius*2, lineH, text, "", 0, "C", false, 0, "")
}
