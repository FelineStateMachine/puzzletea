package takuzuplus

import (
	"codeberg.org/go-pdf/fpdf"
	"github.com/FelineStateMachine/puzzletea/pdfexport"
	"github.com/FelineStateMachine/puzzletea/takuzu"
)

type printAdapter struct{}

var PDFPrintAdapter = printAdapter{}

var takuzuPlusRules = []string{
	"No three equal adjacent in any row or column.",
	"Each row/column has equal 0 and 1 counts, and rows/columns are unique.",
	"= means equal neighbors; x means opposite neighbors.",
}

func (printAdapter) CanonicalGameType() string { return "Takuzu+" }
func (printAdapter) Aliases() []string         { return []string{"takuzu plus", "binario+", "binario plus"} }

func (printAdapter) BuildPDFPayload(save []byte) (any, error) {
	return pdfexport.ParseTakuzuPlusPrintData(save)
}

func (printAdapter) RenderPDFBody(pdf *fpdf.Fpdf, payload any) error {
	switch data := payload.(type) {
	case *pdfexport.TakuzuData:
		takuzu.RenderTakuzuPDFBodyWithRules(pdf, data, takuzuPlusRules)
	}
	return nil
}
