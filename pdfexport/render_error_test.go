package pdfexport

import (
	"errors"
	"path/filepath"
	"strings"
	"testing"

	"codeberg.org/go-pdf/fpdf"
	"github.com/FelineStateMachine/puzzletea/game"
)

func TestWritePDFReturnsAdapterRenderError(t *testing.T) {
	const category = "Render Error Test"
	game.RegisterPrintAdapter(renderErrorTestAdapter{category: category})

	output := filepath.Join(t.TempDir(), "error.pdf")
	puzzles := []Puzzle{
		{
			Category:     category,
			Name:         "broken-page",
			Index:        1,
			PrintPayload: struct{}{},
		},
	}

	err := WritePDF(output, nil, puzzles, RenderConfig{VolumeNumber: 1})
	if err == nil {
		t.Fatal("expected adapter render error")
	}
	if !strings.Contains(err.Error(), "failed to render body") {
		t.Fatalf("error = %q, want render failure context", err.Error())
	}
	if !strings.Contains(err.Error(), "broken-page") {
		t.Fatalf("error = %q, want puzzle name context", err.Error())
	}
}

type renderErrorTestAdapter struct {
	category string
}

func (a renderErrorTestAdapter) CanonicalGameType() string           { return a.category }
func (a renderErrorTestAdapter) Aliases() []string                   { return nil }
func (a renderErrorTestAdapter) BuildPDFPayload([]byte) (any, error) { return nil, nil }
func (a renderErrorTestAdapter) RenderPDFBody(*fpdf.Fpdf, any) error {
	return errors.New("failed to render body")
}
