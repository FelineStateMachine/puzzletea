package ui

import (
	"testing"

	"github.com/FelineStateMachine/puzzletea/store"

	"charm.land/bubbles/v2/list"
)

// --- MenuItem (P0) ---

func TestMenuItemTitle(t *testing.T) {
	item := MenuItem{ItemTitle: "Generate", Desc: "a new puzzle."}
	if got := item.Title(); got != "Generate" {
		t.Errorf("Title() = %q, want %q", got, "Generate")
	}
}

func TestMenuItemDescription(t *testing.T) {
	item := MenuItem{ItemTitle: "Generate", Desc: "a new puzzle."}
	if got := item.Description(); got != "a new puzzle." {
		t.Errorf("Description() = %q, want %q", got, "a new puzzle.")
	}
}

func TestMenuItemFilterValue(t *testing.T) {
	item := MenuItem{ItemTitle: "Generate", Desc: "a new puzzle."}
	if got := item.FilterValue(); got != "Generate" {
		t.Errorf("FilterValue() = %q, want %q", got, "Generate")
	}
}

// --- FormatStatus (P0) ---

func TestFormatStatus(t *testing.T) {
	tests := []struct {
		status store.GameStatus
		want   string
	}{
		{store.StatusNew, "New"},
		{store.StatusInProgress, "In Progress"},
		{store.StatusCompleted, "Completed"},
		{store.StatusAbandoned, "abandoned"},
	}

	for _, tt := range tests {
		t.Run(string(tt.status), func(t *testing.T) {
			got := FormatStatus(tt.status)
			if got != tt.want {
				t.Errorf("FormatStatus(%q) = %q, want %q", tt.status, got, tt.want)
			}
		})
	}
}

// --- InitList (P1) ---

func TestInitList(t *testing.T) {
	items := []list.Item{
		MenuItem{ItemTitle: "One", Desc: "first"},
		MenuItem{ItemTitle: "Two", Desc: "second"},
	}

	l := InitList(items, "Test Menu")

	if l.Title != "Test Menu" {
		t.Errorf("Title = %q, want %q", l.Title, "Test Menu")
	}

	allItems := l.Items()
	if len(allItems) != 2 {
		t.Fatalf("Items count = %d, want 2", len(allItems))
	}

	first := allItems[0].(MenuItem)
	if first.Title() != "One" {
		t.Errorf("first item Title = %q, want %q", first.Title(), "One")
	}
}

func TestInitListDisablesQuitAndFilter(t *testing.T) {
	l := InitList(nil, "Empty")

	// Filtering should be disabled.
	if l.FilteringEnabled() {
		t.Error("expected filtering to be disabled")
	}

	// ShowHelp should be disabled.
	if l.ShowHelp() {
		t.Error("expected help to be disabled")
	}
}
