package app

import (
	"testing"

	"github.com/FelineStateMachine/puzzletea/game"
	"github.com/FelineStateMachine/puzzletea/puzzle"
	"github.com/FelineStateMachine/puzzletea/registry"
)

type testMode struct {
	game.BaseMode
}

func newTestMode(title string) testMode {
	return testMode{BaseMode: game.NewBaseMode(title, title+" description")}
}

func TestModeDisplayTitlesStripUniqueGridSizes(t *testing.T) {
	t.Parallel()

	entry := registry.Entry{
		Definition: puzzle.Definition{Name: "Example"},
		Modes: []registry.ModeEntry{
			{Definition: puzzle.ModeDef{Title: "Mini 5x5", Description: "Mini 5x5 description"}},
			{Definition: puzzle.ModeDef{Title: "Hard 10x10", Description: "Hard 10x10 description"}},
		},
	}

	got := modeDisplayTitles(entry)
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

	entry := registry.Entry{
		Definition: puzzle.Definition{Name: "Example"},
		Modes: []registry.ModeEntry{
			{Definition: puzzle.ModeDef{Title: "Easy 5x5", Description: "Easy 5x5 description"}},
			{Definition: puzzle.ModeDef{Title: "Easy 10x10", Description: "Easy 10x10 description"}},
			{Definition: puzzle.ModeDef{Title: "Hard 10x10", Description: "Hard 10x10 description"}},
		},
	}

	got := modeDisplayTitles(entry)
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
	entry := registry.Entry{
		Definition: puzzle.Definition{Name: "Example"},
		Modes: []registry.ModeEntry{{
			Definition: puzzle.ModeDef{Title: original.Title(), Description: original.Description()},
		}},
	}

	items := buildModeDisplayItems(entry)
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

	mode, ok := unwrapModeDisplayItem(items[0]).(registry.ModeEntry)
	if !ok {
		t.Fatal("unwrapModeDisplayItem() did not return a registry.ModeEntry")
	}
	if got := mode.Definition.Title; got != "Hard 10x10" {
		t.Fatalf("unwrapped mode Title() = %q, want %q", got, "Hard 10x10")
	}
}
