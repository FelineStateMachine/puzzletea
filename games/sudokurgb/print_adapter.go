package sudokurgb

import (
	"strconv"

	"codeberg.org/go-pdf/fpdf"
	"github.com/FelineStateMachine/puzzletea/export/pdf"
)

type printAdapter struct{}

var PDFPrintAdapter = printAdapter{}

func (printAdapter) CanonicalGameType() string { return "Sudoku RGB" }
func (printAdapter) Aliases() []string {
	return []string{"sudoku rgb", "rgb sudoku", "ripeto", "sudoku ripeto"}
}

func (printAdapter) BuildPDFPayload(save []byte) (any, error) {
	return pdfexport.ParseSudokuRGBPrintData(save)
}

func (printAdapter) RenderPDFBody(pdf *fpdf.Fpdf, payload any) error {
	switch data := payload.(type) {
	case *pdfexport.SudokuData:
		renderSudokuRGBPage(pdf, data)
	}
	return nil
}

func renderSudokuRGBPage(pdf *fpdf.Fpdf, data *pdfexport.SudokuData) {
	if data == nil {
		return
	}

	pageW, pageH := pdf.GetPageSize()
	pageNo := pdf.PageNo()
	body := pdfexport.PuzzleBodyRect(pageW, pageH, pageNo)
	rules := []string{"Fill rows, columns, and 3x3 boxes with three 1s, three 2s, and three 3s"}
	ruleLines := pdfexport.InstructionLineCount(pdf, body.W, rules)
	area := pdfexport.PuzzleBoardRect(pageW, pageH, pageNo, ruleLines)
	cellSize := pdfexport.FitSudokuCellSize(9, 9, area)
	if cellSize <= 0 {
		return
	}

	boardH := 9.0 * cellSize
	startX, startY := pdfexport.CenteredOrigin(area, 9, 9, cellSize)

	drawSudokuGridLines(pdf, startX, startY, cellSize)
	drawSudokuRGBGivens(pdf, startX, startY, cellSize, data.Givens)

	ruleY := pdfexport.InstructionY(startY+boardH, pageH, ruleLines)
	pdfexport.RenderInstructions(pdf, body.X, ruleY, body.W, rules)
}

func drawSudokuGridLines(pdf *fpdf.Fpdf, startX, startY, cellSize float64) {
	pdf.SetDrawColor(25, 25, 25)

	for i := range 10 {
		x := startX + float64(i)*cellSize
		pdf.SetLineWidth(sudokuLineWidthFor(i))
		pdf.Line(x, startY, x, startY+9.0*cellSize)
	}

	for i := range 10 {
		y := startY + float64(i)*cellSize
		pdf.SetLineWidth(sudokuLineWidthFor(i))
		pdf.Line(startX, y, startX+9.0*cellSize, y)
	}
}

func sudokuLineWidthFor(index int) float64 {
	switch {
	case index == 0 || index == 9:
		return 0.72
	case index%3 == 0:
		return 0.56
	default:
		return pdfexport.ThinGridLineMM
	}
}

func drawSudokuRGBGivens(pdf *fpdf.Fpdf, startX, startY, cellSize float64, givens [9][9]int) {
	fontSize := pdfexport.StandardCellFontSize(cellSize, 0.62)
	lineH := fontSize * 0.85
	pdf.SetFont(pdfexport.SansFontFamily, "B", fontSize)
	pdf.SetTextColor(pdfexport.PrimaryTextGray, pdfexport.PrimaryTextGray, pdfexport.PrimaryTextGray)

	for y := range 9 {
		for x := range 9 {
			value := givens[y][x]
			if value < 1 || value > valueCount {
				continue
			}
			cellX := startX + float64(x)*cellSize
			cellY := startY + float64(y)*cellSize
			pdf.SetXY(cellX, cellY+(cellSize-lineH)/2)
			pdf.CellFormat(cellSize, lineH, strconv.Itoa(value), "", 0, "C", false, 0, "")
		}
	}
}
