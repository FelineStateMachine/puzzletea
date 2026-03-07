package app

import (
	"testing"

	"charm.land/bubbles/v2/list"
	"github.com/FelineStateMachine/puzzletea/game"
)

type testMode struct {
	game.BaseMode
}

func newTestMode(title string) testMode {
	return testMode{BaseMode: game.NewBaseMode(title, title+" description")}
}

func TestModeDisplayTitlesStripUniqueGridSizes(t *testing.T) {
	t.Parallel()

	cat := game.Category{
		Name: "Example",
		Modes: []list.Item{
			newTestMode("Mini 5x5"),
			newTestMode("Hard 10x10"),
		},
	}

	got := modeDisplayTitles(cat)
	want := []string{"Mini", "Hard"}

	if len(got) != len(want) {
		t.Fatalf("len(modeDisplayTitles()) = %d, want %d", len(got), len(want))
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("modeDisplayTitles()[%d] = %q, want %q", i, got[i], want[i])
		}
	}
}

func TestModeDisplayTitlesKeepDuplicateBaseNames(t *testing.T) {
	t.Parallel()

	cat := game.Category{
		Name: "Example",
		Modes: []list.Item{
			newTestMode("Easy 5x5"),
			newTestMode("Easy 10x10"),
			newTestMode("Hard 10x10"),
		},
	}

	got := modeDisplayTitles(cat)
	want := []string{"Easy 5x5", "Easy 10x10", "Hard"}

	if len(got) != len(want) {
		t.Fatalf("len(modeDisplayTitles()) = %d, want %d", len(got), len(want))
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("modeDisplayTitles()[%d] = %q, want %q", i, got[i], want[i])
		}
	}
}

func TestBuildModeDisplayItemsPreservesOriginalMode(t *testing.T) {
	t.Parallel()

	original := newTestMode("Hard 10x10")
	cat := game.Category{
		Name:  "Example",
		Modes: []list.Item{original},
	}

	items := buildModeDisplayItems(cat)
	if len(items) != 1 {
		t.Fatalf("len(buildModeDisplayItems()) = %d, want 1", len(items))
	}

	displayMode, ok := items[0].(game.Mode)
	if !ok {
		t.Fatal("buildModeDisplayItems()[0] does not implement game.Mode")
	}
	if got := displayMode.Title(); got != "Hard" {
		t.Fatalf("display item Title() = %q, want %q", got, "Hard")
	}

	mode, ok := unwrapModeDisplayItem(items[0]).(game.Mode)
	if !ok {
		t.Fatal("unwrapModeDisplayItem() did not return a game.Mode")
	}
	if got := mode.Title(); got != "Hard 10x10" {
		t.Fatalf("unwrapped mode Title() = %q, want %q", got, "Hard 10x10")
	}
}
