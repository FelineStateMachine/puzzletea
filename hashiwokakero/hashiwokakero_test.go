package hashiwokakero

import (
	"encoding/json"
	"testing"
)

// --- helpers ---

// fourCornerPuzzle builds a 4x4 puzzle with islands at the four corners:
//
//	0 . . 1
//	. . . .
//	. . . .
//	2 . . 3
func fourCornerPuzzle() *Puzzle {
	return &Puzzle{
		Width:  4,
		Height: 4,
		Islands: []Island{
			{ID: 0, X: 0, Y: 0, Required: 2},
			{ID: 1, X: 3, Y: 0, Required: 2},
			{ID: 2, X: 0, Y: 3, Required: 2},
			{ID: 3, X: 3, Y: 3, Required: 2},
		},
	}
}

// linePuzzle builds a horizontal line of islands: 0 -- 1 -- 2 -- 3
func linePuzzle() *Puzzle {
	return &Puzzle{
		Width:  7,
		Height: 1,
		Islands: []Island{
			{ID: 0, X: 0, Y: 0, Required: 1},
			{ID: 1, X: 2, Y: 0, Required: 2},
			{ID: 2, X: 4, Y: 0, Required: 2},
			{ID: 3, X: 6, Y: 0, Required: 1},
		},
	}
}

// --- GetBridge (P0) ---

func TestGetBridge(t *testing.T) {
	t.Run("bridge exists", func(t *testing.T) {
		p := fourCornerPuzzle()
		p.SetBridge(0, 1, 1)

		b := p.GetBridge(0, 1)
		if b == nil {
			t.Fatal("expected bridge, got nil")
		}
		if b.Count != 1 {
			t.Errorf("Count = %d, want 1", b.Count)
		}
	})

	t.Run("order independence", func(t *testing.T) {
		p := fourCornerPuzzle()
		p.SetBridge(0, 1, 1)

		b := p.GetBridge(1, 0)
		if b == nil {
			t.Fatal("GetBridge(1,0) should find bridge set as (0,1)")
		}
		if b.Count != 1 {
			t.Errorf("Count = %d, want 1", b.Count)
		}
	})

	t.Run("no bridge", func(t *testing.T) {
		p := fourCornerPuzzle()
		p.SetBridge(0, 1, 1)

		if b := p.GetBridge(0, 2); b != nil {
			t.Errorf("expected nil, got bridge with Count=%d", b.Count)
		}
	})
}

// --- SetBridge (P0) ---

func TestSetBridge(t *testing.T) {
	t.Run("create single bridge", func(t *testing.T) {
		p := fourCornerPuzzle()
		p.SetBridge(0, 1, 1)

		b := p.GetBridge(0, 1)
		if b == nil || b.Count != 1 {
			t.Fatalf("expected Count=1, got %v", b)
		}
	})

	t.Run("create double bridge", func(t *testing.T) {
		p := fourCornerPuzzle()
		p.SetBridge(0, 1, 2)

		b := p.GetBridge(0, 1)
		if b == nil || b.Count != 2 {
			t.Fatalf("expected Count=2, got %v", b)
		}
	})

	t.Run("update bridge", func(t *testing.T) {
		p := fourCornerPuzzle()
		p.SetBridge(0, 1, 1)
		p.SetBridge(0, 1, 2)

		b := p.GetBridge(0, 1)
		if b == nil || b.Count != 2 {
			t.Fatalf("expected Count=2 after update, got %v", b)
		}
	})

	t.Run("remove bridge", func(t *testing.T) {
		p := fourCornerPuzzle()
		p.SetBridge(0, 1, 1)
		p.SetBridge(0, 1, 0)

		if b := p.GetBridge(0, 1); b != nil {
			t.Errorf("expected bridge removed, got Count=%d", b.Count)
		}
	})

	t.Run("remove nonexistent", func(t *testing.T) {
		p := fourCornerPuzzle()
		p.SetBridge(5, 6, 0) // must not panic
	})

	t.Run("normalizes order", func(t *testing.T) {
		p := fourCornerPuzzle()
		p.SetBridge(3, 1, 1)

		if len(p.Bridges) != 1 {
			t.Fatalf("expected 1 bridge, got %d", len(p.Bridges))
		}
		if p.Bridges[0].Island1 != 1 || p.Bridges[0].Island2 != 3 {
			t.Errorf("bridge stored as (%d,%d), want (1,3)", p.Bridges[0].Island1, p.Bridges[0].Island2)
		}
	})

	t.Run("invalidates caches", func(t *testing.T) {
		p := fourCornerPuzzle()
		// Trigger cache builds.
		_ = p.BridgeCount(0)
		_ = p.FindIslandAt(0, 0)
		_ = p.CellContent(0, 0)

		p.SetBridge(0, 1, 1)

		if p.cellCache != nil {
			t.Error("cellCache should be nil after SetBridge")
		}
		if p.posIndex != nil {
			t.Error("posIndex should be nil after SetBridge")
		}
		if p.bridgeCounts != nil {
			t.Error("bridgeCounts should be nil after SetBridge")
		}
	})
}

