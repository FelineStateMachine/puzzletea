package takuzu

import (
	"github.com/FelineStateMachine/puzzletea/game"
	"github.com/FelineStateMachine/puzzletea/pdfexport"
	"github.com/go-pdf/fpdf"
)

type printAdapter struct{}

func (printAdapter) CanonicalGameType() string { return "Takuzu" }
func (printAdapter) Aliases() []string         { return []string{"takuzu"} }

func (printAdapter) BuildPDFPayload(save []byte) (any, error) {
	return pdfexport.ParseTakuzuPrintData(save)
}

func (printAdapter) RenderPDFBody(pdf *fpdf.Fpdf, payload any) error {
	switch data := payload.(type) {
	case *pdfexport.TakuzuData:
		pdfexport.RenderTakuzuPage(pdf, data)
	}
	return nil
}

func init() {
	game.RegisterPrintAdapter(printAdapter{})
}
