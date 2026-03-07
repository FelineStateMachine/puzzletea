package app

import (
	"testing"

	"github.com/FelineStateMachine/puzzletea/game"
	"github.com/FelineStateMachine/puzzletea/theme"
)

func TestHelpMarkdownStyleUsesPaletteColors(t *testing.T) {
	if err := theme.Apply(""); err != nil {
		t.Fatalf("apply default theme: %v", err)
	}
	t.Cleanup(func() { _ = theme.Apply("") })

	p := theme.Current()
	style := helpMarkdownStyle(p)

	if got := *style.H1.Color; got != hexColor(p.Accent) {
		t.Fatalf("H1 color = %q, want %q", got, hexColor(p.Accent))
	}
	if got := *style.Code.StylePrimitive.BackgroundColor; got != hexColor(p.Surface) {
		t.Fatalf("inline code background = %q, want %q", got, hexColor(p.Surface))
	}
	if got := *style.Link.Color; got != hexColor(p.Highlight) {
		t.Fatalf("link color = %q, want %q", got, hexColor(p.Highlight))
	}
}

func TestHelpMarkdownThemeKeyChangesWithTheme(t *testing.T) {
	if err := theme.Apply(""); err != nil {
		t.Fatalf("apply default theme: %v", err)
	}
	t.Cleanup(func() { _ = theme.Apply("") })

	defaultKey := helpMarkdownThemeKey(theme.Current())

	if err := theme.Apply("Dracula"); err != nil {
		t.Fatalf("apply Dracula theme: %v", err)
	}

	draculaKey := helpMarkdownThemeKey(theme.Current())
	if draculaKey == defaultKey {
		t.Fatalf("theme key did not change across themes: %q", draculaKey)
	}
}

func TestUpdateHelpDetailViewportRebuildsRendererAfterThemeChange(t *testing.T) {
	if err := theme.Apply(""); err != nil {
		t.Fatalf("apply default theme: %v", err)
	}
	t.Cleanup(func() { _ = theme.Apply("") })

	m := model{
		width:  100,
		height: 30,
		help: helpState{
			category: game.Category{
				Name: "Nonogram",
				Help: "# Nonogram\n\n- clue\n",
			},
		},
	}

	m = m.updateHelpDetailViewport()
	if m.help.renderer == nil {
		t.Fatal("expected initial help renderer")
	}

	initialRenderer := m.help.renderer
	initialThemeKey := m.help.rendererTheme

	m = m.updateHelpDetailViewport()
	if m.help.renderer != initialRenderer {
		t.Fatal("expected help renderer to be reused when width and theme are unchanged")
	}

	if err := theme.Apply("Dracula"); err != nil {
		t.Fatalf("apply Dracula theme: %v", err)
	}

	m = m.updateHelpDetailViewport()
	if m.help.renderer == nil {
		t.Fatal("expected help renderer after theme change")
	}
	if m.help.renderer == initialRenderer {
		t.Fatal("expected help renderer to rebuild after theme change")
	}
	if m.help.rendererTheme == initialThemeKey {
		t.Fatalf("renderer theme key = %q, want change from %q", m.help.rendererTheme, initialThemeKey)
	}
}