// --- BridgeCount (P0) ---

func TestBridgeCount(t *testing.T) {
	t.Run("single bridge", func(t *testing.T) {
		p := fourCornerPuzzle()
		p.SetBridge(0, 1, 1)

		if got := p.BridgeCount(0); got != 1 {
			t.Errorf("BridgeCount(0) = %d, want 1", got)
		}
	})

	t.Run("double bridge", func(t *testing.T) {
		p := fourCornerPuzzle()
		p.SetBridge(0, 1, 2)

		if got := p.BridgeCount(0); got != 2 {
			t.Errorf("BridgeCount(0) = %d, want 2", got)
		}
	})

	t.Run("multiple bridges", func(t *testing.T) {
		p := fourCornerPuzzle()
		p.SetBridge(0, 1, 1)
		p.SetBridge(0, 2, 2)

		if got := p.BridgeCount(0); got != 3 {
			t.Errorf("BridgeCount(0) = %d, want 3", got)
		}
	})

	t.Run("no bridges", func(t *testing.T) {
		p := fourCornerPuzzle()

		if got := p.BridgeCount(0); got != 0 {
			t.Errorf("BridgeCount(0) = %d, want 0", got)
		}
	})

	t.Run("invalid ID", func(t *testing.T) {
		p := fourCornerPuzzle()

		if got := p.BridgeCount(999); got != 0 {
			t.Errorf("BridgeCount(999) = %d, want 0", got)
		}
	})

	t.Run("counts both endpoints", func(t *testing.T) {
		p := fourCornerPuzzle()
		p.SetBridge(0, 1, 1)

		if got := p.BridgeCount(1); got != 1 {
			t.Errorf("BridgeCount(1) = %d, want 1", got)
		}
	})
}

// --- WouldCross (P0) ---

