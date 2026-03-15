package netwalk

import (
	"math"
	"math/rand/v2"
	"strings"
	"testing"

	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"github.com/FelineStateMachine/puzzletea/game"
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

func TestDefaultKeyMapUsesEnterForLockAndSpaceForRotate(t *testing.T) {
	if !key.Matches(
		keyPress("space"),
		DefaultKeyMap.Rotate,
	) {
		t.Fatal("space should match rotate binding")
	}
	if key.Matches(
		keyPress("enter"),
		DefaultKeyMap.Rotate,
	) {
		t.Fatal("enter should not match rotate binding")
	}
	if !key.Matches(
		keyPress("enter"),
		DefaultKeyMap.Lock,
	) {
		t.Fatal("enter should match lock binding")
	}
	if key.Matches(
		keyPress("l"),
		DefaultKeyMap.Lock,
	) {
		t.Fatal("l should not match lock binding")
	}
}

func TestFrontierWeightPrefersPackedCandidatesOnHardProfiles(t *testing.T) {
	active := map[point]struct{}{
		{X: 2, Y: 2}: {},
		{X: 1, Y: 1}: {},
		{X: 3, Y: 1}: {},
	}
	adjacency := map[point]directionMask{
		{X: 2, Y: 2}: 0,
		{X: 1, Y: 1}: 0,
		{X: 3, Y: 1}: 0,
	}
	bounds := activeBounds{minX: 0, maxX: 4, minY: 0, maxY: 4}

	packed := frontierWeight(
		5,
		frontierEdge{from: point{X: 2, Y: 2}, to: point{X: 2, Y: 1}},
		active,
		adjacency,
		bounds,
		hardProfile,
	)
	isolated := frontierWeight(
		5,
		frontierEdge{from: point{X: 2, Y: 2}, to: point{X: 2, Y: 3}},
		active,
		adjacency,
		bounds,
		hardProfile,
	)
	if packed <= isolated {
		t.Fatalf("packed frontier weight = %d, want > isolated %d", packed, isolated)
	}
}

func TestSpanGrowthScoreRewardsExpansionBeforeTarget(t *testing.T) {
	bounds := activeBounds{minX: 2, maxX: 4, minY: 2, maxY: 4}

	growing := spanGrowthScore(9, point{X: 1, Y: 4}, bounds, mediumProfile)
	stable := spanGrowthScore(9, point{X: 3, Y: 3}, bounds, mediumProfile)
	if growing <= stable {
		t.Fatalf("expanding span score = %d, want > stable %d", growing, stable)
	}
}

func TestNetwalkModeDensityProgression(t *testing.T) {
	modes := netwalkModesFromRegistry(t)
	fill := make([]float64, len(modes))
	junctions := make([]float64, len(modes))

	for i, mode := range modes {
		fill[i], junctions[i] = sampleModeMetrics(t, mode, 12)
	}

	for i := 1; i < len(fill); i++ {
		if fill[i] <= fill[i-1] {
			t.Fatalf("fill ratio[%d] = %.3f, want > %.3f", i, fill[i], fill[i-1])
		}
	}
	for i := 2; i < len(junctions); i++ {
		if junctions[i] <= junctions[i-1] {
			t.Fatalf("junction avg[%d] = %.3f, want > %.3f", i, junctions[i], junctions[i-1])
		}
	}

	for i, mode := range modes {
		target := float64(targetActiveFromFillRatio(mode.Size, mode.FillRatio)) / float64(mode.Size*mode.Size)
		if math.Abs(fill[i]-target) > 1e-9 {
			t.Fatalf("%s fill ratio = %.3f, want %.3f", mode.Title(), fill[i], target)
		}
	}
}

func TestCellRowsShowDirectionalRootsAndLeaves(t *testing.T) {
	m := Model{}

	m.puzzle = newPuzzle(3)
	m.puzzle.Tiles[1][1] = tile{Kind: serverCell}
	m.puzzle.Tiles[1][2] = tile{Kind: nodeCell}
	m.puzzle.Tiles[0][1] = tile{Kind: nodeCell}
	m.state.rotatedMasks = make([][]directionMask, 3)
	for y := range 3 {
		m.state.rotatedMasks[y] = make([]directionMask, 3)
	}
	m.state.rotatedMasks[1][1] = south
	m.state.rotatedMasks[1][2] = north
	m.state.rotatedMasks[0][1] = north | east | south

	if got := cellRows(m, 1, 1); got != [cellHeight]string{"     ", "  ◆  ", "  │  "} {
		t.Fatalf("south root rows = %#v", got)
	}
	if got := cellRows(m, 2, 1); got != [cellHeight]string{"  │  ", "  ●  ", "     "} {
		t.Fatalf("north leaf rows = %#v", got)
	}
	if got := cellRows(m, 1, 0); got != [cellHeight]string{"  │  ", "  ├──", "  │  "} {
		t.Fatalf("tee rows = %#v", got)
	}
}

func TestGridViewUsesTallerFrameWithoutInteriorBoxes(t *testing.T) {
	m := Model{
		puzzle: newPuzzle(2),
	}
	m.recompute()

	lines := strings.Split(ansi.Strip(gridView(m)), "\n")
	if len(lines) != 8 {
		t.Fatalf("rendered line count = %d, want 8", len(lines))
	}
	for _, idx := range []int{1, 2, 3, 4, 5, 6} {
		if got := strings.Count(lines[idx], "│"); got != 2 {
			t.Fatalf("line %d has %d vertical borders, want outer frame only", idx, got)
		}
	}
}

func TestGridViewShowsCursorGlyphsOnBlankCells(t *testing.T) {
	m := Model{
		puzzle: newPuzzle(2),
		cursor: game.Cursor{X: 1, Y: 1},
	}
	m.recompute()

	view := ansi.Strip(gridView(m))
	if !strings.Contains(view, "▸   ◂") {
		t.Fatalf("blank cursor markers missing from view:\n%s", view)
	}
}

func netwalkModesFromRegistry(t *testing.T) []NetwalkMode {
	t.Helper()

	modes := make([]NetwalkMode, 0, len(Modes))
	for i, mode := range Modes {
		netwalkMode, ok := mode.(NetwalkMode)
		if !ok {
			t.Fatalf("mode %d has type %T, want NetwalkMode", i, mode)
		}
		modes = append(modes, netwalkMode)
	}
	return modes
}

func sampleModeMetrics(
	t *testing.T,
	mode NetwalkMode,
	samples int,
) (float64, float64) {
	t.Helper()

	var totalFill float64
	var totalJunctions float64
	for i := range samples {
		rng := rand.New(rand.NewPCG(uint64(1000+i), uint64(7000+i)))
		puzzle, err := GenerateSeededWithDensity(mode.Size, mode.FillRatio, mode.Profile, rng)
		if err != nil {
			t.Fatalf("GenerateSeededWithDensity(%q) error = %v", mode.Title(), err)
		}
		totalFill += puzzleFillRatio(puzzle)
		totalJunctions += float64(puzzleJunctionCount(puzzle))
	}
	return totalFill / float64(samples), totalJunctions / float64(samples)
}

func puzzleFillRatio(p Puzzle) float64 {
	if p.Size <= 0 {
		return 0
	}
	active := 0
	for y := range p.Size {
		for x := range p.Size {
			if isActive(p.Tiles[y][x]) {
				active++
			}
		}
	}
	return float64(active) / float64(p.Size*p.Size)
}

func puzzleJunctionCount(p Puzzle) int {
	count := 0
	for y := range p.Size {
		for x := range p.Size {
			if !isActive(p.Tiles[y][x]) {
				continue
			}
			if degree(p.Tiles[y][x].BaseMask) >= 3 {
				count++
			}
		}
	}
	return count
}

func keyPress(value string) tea.KeyPressMsg {
	switch value {
	case "enter":
		return tea.KeyPressMsg{Code: tea.KeyEnter}
	case "space":
		return tea.KeyPressMsg{Code: tea.KeySpace, Text: " "}
	default:
		r := []rune(value)
		return tea.KeyPressMsg{Code: r[0], Text: value}
	}
}
