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
