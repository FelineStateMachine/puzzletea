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
	"Rows/columns have equal 0s and 1s, and all rows/columns are unique.",
	"= means same; x means different.",
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
