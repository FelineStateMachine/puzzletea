package pdfexport

import (
	"reflect"
	"testing"

	"codeberg.org/go-pdf/fpdf"
)

func newDifficultyTestPDF(t *testing.T) *fpdf.Fpdf {
	t.Helper()

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
	return pdf
}

func TestScoreToDifficultyStarStates(t *testing.T) {
	tests := []struct {
		name  string
		score float64
		want  []difficultyStarState
	}{
		{
			name:  "zero score is empty",
			score: 0.0,
			want:  []difficultyStarState{difficultyStarEmpty, difficultyStarEmpty, difficultyStarEmpty, difficultyStarEmpty, difficultyStarEmpty},
		},
		{
			name:  "one tenth is half then empty",
			score: 0.1,
			want:  []difficultyStarState{difficultyStarHalf, difficultyStarEmpty, difficultyStarEmpty, difficultyStarEmpty, difficultyStarEmpty},
		},
		{
			name:  "three tenths is full half then empty",
			score: 0.3,
			want:  []difficultyStarState{difficultyStarFull, difficultyStarHalf, difficultyStarEmpty, difficultyStarEmpty, difficultyStarEmpty},
		},
		{
			name:  "five tenths is two and a half stars",
			score: 0.5,
			want:  []difficultyStarState{difficultyStarFull, difficultyStarFull, difficultyStarHalf, difficultyStarEmpty, difficultyStarEmpty},
		},
		{
			name:  "eight tenths is four full stars",
			score: 0.8,
			want:  []difficultyStarState{difficultyStarFull, difficultyStarFull, difficultyStarFull, difficultyStarFull, difficultyStarEmpty},
		},
		{
			name:  "one is five full stars",
			score: 1.0,
			want:  []difficultyStarState{difficultyStarFull, difficultyStarFull, difficultyStarFull, difficultyStarFull, difficultyStarFull},
		},
		{
			name:  "negative clamps low",
			score: -1.0,
			want:  []difficultyStarState{difficultyStarEmpty, difficultyStarEmpty, difficultyStarEmpty, difficultyStarEmpty, difficultyStarEmpty},
		},
		{
			name:  "above one clamps high",
			score: 2.0,
			want:  []difficultyStarState{difficultyStarFull, difficultyStarFull, difficultyStarFull, difficultyStarFull, difficultyStarFull},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := scoreToDifficultyStarStates(tt.score)
			if !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("states = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPuzzleDifficultySubtitleLayoutWidthNormalMode(t *testing.T) {
	pdf := newDifficultyTestPDF(t)
	layout := puzzleDifficultySubtitleLayout(pdf, Puzzle{
		ModeSelection:   "Expert",
		DifficultyScore: 0.5,
	})
	if layout.labelText != "Mode: Expert | Difficulty:" {
		t.Fatalf("label = %q", layout.labelText)
	}
	if layout.totalWidth <= 0 {
		t.Fatalf("total width = %.3f, want positive", layout.totalWidth)
	}
	if layout.totalWidth > halfLetterWidthMM {
		t.Fatalf("total width = %.3f, want <= %.3f", layout.totalWidth, halfLetterWidthMM)
	}
}

func TestPuzzleDifficultySubtitleLayoutWidthMixedMode(t *testing.T) {
	pdf := newDifficultyTestPDF(t)
	layout := puzzleDifficultySubtitleLayout(pdf, Puzzle{
		ModeSelection:   "Mixed Modes",
		DifficultyScore: 0.3,
	})
	if layout.labelText != "Difficulty:" {
		t.Fatalf("label = %q", layout.labelText)
	}
	if layout.totalWidth <= 0 {
		t.Fatalf("total width = %.3f, want positive", layout.totalWidth)
	}
	if layout.totalWidth > halfLetterWidthMM {
		t.Fatalf("total width = %.3f, want <= %.3f", layout.totalWidth, halfLetterWidthMM)
	}
}

func TestPuzzleDifficultySubtitleLayoutStatesSequence(t *testing.T) {
	pdf := newDifficultyTestPDF(t)
	layout := puzzleDifficultySubtitleLayout(pdf, Puzzle{
		ModeSelection:   "Tricky",
		DifficultyScore: 0.3,
	})
	want := []difficultyStarState{
		difficultyStarFull,
		difficultyStarHalf,
		difficultyStarEmpty,
		difficultyStarEmpty,
		difficultyStarEmpty,
	}
	if !reflect.DeepEqual(layout.starStates, want) {
		t.Fatalf("states = %v, want %v", layout.starStates, want)
	}
}

func TestDifficultyStarPointsReturnsTenVertices(t *testing.T) {
	points := difficultyStarPoints(10, 20, difficultyStarSizeMM)
	if len(points) != 10 {
		t.Fatalf("vertex count = %d, want 10", len(points))
	}
}
