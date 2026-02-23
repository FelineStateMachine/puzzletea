package nonogram

import (
	"github.com/FelineStateMachine/puzzletea/game"
	"github.com/FelineStateMachine/puzzletea/pdfexport"
	"github.com/go-pdf/fpdf"
)

type printAdapter struct{}

func (printAdapter) CanonicalGameType() string { return "Nonogram" }
func (printAdapter) Aliases() []string         { return []string{"nonogram"} }

func (printAdapter) BuildPDFPayload(save []byte) (any, error) {
	return pdfexport.ParseNonogramPrintData(save)
}

func (printAdapter) RenderPDFBody(pdf *fpdf.Fpdf, payload any) error {
	switch data := payload.(type) {
	case *pdfexport.NonogramData:
		pdfexport.RenderNonogramPage(pdf, data)
	}
	return nil
}

func init() {
	game.RegisterPrintAdapter(printAdapter{})
}
