package takuzu

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/FelineStateMachine/puzzletea/game"
	"github.com/FelineStateMachine/puzzletea/pdfexport"
	"github.com/go-pdf/fpdf"
)

type printAdapter struct{}

func (printAdapter) CanonicalGameType() string { return "Takuzu" }
func (printAdapter) Aliases() []string         { return []string{"takuzu"} }

func (printAdapter) RenderMarkdownSnippet(save []byte) (string, error) {
	var data Save
	if err := json.Unmarshal(save, &data); err != nil {
		return "", fmt.Errorf("decode takuzu save: %w", err)
	}

	cells := game.MakeStringGrid(data.Size, data.Size, ".")
	stateRows := game.SplitLines(data.State)
	providedRows := game.SplitLines(data.Provided)
	for y := range len(cells) {
		for x := range len(cells[y]) {
			provided := y < len(providedRows) && x < len(providedRows[y]) && providedRows[y][x] == '#'
			if !provided {
				continue
			}
			if y >= len(stateRows) || x >= len(stateRows[y]) {
				continue
			}
			switch stateRows[y][x] {
			case '0', '1':
				cells[y][x] = string(stateRows[y][x])
			}
		}
	}

	var b strings.Builder
	b.WriteString("### Given Grid\n\n")
	b.WriteString(game.RenderGridTable(cells))
	b.WriteString("\n\n")
	b.WriteString("Goal: fill with 0/1 so no three equal adjacent cells appear, ")
	b.WriteString("each row/column has equal 0 and 1 counts, and rows/columns are unique.")
	return b.String(), nil
}

func (printAdapter) BuildPDFPayload(save []byte, snippet string) (any, error) {
	payload, err := pdfexport.ParseTakuzuPrintData(save)
	if err != nil {
		return nil, err
	}
	if !game.IsNilPrintPayload(payload) {
		return payload, nil
	}
	if strings.TrimSpace(snippet) == "" {
		return nil, nil
	}

	_, table, err := pdfexport.ParsePrintableFromSnippet("Takuzu", snippet)
	if err != nil {
		return nil, err
	}
	return table, nil
}

func (printAdapter) RenderPDFBody(pdf *fpdf.Fpdf, payload any) error {
	switch data := payload.(type) {
	case *pdfexport.TakuzuData:
		pdfexport.RenderTakuzuPage(pdf, data)
	case *pdfexport.GridTable:
		pdfexport.RenderGridTablePage(pdf, data)
	}
	return nil
}

func init() {
	game.RegisterPrintAdapter(printAdapter{})
}
