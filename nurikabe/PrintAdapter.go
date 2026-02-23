package nurikabe

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/FelineStateMachine/puzzletea/game"
	"github.com/FelineStateMachine/puzzletea/pdfexport"
	"github.com/go-pdf/fpdf"
)

type printAdapter struct{}

func (printAdapter) CanonicalGameType() string { return "Nurikabe" }
func (printAdapter) Aliases() []string         { return []string{"nurikabe"} }

func (printAdapter) RenderMarkdownSnippet(save []byte) (string, error) {
	var data Save
	if err := json.Unmarshal(save, &data); err != nil {
		return "", fmt.Errorf("decode nurikabe save: %w", err)
	}

	clues := game.ParseCSVIntGrid(data.Clues, data.Width, data.Height)
	cells := game.MakeStringGrid(data.Width, data.Height, ".")
	for y := range len(cells) {
		for x := range len(cells[y]) {
			if y < len(clues) && x < len(clues[y]) && clues[y][x] > 0 {
				cells[y][x] = strconv.Itoa(clues[y][x])
			}
		}
	}

	var b strings.Builder
	b.WriteString("### Clue Grid\n\n")
	b.WriteString(game.RenderGridTable(cells))
	b.WriteString("\n\n")
	b.WriteString("Goal: build one connected sea while each numbered island has the exact size of its clue.")
	return b.String(), nil
}

func (printAdapter) BuildPDFPayload(save []byte, snippet string) (any, error) {
	payload, err := pdfexport.ParseNurikabePrintData(save)
	if err != nil {
		return nil, err
	}
	if !game.IsNilPrintPayload(payload) {
		return payload, nil
	}
	if strings.TrimSpace(snippet) == "" {
		return nil, nil
	}

	_, table, err := pdfexport.ParsePrintableFromSnippet("Nurikabe", snippet)
	if err != nil {
		return nil, err
	}
	return table, nil
}

func (printAdapter) RenderPDFBody(pdf *fpdf.Fpdf, payload any) error {
	switch data := payload.(type) {
	case *pdfexport.NurikabeData:
		pdfexport.RenderNurikabePage(pdf, data)
	case *pdfexport.GridTable:
		pdfexport.RenderGridTablePage(pdf, data)
	}
	return nil
}

func init() {
	game.RegisterPrintAdapter(printAdapter{})
}
