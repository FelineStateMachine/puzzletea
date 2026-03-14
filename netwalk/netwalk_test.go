package netwalk

import (
	"math/rand/v2"
	"strings"
	"testing"

	"github.com/charmbracelet/x/ansi"
)

func TestRotateMask(t *testing.T) {
	mask := north | east
	if got := rotateMask(mask, 1); got != east|south {
		t.Fatalf("rotateMask(..., 1) = %v, want %v", got, east|south)
	}
	if got := rotateMask(mask, 2); got != south|west {
		t.Fatalf("rotateMask(..., 2) = %v, want %v", got, south|west)
	}
}

func TestAnalyzePuzzleSolvedAndDangling(t *testing.T) {
	puzzle := newPuzzle(2)
	puzzle.Root = point{X: 0, Y: 0}
	puzzle.Tiles[0][0] = tile{BaseMask: east, Kind: serverCell}
	puzzle.Tiles[0][1] = tile{BaseMask: west, Kind: nodeCell}

	state := analyzePuzzle(puzzle)
	if !state.solved {
		t.Fatal("expected simple 2-cell puzzle to be solved")
	}

	puzzle.Tiles[0][1].Rotation = 1
	state = analyzePuzzle(puzzle)
	if state.solved {
		t.Fatal("rotated puzzle should not be solved")
	}
	if state.dangling == 0 {
		t.Fatal("expected dangling connectors after bad rotation")
	}
}

func TestSaveImportRoundTrip(t *testing.T) {
	puzzle := newPuzzle(3)
	puzzle.Root = point{X: 1, Y: 1}
	puzzle.Tiles[1][1] = tile{BaseMask: north | east | south, Rotation: 1, InitialRotation: 2, Kind: serverCell, Locked: true}
	puzzle.Tiles[0][1] = tile{BaseMask: south, Rotation: 3, InitialRotation: 3, Kind: nodeCell}
	puzzle.Tiles[1][2] = tile{BaseMask: west, Rotation: 0, InitialRotation: 1, Kind: nodeCell}
	puzzle.Tiles[2][1] = tile{BaseMask: north, Rotation: 2, InitialRotation: 2, Kind: nodeCell}

	m := Model{puzzle: puzzle, keys: DefaultKeyMap, modeTitle: "Test"}
	m.recompute()
	save, err := m.GetSave()
	if err != nil {
		t.Fatalf("GetSave() error = %v", err)
	}

	imported, err := ImportModel(save)
	if err != nil {
		t.Fatalf("ImportModel() error = %v", err)
	}

	if imported.puzzle.Size != puzzle.Size {
		t.Fatalf("size = %d, want %d", imported.puzzle.Size, puzzle.Size)
	}
	if imported.puzzle.Root != puzzle.Root {
		t.Fatalf("root = %+v, want %+v", imported.puzzle.Root, puzzle.Root)
	}
	if !imported.puzzle.Tiles[1][1].Locked {
		t.Fatal("expected lock state to round-trip")
	}
	if imported.puzzle.Tiles[1][1].Rotation != 1 || imported.puzzle.Tiles[1][1].InitialRotation != 2 {
		t.Fatal("expected rotations to round-trip")
	}
}

func TestGenerateSeededDeterministic(t *testing.T) {
	rngA := rand.New(rand.NewPCG(10, 20))
	rngB := rand.New(rand.NewPCG(10, 20))

	a, err := GenerateSeeded(7, 14, rngA)
	if err != nil {
		t.Fatalf("GenerateSeeded() error = %v", err)
	}
	b, err := GenerateSeeded(7, 14, rngB)
	if err != nil {
		t.Fatalf("GenerateSeeded() error = %v", err)
	}

	if got, want := encodeMaskRows(a.Tiles), encodeMaskRows(b.Tiles); got != want {
		t.Fatalf("mask encoding mismatch\n got %q\nwant %q", got, want)
	}
	if got, want := encodeRotationRows(a.Tiles, true), encodeRotationRows(b.Tiles, true); got != want {
		t.Fatalf("initial rotations mismatch\n got %q\nwant %q", got, want)
	}
}

func TestCellTextUsesStarAndDot(t *testing.T) {
	m := Model{}

	rootTile := tile{Kind: serverCell}
	leafTile := tile{Kind: nodeCell}
	junctionTile := tile{Kind: nodeCell}

	m.puzzle = newPuzzle(3)
	m.puzzle.Tiles[1][1] = rootTile
	m.puzzle.Tiles[1][2] = leafTile
	m.puzzle.Tiles[0][1] = junctionTile
	m.state.rotatedMasks = make([][]directionMask, 3)
	for y := range 3 {
		m.state.rotatedMasks[y] = make([]directionMask, 3)
	}
	m.state.rotatedMasks[1][1] = east | west
	m.state.rotatedMasks[1][2] = west
	m.state.rotatedMasks[0][1] = east | south | west

	if got := cellText(m, 1, 1); got != "──★──" {
		t.Fatalf("root cellText = %q, want %q", got, "──★──")
	}
	if got := cellText(m, 2, 1); got != "──•  " {
		t.Fatalf("leaf cellText = %q, want %q", got, "──•  ")
	}
	if got := cellText(m, 1, 0); got != "──┬──" {
		t.Fatalf("junction cellText = %q, want %q", got, "──┬──")
	}
}

func TestBridgeTextSpansSeparators(t *testing.T) {
	m := Model{}
	m.puzzle = newPuzzle(2)
	m.state.rotatedMasks = [][]directionMask{
		{east, west},
		{south, north},
	}

	if got := verticalBridgeText(m, 1, 0); got != "─" {
		t.Fatalf("verticalBridgeText = %q, want %q", got, "─")
	}

	m.state.rotatedMasks = [][]directionMask{
		{south, 0},
		{north, 0},
	}
	if got := horizontalBridgeText(m, 0, 1); got != "  │  " {
		t.Fatalf("horizontalBridgeText = %q, want %q", got, "  │  ")
	}
}

func TestGridViewUsesMinimalFrameWithoutInteriorBoxes(t *testing.T) {
	m := Model{
		puzzle: newPuzzle(2),
	}
	m.recompute()

	lines := strings.Split(ansi.Strip(gridView(m)), "\n")
	if len(lines) != 5 {
		t.Fatalf("rendered line count = %d, want 5", len(lines))
	}
	for _, idx := range []int{1, 2, 3} {
		if got := strings.Count(lines[idx], "│"); got != 2 {
			t.Fatalf("line %d has %d vertical borders, want outer frame only", idx, got)
		}
	}
}
