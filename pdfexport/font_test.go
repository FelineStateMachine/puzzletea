package pdfexport

import "testing"

func TestStandardCellFontSizeBounds(t *testing.T) {
	tests := []struct {
		name     string
		cellSize float64
		scale    float64
		want     float64
	}{
		{
			name:     "clamps low",
			cellSize: 3.0,
			scale:    0.6,
			want:     5.2,
		},
		{
			name:     "keeps in range",
			cellSize: 10.0,
			scale:    0.6,
			want:     6.0,
		},
		{
			name:     "clamps high",
			cellSize: 20.0,
			scale:    0.7,
			want:     8.2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := standardCellFontSize(tt.cellSize, tt.scale)
			if got != tt.want {
				t.Fatalf("font size = %.3f, want %.3f", got, tt.want)
			}
		})
	}
}

func TestClampStandardCellFontSizeBounds(t *testing.T) {
	tests := []struct {
		name string
		in   float64
		want float64
	}{
		{
			name: "below min",
			in:   3.9,
			want: 5.2,
		},
		{
			name: "in range",
			in:   6.5,
			want: 6.5,
		},
		{
			name: "above max",
			in:   9.1,
			want: 8.2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := clampStandardCellFontSize(tt.in)
			if got != tt.want {
				t.Fatalf("clamp = %.3f, want %.3f", got, tt.want)
			}
		})
	}
}