func TestWouldCross(t *testing.T) {
	t.Run("no crossing parallel horizontal", func(t *testing.T) {
		// Two horizontal bridges on different rows can't cross.
		//  0 --- 1
		//  2 --- 3
		p := fourCornerPuzzle()
		p.SetBridge(0, 1, 1) // horizontal top
		if p.WouldCross(2, 3) {
			t.Error("parallel horizontal bridges should not cross")
		}
	})

	t.Run("no crossing parallel vertical", func(t *testing.T) {
		// Two vertical bridges on different columns.
		p := fourCornerPuzzle()
		p.SetBridge(0, 2, 1) // vertical left
		if p.WouldCross(1, 3) {
			t.Error("parallel vertical bridges should not cross")
		}
	})

	t.Run("horizontal crosses vertical", func(t *testing.T) {
		// Cross pattern: vertical bridge 0-2, then check horizontal 1-? crossing it.
		// Need a puzzle with an island in the middle area.
		p := &Puzzle{
			Width:  5,
			Height: 5,
			Islands: []Island{
				{ID: 0, X: 2, Y: 0, Required: 2},
				{ID: 1, X: 0, Y: 2, Required: 2},
				{ID: 2, X: 2, Y: 4, Required: 2},
				{ID: 3, X: 4, Y: 2, Required: 2},
			},
		}
		p.SetBridge(0, 2, 1) // vertical bridge through (2,1),(2,2),(2,3)
		if !p.WouldCross(1, 3) {
			t.Error("horizontal bridge should cross existing vertical bridge")
		}
	})

	t.Run("vertical crosses horizontal", func(t *testing.T) {
		p := &Puzzle{
			Width:  5,
			Height: 5,
			Islands: []Island{
				{ID: 0, X: 2, Y: 0, Required: 2},
				{ID: 1, X: 0, Y: 2, Required: 2},
				{ID: 2, X: 2, Y: 4, Required: 2},
				{ID: 3, X: 4, Y: 2, Required: 2},
			},
		}
		p.SetBridge(1, 3, 1) // horizontal bridge through (1,2),(2,2),(3,2)
		if !p.WouldCross(0, 2) {
			t.Error("vertical bridge should cross existing horizontal bridge")
		}
	})

	t.Run("island in path", func(t *testing.T) {
		// Island sitting between two endpoints blocks the bridge.
		p := &Puzzle{
			Width:  5,
			Height: 1,
			Islands: []Island{
				{ID: 0, X: 0, Y: 0, Required: 2},
				{ID: 1, X: 2, Y: 0, Required: 2},
				{ID: 2, X: 4, Y: 0, Required: 2},
			},
		}
		if !p.WouldCross(0, 2) {
			t.Error("bridge should be blocked by island in the middle")
		}
	})

	t.Run("endpoints dont count as crossing", func(t *testing.T) {
		p := fourCornerPuzzle()
		// No bridges set — the endpoints themselves are not obstacles.
		if p.WouldCross(0, 1) {
			t.Error("empty path between endpoints should not cross")
		}
	})

	t.Run("non-aligned islands", func(t *testing.T) {
		p := &Puzzle{
			Width:  4,
			Height: 4,
			Islands: []Island{
				{ID: 0, X: 0, Y: 0, Required: 1},
				{ID: 1, X: 1, Y: 1, Required: 1},
			},
		}
		if !p.WouldCross(0, 1) {
			t.Error("diagonal (non-aligned) islands should return true")
		}
	})

	t.Run("nil island", func(t *testing.T) {
		p := fourCornerPuzzle()
		if !p.WouldCross(0, 99) {
			t.Error("bad island ID should return true")
		}
	})

	t.Run("adjacent islands no obstacles", func(t *testing.T) {
		p := &Puzzle{
			Width:  3,
			Height: 1,
			Islands: []Island{
				{ID: 0, X: 0, Y: 0, Required: 1},
				{ID: 1, X: 1, Y: 0, Required: 1},
			},
		}
		if p.WouldCross(0, 1) {
			t.Error("adjacent islands with nothing between should not cross")
		}
	})
}

// --- IsConnected (P0) ---

func TestIsConnected(t *testing.T) {
	t.Run("all connected chain", func(t *testing.T) {
		p := linePuzzle()
		p.SetBridge(0, 1, 1)
		p.SetBridge(1, 2, 1)
		p.SetBridge(2, 3, 1)

		if !p.IsConnected() {
			t.Error("chain of bridges should be connected")
		}
	})

	t.Run("disconnected pair", func(t *testing.T) {
		p := fourCornerPuzzle()
		p.SetBridge(0, 1, 1) // only top two connected

		if p.IsConnected() {
			t.Error("only two of four islands connected should not be connected")
		}
	})

	t.Run("single island", func(t *testing.T) {
		p := &Puzzle{
			Width:   3,
			Height:  3,
			Islands: []Island{{ID: 0, X: 1, Y: 1, Required: 0}},
		}
		if !p.IsConnected() {
			t.Error("single island should be connected")
		}
	})

	t.Run("no bridges multiple islands", func(t *testing.T) {
		p := fourCornerPuzzle()
		if p.IsConnected() {
			t.Error("multiple islands with no bridges should not be connected")
		}
	})

	t.Run("multiple components", func(t *testing.T) {
		p := fourCornerPuzzle()
		p.SetBridge(0, 1, 1) // top pair
		p.SetBridge(2, 3, 1) // bottom pair

		if p.IsConnected() {
			t.Error("two separate pairs should not be connected")
		}
	})

	t.Run("star topology", func(t *testing.T) {
		// Center island connected to 4 spokes.
		p := &Puzzle{
			Width:  5,
			Height: 5,
			Islands: []Island{
				{ID: 0, X: 2, Y: 2, Required: 4},
				{ID: 1, X: 2, Y: 0, Required: 1},
				{ID: 2, X: 4, Y: 2, Required: 1},
				{ID: 3, X: 2, Y: 4, Required: 1},
				{ID: 4, X: 0, Y: 2, Required: 1},
			},
		}
		p.SetBridge(0, 1, 1)
		p.SetBridge(0, 2, 1)
		p.SetBridge(0, 3, 1)
		p.SetBridge(0, 4, 1)

		if !p.IsConnected() {
			t.Error("star topology should be connected")
		}
	})
}

