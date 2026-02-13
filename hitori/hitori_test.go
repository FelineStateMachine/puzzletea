package hitori

import (
	"testing"
)

func TestGridClone(t *testing.T) {
	original := grid{
		{'1', '2', '3'},
		{'4', '5', '6'},
		{'7', '8', '9'},
	}
	cloned := original.clone()

	if len(cloned) != len(original) {
		t.Fatalf("clone length mismatch: got %d, want %d", len(cloned), len(original))
	}

	cloned[0][0] = '2'
	if original[0][0] == cloned[0][0] {
		t.Error("clone is not a deep copy")
	}
}

func TestCreateEmptyState(t *testing.T) {
	state := createEmptyState(5)
	expected := "     \n     \n     \n     \n     "
	if string(state) != expected {
		t.Errorf("createEmptyState mismatch:\ngot:\n%q\nwant:\n%q", string(state), expected)
	}
}

func TestCheckConstraints(t *testing.T) {
	tests := []struct {
		name string
		grid grid
		size int
		want bool
	}{
		{
			name: "valid no duplicates row",
			grid: grid{
				{'1', '2', '3'},
				{'2', '3', '1'},
				{'3', '1', '2'},
			},
			size: 3,
			want: true,
		},
		{
			name: "invalid duplicate in row",
			grid: grid{
				{'1', '1', '3'},
				{'2', '3', '1'},
				{'3', '2', '1'},
			},
			size: 3,
			want: false,
		},
		{
			name: "invalid duplicate in column",
			grid: grid{
				{'1', '2', '3'},
				{'2', '3', '1'},
				{'1', '1', '2'},
			},
			size: 3,
			want: false,
		},
		{
			name: "valid with shaded cells",
			grid: grid{
				{'#', '2', '3'},
				{'2', '#', '1'},
				{'3', '1', '#'},
			},
			size: 3,
			want: true,
		},
		{
			name: "invalid adjacent shaded",
			grid: grid{
				{'#', '#', '3'},
				{'2', '3', '1'},
				{'3', '1', '2'},
			},
			size: 3,
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := checkConstraints(tt.grid, tt.size)
			if got != tt.want {
				t.Errorf("checkConstraints() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsConnected(t *testing.T) {
	tests := []struct {
		name string
		grid grid
		size int
		want bool
	}{
		{
			name: "connected all cells",
			grid: grid{
				{'1', '2', '3'},
				{'2', '3', '1'},
				{'3', '1', '2'},
			},
			size: 3,
			want: true,
		},
		{
			name: "connected with shaded - diagonal only",
			grid: grid{
				{'1', '#', '2'},
				{'#', '#', '#'},
				{'3', '#', '4'},
			},
			size: 3,
			want: false,
		},
		{
			name: "not connected separated groups",
			grid: grid{
				{'1', '#', '#'},
				{'#', '#', '#'},
				{'#', '#', '2'},
			},
			size: 3,
			want: false,
		},
		{
			name: "all shaded",
			grid: grid{
				{'#', '#', '#'},
				{'#', '#', '#'},
				{'#', '#', '#'},
			},
			size: 3,
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isConnected(tt.grid, tt.size)
			if got != tt.want {
				t.Errorf("isConnected() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGridString(t *testing.T) {
	g := grid{
		{'1', '2', '3'},
		{'4', '5', '6'},
	}
	expected := "123\n456"
	if g.String() != expected {
		t.Errorf("grid.String() = %q, want %q", g.String(), expected)
	}
}

func TestNewGrid(t *testing.T) {
	s := state("123\n456")
	g := newGrid(s)
	if len(g) != 2 || len(g[0]) != 3 {
		t.Errorf("newGrid size mismatch: got %dx%d", len(g), len(g[0]))
	}
	if g[0][0] != '1' || g[1][2] != '6' {
		t.Errorf("newGrid content mismatch")
	}
}

func TestSerializeProvided(t *testing.T) {
	p := [][]bool{
		{true, false, true},
		{false, true, false},
		{true, false, true},
	}
	result := serializeProvided(p)
	expected := "#.#\n.#.\n#.#"
	if result != expected {
		t.Errorf("serializeProvided = %q, want %q", result, expected)
	}
}

func TestDeserializeProvided(t *testing.T) {
	s := "#.#\n.#.\n#.#"
	p := deserializeProvided(s, 3)
	if len(p) != 3 || len(p[0]) != 3 {
		t.Fatalf("deserializeProvided size mismatch")
	}
	if p[0][0] != true || p[0][1] != false || p[0][2] != true {
		t.Errorf("deserializeProvided values mismatch")
	}
}

func TestGeneratePuzzle(t *testing.T) {
	mode := NewMode("Test", "test", 5, 0.6)
	puzzle, provided, err := GeneratePuzzle(mode)
	if err != nil {
		t.Fatal(err)
	}

	if len(puzzle) != 5 {
		t.Errorf("puzzle size = %d, want 5", len(puzzle))
	}
	if len(provided) != 5 {
		t.Errorf("provided size = %d, want 5", len(provided))
	}

	clueCount := 0
	for y := 0; y < 5; y++ {
		for x := 0; x < 5; x++ {
			if provided[y][x] && puzzle[y][x] != shadedCell {
				clueCount++
			}
		}
	}
	_ = clueCount
}

func TestSaveLoadRoundTrip(t *testing.T) {
	mode := NewMode("Test", "test", 5, 0.6)
	puzzle, provided, err := GeneratePuzzle(mode)
	if err != nil {
		t.Fatal(err)
	}

	m, err := New(mode, puzzle, provided)
	if err != nil {
		t.Fatal(err)
	}

	data, err := m.GetSave()
	if err != nil {
		t.Fatal(err)
	}

	loaded, err := ImportModel(data)
	if err != nil {
		t.Fatal(err)
	}

	if loaded.size != m.(Model).size {
		t.Errorf("size mismatch after load")
	}
}
