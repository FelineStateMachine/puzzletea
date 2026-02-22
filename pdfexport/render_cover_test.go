package pdfexport

import (
	"reflect"
	"testing"

	"github.com/go-pdf/fpdf"
)

func TestResolveCoverColorDeterministicWithSeed(t *testing.T) {
	cfgA := RenderConfig{ShuffleSeed: "zine-seed"}
	cfgB := RenderConfig{ShuffleSeed: "zine-seed"}

	colorA := resolveCoverColor(cfgA)
	colorB := resolveCoverColor(cfgB)
	if colorA != colorB {
		t.Fatalf("resolveCoverColor mismatch for identical seed: %+v vs %+v", colorA, colorB)
	}
}

func TestSplitCoverSubtitleLinesClampsToMaxLines(t *testing.T) {
	pdf := fpdf.New("P", "mm", "A4", "")
	if err := registerPDFFonts(pdf); err != nil {
		t.Fatalf("registerPDFFonts error: %v", err)
	}
	pdf.AddPage()
	pdf.SetFont(coverFontFamily, "", 18)

	subtitle := "The Catacombs of the Last Bastion and Other Grim Architectural Delights"
	maxW := 78.0
	got := splitCoverSubtitleLines(pdf, subtitle, maxW, 2)
	if len(got) != 2 {
		t.Fatalf("line count = %d, want 2 (%v)", len(got), got)
	}

	gotAgain := splitCoverSubtitleLines(pdf, subtitle, maxW, 2)
	if !reflect.DeepEqual(got, gotAgain) {
		t.Fatalf("splitCoverSubtitleLines not stable: %v vs %v", got, gotAgain)
	}
}
