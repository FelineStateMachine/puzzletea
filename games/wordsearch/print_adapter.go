package wordsearch

import (
	"strings"

	"codeberg.org/go-pdf/fpdf"
	"github.com/FelineStateMachine/puzzletea/export/pdf"
)

type printAdapter struct{}

var PDFPrintAdapter = printAdapter{}

const (
	wordBankHeaderHeight = 9.0
	wordBankListTopGap   = 2.2
)

func (printAdapter) CanonicalGameType() string { return "Word Search" }
func (printAdapter) Aliases() []string {
	return []string{"word search", "wordsearch"}
}

func (printAdapter) BuildPDFPayload(save []byte) (any, error) {
	return pdfexport.ParseWordSearchPrintData(save)
}

func (printAdapter) RenderPDFBody(pdf *fpdf.Fpdf, payload any) error {
	switch data := payload.(type) {
	case *pdfexport.WordSearchData:
		renderWordSearchPage(pdf, data)
	}
	return nil
}

func renderWordSearchPage(pdf *fpdf.Fpdf, data *pdfexport.WordSearchData) {
	if data == nil || data.Width <= 0 || data.Height <= 0 {
		return
	}

	pageW, pageH := pdf.GetPageSize()
	pageNo := pdf.PageNo()
	body := pdfexport.PuzzleBodyRect(pageW, pageH, pageNo)
	availW := body.W
	availH := body.H

	columnCount := wordSearchColumnCount(data.Width, len(data.Words))
	wordFontSize := pdfexport.PuzzleWordBankFontSize
	wordLineHeight := 4.2
	gridListGap := pdfexport.WordSearchGridGapMM

	estimatedWordLines := estimateWordBankLineCount(pdf, data.Words, columnCount, availW, wordFontSize)
	wordBankHeight := wordBankHeaderHeight + wordBankListTopGap + float64(estimatedWordLines)*wordLineHeight
	maxWordBankHeight := availH * 0.42
	if wordBankHeight > maxWordBankHeight {
		wordBankHeight = maxWordBankHeight
	}
	if wordBankHeight < 16 {
		wordBankHeight = 16
	}

	gridAreaH := availH - wordBankHeight - gridListGap
	if gridAreaH < availH*0.5 {
		gridAreaH = availH * 0.5
	}

	gridArea := pdfexport.Rect{X: body.X, Y: body.Y, W: availW, H: gridAreaH}
	cellSize := pdfexport.FitWordSearchCellSize(data.Width, data.Height, gridArea)
	if cellSize <= 0 {
		return
	}

	gridW := float64(data.Width) * cellSize
	gridH := float64(data.Height) * cellSize
	gridX := body.X + (gridArea.W-gridW)/2
	gridY := body.Y + (gridArea.H-gridH)/2

	drawWordSearchGrid(pdf, data, gridX, gridY, cellSize)
	drawWordBank(
		pdf,
		data.Words,
		body.X,
		gridY+gridH+gridListGap,
		availW,
		pageH-pdfexport.PuzzleBottomInsetMM-(gridY+gridH+gridListGap),
		columnCount,
	)
}

func drawWordSearchGrid(pdf *fpdf.Fpdf, data *pdfexport.WordSearchData, startX, startY, cellSize float64) {
	pdf.SetDrawColor(45, 45, 45)
	pdf.SetLineWidth(pdfexport.ThinGridLineMM)
	for y := range data.Height {
		for x := range data.Width {
			cellX := startX + float64(x)*cellSize
			cellY := startY + float64(y)*cellSize
			pdf.Rect(cellX, cellY, cellSize, cellSize, "D")

			cellText := " "
			if y < len(data.Grid) && x < len(data.Grid[y]) {
				cellText = strings.TrimSpace(strings.ToUpper(data.Grid[y][x]))
			}
			if cellText == "" || cellText == "." {
				continue
			}

			fontSize := pdfexport.StandardCellFontSize(cellSize, 0.74)
			lineH := fontSize * 0.86
			pdf.SetFont(pdfexport.SansFontFamily, "B", fontSize)
			pdf.SetTextColor(pdfexport.PrimaryTextGray, pdfexport.PrimaryTextGray, pdfexport.PrimaryTextGray)
			pdf.SetXY(cellX, cellY+(cellSize-lineH)/2)
			pdf.CellFormat(cellSize, lineH, cellText, "", 0, "C", false, 0, "")
		}
	}

	pdf.SetLineWidth(pdfexport.OuterBorderLineMM)
	pdf.Rect(startX, startY, float64(data.Width)*cellSize, float64(data.Height)*cellSize, "D")
}

