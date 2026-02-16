package game

import (
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
)

// --- Cursor.Move (P0) ---

func TestCursorMove(t *testing.T) {
	keys := DefaultCursorKeyMap

	keyMsg := func(k string) tea.KeyPressMsg {
		return tea.KeyPressMsg{Code: rune(k[0]), Text: k}
	}
	specialKeyMsg := func(k rune) tea.KeyPressMsg {
		return tea.KeyPressMsg{Code: k}
	}

	t.Run("movement within bounds", func(t *testing.T) {
		tests := []struct {
			name      string
			startX    int
			startY    int
			msg       tea.KeyPressMsg
			maxX      int
			maxY      int
			wantX     int
			wantY     int
			wantMoved bool
		}{
			{"move up within bounds", 2, 2, specialKeyMsg(tea.KeyUp), 4, 4, 2, 1, true},
			{"move down within bounds", 2, 2, specialKeyMsg(tea.KeyDown), 4, 4, 2, 3, true},
			{"move left within bounds", 2, 2, specialKeyMsg(tea.KeyLeft), 4, 4, 1, 2, true},
			{"move right within bounds", 2, 2, specialKeyMsg(tea.KeyRight), 4, 4, 3, 2, true},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				c := Cursor{X: tt.startX, Y: tt.startY}
				moved := c.Move(keys, tt.msg, tt.maxX, tt.maxY)
				if moved != tt.wantMoved {
					t.Errorf("Move() = %v, want %v", moved, tt.wantMoved)
				}
				if c.X != tt.wantX || c.Y != tt.wantY {
					t.Errorf("cursor = (%d,%d), want (%d,%d)", c.X, c.Y, tt.wantX, tt.wantY)
				}
			})
		}
	})

	t.Run("blocked at edges", func(t *testing.T) {
		tests := []struct {
			name   string
			startX int
			startY int
			msg    tea.KeyPressMsg
			maxX   int
			maxY   int
		}{
			{"blocked at top edge", 2, 0, specialKeyMsg(tea.KeyUp), 4, 4},
			{"blocked at bottom edge", 2, 4, specialKeyMsg(tea.KeyDown), 4, 4},
			{"blocked at left edge", 0, 2, specialKeyMsg(tea.KeyLeft), 4, 4},
			{"blocked at right edge", 4, 2, specialKeyMsg(tea.KeyRight), 4, 4},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				c := Cursor{X: tt.startX, Y: tt.startY}
				moved := c.Move(keys, tt.msg, tt.maxX, tt.maxY)
				if moved {
					t.Error("Move() = true, want false at edge")
				}
				if c.X != tt.startX || c.Y != tt.startY {
					t.Errorf("cursor moved to (%d,%d), want (%d,%d)", c.X, c.Y, tt.startX, tt.startY)
				}
			})
		}
	})

	t.Run("1x1 grid no movement", func(t *testing.T) {
		directions := []tea.KeyPressMsg{
			specialKeyMsg(tea.KeyUp),
			specialKeyMsg(tea.KeyDown),
			specialKeyMsg(tea.KeyLeft),
			specialKeyMsg(tea.KeyRight),
		}
		for _, msg := range directions {
			c := Cursor{X: 0, Y: 0}
			moved := c.Move(keys, msg, 0, 0)
			if moved {
				t.Errorf("Move() = true on 1x1 grid for key %v", msg)
			}
			if c.X != 0 || c.Y != 0 {
				t.Errorf("cursor = (%d,%d), want (0,0)", c.X, c.Y)
			}
		}
	})

	t.Run("arrow keys work", func(t *testing.T) {
		tests := []struct {
			msg   tea.KeyPressMsg
			wantX int
			wantY int
		}{
			{specialKeyMsg(tea.KeyUp), 2, 1},
			{specialKeyMsg(tea.KeyDown), 2, 3},
			{specialKeyMsg(tea.KeyLeft), 1, 2},
			{specialKeyMsg(tea.KeyRight), 3, 2},
		}
		for _, tt := range tests {
			c := Cursor{X: 2, Y: 2}
			c.Move(keys, tt.msg, 4, 4)
			if c.X != tt.wantX || c.Y != tt.wantY {
				t.Errorf("arrow key: cursor = (%d,%d), want (%d,%d)", c.X, c.Y, tt.wantX, tt.wantY)
			}
		}
	})

	t.Run("WASD keys work", func(t *testing.T) {
		tests := []struct {
			key   string
			wantX int
			wantY int
		}{
			{"w", 2, 1},
			{"s", 2, 3},
			{"a", 1, 2},
			{"d", 3, 2},
		}
		for _, tt := range tests {
			c := Cursor{X: 2, Y: 2}
			c.Move(keys, keyMsg(tt.key), 4, 4)
			if c.X != tt.wantX || c.Y != tt.wantY {
				t.Errorf("WASD %q: cursor = (%d,%d), want (%d,%d)", tt.key, c.X, c.Y, tt.wantX, tt.wantY)
			}
		}
	})

	t.Run("vim keys work", func(t *testing.T) {
		tests := []struct {
			key   string
			wantX int
			wantY int
		}{
			{"k", 2, 1},
			{"j", 2, 3},
			{"h", 1, 2},
			{"l", 3, 2},
		}
		for _, tt := range tests {
			c := Cursor{X: 2, Y: 2}
			c.Move(keys, keyMsg(tt.key), 4, 4)
			if c.X != tt.wantX || c.Y != tt.wantY {
				t.Errorf("vim %q: cursor = (%d,%d), want (%d,%d)", tt.key, c.X, c.Y, tt.wantX, tt.wantY)
			}
		}
	})

	t.Run("unrelated key no movement", func(t *testing.T) {
		c := Cursor{X: 2, Y: 2}
		moved := c.Move(keys, specialKeyMsg(tea.KeyEnter), 4, 4)
		if moved {
			t.Error("Move() = true for unrelated key, want false")
		}
		if c.X != 2 || c.Y != 2 {
			t.Errorf("cursor = (%d,%d), want (2,2)", c.X, c.Y)
		}
	})
}

