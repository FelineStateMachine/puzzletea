package pdfexport

import (
	"math"

	"codeberg.org/go-pdf/fpdf"
)

const (
	halfLetterWidthMM  = 139.7
	halfLetterHeightMM = 215.9
	letterWidthMM      = 279.4
	letterHeightMM     = 215.9

	footerTextGray    = 78
	secondaryTextGray = 60
	ruleTextGray      = 54
	dimTextGray       = 92
	primaryTextGray   = 20

	pageMarginXMM         = 10.0
	bindingGutterExtraMM  = 2.6
	puzzleTopMM           = 28.0
	puzzleBottomInsetMM   = 16.0
	instructionGapMM      = 4.2
	instructionLineHMM    = 4.6
	wordSearchGridListGap = 4.0

	thinGridLineMM    = 0.24
	majorGridLineMM   = 0.34
	outerBorderLineMM = 0.62
)

var logicalPageNumberOverride int

const (
	puzzleTitleFontSize        = 13.0 + pdfFontSizeDelta
	puzzleSubtitleFontSize     = 9.0 + pdfFontSizeDelta
	puzzleInstructionFontSize  = 7.0 + pdfFontSizeDelta
	puzzleWordBankFontSize     = 8.8 + pdfFontSizeDelta
	puzzleWordBankHeadSize     = 9.2 + pdfFontSizeDelta
	difficultyStarSizeMM       = 3.5
	difficultyStarGapMM        = 0.9
	difficultyTextToStarsGapMM = 2.1
	difficultyStarOutlineMM    = 0.26
)

type (
	boardFamily int
	rectMM      struct {
		x float64
		y float64
		w float64
		h float64
	}
)

const (
	boardFamilyCompact boardFamily = iota
	boardFamilySudoku
	boardFamilyHashi
	boardFamilyNonogram
	boardFamilyWordSearch
	boardFamilyTable
)

type boardSizing struct {
	minCell    float64
	maxCell    float64
	targetFill float64
}

func puzzleBodyRect(pageW, pageH float64, pageNo int) rectMM {
	leftMargin, rightMargin := puzzleHorizontalMargins(pageNo)
	return rectMM{
		x: leftMargin,
		y: puzzleTopMM,
		w: pageW - leftMargin - rightMargin,
		h: pageH - puzzleTopMM - puzzleBottomInsetMM,
	}
}

func puzzleBoardRect(pageW, pageH float64, pageNo, ruleLines int) rectMM {
	rect := puzzleBodyRect(pageW, pageH, pageNo)
	if ruleLines > 0 {
		rect.h -= instructionGapMM + float64(ruleLines)*instructionLineHMM
		if rect.h < 0 {
			rect.h = 0
		}
	}
	return rect
}

func puzzleHorizontalMargins(pageNo int) (left, right float64) {
	pageNo = effectiveLogicalPageNumber(pageNo)
	left = pageMarginXMM
	right = pageMarginXMM
	if pageNo <= 1 {
		return left, right
	}
	if pageNo%2 == 0 {
		// Even pages sit on the left side in a spread, so inside edge is right.
		right += bindingGutterExtraMM
		return left, right
	}
	// Odd pages sit on the right side in a spread, so inside edge is left.
	left += bindingGutterExtraMM
	return left, right
}

func effectiveLogicalPageNumber(pageNo int) int {
	if logicalPageNumberOverride > 0 {
		return logicalPageNumberOverride
	}
	return pageNo
}

func withLogicalPageNumber(pageNo int, fn func() error) error {
	previous := logicalPageNumberOverride
	logicalPageNumberOverride = pageNo
	defer func() {
		logicalPageNumberOverride = previous
	}()
	return fn()
}

func centeredOrigin(area rectMM, cols, rows int, cellSize float64) (float64, float64) {
	boardW := float64(cols) * cellSize
	boardH := float64(rows) * cellSize
	return area.x + (area.w-boardW)/2, area.y + (area.h-boardH)/2
}

func instructionY(boardBottom, pageH float64, lineCount int) float64 {
	y := boardBottom + instructionGapMM
	maxY := pageH - puzzleBottomInsetMM - float64(max(lineCount-1, 0))*instructionLineHMM
	if y > maxY {
		return maxY
	}
	return y
}

func fitBoardCellSize(cols, rows int, area rectMM, family boardFamily) float64 {
	if cols <= 0 || rows <= 0 || area.w <= 0 || area.h <= 0 {
		return 0
	}

	sizeCfg := boardSizingFor(family)
	maxFit := math.Min(area.w/float64(cols), area.h/float64(rows))
	size := clampFloat(maxFit, sizeCfg.minCell, sizeCfg.maxCell)
	target := (area.h * sizeCfg.targetFill) / float64(rows)
	if target > size {
		size = math.Min(target, maxFit)
	}

	size = clampFloat(size, sizeCfg.minCell, sizeCfg.maxCell)
	if size > maxFit {
		return maxFit
	}
	return size
}

func boardSizingFor(family boardFamily) boardSizing {
	switch family {
	case boardFamilySudoku:
		return boardSizing{minCell: 10.8, maxCell: 12.8, targetFill: 0.66}
	case boardFamilyHashi:
		return boardSizing{minCell: 6.0, maxCell: 16.2, targetFill: 0.58}
	case boardFamilyNonogram:
		return boardSizing{minCell: 3.6, maxCell: 8.8, targetFill: 0.64}
	case boardFamilyWordSearch:
		return boardSizing{minCell: 4.8, maxCell: 10.2, targetFill: 0.66}
	case boardFamilyTable:
		return boardSizing{minCell: 3.6, maxCell: 11.0, targetFill: 0.60}
	default:
		return boardSizing{minCell: 5.4, maxCell: 13.4, targetFill: 0.58}
	}
}

func clampFloat(v, minV, maxV float64) float64 {
	if v < minV {
		return minV
	}
	if v > maxV {
		return maxV
	}
	return v
}

func setPuzzleTitleStyle(pdf *fpdf.Fpdf) {
	pdf.SetTextColor(primaryTextGray, primaryTextGray, primaryTextGray)
	pdf.SetFont(sansFontFamily, "B", puzzleTitleFontSize)
}

func setPuzzleSubtitleStyle(pdf *fpdf.Fpdf) {
	pdf.SetTextColor(secondaryTextGray, secondaryTextGray, secondaryTextGray)
	pdf.SetFont(sansFontFamily, "", puzzleSubtitleFontSize)
}

func setInstructionStyle(pdf *fpdf.Fpdf) {
	pdf.SetTextColor(ruleTextGray, ruleTextGray, ruleTextGray)
	pdf.SetFont(sansFontFamily, "", puzzleInstructionFontSize)
}
