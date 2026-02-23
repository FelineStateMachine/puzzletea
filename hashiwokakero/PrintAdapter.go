package hashiwokakero

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

func (printAdapter) CanonicalGameType() string { return "Hashiwokakero" }
func (printAdapter) Aliases() []string {
	return []string{"hashi", "hashiwokakero", "hashi wokakero"}
}

func (printAdapter) RenderMarkdownSnippet(save []byte) (string, error) {
	var data Save
	if err := json.Unmarshal(save, &data); err != nil {
		return "", fmt.Errorf("decode hashiwokakero save: %w", err)
	}

	cells := game.MakeStringGrid(data.Width, data.Height, ".")
	for _, island := range data.Islands {
		if island.Y < 0 || island.Y >= len(cells) {
			continue
		}
		if island.X < 0 || island.X >= len(cells[island.Y]) {
			continue
		}
		cells[island.Y][island.X] = strconv.Itoa(island.Required)
	}

	var b strings.Builder
	b.WriteString("### Grid\n\n")
	b.WriteString(game.RenderGridTable(cells))
	b.WriteString("\n\n")
	b.WriteString("Rules: connect numbered islands with horizontal/vertical bridges. ")
	b.WriteString("Use up to two bridges per connection and never cross bridges.")
	return b.String(), nil
}

func (printAdapter) BuildPDFPayload(save []byte, snippet string) (any, error) {
	payload, err := pdfexport.ParseHashiPrintData(save)
	if err != nil {
		return nil, err
	}
	if !game.IsNilPrintPayload(payload) {
		return payload, nil
	}
	if strings.TrimSpace(snippet) == "" {
		return nil, nil
	}

	_, table, err := pdfexport.ParsePrintableFromSnippet("Hashiwokakero", snippet)
	if err != nil {
		return nil, err
	}
	return table, nil
}

func (printAdapter) RenderPDFBody(pdf *fpdf.Fpdf, payload any) error {
	switch data := payload.(type) {
	case *pdfexport.HashiData:
		pdfexport.RenderHashiPage(pdf, data)
	case *pdfexport.GridTable:
		pdfexport.RenderGridTablePage(pdf, data)
	}
	return nil
}

func init() {
	game.RegisterPrintAdapter(printAdapter{})
}