// --- IsSolved (P0) ---

func TestIsSolved(t *testing.T) {
	t.Run("fully solved", func(t *testing.T) {
		// Simple: two islands requiring 1 bridge each.
		p := &Puzzle{
			Width:  3,
			Height: 1,
			Islands: []Island{
				{ID: 0, X: 0, Y: 0, Required: 1},
				{ID: 1, X: 2, Y: 0, Required: 1},
			},
		}
		p.SetBridge(0, 1, 1)

		if !p.IsSolved() {
			t.Error("puzzle should be solved")
		}
	})

	t.Run("satisfied but disconnected", func(t *testing.T) {
		// Four islands: each requires 1 bridge. Connect as two pairs.
		p := &Puzzle{
			Width:  5,
			Height: 5,
			Islands: []Island{
				{ID: 0, X: 0, Y: 0, Required: 1},
				{ID: 1, X: 2, Y: 0, Required: 1},
				{ID: 2, X: 0, Y: 4, Required: 1},
				{ID: 3, X: 2, Y: 4, Required: 1},
			},
		}
		p.SetBridge(0, 1, 1)
		p.SetBridge(2, 3, 1)

		if p.IsSolved() {
			t.Error("satisfied but disconnected should not be solved")
		}
	})

	t.Run("connected but unsatisfied", func(t *testing.T) {
		p := &Puzzle{
			Width:  3,
			Height: 1,
			Islands: []Island{
				{ID: 0, X: 0, Y: 0, Required: 2},
				{ID: 1, X: 2, Y: 0, Required: 2},
			},
		}
		p.SetBridge(0, 1, 1) // only 1, need 2

		if p.IsSolved() {
			t.Error("connected but unsatisfied should not be solved")
		}
	})

	t.Run("over-satisfied", func(t *testing.T) {
		p := &Puzzle{
			Width:  3,
			Height: 1,
			Islands: []Island{
				{ID: 0, X: 0, Y: 0, Required: 1},
				{ID: 1, X: 2, Y: 0, Required: 1},
			},
		}
		p.SetBridge(0, 1, 2)

		if p.IsSolved() {
			t.Error("over-satisfied should not be solved")
		}
	})

	t.Run("empty puzzle no bridges", func(t *testing.T) {
		p := fourCornerPuzzle()
		if p.IsSolved() {
			t.Error("no bridges placed should not be solved")
		}
	})
}

// --- FindIslandAt (P1) ---

func TestFindIslandAt(t *testing.T) {
	t.Run("island exists", func(t *testing.T) {
		p := fourCornerPuzzle()
		isl := p.FindIslandAt(3, 0)
		if isl == nil {
			t.Fatal("expected island at (3,0)")
		}
		if isl.ID != 1 {
			t.Errorf("ID = %d, want 1", isl.ID)
		}
	})

	t.Run("no island", func(t *testing.T) {
		p := fourCornerPuzzle()
		if isl := p.FindIslandAt(1, 1); isl != nil {
			t.Errorf("expected nil at (1,1), got island %d", isl.ID)
		}
	})

	t.Run("lazy index build", func(t *testing.T) {
		p := fourCornerPuzzle()
		if p.posIndex != nil {
			t.Fatal("posIndex should be nil before first call")
		}
		_ = p.FindIslandAt(0, 0)
		if p.posIndex == nil {
			t.Error("posIndex should be built after first call")
		}
		// Subsequent call should use the same index.
		_ = p.FindIslandAt(3, 3)
		if p.posIndex == nil {
			t.Error("posIndex should persist across calls")
		}
	})
}

