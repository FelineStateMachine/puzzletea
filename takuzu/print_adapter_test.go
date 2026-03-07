package takuzu

import "testing"

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
			wantMin:  5.2,
			wantMax:  5.2,
		},
		{
			name:     "12x12 remains comfortably readable",
			cellSize: 10.0,
			size:     12,
			wantMin:  6.3,
			wantMax:  6.7,
		},
		{
			name:     "14x14 remains comfortably readable",
			cellSize: 9.0,
			size:     14,
			wantMin:  5.7,
			wantMax:  6.0,
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

func TestTakuzuRelationFontSize(t *testing.T) {
	tests := []struct {
		name     string
		cellSize float64
		size     int
		wantMin  float64
		wantMax  float64
	}{
		{
			name:     "small cells keep a larger minimum for relation clues",
			cellSize: 8.0,
			size:     14,
			wantMin:  6.0,
			wantMax:  6.0,
		},
		{
			name:     "10x10 clues scale above the old minimum",
			cellSize: 11.0,
			size:     10,
			wantMin:  6.3,
			wantMax:  6.5,
		},
		{
			name:     "large cells still respect the shared cap",
			cellSize: 16.0,
			size:     6,
			wantMin:  8.2,
			wantMax:  8.2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := takuzuRelationFontSize(tt.cellSize, tt.size)
			if got < tt.wantMin || got > tt.wantMax {
				t.Fatalf("font size = %.3f, want %.3f..%.3f", got, tt.wantMin, tt.wantMax)
			}
		})
	}
}

func TestTakuzuRelationBackdropSize(t *testing.T) {
	tests := []struct {
		name     string
		cellSize float64
		fontSize float64
		wantMin  float64
		wantMax  float64
	}{
		{
			name:     "adds padding beyond the glyph size",
			cellSize: 11.0,
			fontSize: 6.4,
			wantMin:  7.7,
			wantMax:  7.8,
		},
		{
			name:     "still provides a readable knockout on tighter boards",
			cellSize: 8.0,
			fontSize: 6.0,
			wantMin:  6.9,
			wantMax:  7.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := takuzuRelationBackdropSize(tt.cellSize, tt.fontSize)
			if got < tt.wantMin || got > tt.wantMax {
				t.Fatalf("backdrop size = %.3f, want %.3f..%.3f", got, tt.wantMin, tt.wantMax)
			}
		})
	}
}
