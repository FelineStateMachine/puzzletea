package ui

import (
	"strings"
	"testing"
	"unicode"

	"github.com/FelineStateMachine/puzzletea/store"
	"github.com/FelineStateMachine/puzzletea/theme"

	"charm.land/bubbles/v2/list"
	"github.com/charmbracelet/x/ansi"
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

func TestInitCategoryListDisablesQuitAndFilter(t *testing.T) {
	l := InitCategoryList(nil, "Categories")

	if l.FilteringEnabled() {
		t.Error("expected filtering to be disabled")
	}

	if l.ShowHelp() {
		t.Error("expected help to be disabled")
	}
}

func TestThemeListFilterDoesNotSplitTitle(t *testing.T) {
	longest := longestThemeName(t)

	items := []list.Item{
		MenuItem{ItemTitle: "Short", Desc: "dark theme"},
		MenuItem{ItemTitle: longest, Desc: "dark theme"},
	}
	l := InitThemeList(items, theme.MaxNameLen+4, 12)
	l.Select(1)
	l.SetFilterText(alternatingCasePrefix(longest, 6))

	rendered := ansi.Strip(l.View())
	if !strings.Contains(strings.ToLower(rendered), strings.ToLower(longest)) {
		t.Fatalf("expected rendered output to contain full longest theme name %q\nview:\n%s", longest, rendered)
	}

	nameRunes := []rune(longest)
	if len(nameRunes) >= 8 {
		head := string(nameRunes[:4])
		tail := string(nameRunes[4:8])
		if strings.Contains(strings.ToLower(rendered), strings.ToLower(head+"\n"+tail)) {
			t.Fatalf("expected longest theme title not to split across lines\nview:\n%s", rendered)
		}
	}
}

func TestThemeListWidthStableWithSelectionAndFilter(t *testing.T) {
	longest := longestThemeName(t)

	items := []list.Item{
		MenuItem{ItemTitle: "Dracula", Desc: "dark theme"},
		MenuItem{ItemTitle: longest, Desc: "dark theme"},
	}
	l := InitThemeList(items, theme.MaxNameLen+4, 12)

	baseline := renderedWidth(l.View())
	if baseline <= 0 {
		t.Fatalf("expected positive rendered width, got %d", baseline)
	}

	l.Select(1)
	if got := renderedWidth(l.View()); got != baseline {
		t.Fatalf("expected stable width after selecting longest theme, got %d want %d", got, baseline)
	}

	l.SetFilterText(alternatingCasePrefix(longest, 6))
	if got := renderedWidth(l.View()); got != baseline {
		t.Fatalf("expected stable width with filter applied, got %d want %d", got, baseline)
	}

	l.Select(0)
	if got := renderedWidth(l.View()); got != baseline {
		t.Fatalf("expected stable width after selecting short theme while filtered, got %d want %d", got, baseline)
	}
}

func longestThemeName(t *testing.T) string {
	t.Helper()

	longest := theme.DefaultThemeName
	themes := theme.AllThemes()
	if len(themes) == 0 {
		t.Fatal("expected at least one embedded theme")
	}
	for _, th := range themes {
		if len(th.Name) > len(longest) {
			longest = th.Name
		}
	}
	return longest
}

func alternatingCasePrefix(s string, n int) string {
	r := []rune(s)
	if n > len(r) {
		n = len(r)
	}
	r = r[:n]
	for i, ch := range r {
		if i%2 == 0 {
			r[i] = unicode.ToLower(ch)
		} else {
			r[i] = unicode.ToUpper(ch)
		}
	}
	return string(r)
}

func renderedWidth(s string) int {
	maxW := 0
	for _, line := range strings.Split(s, "\n") {
		if w := ansi.StringWidth(line); w > maxW {
			maxW = w
		}
	}
	return maxW
}