// --- FindIslandByID (P1) ---

func TestFindIslandByID(t *testing.T) {
	t.Run("sequential IDs", func(t *testing.T) {
		p := fourCornerPuzzle()
		isl := p.FindIslandByID(2)
		if isl == nil {
			t.Fatal("expected island with ID=2")
		}
		if isl.X != 0 || isl.Y != 3 {
			t.Errorf("island 2 at (%d,%d), want (0,3)", isl.X, isl.Y)
		}
	})

	t.Run("invalid ID", func(t *testing.T) {
		p := fourCornerPuzzle()
		if isl := p.FindIslandByID(99); isl != nil {
			t.Errorf("expected nil, got island %d", isl.ID)
		}
	})

	t.Run("non-sequential fallback", func(t *testing.T) {
		p := &Puzzle{
			Width:  3,
			Height: 3,
			Islands: []Island{
				{ID: 10, X: 0, Y: 0, Required: 1},
				{ID: 20, X: 2, Y: 0, Required: 1},
				{ID: 30, X: 0, Y: 2, Required: 1},
			},
		}
		isl := p.FindIslandByID(20)
		if isl == nil {
			t.Fatal("expected island with ID=20")
		}
		if isl.X != 2 || isl.Y != 0 {
			t.Errorf("island 20 at (%d,%d), want (2,0)", isl.X, isl.Y)
		}
	})
}

// --- FindAdjacentIsland (P1) ---

func TestFindAdjacentIsland(t *testing.T) {
	t.Run("finds island to the right", func(t *testing.T) {
		p := fourCornerPuzzle()
		adj := p.FindAdjacentIsland(0, 1, 0)
		if adj == nil || adj.ID != 1 {
			t.Errorf("expected island 1, got %v", adj)
		}
	})

	t.Run("finds island below", func(t *testing.T) {
		p := fourCornerPuzzle()
		adj := p.FindAdjacentIsland(0, 0, 1)
		if adj == nil || adj.ID != 2 {
			t.Errorf("expected island 2, got %v", adj)
		}
	})

	t.Run("no island in direction", func(t *testing.T) {
		p := fourCornerPuzzle()
		adj := p.FindAdjacentIsland(0, -1, 0) // left from leftmost
		if adj != nil {
			t.Errorf("expected nil, got island %d", adj.ID)
		}
	})

	t.Run("stops at grid edge", func(t *testing.T) {
		p := fourCornerPuzzle()
		adj := p.FindAdjacentIsland(0, 0, -1) // up from top
		if adj != nil {
			t.Errorf("expected nil, got island %d", adj.ID)
		}
	})

	t.Run("blocked by perpendicular bridge", func(t *testing.T) {
		// Cross puzzle:
		//   . 0 .
		//   1 . 2
		//   . 3 .
		p := &Puzzle{
			Width:  3,
			Height: 3,
			Islands: []Island{
				{ID: 0, X: 1, Y: 0, Required: 1},
				{ID: 1, X: 0, Y: 1, Required: 1},
				{ID: 2, X: 2, Y: 1, Required: 1},
				{ID: 3, X: 1, Y: 2, Required: 1},
			},
		}
		p.SetBridge(1, 2, 1) // horizontal bridge through (1,1)

		// Vertical ray from 0 downward should be blocked at (1,1).
		adj := p.FindAdjacentIsland(0, 0, 1)
		if adj != nil {
			t.Errorf("expected nil (blocked by H bridge), got island %d", adj.ID)
		}
	})

	t.Run("not blocked by parallel bridge", func(t *testing.T) {
		// Two islands vertically aligned with a vertical bridge between another pair
		// in the same column. The ray should pass through a parallel bridge.
		p := &Puzzle{
			Width:  3,
			Height: 5,
			Islands: []Island{
				{ID: 0, X: 0, Y: 0, Required: 1},
				{ID: 1, X: 0, Y: 2, Required: 2},
				{ID: 2, X: 0, Y: 4, Required: 1},
			},
		}
		// Bridge 1-2 creates vertical bridge segments at (0,3).
		p.SetBridge(1, 2, 1)

		// Ray from 0 downward should find island 1 (not blocked by vertical bridge beyond).
		adj := p.FindAdjacentIsland(0, 0, 1)
		if adj == nil || adj.ID != 1 {
			t.Errorf("expected island 1, got %v", adj)
		}
	})
}

