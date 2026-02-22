package pdfexport

import (
	"math"
	"strings"
	"unicode/utf8"

	"github.com/go-pdf/fpdf"
)

func renderHitoriPage(pdf *fpdf.Fpdf, data *HitoriData) {
	if data == nil || data.Size <= 0 {
		return
	}

	size := data.Size
	pageW, pageH := pdf.GetPageSize()
	marginX := 10.0
	top := 28.0
	bottom := 22.0

	availW := pageW - 2*marginX
	availH := pageH - top - bottom
	cellSize := math.Min(availW/float64(size), availH/float64(size))
	cellSize = math.Max(5.0, math.Min(12.0, cellSize))

	blockW := float64(size) * cellSize
	blockH := float64(size) * cellSize
	startX := (pageW - blockW) / 2
	startY := top + (availH-blockH)/2

	pdf.SetDrawColor(55, 55, 55)
	pdf.SetLineWidth(0.16)
	for y := 0; y < size; y++ {
		for x := 0; x < size; x++ {
			cellX := startX + float64(x)*cellSize
			cellY := startY + float64(y)*cellSize
			pdf.Rect(cellX, cellY, cellSize, cellSize, "D")

			text := ""
			if y < len(data.Numbers) && x < len(data.Numbers[y]) {
				text = strings.TrimSpace(data.Numbers[y][x])
			}
			if text == "" || text == "." {
				continue
			}
			drawHitoriCellNumber(pdf, cellX, cellY, cellSize, text)
		}
	}

	pdf.SetDrawColor(35, 35, 35)
	pdf.SetLineWidth(0.62)
	pdf.Rect(startX, startY, blockW, blockH, "D")

	ruleY := startY + blockH + 5
	if ruleY+4 <= pageH-6 {
		pdf.SetTextColor(85, 85, 85)
		pdf.SetFont("Helvetica", "", 7.3)
		pdf.SetXY(marginX, ruleY)
		pdf.CellFormat(
			pageW-2*marginX,
			4,
			"Shade duplicates; shaded cells cannot touch orthogonally; unshaded cells stay connected.",
			"",
			0,
			"C",
			false,
			0,
			"",
		)
	}
}

func drawHitoriCellNumber(pdf *fpdf.Fpdf, x, y, cellSize float64, text string) {
	fontSize := standardCellFontSize(cellSize, 0.58)
	runeCount := utf8.RuneCountInString(text)
	switch {
	case runeCount >= 3:
		fontSize *= 0.7
	case runeCount == 2:
		fontSize *= 0.82
	}
	fontSize = clampStandardCellFontSize(fontSize)

	pdf.SetTextColor(20, 20, 20)
	pdf.SetFont("Helvetica", "B", fontSize)
	lineH := fontSize * 0.92
	pdf.SetXY(x, y+(cellSize-lineH)/2)
	pdf.CellFormat(cellSize, lineH, text, "", 0, "C", false, 0, "")
}

func renderTakuzuPage(pdf *fpdf.Fpdf, data *TakuzuData) {
	if data == nil || data.Size <= 0 {
		return
	}

	size := data.Size
	pageW, pageH := pdf.GetPageSize()
	marginX := 10.0
	top := 28.0
	bottom := 24.0

	availW := pageW - 2*marginX
	availH := pageH - top - bottom
	cellSize := math.Min(availW/float64(size), availH/float64(size))
	cellSize = math.Max(4.4, math.Min(11.0, cellSize))

	blockW := float64(size) * cellSize
	blockH := float64(size) * cellSize
	startX := (pageW - blockW) / 2
	startY := top + (availH-blockH)/2

	pdf.SetDrawColor(60, 60, 60)
	pdf.SetLineWidth(0.14)
	for y := 0; y < size; y++ {
		for x := 0; x < size; x++ {
			cellX := startX + float64(x)*cellSize
			cellY := startY + float64(y)*cellSize
			pdf.Rect(cellX, cellY, cellSize, cellSize, "D")
		}
	}

	if data.GroupEveryTwo {
		pdf.SetDrawColor(145, 145, 145)
		pdf.SetLineWidth(0.24)
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
	pdf.SetLineWidth(0.60)
	pdf.Rect(startX, startY, blockW, blockH, "D")

	ruleY := startY + blockH + 4
	if ruleY+8 <= pageH-6 {
		pdf.SetTextColor(85, 85, 85)
		pdf.SetFont("Helvetica", "", 7.1)
		pdf.SetXY(marginX, ruleY)
		pdf.CellFormat(pageW-2*marginX, 4, "No three equal adjacent in any row or column.", "", 0, "C", false, 0, "")
		pdf.SetXY(marginX, ruleY+4)
		pdf.CellFormat(pageW-2*marginX, 4, "Each row/column has equal 0 and 1 counts, and rows/columns are unique.", "", 0, "C", false, 0, "")
	}
}

func drawTakuzuGiven(pdf *fpdf.Fpdf, x, y, cellSize float64, size int, text string) {
	fontSize := takuzuGivenFontSize(cellSize, size)

	pdf.SetTextColor(15, 15, 15)
	pdf.SetFont("Helvetica", "B", fontSize)
	lineH := fontSize * 0.9
	pdf.SetXY(x, y+(cellSize-lineH)/2)
	pdf.CellFormat(cellSize, lineH, text, "", 0, "C", false, 0, "")
}

func takuzuGivenFontSize(cellSize float64, size int) float64 {
	fontSize := standardCellFontSize(cellSize, 0.68)
	switch {
	case size >= 14:
		fontSize *= 0.94
	case size >= 12:
		fontSize *= 0.97
	}
	return clampStandardCellFontSize(fontSize)
}
