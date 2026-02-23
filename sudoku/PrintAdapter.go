package sudoku

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/FelineStateMachine/puzzletea/game"
	"github.com/FelineStateMachine/puzzletea/pdfexport"
	"github.com/go-pdf/fpdf"
)

type printAdapter struct{}

func (printAdapter) CanonicalGameType() string { return "Sudoku" }
func (printAdapter) Aliases() []string         { return []string{"sudoku"} }

func (printAdapter) RenderMarkdownSnippet(save []byte) (string, error) {
	var data Save
	if err := json.Unmarshal(save, &data); err != nil {
		return "", fmt.Errorf("decode sudoku save: %w", err)
	}

	cells := game.MakeStringGrid(9, 9, ".")
	for _, provided := range data.Provided {
		if provided.Y < 0 || provided.Y >= len(cells) {
			continue
		}
		if provided.X < 0 || provided.X >= len(cells[provided.Y]) {
			continue
		}
		cells[provided.Y][provided.X] = fmt.Sprintf("%d", provided.V)
	}

	var b strings.Builder
	b.WriteString("### Given Grid\n\n")
	b.WriteString(game.RenderGridTable(cells))
	b.WriteString("\n\n")
	b.WriteString("Goal: fill each row, column, and 3x3 box with digits 1-9 exactly once.")
	return b.String(), nil
}

func (printAdapter) BuildPDFPayload(save []byte, snippet string) (any, error) {
	payload, err := pdfexport.ParseSudokuPrintData(save)
	if err != nil {
		return nil, err
	}
	if !game.IsNilPrintPayload(payload) {
		return payload, nil
	}
	if strings.TrimSpace(snippet) == "" {
		return nil, nil
	}

	_, table, err := pdfexport.ParsePrintableFromSnippet("Sudoku", snippet)
	if err != nil {
		return nil, err
	}
	return table, nil
}

func (printAdapter) RenderPDFBody(pdf *fpdf.Fpdf, payload any) error {
	switch data := payload.(type) {
	case *pdfexport.SudokuData:
		pdfexport.RenderSudokuPage(pdf, data)
	case *pdfexport.GridTable:
		pdfexport.RenderGridTablePage(pdf, data)
	}
	return nil
}

func init() {
	game.RegisterPrintAdapter(printAdapter{})
}