// --- CellContent (P1) ---

func TestCellContent(t *testing.T) {
	t.Run("island cell", func(t *testing.T) {
		p := fourCornerPuzzle()
		ci := p.CellContent(0, 0)
		if ci.Kind != cellIsland {
			t.Errorf("Kind = %d, want cellIsland(%d)", ci.Kind, cellIsland)
		}
		if ci.IslandID != 0 {
			t.Errorf("IslandID = %d, want 0", ci.IslandID)
		}
	})

	t.Run("horizontal bridge cell", func(t *testing.T) {
		p := fourCornerPuzzle()
		p.SetBridge(0, 1, 1)

		ci := p.CellContent(1, 0)
		if ci.Kind != cellBridgeH {
			t.Errorf("Kind = %d, want cellBridgeH(%d)", ci.Kind, cellBridgeH)
		}
	})

	t.Run("vertical bridge cell", func(t *testing.T) {
		p := fourCornerPuzzle()
		p.SetBridge(0, 2, 1)

		ci := p.CellContent(0, 1)
		if ci.Kind != cellBridgeV {
			t.Errorf("Kind = %d, want cellBridgeV(%d)", ci.Kind, cellBridgeV)
		}
	})

	t.Run("empty cell", func(t *testing.T) {
		p := fourCornerPuzzle()
		ci := p.CellContent(1, 1)
		if ci.Kind != cellEmpty {
			t.Errorf("Kind = %d, want cellEmpty(%d)", ci.Kind, cellEmpty)
		}
	})

	t.Run("bridge count", func(t *testing.T) {
		p := fourCornerPuzzle()
		p.SetBridge(0, 1, 2)

		ci := p.CellContent(1, 0) // bridge segment between (0,0) and (3,0)
		if ci.BridgeCount != 2 {
			t.Errorf("BridgeCount = %d, want 2", ci.BridgeCount)
		}
	})

	t.Run("lazy rebuild after modification", func(t *testing.T) {
		p := fourCornerPuzzle()
		_ = p.CellContent(1, 0) // build cache (empty)

		p.SetBridge(0, 1, 1) // invalidates cache

		ci := p.CellContent(1, 0) // should rebuild
		if ci.Kind != cellBridgeH {
			t.Errorf("Kind = %d after rebuild, want cellBridgeH(%d)", ci.Kind, cellBridgeH)
		}
	})
}

// --- Save/Load round-trip (P1) ---

