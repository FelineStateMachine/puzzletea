package pdfexport

import "testing"

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
