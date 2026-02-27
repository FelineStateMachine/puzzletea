package shikaku

import (
	"strconv"
	"unicode/utf8"

	"codeberg.org/go-pdf/fpdf"
	"github.com/FelineStateMachine/puzzletea/game"
	"github.com/FelineStateMachine/puzzletea/pdfexport"
)

type printAdapter struct{}

func (printAdapter) CanonicalGameType() string { return "Shikaku" }
func (printAdapter) Aliases() []string         { return []string{"shikaku"} }

func (printAdapter) BuildPDFPayload(save []byte) (any, error) {
	return pdfexport.ParseShikakuPrintData(save)
}

func (printAdapter) RenderPDFBody(pdf *fpdf.Fpdf, payload any) error {
	switch data := payload.(type) {
	case *pdfexport.ShikakuData:
		renderShikakuPage(pdf, data)
	}
	return nil
}

func renderShikakuPage(pdf *fpdf.Fpdf, data *pdfexport.ShikakuData) {
	if data == nil || data.Width <= 0 || data.Height <= 0 {
		return
	}

	pageW, pageH := pdf.GetPageSize()
	pageNo := pdf.PageNo()
	area := pdfexport.PuzzleBoardRect(pageW, pageH, pageNo, 1)
	cellSize := pdfexport.FitCompactCellSize(data.Width, data.Height, area)
	if cellSize <= 0 {
		return
	}

	blockW := float64(data.Width) * cellSize
	blockH := float64(data.Height) * cellSize
	startX, startY := pdfexport.CenteredOrigin(area, data.Width, data.Height, cellSize)

	pdf.SetDrawColor(55, 55, 55)
	pdf.SetLineWidth(pdfexport.ThinGridLineMM)
	for y := 0; y < data.Height; y++ {
		for x := 0; x < data.Width; x++ {
			cellX := startX + float64(x)*cellSize
			cellY := startY + float64(y)*cellSize
			pdf.Rect(cellX, cellY, cellSize, cellSize, "D")

			if y >= len(data.Clues) || x >= len(data.Clues[y]) || data.Clues[y][x] <= 0 {
				continue
			}
			drawShikakuClue(pdf, cellX, cellY, cellSize, data.Clues[y][x])
		}
	}

	pdf.SetDrawColor(35, 35, 35)
	pdf.SetLineWidth(pdfexport.OuterBorderLineMM)
	pdf.Rect(startX, startY, blockW, blockH, "D")

	ruleY := pdfexport.InstructionY(startY+blockH, pageH, 1)
	pdfexport.SetInstructionStyle(pdf)
	pdf.SetXY(area.X, ruleY)
	pdf.CellFormat(
		area.W,
		pdfexport.InstructionLineHMM,
		"Partition into rectangles where each clue equals its rectangle area.",
		"",
		0,
		"C",
		false,
		0,
		"",
	)
}

func drawShikakuClue(pdf *fpdf.Fpdf, x, y, cellSize float64, value int) {
	text := strconv.Itoa(value)
	fontSize := pdfexport.StandardCellFontSize(cellSize, 0.58)
	runeCount := utf8.RuneCountInString(text)
	switch {
	case runeCount >= 3:
		fontSize *= 0.72
	case runeCount == 2:
		fontSize *= 0.84
	}
	fontSize = pdfexport.ClampStandardCellFontSize(fontSize)

	pdf.SetTextColor(pdfexport.PrimaryTextGray, pdfexport.PrimaryTextGray, pdfexport.PrimaryTextGray)
	pdf.SetFont(pdfexport.SansFontFamily, "B", fontSize)
	lineH := fontSize * 0.92
	pdf.SetXY(x, y+(cellSize-lineH)/2)
	pdf.CellFormat(cellSize, lineH, text, "", 0, "C", false, 0, "")
}

func init() {
	game.RegisterPrintAdapter(printAdapter{})
}