// --- BaseMode (P1) ---

func TestBaseMode(t *testing.T) {
	t.Run("normal values", func(t *testing.T) {
		b := NewBaseMode("Easy", "5x5 grid")
		if got := b.Title(); got != "Easy" {
			t.Errorf("Title() = %q, want %q", got, "Easy")
		}
		if got := b.Description(); got != "5x5 grid" {
			t.Errorf("Description() = %q, want %q", got, "5x5 grid")
		}
		if got := b.FilterValue(); got != "Easy 5x5 grid" {
			t.Errorf("FilterValue() = %q, want %q", got, "Easy 5x5 grid")
		}
	})

	t.Run("empty strings", func(t *testing.T) {
		b := NewBaseMode("", "")
		if got := b.Title(); got != "" {
			t.Errorf("Title() = %q, want %q", got, "")
		}
		if got := b.Description(); got != "" {
			t.Errorf("Description() = %q, want %q", got, "")
		}
		if got := b.FilterValue(); got != " " {
			t.Errorf("FilterValue() = %q, want %q", got, " ")
		}
	})
}

// --- Category (P3) ---

func TestCategory(t *testing.T) {
	c := Category{Name: "Sudoku", Desc: "Classic"}

	if got := c.Title(); got != "Sudoku" {
		t.Errorf("Title() = %q, want %q", got, "Sudoku")
	}
	if got := c.Description(); got != "Classic" {
		t.Errorf("Description() = %q, want %q", got, "Classic")
	}
	if got := c.FilterValue(); got != "Sudoku" {
		t.Errorf("FilterValue() = %q, want %q", got, "Sudoku")
	}
}

// --- Registry (P1) ---

