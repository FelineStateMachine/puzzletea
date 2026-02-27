package pdfexport

import (
	"math"
	"testing"
)

func TestThinGridLineFloor(t *testing.T) {
	if thinGridLineMM < 0.22 {
		t.Fatalf("thinGridLineMM = %.2f, want >= 0.22", thinGridLineMM)
	}
}

func TestCenteredOriginKeepsBoardCentered(t *testing.T) {
	pageArea := puzzleBoardRect(halfLetterWidthMM, halfLetterHeightMM, 2, 1)
	tests := []struct {
		name   string
		cols   int
		rows   int
		family boardFamily
	}{
		{name: "compact", cols: 8, rows: 8, family: boardFamilyCompact},
		{name: "sudoku", cols: 9, rows: 9, family: boardFamilySudoku},
		{name: "hashi-span", cols: 6, rows: 6, family: boardFamilyHashi},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cell := fitBoardCellSize(tt.cols, tt.rows, pageArea, tt.family)
			if cell <= 0 {
				t.Fatal("expected positive cell size")
			}
			x, y := centeredOrigin(pageArea, tt.cols, tt.rows, cell)

			centerX := x + float64(tt.cols)*cell/2
			centerY := y + float64(tt.rows)*cell/2
			wantX := pageArea.x + pageArea.w/2
			wantY := pageArea.y + pageArea.h/2
			if diff := math.Abs(centerX - wantX); diff > 0.01 {
				t.Fatalf("centerX diff = %.4f, want <= 0.01", diff)
			}
			if diff := math.Abs(centerY - wantY); diff > 0.01 {
				t.Fatalf("centerY diff = %.4f, want <= 0.01", diff)
			}
		})
	}
}

func TestPuzzleBodyRectMirrorsGutterByParity(t *testing.T) {
	even := puzzleBodyRect(halfLetterWidthMM, halfLetterHeightMM, 2)
	odd := puzzleBodyRect(halfLetterWidthMM, halfLetterHeightMM, 3)

	if even.w <= 0 || odd.w <= 0 {
		t.Fatal("expected positive body widths")
	}
	if math.Abs(even.w-odd.w) > 0.001 {
		t.Fatalf("body widths differ: even=%.3f odd=%.3f", even.w, odd.w)
	}

	evenCenterX := even.x + even.w/2
	oddCenterX := odd.x + odd.w/2
	pageCenterX := halfLetterWidthMM / 2

	if evenCenterX >= pageCenterX {
		t.Fatalf("even page center = %.3f, want left of %.3f", evenCenterX, pageCenterX)
	}
	if oddCenterX <= pageCenterX {
		t.Fatalf("odd page center = %.3f, want right of %.3f", oddCenterX, pageCenterX)
	}
}

func TestSaddleStitchPadCount(t *testing.T) {
	tests := []struct {
		name  string
		pages int
		want  int
	}{
		{name: "already multiple of four", pages: 36, want: 0},
		{name: "remainder one", pages: 35, want: 1},
		{name: "remainder two", pages: 34, want: 2},
		{name: "remainder three", pages: 33, want: 3},
		{name: "non-positive", pages: 0, want: 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := saddleStitchPadCount(tt.pages)
			if got != tt.want {
				t.Fatalf("saddleStitchPadCount(%d) = %d, want %d", tt.pages, got, tt.want)
			}
		})
	}
}

func TestSaddleStitchPadCountForStandardPackLayout(t *testing.T) {
	tests := []struct {
		name       string
		puzzleRows int
		wantPad    int
	}{
		{name: "single puzzle", puzzleRows: 1, wantPad: 0},
		{name: "two puzzles", puzzleRows: 2, wantPad: 3},
		{name: "thirty-two puzzles", puzzleRows: 32, wantPad: 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			totalWithoutPad := tt.puzzleRows + 3 // cover + title + back cover
			got := saddleStitchPadCount(totalWithoutPad)
			if got != tt.wantPad {
				t.Fatalf("pad pages = %d, want %d", got, tt.wantPad)
			}
			if (totalWithoutPad+got)%4 != 0 {
				t.Fatalf("total pages = %d, want multiple of 4", totalWithoutPad+got)
			}
		})
	}
}

func TestSaddleStitchPadCountForTitleOnlyPackLayout(t *testing.T) {
	tests := []struct {
		name       string
		puzzleRows int
		wantPad    int
	}{
		{name: "single puzzle", puzzleRows: 1, wantPad: 2},
		{name: "two puzzles", puzzleRows: 2, wantPad: 1},
		{name: "thirty-two puzzles", puzzleRows: 32, wantPad: 3},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			totalWithoutPad := tt.puzzleRows + 1 // title page + puzzle pages
			got := saddleStitchPadCount(totalWithoutPad)
			if got != tt.wantPad {
				t.Fatalf("pad pages = %d, want %d", got, tt.wantPad)
			}
			if (totalWithoutPad+got)%4 != 0 {
				t.Fatalf("total pages = %d, want multiple of 4", totalWithoutPad+got)
			}
		})
	}
}
