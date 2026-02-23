package shikaku

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

func (printAdapter) CanonicalGameType() string { return "Shikaku" }
func (printAdapter) Aliases() []string         { return []string{"shikaku"} }

func (printAdapter) RenderMarkdownSnippet(save []byte) (string, error) {
	var data Save
	if err := json.Unmarshal(save, &data); err != nil {
		return "", fmt.Errorf("decode shikaku save: %w", err)
	}

	cells := game.MakeStringGrid(data.Width, data.Height, ".")
	for _, clue := range data.Clues {
		if clue.Y < 0 || clue.Y >= len(cells) {
			continue
		}
		if clue.X < 0 || clue.X >= len(cells[clue.Y]) {
			continue
		}
		cells[clue.Y][clue.X] = strconv.Itoa(clue.Value)
	}

	var b strings.Builder
	b.WriteString("### Clue Grid\n\n")
	b.WriteString(game.RenderGridTable(cells))
	b.WriteString("\n\n")
	b.WriteString("Goal: partition the grid into rectangles so each rectangle contains one clue and its area matches that clue.")
	return b.String(), nil
}

func (printAdapter) BuildPDFPayload(save []byte, snippet string) (any, error) {
	payload, err := pdfexport.ParseShikakuPrintData(save)
	if err != nil {
		return nil, err
	}
	if !game.IsNilPrintPayload(payload) {
		return payload, nil
	}
	if strings.TrimSpace(snippet) == "" {
		return nil, nil
	}

	_, table, err := pdfexport.ParsePrintableFromSnippet("Shikaku", snippet)
	if err != nil {
		return nil, err
	}
	return table, nil
}

func (printAdapter) RenderPDFBody(pdf *fpdf.Fpdf, payload any) error {
	switch data := payload.(type) {
	case *pdfexport.ShikakuData:
		pdfexport.RenderShikakuPage(pdf, data)
	case *pdfexport.GridTable:
		pdfexport.RenderGridTablePage(pdf, data)
	}
	return nil
}

func init() {
	game.RegisterPrintAdapter(printAdapter{})
}
