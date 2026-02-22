package pdfexport

import (
	_ "embed"
	"fmt"

	"github.com/go-pdf/fpdf"
)

const (
	sansFontFamily  = "AtkinsonHyperlegibleNext"
	coverFontFamily = "DMSerifDisplay"
)

var (
	//go:embed fonts/AtkinsonHyperlegibleNext-Regular.ttf
	atkinsonRegularTTF []byte

	//go:embed fonts/AtkinsonHyperlegibleNext-Bold.ttf
	atkinsonBoldTTF []byte

	//go:embed fonts/DMSerifDisplay-Regular.ttf
	dmSerifDisplayTTF []byte
)

func registerPDFFonts(pdf *fpdf.Fpdf) error {
	pdf.AddUTF8FontFromBytes(sansFontFamily, "", atkinsonRegularTTF)
	pdf.AddUTF8FontFromBytes(sansFontFamily, "B", atkinsonBoldTTF)
	pdf.AddUTF8FontFromBytes(coverFontFamily, "", dmSerifDisplayTTF)
	if pdf.Err() {
		return fmt.Errorf("register pdf fonts: %w", pdf.Error())
	}
	return nil
}
