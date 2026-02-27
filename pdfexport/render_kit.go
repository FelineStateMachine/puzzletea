package pdfexport

import "codeberg.org/go-pdf/fpdf"

type Rect struct {
	X float64
	Y float64
	W float64
	H float64
}

const (
	SansFontFamily = sansFontFamily

	PrimaryTextGray   = primaryTextGray
	SecondaryTextGray = secondaryTextGray
	RuleTextGray      = ruleTextGray
	DimTextGray       = dimTextGray

	ThinGridLineMM    = thinGridLineMM
	MajorGridLineMM   = majorGridLineMM
	OuterBorderLineMM = outerBorderLineMM

	InstructionLineHMM = instructionLineHMM

	PuzzleBottomInsetMM = puzzleBottomInsetMM
	WordSearchGridGapMM = wordSearchGridListGap

	PuzzleWordBankFontSize = puzzleWordBankFontSize
	PuzzleWordBankHeadSize = puzzleWordBankHeadSize
)

func PuzzleBodyRect(pageW, pageH float64, pageNo int) Rect {
	area := puzzleBodyRect(pageW, pageH, pageNo)
	return Rect{X: area.x, Y: area.y, W: area.w, H: area.h}
}

func PuzzleBoardRect(pageW, pageH float64, pageNo, ruleLines int) Rect {
	area := puzzleBoardRect(pageW, pageH, pageNo, ruleLines)
	return Rect{X: area.x, Y: area.y, W: area.w, H: area.h}
}

func CenteredOrigin(area Rect, cols, rows int, cellSize float64) (float64, float64) {
	internal := rectMM{x: area.X, y: area.Y, w: area.W, h: area.H}
	return centeredOrigin(internal, cols, rows, cellSize)
}

func InstructionY(boardBottom, pageH float64, lineCount int) float64 {
	return instructionY(boardBottom, pageH, lineCount)
}

func FitCompactCellSize(cols, rows int, area Rect) float64 {
	internal := rectMM{x: area.X, y: area.Y, w: area.W, h: area.H}
	return fitBoardCellSize(cols, rows, internal, boardFamilyCompact)
}

func FitSudokuCellSize(cols, rows int, area Rect) float64 {
	internal := rectMM{x: area.X, y: area.Y, w: area.W, h: area.H}
	return fitBoardCellSize(cols, rows, internal, boardFamilySudoku)
}

func FitHashiCellSize(cols, rows int, area Rect) float64 {
	internal := rectMM{x: area.X, y: area.Y, w: area.W, h: area.H}
	return fitBoardCellSize(cols, rows, internal, boardFamilyHashi)
}

func FitNonogramCellSize(cols, rows int, area Rect) float64 {
	internal := rectMM{x: area.X, y: area.Y, w: area.W, h: area.H}
	return fitBoardCellSize(cols, rows, internal, boardFamilyNonogram)
}

func FitWordSearchCellSize(cols, rows int, area Rect) float64 {
	internal := rectMM{x: area.X, y: area.Y, w: area.W, h: area.H}
	return fitBoardCellSize(cols, rows, internal, boardFamilyWordSearch)
}

func FitTableCellSize(cols, rows int, area Rect) float64 {
	internal := rectMM{x: area.X, y: area.Y, w: area.W, h: area.H}
	return fitBoardCellSize(cols, rows, internal, boardFamilyTable)
}

func StandardCellFontSize(cellSize, scale float64) float64 {
	return standardCellFontSize(cellSize, scale)
}

func ClampStandardCellFontSize(fontSize float64) float64 {
	return clampStandardCellFontSize(fontSize)
}

func SetInstructionStyle(pdf *fpdf.Fpdf) {
	setInstructionStyle(pdf)
}
