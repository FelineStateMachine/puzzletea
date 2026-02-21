package theme

import (
	"fmt"
	"image/color"
	"testing"
)

// --- Embedded theme loading (P0) ---

func TestParseEmbeddedThemes(t *testing.T) {
	themes := AllThemes()
	if len(themes) == 0 {
		t.Fatal("expected at least one embedded theme")
	}
	// Spot-check a well-known theme exists.
	found := false
	for _, th := range themes {
		if th.Name == "Dracula" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected to find Dracula theme in embedded themes")
	}
}

func TestThemeNames(t *testing.T) {
	names := ThemeNames()
	if len(names) < 2 {
		t.Fatalf("expected at least 2 theme names, got %d", len(names))
	}
	if names[0] != DefaultThemeName {
		t.Errorf("first theme name should be %q, got %q", DefaultThemeName, names[0])
	}
}

func TestMaxNameLen(t *testing.T) {
	maxFromAllThemes := len(DefaultThemeName)
	for _, th := range AllThemes() {
		if len(th.Name) > maxFromAllThemes {
			maxFromAllThemes = len(th.Name)
		}
	}
	if MaxNameLen != maxFromAllThemes {
		t.Errorf("MaxNameLen = %d, want %d from AllThemes", MaxNameLen, maxFromAllThemes)
	}

	maxFromThemeNames := 0
	for _, name := range ThemeNames() {
		if len(name) > maxFromThemeNames {
			maxFromThemeNames = len(name)
		}
	}
	if MaxNameLen != maxFromThemeNames {
		t.Errorf("MaxNameLen = %d, want %d from ThemeNames", MaxNameLen, maxFromThemeNames)
	}
}

func TestLookupTheme(t *testing.T) {
	th := LookupTheme("Dracula")
	if th == nil {
		t.Fatal("expected to find Dracula theme")
	}
	if th.Name != "Dracula" {
		t.Errorf("expected name Dracula, got %q", th.Name)
	}
	if !th.Meta.IsDark {
		t.Error("expected Dracula to be a dark theme")
	}
}

func TestLookupThemeCaseInsensitive(t *testing.T) {
	th := LookupTheme("dracula")
	if th == nil {
		t.Fatal("expected case-insensitive lookup to find Dracula")
	}
}

func TestLookupThemeNotFound(t *testing.T) {
	th := LookupTheme("nonexistent-theme-xyz")
	if th != nil {
		t.Error("expected nil for unknown theme")
	}
}

// --- Palette derivation (P0) ---

func TestDefaultPaletteNonNil(t *testing.T) {
	p := defaultPalette()
	if p.BG == nil {
		t.Error("default palette BG should not be nil")
	}
	if p.FG == nil {
		t.Error("default palette FG should not be nil")
	}
	if p.Accent == nil {
		t.Error("default palette Accent should not be nil")
	}
}

func TestDerivePaletteFromTheme(t *testing.T) {
	th := LookupTheme("Dracula")
	if th == nil {
		t.Fatal("Dracula theme not found")
	}
	p := th.Palette()
	if p.BG == nil {
		t.Error("Dracula palette BG should not be nil")
	}
	if p.Success == nil {
		t.Error("Dracula palette Success should not be nil")
	}
	// All 16 ANSI slots should be populated.
	for i, c := range p.ANSI {
		if c == nil {
			t.Errorf("Dracula palette ANSI[%d] should not be nil", i)
		}
	}
}

// --- Apply / Current (P0) ---

func TestApplyDefault(t *testing.T) {
	if err := Apply(""); err != nil {
		t.Fatalf("Apply empty should succeed: %v", err)
	}
	p := Current()
	if p.BG == nil {
		t.Error("current palette BG should not be nil after applying default")
	}
}

func TestApplyNamedTheme(t *testing.T) {
	if err := Apply("Dracula"); err != nil {
		t.Fatalf("Apply Dracula should succeed: %v", err)
	}
	p := Current()
	if p.BG == nil {
		t.Error("current palette BG should not be nil after applying Dracula")
	}
	// Restore default for other tests.
	_ = Apply("")
}

