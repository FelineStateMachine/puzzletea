package hitori

import (
	"strings"
	"unicode/utf8"

	"codeberg.org/go-pdf/fpdf"
	"github.com/FelineStateMachine/puzzletea/pdfexport"
)

type printAdapter struct{}

func (printAdapter) CanonicalGameType() string { return "Hitori" }
func (printAdapter) Aliases() []string         { return []string{"hitori"} }

func (printAdapter) BuildPDFPayload(save []byte) (any, error) {
	return pdfexport.ParseHitoriPrintData(save)
}

func (printAdapter) RenderPDFBody(pdf *fpdf.Fpdf, payload any) error {
	switch data := payload.(type) {
	case *pdfexport.HitoriData:
		renderHitoriPage(pdf, data)
	}
	return nil
}

func renderHitoriPage(pdf *fpdf.Fpdf, data *pdfexport.HitoriData) {
	if data == nil || data.Size <= 0 {
		return
	}

	size := data.Size
	pageW, pageH := pdf.GetPageSize()
	pageNo := pdf.PageNo()
	area := pdfexport.PuzzleBoardRect(pageW, pageH, pageNo, 1)
	cellSize := pdfexport.FitCompactCellSize(size, size, area)
	if cellSize <= 0 {
		return
	}

	blockW := float64(size) * cellSize
	blockH := float64(size) * cellSize
	startX, startY := pdfexport.CenteredOrigin(area, size, size, cellSize)

	pdf.SetDrawColor(55, 55, 55)
	pdf.SetLineWidth(pdfexport.ThinGridLineMM)
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
	pdf.SetLineWidth(pdfexport.OuterBorderLineMM)
	pdf.Rect(startX, startY, blockW, blockH, "D")

	ruleY := pdfexport.InstructionY(startY+blockH, pageH, 1)
	pdfexport.SetInstructionStyle(pdf)
	pdf.SetXY(area.X, ruleY)
	pdf.CellFormat(
		area.W,
		pdfexport.InstructionLineHMM,
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
	fontSize := pdfexport.StandardCellFontSize(cellSize, 0.58)
	runeCount := utf8.RuneCountInString(text)
	switch {
	case runeCount >= 3:
		fontSize *= 0.7
	case runeCount == 2:
		fontSize *= 0.82
	}
	fontSize = pdfexport.ClampStandardCellFontSize(fontSize)

	pdf.SetTextColor(pdfexport.PrimaryTextGray, pdfexport.PrimaryTextGray, pdfexport.PrimaryTextGray)
	pdf.SetFont(pdfexport.SansFontFamily, "B", fontSize)
	lineH := fontSize * 0.92
	pdf.SetXY(x, y+(cellSize-lineH)/2)
	pdf.CellFormat(cellSize, lineH, text, "", 0, "C", false, 0, "")
}

func init() {
	pdfexport.RegisterPrintAdapter(printAdapter{})
}
