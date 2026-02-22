package pdfexport

import (
	"math"
	"strconv"
	"unicode/utf8"

	"github.com/go-pdf/fpdf"
)

func renderNurikabePage(pdf *fpdf.Fpdf, data *NurikabeData) {
	if data == nil || data.Width <= 0 || data.Height <= 0 {
		return
	}

	pageW, pageH := pdf.GetPageSize()
	marginX := 10.0
	top := 28.0
	bottom := 22.0

	availW := pageW - 2*marginX
	availH := pageH - top - bottom
	cellSize := math.Min(availW/float64(data.Width), availH/float64(data.Height))
	cellSize = math.Max(4.8, math.Min(12.0, cellSize))

	blockW := float64(data.Width) * cellSize
	blockH := float64(data.Height) * cellSize
	startX := (pageW - blockW) / 2
	startY := top + (availH-blockH)/2

	pdf.SetDrawColor(55, 55, 55)
	pdf.SetLineWidth(0.15)
	for y := 0; y < data.Height; y++ {
		for x := 0; x < data.Width; x++ {
			cellX := startX + float64(x)*cellSize
			cellY := startY + float64(y)*cellSize
			pdf.Rect(cellX, cellY, cellSize, cellSize, "D")

			if y >= len(data.Clues) || x >= len(data.Clues[y]) || data.Clues[y][x] <= 0 {
				continue
			}
			drawRectanglePuzzleClue(pdf, cellX, cellY, cellSize, data.Clues[y][x])
		}
	}

	pdf.SetDrawColor(35, 35, 35)
	pdf.SetLineWidth(0.62)
	pdf.Rect(startX, startY, blockW, blockH, "D")

	ruleY := startY + blockH + 5
	if ruleY+4 <= pageH-6 {
		pdf.SetTextColor(85, 85, 85)
		pdf.SetFont("Helvetica", "", 7.2)
		pdf.SetXY(marginX, ruleY)
		pdf.CellFormat(pageW-2*marginX, 4, "Expand each numbered island to its size; connect all sea cells into one wall.", "", 0, "C", false, 0, "")
	}
}

func renderShikakuPage(pdf *fpdf.Fpdf, data *ShikakuData) {
	if data == nil || data.Width <= 0 || data.Height <= 0 {
		return
	}

	pageW, pageH := pdf.GetPageSize()
	marginX := 10.0
	top := 28.0
	bottom := 22.0

	availW := pageW - 2*marginX
	availH := pageH - top - bottom
	cellSize := math.Min(availW/float64(data.Width), availH/float64(data.Height))
	cellSize = math.Max(4.8, math.Min(12.0, cellSize))

	blockW := float64(data.Width) * cellSize
	blockH := float64(data.Height) * cellSize
	startX := (pageW - blockW) / 2
	startY := top + (availH-blockH)/2

	pdf.SetDrawColor(55, 55, 55)
	pdf.SetLineWidth(0.15)
	for y := 0; y < data.Height; y++ {
		for x := 0; x < data.Width; x++ {
			cellX := startX + float64(x)*cellSize
			cellY := startY + float64(y)*cellSize
			pdf.Rect(cellX, cellY, cellSize, cellSize, "D")

			if y >= len(data.Clues) || x >= len(data.Clues[y]) || data.Clues[y][x] <= 0 {
				continue
			}
			drawRectanglePuzzleClue(pdf, cellX, cellY, cellSize, data.Clues[y][x])
		}
	}

	pdf.SetDrawColor(35, 35, 35)
	pdf.SetLineWidth(0.62)
	pdf.Rect(startX, startY, blockW, blockH, "D")

	ruleY := startY + blockH + 5
	if ruleY+4 <= pageH-6 {
		pdf.SetTextColor(85, 85, 85)
		pdf.SetFont("Helvetica", "", 7.2)
		pdf.SetXY(marginX, ruleY)
		pdf.CellFormat(pageW-2*marginX, 4, "Partition into rectangles where each clue equals its rectangle area.", "", 0, "C", false, 0, "")
	}
}

func drawRectanglePuzzleClue(pdf *fpdf.Fpdf, x, y, cellSize float64, value int) {
	text := strconv.Itoa(value)
	fontSize := standardCellFontSize(cellSize, 0.58)
	runeCount := utf8.RuneCountInString(text)
	switch {
	case runeCount >= 3:
		fontSize *= 0.72
	case runeCount == 2:
		fontSize *= 0.84
	}
	fontSize = clampStandardCellFontSize(fontSize)

	pdf.SetTextColor(20, 20, 20)
	pdf.SetFont("Helvetica", "B", fontSize)
	lineH := fontSize * 0.92
	pdf.SetXY(x, y+(cellSize-lineH)/2)
	pdf.CellFormat(cellSize, lineH, text, "", 0, "C", false, 0, "")
}