func TestApplyUnknownTheme(t *testing.T) {
	err := Apply("nonexistent-theme-xyz")
	if err == nil {
		t.Error("expected error for unknown theme")
	}
}

// --- Contrast enforcement (P0) ---

func TestContrastRatioKnownValues(t *testing.T) {
	// Black on white should be 21:1.
	black := color.NRGBA{R: 0, G: 0, B: 0, A: 255}
	white := color.NRGBA{R: 255, G: 255, B: 255, A: 255}
	got := contrastRatio(black, white)
	if got < 20.9 || got > 21.1 {
		t.Errorf("black-on-white contrast = %.2f, want ~21.0", got)
	}
	// Same color should be 1:1.
	got = contrastRatio(black, black)
	if got < 0.99 || got > 1.01 {
		t.Errorf("same-color contrast = %.2f, want 1.0", got)
	}
}

func TestDefaultPaletteContrast(t *testing.T) {
	p := defaultPalette()
	assertPaletteContrast(t, "default", p)
}

func TestAllThemePaletteContrast(t *testing.T) {
	for _, th := range AllThemes() {
		p := th.Palette()
		t.Run(th.Name, func(t *testing.T) {
			assertPaletteContrast(t, th.Name, p)
		})
	}
}

// assertPaletteContrast checks that every critical FG/BG pairing in a palette
// meets WCAG AA minimums. Bold/large text uses 3:1; normal text uses 4.5:1.
func assertPaletteContrast(t *testing.T, name string, p Palette) {
	t.Helper()

	type pair struct {
		label    string
		fg, bg   color.Color
		minRatio float64
	}

	pairs := []pair{
		// FG tokens on main BG (normal text → 4.5:1)
		{"FG on BG", p.FG, p.BG, minContrastNormal},
		{"TextDim on BG", p.TextDim, p.BG, minContrastNormal},
		{"Accent on BG", p.Accent, p.BG, minContrastNormal},
		{"AccentSoft on BG", p.AccentSoft, p.BG, minContrastNormal},
		{"Info on BG", p.Info, p.BG, minContrastNormal},
		{"Error on BG", p.Error, p.BG, minContrastNormal},
		{"Success on BG", p.Success, p.BG, minContrastNormal},
		{"Given on BG", p.Given, p.BG, minContrastNormal},
		{"Linked on BG", p.Linked, p.BG, minContrastNormal},
		{"Secondary on BG", p.Secondary, p.BG, minContrastNormal},
		{"Tertiary on BG", p.Tertiary, p.BG, minContrastNormal},
		{"Highlight on BG", p.Highlight, p.BG, minContrastNormal},
		// FG tokens on derived BGs (bold text → 3:1)
		{"AccentText on AccentBG (cursor)", p.AccentText, p.AccentBG, minContrastLarge},
		{"SolvedFG on SuccessBG (solved text)", p.SolvedFG, p.SuccessBG, minContrastNormal},
		{"SuccessBorder on SuccessBG (solved)", p.SuccessBorder, p.SuccessBG, minContrastLarge},
	}

	for _, pp := range pairs {
		cr := contrastRatio(pp.fg, pp.bg)
		if cr < pp.minRatio {
			t.Errorf("[%s] %s: contrast %.2f:1 < %.1f:1 minimum",
				name, pp.label, cr, pp.minRatio)
		}
	}
}

func TestMidTone(t *testing.T) {
	black := color.NRGBA{R: 0, G: 0, B: 0, A: 255}
	white := color.NRGBA{R: 255, G: 255, B: 255, A: 255}
	mid := MidTone(black, white)
	r, g, b, _ := mid.RGBA()
	// Should be roughly 127/128 for each channel.
	got := fmt.Sprintf("%d,%d,%d", r>>8, g>>8, b>>8)
	if got != "127,127,127" {
		t.Errorf("midTone(black, white) = %s, want 127,127,127", got)
	}
}
