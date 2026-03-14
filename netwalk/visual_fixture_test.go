package netwalk

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/ansi"
)

func TestVisualFixtureJSONLMatchesCommittedFile(t *testing.T) {
	want, err := VisualFixtureJSONL()
	if err != nil {
		t.Fatalf("VisualFixtureJSONL() error = %v", err)
	}

	path := filepath.Join("testdata", "visual_states.jsonl")
	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read fixture file: %v", err)
	}

	if string(got) != string(want) {
		t.Fatalf("committed fixture is out of sync; regenerate %s", path)
	}
}

func TestVisualFixtureCasesCoverExpectedViews(t *testing.T) {
	wantNames := []string{
		"cursor-root-horizontal",
		"leaf-gallery",
		"straight-and-corner-gallery",
		"tee-and-cross-gallery",
		"connected-horizontal-bridge",
		"connected-vertical-bridge",
		"disconnected-default-foreground",
		"dangling-error-state",
		"locked-root-cursor",
		"solved-with-empty-cells",
	}
	if len(visualFixtureCases) != len(wantNames) {
		t.Fatalf("fixture case count = %d, want %d", len(visualFixtureCases), len(wantNames))
	}
	for i, want := range wantNames {
		if got := visualFixtureCases[i].name; got != want {
			t.Fatalf("fixture case %d = %q, want %q", i, got, want)
		}
	}
}

func TestVisualFixtureRepresentativeViews(t *testing.T) {
	t.Run("dangling case shows error status", func(t *testing.T) {
		view := fixtureView(t, "dangling-error-state")
		if !strings.Contains(view, "dangling 1") {
			t.Fatalf("expected dangling status in view:\n%s", view)
		}
	})

	t.Run("locked root case reports lock count", func(t *testing.T) {
		view := fixtureView(t, "locked-root-cursor")
		if !strings.Contains(view, "locks 1") {
			t.Fatalf("expected lock count in view:\n%s", view)
		}
	})

	t.Run("solved case shows solved badge", func(t *testing.T) {
		view := fixtureView(t, "solved-with-empty-cells")
		if !strings.Contains(view, "SOLVED") {
			t.Fatalf("expected solved badge in view:\n%s", view)
		}
	})
}

func fixtureView(t *testing.T, name string) string {
	t.Helper()

	for _, fixture := range visualFixtureCases {
		if fixture.name != name {
			continue
		}

		save, err := visualFixtureSave(fixture.puzzle)
		if err != nil {
			t.Fatalf("visualFixtureSave(%q) error = %v", name, err)
		}
		model, err := ImportModel(save)
		if err != nil {
			t.Fatalf("ImportModel(%q) error = %v", name, err)
		}
		g, _ := model.Update(tea.WindowSizeMsg{Width: 120, Height: 40})
		return ansi.Strip(g.View())
	}

	t.Fatalf("fixture %q not found", name)
	return ""
}
