package hitori

import (
	"github.com/FelineStateMachine/puzzletea/game"
	"github.com/FelineStateMachine/puzzletea/pdfexport"
	"github.com/go-pdf/fpdf"
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
		pdfexport.RenderHitoriPage(pdf, data)
	}
	return nil
}

func init() {
	game.RegisterPrintAdapter(printAdapter{})
}
