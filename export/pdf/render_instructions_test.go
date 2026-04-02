package pdfexport

import (
	"testing"

	"codeberg.org/go-pdf/fpdf"
)

func TestWrapInstructionLinesWrapsToSafeWidth(t *testing.T) {
	pdf := fpdf.NewCustom(&fpdf.InitType{
		OrientationStr: "P",
		UnitStr:        "mm",
		Size: fpdf.SizeType{
			Wd: halfLetterWidthMM,
			Ht: halfLetterHeightMM,
		},
	})
	if err := registerPDFFonts(pdf); err != nil {
		t.Fatalf("registerPDFFonts error: %v", err)
	}
	pdf.AddPage()

	body := PuzzleBodyRect(halfLetterWidthMM, halfLetterHeightMM, 2)
	lines := wrapInstructionLines(
		pdf,
		body.W,
		[]string{"Shade duplicates; shaded cells cannot touch orthogonally; unshaded cells stay connected."},
	)
	if len(lines) < 2 {
		t.Fatalf("line count = %d, want at least 2 (%v)", len(lines), lines)
	}

	setInstructionStyle(pdf)
	maxWidth := body.W - instructionWrapInsetMM*2
	for _, line := range lines {
		if got := pdf.GetStringWidth(line); got > maxWidth+0.01 {
			t.Fatalf("line %q width = %.2f, want <= %.2f", line, got, maxWidth)
		}
	}
}

func TestWrapInstructionLinesKeepsShortRulesSingleLine(t *testing.T) {
	pdf := fpdf.NewCustom(&fpdf.InitType{
		OrientationStr: "P",
		UnitStr:        "mm",
		Size: fpdf.SizeType{
			Wd: halfLetterWidthMM,
			Ht: halfLetterHeightMM,
		},
	})
	if err := registerPDFFonts(pdf); err != nil {
		t.Fatalf("registerPDFFonts error: %v", err)
	}
	pdf.AddPage()

	body := PuzzleBodyRect(halfLetterWidthMM, halfLetterHeightMM, 2)
	lines := wrapInstructionLines(pdf, body.W, []string{"Fill rows, columns, and 3x3 boxes with 1-9"})
	if len(lines) != 1 {
		t.Fatalf("line count = %d, want 1 (%v)", len(lines), lines)
	}
}

func TestTakuzuPlusRuleCopyWrapsToThreePhysicalLines(t *testing.T) {
	pdf := fpdf.NewCustom(&fpdf.InitType{
		OrientationStr: "P",
		UnitStr:        "mm",
		Size: fpdf.SizeType{
			Wd: halfLetterWidthMM,
			Ht: halfLetterHeightMM,
		},
	})
	if err := registerPDFFonts(pdf); err != nil {
		t.Fatalf("registerPDFFonts error: %v", err)
	}
	pdf.AddPage()

	body := PuzzleBodyRect(halfLetterWidthMM, halfLetterHeightMM, 2)
	lines := wrapInstructionLines(pdf, body.W, []string{
		"No three equal adjacent in any row or column.",
		"Rows/columns have equal 0s and 1s, and all rows/columns are unique.",
		"= means same; x means different.",
	})
	if len(lines) != 3 {
		t.Fatalf("line count = %d, want 3 (%v)", len(lines), lines)
	}
}
