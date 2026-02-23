package hitori

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/FelineStateMachine/puzzletea/game"
	"github.com/FelineStateMachine/puzzletea/pdfexport"
	"github.com/go-pdf/fpdf"
)

type printAdapter struct{}

func (printAdapter) CanonicalGameType() string { return "Hitori" }
func (printAdapter) Aliases() []string         { return []string{"hitori"} }

func (printAdapter) RenderMarkdownSnippet(save []byte) (string, error) {
	var data Save
	if err := json.Unmarshal(save, &data); err != nil {
		return "", fmt.Errorf("decode hitori save: %w", err)
	}

	rows := game.SplitLines(data.Numbers)
	targetHeight := data.Size
	if targetHeight < len(rows) {
		targetHeight = len(rows)
	}
	if targetHeight < 1 {
		targetHeight = 1
	}

	cells := make([][]string, 0, targetHeight)
	for y := range targetHeight {
		row := []rune{}
		if y < len(rows) {
			row = []rune(rows[y])
		}
		targetWidth := data.Size
		if targetWidth < len(row) {
			targetWidth = len(row)
		}
		if targetWidth < 1 {
			targetWidth = 1
		}

		cellsRow := make([]string, targetWidth)
		for x := range len(cellsRow) {
			cellsRow[x] = "."
			if x < len(row) && row[x] != ' ' {
				cellsRow[x] = string(row[x])
			}
		}
		cells = append(cells, cellsRow)
	}

	var b strings.Builder
	b.WriteString("### Grid\n\n")
	b.WriteString(game.RenderGridTable(cells))
	b.WriteString("\n\n")
	b.WriteString("Goal: shade cells so no row or column has duplicate unshaded values, ")
	b.WriteString("shaded cells do not touch orthogonally, and all unshaded cells stay connected.")
	return b.String(), nil
}

func (printAdapter) BuildPDFPayload(save []byte, snippet string) (any, error) {
	payload, err := pdfexport.ParseHitoriPrintData(save)
	if err != nil {
		return nil, err
	}
	if !game.IsNilPrintPayload(payload) {
		return payload, nil
	}
	if strings.TrimSpace(snippet) == "" {
		return nil, nil
	}

	_, table, err := pdfexport.ParsePrintableFromSnippet("Hitori", snippet)
	if err != nil {
		return nil, err
	}
	return table, nil
}

func (printAdapter) RenderPDFBody(pdf *fpdf.Fpdf, payload any) error {
	switch data := payload.(type) {
	case *pdfexport.HitoriData:
		pdfexport.RenderHitoriPage(pdf, data)
	case *pdfexport.GridTable:
		pdfexport.RenderGridTablePage(pdf, data)
	}
	return nil
}

func init() {
	game.RegisterPrintAdapter(printAdapter{})
}
