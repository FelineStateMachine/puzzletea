package spellpuzzle

import (
	"encoding/json"
	"fmt"
	"strings"

	"codeberg.org/go-pdf/fpdf"
	"github.com/FelineStateMachine/puzzletea/export/pdf"
)

type printAdapter struct{}

var PDFPrintAdapter = printAdapter{}

type printPayload struct {
	Board      board
	Placements []WordPlacement
	Bank       []rune
}

type printLayout struct {
	cellSize float64
	boardX   float64
	boardY   float64
	tileSize float64
	tileGap  float64
	bankX    float64
	bankY    float64
}

func (printAdapter) CanonicalGameType() string { return "Spell Puzzle" }
func (printAdapter) Aliases() []string {
	return []string{"spell", "spellpuzzle"}
}

func (printAdapter) BuildPDFPayload(save []byte) (any, error) {
	if len(strings.TrimSpace(string(save))) == 0 {
		return nil, nil
	}

	var exported Save
	if err := json.Unmarshal(save, &exported); err != nil {
		return nil, fmt.Errorf("decode spell puzzle save: %w", err)
	}

	placements := append([]WordPlacement(nil), exported.Placements...)
	for i := range placements {
		placements[i].Found = false
	}
	placements = normalizePlacements(placements)
	if len(placements) == 0 {
		return nil, nil
	}

	board := buildBoard(placements)
	if board.Width <= 0 || board.Height <= 0 {
		return nil, nil
	}

	return &printPayload{
		Board:      board,
		Placements: placements,
		Bank:       []rune(exported.Bank),
	}, nil
}

func (printAdapter) RenderPDFBody(pdf *fpdf.Fpdf, payload any) error {
	switch data := payload.(type) {
	case *printPayload:
		renderSpellPuzzlePage(pdf, data)
	}
	return nil
}

func renderSpellPuzzlePage(pdf *fpdf.Fpdf, data *printPayload) {
	if data == nil || data.Board.Width <= 0 || data.Board.Height <= 0 {
		return
	}

	pageW, pageH := pdf.GetPageSize()
	pageNo := pdf.PageNo()
	body := pdfexport.PuzzleBodyRect(pageW, pageH, pageNo)
	rules := []string{"Form words from the letter bank to fill the crossword"}
	ruleLines := pdfexport.InstructionLineCount(pdf, body.W, rules)
	area := pdfexport.PuzzleBoardRect(pageW, pageH, pageNo, ruleLines)
	layout, ok := computePrintLayout(area, data)
	if !ok {
		return
	}

	drawSpellPuzzleBoard(pdf, data.Board, layout)
	drawSpellPuzzleBank(pdf, data.Bank, layout)

	contentBottom := layout.bankY + layout.tileSize
	ruleY := pdfexport.InstructionY(contentBottom, pageH, ruleLines)
	pdfexport.RenderInstructions(pdf, body.X, ruleY, body.W, rules)
}

func computePrintLayout(area pdfexport.Rect, data *printPayload) (printLayout, bool) {
	if data == nil || data.Board.Width <= 0 || data.Board.Height <= 0 {
		return printLayout{}, false
	}

	letterCount := max(len(data.Bank), 1)
	tileGap := 2.5
	tileSize := min(9.5, (area.W-float64(letterCount-1)*tileGap)/float64(letterCount))
	if tileSize < 5.2 {
		tileGap = 1.4
		tileSize = (area.W - float64(letterCount-1)*tileGap) / float64(letterCount)
	}
	if tileSize <= 0 {
		return printLayout{}, false
	}

	bankSectionHeight := tileSize + 6.0
	boardArea := pdfexport.Rect{
		X: area.X,
		Y: area.Y,
		W: area.W,
		H: area.H - bankSectionHeight - 4.0,
	}
	if boardArea.H <= 0 {
		return printLayout{}, false
	}

	cellSize := pdfexport.FitCompactCellSize(data.Board.Width, data.Board.Height, boardArea)
	if cellSize <= 0 {
		return printLayout{}, false
	}

	boardX, boardY := pdfexport.CenteredOrigin(boardArea, data.Board.Width, data.Board.Height, cellSize)
	bankWidth := float64(len(data.Bank))*tileSize + float64(max(len(data.Bank)-1, 0))*tileGap
	bankX := area.X + (area.W-bankWidth)/2
	bankY := boardArea.Y + boardArea.H + 4.0 + (bankSectionHeight-tileSize)/2

	return printLayout{
		cellSize: cellSize,
		boardX:   boardX,
		boardY:   boardY,
		tileSize: tileSize,
		tileGap:  tileGap,
		bankX:    bankX,
		bankY:    bankY,
	}, true
}

