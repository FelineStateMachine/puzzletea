package sudoku

import (
	"github.com/FelineStateMachine/puzzletea/game"
	"github.com/FelineStateMachine/puzzletea/pdfexport"
	"github.com/go-pdf/fpdf"
)

type printAdapter struct{}

func (printAdapter) CanonicalGameType() string { return "Sudoku" }
func (printAdapter) Aliases() []string         { return []string{"sudoku"} }

func (printAdapter) BuildPDFPayload(save []byte) (any, error) {
	return pdfexport.ParseSudokuPrintData(save)
}

func (printAdapter) RenderPDFBody(pdf *fpdf.Fpdf, payload any) error {
	switch data := payload.(type) {
	case *pdfexport.SudokuData:
		pdfexport.RenderSudokuPage(pdf, data)
	}
	return nil
}

func init() {
	game.RegisterPrintAdapter(printAdapter{})
}
