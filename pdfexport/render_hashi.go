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
	marginX := 10.0
	top := 28.0
	bottom := 22.0

	availW := pageW - 2*marginX
	availH := pageH - top - bottom

	spanX := max(data.Width-1, 1)
	spanY := max(data.Height-1, 1)
	step := math.Min(availW/float64(spanX), availH/float64(spanY))
	step = math.Max(5.2, math.Min(16.0, step))

	boardW := float64(spanX) * step
	boardH := float64(spanY) * step
	originX := (pageW - boardW) / 2
	originY := top + (availH-boardH)/2

	drawHashiGuideDots(pdf, originX, originY, data.Width, data.Height, step)
	drawHashiIslands(pdf, originX, originY, step, data.Islands)

	ruleY := originY + boardH + 5
	if ruleY+4 <= pageH-6 {
		pdf.SetTextColor(85, 85, 85)
		pdf.SetFont("Helvetica", "", 7.2)
		pdf.SetXY(marginX, ruleY)
		pdf.CellFormat(pageW-2*marginX, 4, "Connect islands horizontally/vertically with up to two bridges and no crossings.", "", 0, "C", false, 0, "")
	}
}

func drawHashiGuideDots(pdf *fpdf.Fpdf, originX, originY float64, width, height int, step float64) {
	pdf.SetFillColor(235, 235, 235)
	r := math.Max(0.20, math.Min(0.55, step*0.035))
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			cx := originX + float64(x)*step
			cy := originY + float64(y)*step
			pdf.Circle(cx, cy, r, "F")
		}
	}
}

func drawHashiIslands(pdf *fpdf.Fpdf, originX, originY, step float64, islands []HashiIsland) {
	radius := math.Max(1.4, math.Min(3.2, step*0.23))
	pdf.SetDrawColor(20, 20, 20)
	pdf.SetFillColor(255, 255, 255)
	pdf.SetLineWidth(0.26)

	for _, island := range islands {
		cx := originX + float64(island.X)*step
		cy := originY + float64(island.Y)*step
		pdf.Circle(cx, cy, radius, "DF")
		drawHashiIslandNumber(pdf, cx, cy, radius, island.Required)
	}
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

	pdf.SetTextColor(18, 18, 18)
	pdf.SetFont("Helvetica", "B", fontSize)
	lineH := fontSize * 0.88
	pdf.SetXY(cx-radius, cy-lineH/2)
	pdf.CellFormat(radius*2, lineH, text, "", 0, "C", false, 0, "")
}
