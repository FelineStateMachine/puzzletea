package takuzu

import (
	"testing"

	"github.com/FelineStateMachine/puzzletea/pdfexport"
)

func TestTakuzuGivenFontSize(t *testing.T) {
	tests := []struct {
		name     string
		cellSize float64
		size     int
		wantMin  float64
		wantMax  float64
	}{
		{
			name:     "small cell keeps readable minimum",
			cellSize: 3.0,
			size:     14,
			wantMin:  8.2,
			wantMax:  8.2,
		},
		{
			name:     "12x12 remains comfortably readable",
			cellSize: 10.0,
			size:     12,
			wantMin:  9.5,
			wantMax:  9.6,
		},
		{
			name:     "14x14 remains comfortably readable",
			cellSize: 9.0,
			size:     14,
			wantMin:  8.5,
			wantMax:  8.6,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := takuzuGivenFontSize(tt.cellSize, tt.size)
			if got < tt.wantMin || got > tt.wantMax {
				t.Fatalf("font size = %.3f, want %.3f..%.3f", got, tt.wantMin, tt.wantMax)
			}
		})
	}
}

func TestTakuzuRelationSizing_FontSize(t *testing.T) {
	tests := []struct {
		name     string
		cellSize float64
		size     int
		wantMin  float64
		wantMax  float64
	}{
		{
			name:     "14x14 scales below the old 9 point floor",
			cellSize: 8.0,
			size:     14,
			wantMin:  6.2,
			wantMax:  6.2,
		},
		{
			name:     "12x12 scales below the old 9 point floor",
			cellSize: 11.0,
			size:     12,
			wantMin:  6.2,
			wantMax:  6.2,
		},
		{
			name:     "larger cells still respect the new cap",
			cellSize: 16.0,
			size:     6,
			wantMin:  7.6,
			wantMax:  7.7,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := takuzuRelationSizingFor(tt.cellSize, tt.size)
			if got.fontSize < tt.wantMin || got.fontSize > tt.wantMax {
				t.Fatalf("font size = %.3f, want %.3f..%.3f", got.fontSize, tt.wantMin, tt.wantMax)
			}
		})
	}
}

func TestTakuzuRelationSizing_BackdropSize(t *testing.T) {
	tests := []struct {
		name     string
		cellSize float64
		size     int
		wantMin  float64
		wantMax  float64
	}{
		{
			name:     "caps the knockout on denser 12x12 boards",
			cellSize: 11.0,
			size:     12,
			wantMin:  6.3,
			wantMax:  6.4,
		},
		{
			name:     "allows slightly larger knockouts when space is available",
			cellSize: 16.0,
			size:     6,
			wantMin:  8.2,
			wantMax:  8.4,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := takuzuRelationSizingFor(tt.cellSize, tt.size)
			if got.backdropSize < tt.wantMin || got.backdropSize > tt.wantMax {
				t.Fatalf("backdrop size = %.3f, want %.3f..%.3f", got.backdropSize, tt.wantMin, tt.wantMax)
			}
		})
	}
}

func TestTakuzuRelationSizing_BackdropStaysWithin12x12Cap(t *testing.T) {
	area := pdfexport.PuzzleBoardRect(139.7, 215.9, 2, 3)
	cellSize := pdfexport.FitCompactCellSize(12, 12, area)
	if cellSize <= 0 {
		t.Fatal("expected positive cell size")
	}

	got := takuzuRelationSizingFor(cellSize, 12)
	if got.backdropSize > cellSize*0.58+0.001 {
		t.Fatalf("backdrop size = %.3f, want <= %.3f", got.backdropSize, cellSize*0.58)
	}
}