func TestSaveLoadRoundTrip(t *testing.T) {
	t.Run("full round trip", func(t *testing.T) {
		original := &Model{
			puzzle: Puzzle{
				Width:  7,
				Height: 7,
				Islands: []Island{
					{ID: 0, X: 0, Y: 0, Required: 2},
					{ID: 1, X: 3, Y: 0, Required: 3},
					{ID: 2, X: 6, Y: 0, Required: 1},
					{ID: 3, X: 0, Y: 3, Required: 2},
				},
				Bridges: []Bridge{
					{Island1: 0, Island2: 1, Count: 1},
					{Island1: 1, Island2: 2, Count: 2},
					{Island1: 0, Island2: 3, Count: 1},
				},
			},
			cursorIsland: 1,
			keys:         DefaultKeyMap,
		}

		data, err := original.GetSave()
		if err != nil {
			t.Fatal(err)
		}

		got, err := ImportModel(data)
		if err != nil {
			t.Fatal(err)
		}

		// Dimensions.
		if got.puzzle.Width != 7 || got.puzzle.Height != 7 {
			t.Errorf("dimensions = %dx%d, want 7x7", got.puzzle.Width, got.puzzle.Height)
		}

		// Islands.
		if len(got.puzzle.Islands) != 4 {
			t.Fatalf("island count = %d, want 4", len(got.puzzle.Islands))
		}
		for i, isl := range original.puzzle.Islands {
			g := got.puzzle.Islands[i]
			if g.ID != isl.ID || g.X != isl.X || g.Y != isl.Y || g.Required != isl.Required {
				t.Errorf("island[%d] = %+v, want %+v", i, g, isl)
			}
		}

		// Bridges.
		if len(got.puzzle.Bridges) != 3 {
			t.Fatalf("bridge count = %d, want 3", len(got.puzzle.Bridges))
		}
		for i, b := range original.puzzle.Bridges {
			g := got.puzzle.Bridges[i]
			if g.Island1 != b.Island1 || g.Island2 != b.Island2 || g.Count != b.Count {
				t.Errorf("bridge[%d] = %+v, want %+v", i, g, b)
			}
		}
	})

	t.Run("caches rebuilt on access", func(t *testing.T) {
		m := &Model{
			puzzle: Puzzle{
				Width:  3,
				Height: 1,
				Islands: []Island{
					{ID: 0, X: 0, Y: 0, Required: 1},
					{ID: 1, X: 2, Y: 0, Required: 1},
				},
				Bridges: []Bridge{
					{Island1: 0, Island2: 1, Count: 1},
				},
			},
			keys: DefaultKeyMap,
		}

		data, err := m.GetSave()
		if err != nil {
			t.Fatal(err)
		}

		got, err := ImportModel(data)
		if err != nil {
			t.Fatal(err)
		}

		// Caches should be nil after import; accessing BridgeCount triggers rebuild.
		if got.puzzle.bridgeCounts != nil {
			t.Error("bridgeCounts should be nil immediately after import")
		}
		if cnt := got.puzzle.BridgeCount(0); cnt != 1 {
			t.Errorf("BridgeCount(0) = %d, want 1", cnt)
		}
	})

	t.Run("invalid JSON", func(t *testing.T) {
		_, err := ImportModel([]byte("not json"))
		if err == nil {
			t.Fatal("expected error for invalid JSON")
		}
	})
}

// --- placeIslands (P2) ---

func TestPlaceIslands(t *testing.T) {
	t.Run("places correct count", func(t *testing.T) {
		for range 10 {
			islands := placeIslands(7, 7, 6)
			if islands == nil {
				continue // may fail on rare occasions
			}
			if len(islands) != 6 {
				t.Errorf("len(islands) = %d, want 6", len(islands))
			}
		}
	})

	t.Run("no diagonal adjacency", func(t *testing.T) {
		for range 20 {
			islands := placeIslands(7, 7, 6)
			if islands == nil {
				continue
			}
			for i := 0; i < len(islands); i++ {
				for j := i + 1; j < len(islands); j++ {
					dx := abs(islands[i].X - islands[j].X)
					dy := abs(islands[i].Y - islands[j].Y)
					if dx <= 1 && dy <= 1 && dx > 0 && dy > 0 {
						t.Errorf("islands %d(%d,%d) and %d(%d,%d) are diagonally adjacent",
							islands[i].ID, islands[i].X, islands[i].Y,
							islands[j].ID, islands[j].X, islands[j].Y)
					}
				}
			}
		}
	})

	t.Run("returns nil when impossible", func(t *testing.T) {
		islands := placeIslands(3, 3, 100)
		if islands != nil {
			t.Errorf("expected nil for 100 islands on 3x3, got %d islands", len(islands))
		}
	})

	t.Run("all within bounds", func(t *testing.T) {
		for range 20 {
			islands := placeIslands(7, 7, 6)
			if islands == nil {
				continue
			}
			for _, isl := range islands {
				if isl.X < 0 || isl.X >= 7 || isl.Y < 0 || isl.Y >= 7 {
					t.Errorf("island %d at (%d,%d) out of bounds", isl.ID, isl.X, isl.Y)
				}
			}
		}
	})
}

