package nonogram

import (
	"math"
	"testing"

	"github.com/FelineStateMachine/puzzletea/pdfexport"
)

func TestNormalizeNonogramHintsForRender(t *testing.T) {
	tests := []struct {
		name  string
		hints [][]int
		size  int
		want  [][]int
	}{
		{
			name:  "empty source defaults to zeros",
			hints: nil,
			size:  3,
			want:  [][]int{{0}, {0}, {0}},
		},
		{
			name:  "preserves provided hint rows",
			hints: [][]int{{3, 1}, {}, {2}},
			size:  3,
			want:  [][]int{{3, 1}, {0}, {2}},
		},
		{
			name:  "pads beyond provided rows",
			hints: [][]int{{1}},
			size:  3,
			want:  [][]int{{1}, {0}, {0}},
		},
		{
			name:  "non-positive size returns nil",
			hints: [][]int{{1}},
			size:  0,
			want:  nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := normalizeNonogramHintsForRender(tt.hints, tt.size)
			if len(got) != len(tt.want) {
				t.Fatalf("rows = %d, want %d", len(got), len(tt.want))
			}
			for row := range len(tt.want) {
				if len(got[row]) != len(tt.want[row]) {
					t.Fatalf("row %d len = %d, want %d", row, len(got[row]), len(tt.want[row]))
				}
				for col := range len(tt.want[row]) {
					if got[row][col] != tt.want[row][col] {
						t.Fatalf("row %d col %d = %d, want %d", row, col, got[row][col], tt.want[row][col])
					}
				}
			}
		})
	}
}

func TestNormalizeNonogramHintsForRenderCopiesRows(t *testing.T) {
	src := [][]int{{2, 1}}

	got := normalizeNonogramHintsForRender(src, 1)
	if got == nil || len(got) != 1 {
		t.Fatalf("unexpected normalized hints: %#v", got)
	}

	src[0][0] = 9
	if got[0][0] != 2 {
		t.Fatalf("normalized hint should be copied, got %d want 2", got[0][0])
	}
}

func TestLayoutNonogramCentersGrid(t *testing.T) {
	tests := []struct {
		name       string
		rowHintCol int
		colHintRow int
	}{
		{name: "shallow hints", rowHintCol: 1, colHintRow: 1},
		{name: "deep hints", rowHintCol: 5, colHintRow: 4},
	}

	const (
		pageNo = 3
		pageW  = 139.7
		pageH  = 215.9
	)

	boardArea := pdfexport.PuzzleBoardRect(pageW, pageH, pageNo, 1)
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			layout := layoutNonogram(
				pageW,
				pageH,
				pageNo,
				10,
				10,
				tt.rowHintCol,
				tt.colHintRow,
			)
			if layout.cellSize <= 0 {
				t.Fatal("expected non-zero cell size")
			}

			gridW := 10.0 * layout.cellSize
			gridH := 10.0 * layout.cellSize
			centerX := layout.gridX + gridW/2
			centerY := layout.gridY + gridH/2
			wantX := boardArea.X + boardArea.W/2
			wantY := boardArea.Y + boardArea.H/2

			if diff := math.Abs(centerX - wantX); diff > 0.8 {
				t.Fatalf("grid centerX diff = %.3f, want <= 0.8", diff)
			}
			if diff := math.Abs(centerY - wantY); diff > 0.8 {
				t.Fatalf("grid centerY diff = %.3f, want <= 0.8", diff)
			}

			fullW := float64(tt.rowHintCol+10) * layout.cellSize
			fullH := float64(tt.colHintRow+10) * layout.cellSize
			if layout.hintStartX < boardArea.X-0.01 {
				t.Fatalf("hintStartX = %.3f, want >= %.3f", layout.hintStartX, boardArea.X)
			}
			if layout.hintStartY < boardArea.Y-0.01 {
				t.Fatalf("hintStartY = %.3f, want >= %.3f", layout.hintStartY, boardArea.Y)
			}
			if right := layout.hintStartX + fullW; right > boardArea.X+boardArea.W+0.01 {
				t.Fatalf("hint block right = %.3f, want <= %.3f", right, boardArea.X+boardArea.W)
			}
			if bottom := layout.hintStartY + fullH; bottom > boardArea.Y+boardArea.H+0.01 {
				t.Fatalf("hint block bottom = %.3f, want <= %.3f", bottom, boardArea.Y+boardArea.H)
			}
		})
	}
}