func TestRegistry(t *testing.T) {
	// Save and restore the global registry to avoid pollution.
	origRegistry := make(map[string]func([]byte) (Gamer, error))
	for k, v := range Registry {
		origRegistry[k] = v
	}
	t.Cleanup(func() {
		Registry = origRegistry
	})

	t.Run("Register adds function", func(t *testing.T) {
		called := false
		fn := func(data []byte) (Gamer, error) {
			called = true
			return nil, nil
		}
		Register("test-game", fn)
		if Registry["test-game"] == nil {
			t.Fatal("expected registry entry to be set")
		}
		if _, err := Registry["test-game"](nil); err != nil {
			t.Fatal(err)
		}
		if !called {
			t.Error("expected registered function to be called")
		}
	})

	t.Run("Register overwrites", func(t *testing.T) {
		first := false
		second := false
		Register("overwrite", func([]byte) (Gamer, error) {
			first = true
			return nil, nil
		})
		Register("overwrite", func([]byte) (Gamer, error) {
			second = true
			return nil, nil
		})
		if _, err := Registry["overwrite"](nil); err != nil {
			t.Fatal(err)
		}
		if first {
			t.Error("first function should not have been called")
		}
		if !second {
			t.Error("second function should have been called")
		}
	})

	t.Run("lookup missing key", func(t *testing.T) {
		if Registry["nonexistent"] != nil {
			t.Error("expected nil for missing key")
		}
	})
}

// --- TitleBarView (P3) ---

func TestTitleBarView(t *testing.T) {
	t.Run("unsolved game", func(t *testing.T) {
		result := TitleBarView("Sudoku", "Easy", false)
		if !strings.Contains(result, "Sudoku") {
			t.Error("expected result to contain game name")
		}
		if !strings.Contains(result, "Easy") {
			t.Error("expected result to contain mode name")
		}
		if strings.Contains(result, "SOLVED") {
			t.Error("expected result to not contain SOLVED")
		}
	})

	t.Run("solved game", func(t *testing.T) {
		result := TitleBarView("Sudoku", "Easy", true)
		if !strings.Contains(result, "SOLVED") {
			t.Error("expected result to contain SOLVED")
		}
	})
}

// --- DebugHeader (P3) ---

func TestDebugHeader(t *testing.T) {
	t.Run("with rows", func(t *testing.T) {
		rows := [][2]string{
			{"Status", "In Progress"},
			{"Size", "5x5"},
		}
		result := DebugHeader("TestGame", rows)
		if !strings.Contains(result, "TestGame") {
			t.Error("expected result to contain title")
		}
		if !strings.Contains(result, "Status") || !strings.Contains(result, "In Progress") {
			t.Error("expected result to contain first key-value pair")
		}
		if !strings.Contains(result, "Size") || !strings.Contains(result, "5x5") {
			t.Error("expected result to contain second key-value pair")
		}
	})

	t.Run("empty rows", func(t *testing.T) {
		result := DebugHeader("TestGame", nil)
		if !strings.Contains(result, "TestGame") {
			t.Error("expected result to contain title")
		}
	})
}

// --- DebugTable (P3) ---

func TestDebugTable(t *testing.T) {
	t.Run("normal table", func(t *testing.T) {
		headers := []string{"Col", "Hint", "Current"}
		rows := [][]string{
			{"0", "1,2", "1,2"},
			{"1", "3", "3"},
		}
		result := DebugTable("Tomography", headers, rows)
		if !strings.Contains(result, "Tomography") {
			t.Error("expected result to contain heading")
		}
		for _, h := range headers {
			if !strings.Contains(result, h) {
				t.Errorf("expected result to contain header %q", h)
			}
		}
		if !strings.Contains(result, "1,2") {
			t.Error("expected result to contain row values")
		}
	})

	t.Run("empty rows", func(t *testing.T) {
		headers := []string{"Col", "Hint"}
		result := DebugTable("Heading", headers, nil)
		if !strings.Contains(result, "Heading") {
			t.Error("expected result to contain heading")
		}
		for _, h := range headers {
			if !strings.Contains(result, h) {
				t.Errorf("expected result to contain header %q", h)
			}
		}
	})
}
