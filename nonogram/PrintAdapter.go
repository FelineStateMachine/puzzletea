package nonogram

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/FelineStateMachine/puzzletea/game"
	"github.com/FelineStateMachine/puzzletea/pdfexport"
	"github.com/go-pdf/fpdf"
)

type printAdapter struct{}

func (printAdapter) CanonicalGameType() string { return "Nonogram" }
func (printAdapter) Aliases() []string         { return []string{"nonogram"} }

func (printAdapter) RenderMarkdownSnippet(save []byte) (string, error) {
	var data Save
	if err := json.Unmarshal(save, &data); err != nil {
		return "", fmt.Errorf("decode nonogram save: %w", err)
	}

	width := data.Width
	height := data.Height
	if width <= 0 {
		width = len(data.ColHints)
	}
	if height <= 0 {
		height = len(data.RowHints)
	}
	if width <= 0 || height <= 0 {
		return "### Puzzle Grid with Integrated Hints\n\n_(empty grid)_", nil
	}

	rowHints := game.NormalizeNonogramHints(data.RowHints, height)
	colHints := game.NormalizeNonogramHints(data.ColHints, width)
	rowHintCols := game.MaxNonogramHintLen(rowHints)
	colHintRows := game.MaxNonogramHintLen(colHints)
	if rowHintCols < 1 {
		rowHintCols = 1
	}
	if colHintRows < 1 {
		colHintRows = 1
	}

	var b strings.Builder
	b.WriteString("### Puzzle Grid with Integrated Hints\n\n")
	b.WriteString(game.RenderNonogramTable(rowHints, colHints, width, height, rowHintCols, colHintRows))
	b.WriteString("\n\n")
	b.WriteString("Row hints are right-aligned beside each row. ")
	b.WriteString("Column hints are stacked above each column and bottom-aligned to the grid.")
	return b.String(), nil
}

func (printAdapter) BuildPDFPayload(save []byte, snippet string) (any, error) {
	payload, err := pdfexport.ParseNonogramPrintData(save)
	if err != nil {
		return nil, err
	}
	if !game.IsNilPrintPayload(payload) {
		return payload, nil
	}
	if strings.TrimSpace(snippet) == "" {
		return nil, nil
	}

	nonogram, _, err := pdfexport.ParsePrintableFromSnippet("Nonogram", snippet)
	if err != nil {
		return nil, err
	}
	return nonogram, nil
}

func (printAdapter) RenderPDFBody(pdf *fpdf.Fpdf, payload any) error {
	switch data := payload.(type) {
	case *pdfexport.NonogramData:
		pdfexport.RenderNonogramPage(pdf, data)
	case *pdfexport.GridTable:
		pdfexport.RenderGridTablePage(pdf, data)
	}
	return nil
}

func init() {
	game.RegisterPrintAdapter(printAdapter{})
}
