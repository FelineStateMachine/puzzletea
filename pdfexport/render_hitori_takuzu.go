package pdfexport

import (
	"strings"
	"unicode/utf8"

	"codeberg.org/go-pdf/fpdf"
)

func renderHitoriPage(pdf *fpdf.Fpdf, data *HitoriData) {
	if data == nil || data.Size <= 0 {
		return
	}

	size := data.Size
	pageW, pageH := pdf.GetPageSize()
	pageNo := pdf.PageNo()
	area := puzzleBoardRect(pageW, pageH, pageNo, 1)
	cellSize := fitBoardCellSize(size, size, area, boardFamilyCompact)
	if cellSize <= 0 {
		return
	}

	blockW := float64(size) * cellSize
	blockH := float64(size) * cellSize
	startX, startY := centeredOrigin(area, size, size, cellSize)

	pdf.SetDrawColor(55, 55, 55)
	pdf.SetLineWidth(thinGridLineMM)
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
	pdf.SetLineWidth(outerBorderLineMM)
	pdf.Rect(startX, startY, blockW, blockH, "D")

	ruleY := instructionY(startY+blockH, pageH, 1)
	setInstructionStyle(pdf)
	pdf.SetXY(area.x, ruleY)
	pdf.CellFormat(
		area.w,
		instructionLineHMM,
		"Shade duplicates; shaded cells cannot touch orthogonally; unshaded cells stay connected.",
		"",
		0,
		"C",
		false,
		0,
		"",
	)
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

	pdf.SetTextColor(primaryTextGray, primaryTextGray, primaryTextGray)
	pdf.SetFont(sansFontFamily, "B", fontSize)
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
	pageNo := pdf.PageNo()
	area := puzzleBoardRect(pageW, pageH, pageNo, 2)
	cellSize := fitBoardCellSize(size, size, area, boardFamilyCompact)
	if cellSize <= 0 {
		return
	}

	blockW := float64(size) * cellSize
	blockH := float64(size) * cellSize
	startX, startY := centeredOrigin(area, size, size, cellSize)

	pdf.SetDrawColor(60, 60, 60)
	pdf.SetLineWidth(thinGridLineMM)
	for y := 0; y < size; y++ {
		for x := 0; x < size; x++ {
			cellX := startX + float64(x)*cellSize
			cellY := startY + float64(y)*cellSize
			pdf.Rect(cellX, cellY, cellSize, cellSize, "D")
		}
	}

	if data.GroupEveryTwo {
		pdf.SetDrawColor(130, 130, 130)
		pdf.SetLineWidth(majorGridLineMM)
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
	pdf.SetLineWidth(outerBorderLineMM)
	pdf.Rect(startX, startY, blockW, blockH, "D")

	ruleY := instructionY(startY+blockH, pageH, 2)
	setInstructionStyle(pdf)
	pdf.SetXY(area.x, ruleY)
	pdf.CellFormat(
		area.w,
		instructionLineHMM,
		"No three equal adjacent in any row or column.",
		"",
		0,
		"C",
		false,
		0,
		"",
	)
	pdf.SetXY(area.x, ruleY+instructionLineHMM)
	pdf.CellFormat(
		area.w,
		instructionLineHMM,
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

	pdf.SetTextColor(primaryTextGray, primaryTextGray, primaryTextGray)
	pdf.SetFont(sansFontFamily, "B", fontSize)
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