func drawSpellPuzzleBoard(pdf *fpdf.Fpdf, b board, layout printLayout) {
	pdf.SetFillColor(255, 255, 255)
	for y := range b.Height {
		for x := range b.Width {
			cell := b.Cells[y][x]
			if !cell.Occupied {
				continue
			}
			cellX := layout.boardX + float64(x)*layout.cellSize
			cellY := layout.boardY + float64(y)*layout.cellSize
			pdf.Rect(cellX, cellY, layout.cellSize, layout.cellSize, "F")
		}
	}

	pdf.SetDrawColor(150, 150, 150)
	pdf.SetLineWidth(pdfexport.ThinGridLineMM)
	for y := range b.Height {
		for x := range b.Width {
			cell := b.Cells[y][x]
			if !cell.Occupied {
				continue
			}
			cellX := layout.boardX + float64(x)*layout.cellSize
			cellY := layout.boardY + float64(y)*layout.cellSize
			pdf.Rect(cellX, cellY, layout.cellSize, layout.cellSize, "D")
		}
	}

	pdf.SetDrawColor(25, 25, 25)
	pdf.SetLineWidth(pdfexport.MajorGridLineMM)
	for y := range b.Height {
		for x := range b.Width {
			cell := b.Cells[y][x]
			if !cell.Occupied {
				continue
			}

			cellX := layout.boardX + float64(x)*layout.cellSize
			cellY := layout.boardY + float64(y)*layout.cellSize
			if b.hasHorizontalEdge(x, y) {
				pdf.Line(cellX, cellY, cellX+layout.cellSize, cellY)
			}
			if b.hasVerticalEdge(x, y) {
				pdf.Line(cellX, cellY, cellX, cellY+layout.cellSize)
			}
			if b.hasHorizontalEdge(x, y+1) {
				pdf.Line(cellX, cellY+layout.cellSize, cellX+layout.cellSize, cellY+layout.cellSize)
			}
			if b.hasVerticalEdge(x+1, y) {
				pdf.Line(cellX+layout.cellSize, cellY, cellX+layout.cellSize, cellY+layout.cellSize)
			}
		}
	}
}

func drawSpellPuzzleBank(pdf *fpdf.Fpdf, bank []rune, layout printLayout) {
	if len(bank) == 0 {
		return
	}

	fontSize := pdfexport.ClampStandardCellFontSize(pdfexport.StandardCellFontSize(layout.tileSize, 0.56))
	lineH := fontSize * 0.9
	pdf.SetFont(pdfexport.SansFontFamily, "B", fontSize)
	pdf.SetTextColor(pdfexport.PrimaryTextGray, pdfexport.PrimaryTextGray, pdfexport.PrimaryTextGray)
	pdf.SetDrawColor(45, 45, 45)
	pdf.SetLineWidth(pdfexport.ThinGridLineMM)

	for i, letter := range bank {
		x := layout.bankX + float64(i)*(layout.tileSize+layout.tileGap)
		pdf.Rect(x, layout.bankY, layout.tileSize, layout.tileSize, "D")
		pdf.SetXY(x, layout.bankY+(layout.tileSize-lineH)/2)
		pdf.CellFormat(layout.tileSize, lineH, strings.ToUpper(string(letter)), "", 0, "C", false, 0, "")
	}
}
