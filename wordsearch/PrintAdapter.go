package wordsearch

import (
	"github.com/FelineStateMachine/puzzletea/game"
	"github.com/FelineStateMachine/puzzletea/pdfexport"
	"github.com/go-pdf/fpdf"
)

type printAdapter struct{}

func (printAdapter) CanonicalGameType() string { return "Word Search" }
func (printAdapter) Aliases() []string {
	return []string{"word search", "wordsearch"}
}

func (printAdapter) BuildPDFPayload(save []byte) (any, error) {
	return pdfexport.ParseWordSearchPrintData(save)
}

func (printAdapter) RenderPDFBody(pdf *fpdf.Fpdf, payload any) error {
	switch data := payload.(type) {
	case *pdfexport.WordSearchData:
		pdfexport.RenderWordSearchPage(pdf, data)
	}
	return nil
}

func init() {
	game.RegisterPrintAdapter(printAdapter{})
}