// --- isDirectlyConnectable (P2) ---

func TestIsDirectlyConnectable(t *testing.T) {
	t.Run("aligned no obstacle", func(t *testing.T) {
		p := &Puzzle{
			Width:  5,
			Height: 1,
			Islands: []Island{
				{ID: 0, X: 0, Y: 0, Required: 1},
				{ID: 1, X: 4, Y: 0, Required: 1},
			},
		}
		if !isDirectlyConnectable(p, p.Islands[0], p.Islands[1]) {
			t.Error("aligned islands with no obstacle should be connectable")
		}
	})

	t.Run("aligned island between", func(t *testing.T) {
		p := linePuzzle()
		// Islands 0,1,2,3 at x=0,2,4,6. Check if 0 and 2 are directly connectable
		// (island 1 at x=2 sits between them... actually 1 is at the endpoint).
		// Instead: 0 at x=0 and 3 at x=6 — islands 1(x=2) and 2(x=4) are between.
		if isDirectlyConnectable(p, p.Islands[0], p.Islands[3]) {
			t.Error("island between should block direct connection")
		}
	})

	t.Run("not aligned", func(t *testing.T) {
		p := fourCornerPuzzle()
		// Islands 0(0,0) and 3(3,3) are diagonal — not aligned.
		if isDirectlyConnectable(p, p.Islands[0], p.Islands[3]) {
			t.Error("diagonal pair should not be connectable")
		}
	})

	t.Run("adjacent distance 1", func(t *testing.T) {
		p := &Puzzle{
			Width:  2,
			Height: 1,
			Islands: []Island{
				{ID: 0, X: 0, Y: 0, Required: 1},
				{ID: 1, X: 1, Y: 0, Required: 1},
			},
		}
		if !isDirectlyConnectable(p, p.Islands[0], p.Islands[1]) {
			t.Error("adjacent islands should be connectable")
		}
	})
}

// --- GeneratePuzzle (P2) ---

func TestGeneratePuzzle(t *testing.T) {
	mode := NewMode("Test", "test mode", 7, 7, 8, 10)

	t.Run("produces valid puzzle", func(t *testing.T) {
		puzzle, err := GeneratePuzzle(mode)
		if err != nil {
			t.Fatal(err)
		}
		for _, isl := range puzzle.Islands {
			if isl.Required < 1 {
				t.Errorf("island %d has Required=%d, want >= 1", isl.ID, isl.Required)
			}
		}
	})

	t.Run("bridges cleared", func(t *testing.T) {
		puzzle, err := GeneratePuzzle(mode)
		if err != nil {
			t.Fatal(err)
		}
		if len(puzzle.Bridges) != 0 {
			t.Errorf("len(Bridges) = %d, want 0", len(puzzle.Bridges))
		}
	})

	t.Run("correct dimensions", func(t *testing.T) {
		puzzle, err := GeneratePuzzle(mode)
		if err != nil {
			t.Fatal(err)
		}
		if puzzle.Width != 7 || puzzle.Height != 7 {
			t.Errorf("dimensions = %dx%d, want 7x7", puzzle.Width, puzzle.Height)
		}
	})
}

// --- Save JSON structure (P1) ---

func TestSaveJSON(t *testing.T) {
	m := Model{
		puzzle: Puzzle{
			Width:  4,
			Height: 4,
			Islands: []Island{
				{ID: 0, X: 0, Y: 0, Required: 2},
			},
			Bridges: []Bridge{
				{Island1: 0, Island2: 1, Count: 1},
			},
		},
		keys: DefaultKeyMap,
	}

	data, err := m.GetSave()
	if err != nil {
		t.Fatal(err)
	}

	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatal(err)
	}

	for _, key := range []string{"width", "height", "islands", "bridges"} {
		if _, ok := raw[key]; !ok {
			t.Errorf("missing key %q in save JSON", key)
		}
	}
}