func drawWordBank(pdf *fpdf.Fpdf, words []string, x, y, width, height float64, columns int) {
	if height <= 0 {
		return
	}

	pdf.SetTextColor(40, 40, 40)
	pdf.SetFont(pdfexport.SansFontFamily, "B", pdfexport.PuzzleWordBankHeadSize)
	pdf.SetXY(x, y)
	pdf.CellFormat(width, 4.8, "Word Bank", "", 0, "L", false, 0, "")

	pdf.SetFont(pdfexport.SansFontFamily, "", pdfexport.PuzzleWordBankFontSize)
	pdf.SetTextColor(pdfexport.RuleTextGray, pdfexport.RuleTextGray, pdfexport.RuleTextGray)
	pdf.SetXY(x, y+4.8)
	pdf.CellFormat(width, 4.2, "Words may run in all 8 directions", "", 0, "L", false, 0, "")

	listY := y + wordBankHeaderHeight + wordBankListTopGap
	if len(words) == 0 {
		pdf.SetFont(pdfexport.SansFontFamily, "", pdfexport.PuzzleWordBankHeadSize)
		pdf.SetTextColor(pdfexport.SecondaryTextGray, pdfexport.SecondaryTextGray, pdfexport.SecondaryTextGray)
		pdf.SetXY(x, listY)
		pdf.CellFormat(width, 4.6, "(word list unavailable)", "", 0, "L", false, 0, "")
		return
	}

	columnGap := 4.0
	if columns < 1 {
		columns = 1
	}
	colWidth := (width - float64(columns-1)*columnGap) / float64(columns)
	if colWidth <= 0 {
		colWidth = width
		columns = 1
	}

	colLines := layoutWordBankColumns(pdf, words, columns, colWidth)
	lineHeight := 4.1
	maxLines := int((height - (listY - y)) / lineHeight)
	if maxLines <= 0 {
		return
	}

	pdf.SetTextColor(pdfexport.PrimaryTextGray, pdfexport.PrimaryTextGray, pdfexport.PrimaryTextGray)
	pdf.SetFont(pdfexport.SansFontFamily, "B", pdfexport.PuzzleWordBankHeadSize)
	for c := range columns {
		colX := x + float64(c)*(colWidth+columnGap)
		curY := listY
		lines := colLines[c]
		for _, line := range lines {
			if int((curY-listY)/lineHeight) >= maxLines {
				break
			}
			pdf.SetXY(colX, curY)
			pdf.CellFormat(colWidth, lineHeight, line, "", 0, "L", false, 0, "")
			curY += lineHeight
		}
	}
}

func wordSearchColumnCount(width, wordCount int) int {
	if width >= 20 || wordCount > 18 {
		return 3
	}
	if wordCount <= 0 {
		return 1
	}
	return 2
}

func estimateWordBankLineCount(pdf *fpdf.Fpdf, words []string, columns int, availW, fontSize float64) int {
	if len(words) == 0 {
		return 1
	}
	if columns < 1 {
		columns = 1
	}

	colWidth := (availW - float64(columns-1)*4.0) / float64(columns)
	if colWidth <= 0 {
		colWidth = availW
	}

	pdf.SetFont(pdfexport.SansFontFamily, "", fontSize)
	lineCounts := make([]int, columns)
	for _, word := range words {
		text := strings.ToUpper(strings.TrimSpace(word))
		if text == "" {
			continue
		}
		lines := pdf.SplitLines([]byte(text), colWidth)
		if len(lines) == 0 {
			lines = [][]byte{[]byte(text)}
		}
		idx := minLineCountColumn(lineCounts)
		lineCounts[idx] += len(lines)
	}

	maxLines := 0
	for _, lineCount := range lineCounts {
		if lineCount > maxLines {
			maxLines = lineCount
		}
	}
	if maxLines == 0 {
		return 1
	}
	return maxLines
}

func layoutWordBankColumns(pdf *fpdf.Fpdf, words []string, columns int, colWidth float64) [][]string {
	colLines := make([][]string, columns)
	if len(words) == 0 || columns <= 0 {
		return colLines
	}

	lineCounts := make([]int, columns)
	for _, word := range words {
		text := strings.ToUpper(strings.TrimSpace(word))
		if text == "" {
			continue
		}
		wrapped := pdf.SplitLines([]byte(text), colWidth)
		if len(wrapped) == 0 {
			wrapped = [][]byte{[]byte(text)}
		}

		idx := minLineCountColumn(lineCounts)
		for _, line := range wrapped {
			colLines[idx] = append(colLines[idx], string(line))
		}
		lineCounts[idx] += len(wrapped)
	}

	return colLines
}

func minLineCountColumn(lineCounts []int) int {
	idx := 0
	for i := 1; i < len(lineCounts); i++ {
		if lineCounts[i] < lineCounts[idx] {
			idx = i
		}
	}
	return idx
}
