package wordsearch

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/FelineStateMachine/puzzletea/game"
	"github.com/FelineStateMachine/puzzletea/pdfexport"
	"github.com/go-pdf/fpdf"
)

type printAdapter struct{}

func (printAdapter) CanonicalGameType() string { return "Word Search" }
func (printAdapter) Aliases() []string {
	return []string{"word search", "wordsearch"}
}

func (printAdapter) RenderMarkdownSnippet(save []byte) (string, error) {
	var data Save
	if err := json.Unmarshal(save, &data); err != nil {
		return "", fmt.Errorf("decode word search save: %w", err)
	}

	rows := game.SplitLines(data.Grid)
	height := data.Height
	if height < len(rows) {
		height = len(rows)
	}
	width := data.Width
	for _, row := range rows {
		if n := len([]rune(row)); n > width {
			width = n
		}
	}

	cells := game.MakeStringGrid(width, height, ".")
	for y := range len(cells) {
		if y >= len(rows) {
			continue
		}
		rowRunes := []rune(rows[y])
		for x := range len(cells[y]) {
			if x < len(rowRunes) {
				cells[y][x] = string(rowRunes[x])
			}
		}
	}

	var b strings.Builder
	b.WriteString("### Grid\n\n")
	b.WriteString(game.RenderGridTable(cells))
	b.WriteString("\n\n")
	b.WriteString("### Word List\n\n")
	b.WriteString("| # | Word |\n")
	b.WriteString("| --- | --- |\n")
	for i, word := range data.Words {
		fmt.Fprintf(&b, "| %d | %s |\n", i+1, game.EscapeMarkdownCell(word.Text))
	}
	if len(data.Words) == 0 {
		b.WriteString("| 1 | (none) |\n")
	}
	b.WriteString("\nGoal: find all listed words in the grid.")
	return b.String(), nil
}

func (printAdapter) BuildPDFPayload(save []byte, snippet string) (any, error) {
	payload, err := pdfexport.ParseWordSearchPrintData(save)
	if err != nil {
		return nil, err
	}
	if !game.IsNilPrintPayload(payload) {
		return payload, nil
	}
	if strings.TrimSpace(snippet) == "" {
		return nil, nil
	}

	_, table, err := pdfexport.ParsePrintableFromSnippet("Word Search", snippet)
	if err != nil {
		return nil, err
	}
	return table, nil
}

func (printAdapter) RenderPDFBody(pdf *fpdf.Fpdf, payload any) error {
	switch data := payload.(type) {
	case *pdfexport.WordSearchData:
		pdfexport.RenderWordSearchPage(pdf, data)
	case *pdfexport.GridTable:
		pdfexport.RenderGridTablePage(pdf, data)
	}
	return nil
}

func init() {
	game.RegisterPrintAdapter(printAdapter{})
}
