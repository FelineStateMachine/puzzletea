package theme

import (
	"image/color"

	"charm.land/lipgloss/v2"
)

// Palette holds the semantic color tokens derived from a terminal color scheme.
// Every style in the application reads from the active palette via [Current].
//
// All ANSI-sourced tokens are foreground/text-role colors.  Background-role
// colors (AccentBG, SuccessBG, ErrorBG, SelectionBG, HighlightBG) are derived
// by blending a FG-role color toward BG or Surface, so they always match the
// theme's luminance range.
type Palette struct {
	// Surfaces
	BG      color.Color // main background (theme Background)
	FG      color.Color // primary text (theme Foreground)
	Surface color.Color // subtle elevation (crosshair BG tint)

	// Brand / accent  (FG-role from ANSI)
	Accent     color.Color // primary accent (ANSI 3 — yellow)
	AccentSoft color.Color // muted accent (ANSI 11 — brightYellow)
	AccentText color.Color // text on AccentBG (ANSI 7 — white)

	// Text
	TextDim color.Color // secondary / muted text (derived)

	// Chrome
	Border color.Color // borders, separators (ANSI 0 — black)

	// Game FG-role tokens  (all from ANSI slots, enforced against BG)
	Success       color.Color // success/solved FG text (ANSI 2 — green)
	SuccessBorder color.Color // solved border FG (ANSI 10 — brightGreen)
	SolvedFG      color.Color // high-contrast text on state BGs (ANSI 15 — brightWhite)
	Error         color.Color // conflict/error FG text (ANSI 1 — red)
	Info          color.Color // hints, informational text (ANSI 4 — blue)
	Given         color.Color // immutable/provided cell FG (ANSI 5 — purple)
	Linked        color.Color // connected/related element FG (ANSI 6 — cyan)
	Highlight     color.Color // selection/highlight FG (ANSI 14 — brightCyan)
	Secondary     color.Color // second value type (ANSI 12 — brightBlue)
	Tertiary      color.Color // third differentiator (ANSI 13 — brightPurple)

	// Derived background-role tokens  (blended, not from ANSI slots)
	AccentBG    color.Color // cursor / active element BG
	SuccessBG   color.Color // solved cell BG
	ErrorBG     color.Color // conflict cell BG
	SelectionBG color.Color // drag / selection BG
	HighlightBG color.Color // adjacent / neighbor highlight BG

	// Full 16-color ANSI palette for puzzle-specific use (e.g. shikaku rects).
	// Order: black, red, green, yellow, blue, purple, cyan, white,
	//        brightBlack..brightWhite.
	ANSI [16]color.Color
}

// ThemeColors returns a curated slice of distinct, colorful palette colors
// suitable for differentiating themed UI elements (cards, hover bars, puzzle
// regions, etc). The colors are drawn from chromatic ANSI slots (skipping
// black/white/gray), giving up to 10 visually distinct options per theme.
func (p Palette) ThemeColors() []color.Color {
	return []color.Color{
		p.ANSI[3],  // yellow
		p.ANSI[4],  // blue
		p.ANSI[2],  // green
		p.ANSI[5],  // purple
		p.ANSI[6],  // cyan
		p.ANSI[1],  // red
		p.ANSI[11], // brightYellow
		p.ANSI[12], // brightBlue
		p.ANSI[14], // brightCyan
		p.ANSI[13], // brightPurple
	}
}

// Blend ratios for derived background colors. Each ratio controls how much
// of the source FG color is mixed into the base (BG or Surface) to produce
// the derived BG. Higher values produce more vivid tinted backgrounds.
const (
	blendAccentBG    = 0.45 // cursor / active element
	blendSuccessBG   = 0.30 // solved cells
	blendErrorBG     = 0.25 // conflict cells
	blendSelectionBG = 0.35 // drag selection
	blendHighlightBG = 0.30 // adjacent / neighbor highlight
)

// derivePalette maps a Theme's 16 terminal colors onto semantic tokens and
// enforces minimum WCAG contrast ratios for every critical FG/BG pairing.
//
// All ANSI slots are mapped to FG-role tokens. Background-role tokens are
// derived by blending a FG color toward BG, respecting the theme's luminance.
func derivePalette(t Theme) Palette {
	bg := lipgloss.Color(t.Background)
	fg := lipgloss.Color(t.Foreground)

	ansi := [16]color.Color{
		lipgloss.Color(t.Black),
		lipgloss.Color(t.Red),
		lipgloss.Color(t.Green),
		lipgloss.Color(t.Yellow),
		lipgloss.Color(t.Blue),
		lipgloss.Color(t.Purple),
		lipgloss.Color(t.Cyan),
		lipgloss.Color(t.White),
		lipgloss.Color(t.BrightBlack),
		lipgloss.Color(t.BrightRed),
		lipgloss.Color(t.BrightGreen),
		lipgloss.Color(t.BrightYellow),
		lipgloss.Color(t.BrightBlue),
		lipgloss.Color(t.BrightPurple),
		lipgloss.Color(t.BrightCyan),
		lipgloss.Color(t.BrightWhite),
	}

	// --- Non-chromatic luminance ramp ---
	surface := ansi[8]         // brightBlack — subtle elevation
	textDim := MidTone(bg, fg) // halfway between BG and FG
	border := ansi[0]          // black — grid chrome

	// If border is too close to BG, swap to brightBlack.
	if contrastRatio(border, bg) < 1.5 {
		border = surface
	}

	// --- FG-role chromatic tokens (all from ANSI slots) ---
	accent := ansi[3]      // yellow  — primary accent
	success := ansi[2]     // green   — success/solved FG
	errorFG := ansi[1]     // red     — conflict/error FG
	linked := ansi[6]      // cyan    — connected elements
	highlight := ansi[14]  // bCyan   — selection/highlight FG
	secondary := ansi[12]  // bBlue   — second value type
	highlightSrc := linked // source color for HighlightBG blend
	_ = highlightSrc

	// --- Derived BG-role tokens (blended toward BG) ---
	accentBG := Blend(bg, accent, blendAccentBG)
	successBG := Blend(bg, success, blendSuccessBG)
	errorBG := Blend(bg, errorFG, blendErrorBG)
	selectionBG := Blend(surface, highlight, blendSelectionBG)
	highlightBG := Blend(surface, linked, blendHighlightBG)

	p := Palette{
		BG:      bg,
		FG:      fg,
		Surface: surface,

		Accent:     accent,
		AccentSoft: ansi[11], // brightYellow
		AccentText: ansi[7],  // white — text on AccentBG

		TextDim: textDim,
		Border:  border,

		Success:       success,
		SuccessBorder: ansi[10], // brightGreen
		SolvedFG:      ansi[15], // brightWhite — text on state BGs
		Error:         errorFG,
		Info:          ansi[12], // brightBlue
		Given:         ansi[5],  // purple
		Linked:        linked,
		Highlight:     highlight,
		Secondary:     secondary,
		Tertiary:      ansi[13], // brightPurple

		AccentBG:    accentBG,
		SuccessBG:   successBG,
		ErrorBG:     errorBG,
		SelectionBG: selectionBG,
		HighlightBG: highlightBG,

		ANSI: ansi,
	}

	return enforcePaletteContrast(p)
}

// enforcePaletteContrast ensures every critical FG-on-BG pairing in p meets
// WCAG AA minimums. It nudges foreground colors toward the palette's brightest
// or darkest ANSI slot when the original contrast is insufficient.
//
// FG-role tokens are enforced against BG (the main canvas). Tokens used on
// derived BGs (AccentBG, SuccessBG, ErrorBG) are enforced against those
// derived backgrounds after the FG tokens are finalized.
func enforcePaletteContrast(p Palette) Palette {
	// --- Phase 1: FG-role tokens against the main BG (normal text → 4.5:1) ---

	p.Accent = ensurePairContrast(p.Accent, p.BG, p, minContrastNormal)
	p.AccentSoft = ensurePairContrast(p.AccentSoft, p.BG, p, minContrastNormal)
	p.FG = ensurePairContrast(p.FG, p.BG, p, minContrastNormal)
	p.TextDim = ensurePairContrast(p.TextDim, p.BG, p, minContrastNormal)
	p.Info = ensurePairContrast(p.Info, p.BG, p, minContrastNormal)
	p.Error = ensurePairContrast(p.Error, p.BG, p, minContrastNormal)
	p.Success = ensurePairContrast(p.Success, p.BG, p, minContrastNormal)
	p.Given = ensurePairContrast(p.Given, p.BG, p, minContrastNormal)
	p.Linked = ensurePairContrast(p.Linked, p.BG, p, minContrastNormal)
	p.Secondary = ensurePairContrast(p.Secondary, p.BG, p, minContrastNormal)
	p.Tertiary = ensurePairContrast(p.Tertiary, p.BG, p, minContrastNormal)
	p.Highlight = ensurePairContrast(p.Highlight, p.BG, p, minContrastNormal)

	// --- Phase 2: re-derive BGs after FG colors are finalized ---
	// (FG nudging may have shifted the source colors used for blending.)

	p.AccentBG = Blend(p.BG, p.Accent, blendAccentBG)
	p.SuccessBG = Blend(p.BG, p.Success, blendSuccessBG)
	p.ErrorBG = Blend(p.BG, p.Error, blendErrorBG)
	p.SelectionBG = Blend(p.Surface, p.Highlight, blendSelectionBG)
	p.HighlightBG = Blend(p.Surface, p.Linked, blendHighlightBG)

	// --- Phase 3: FG tokens on derived BGs (bold text → 3:1) ---

	p.AccentText = ensurePairContrast(p.AccentText, p.AccentBG, p, minContrastLarge)
	p.SolvedFG = ensurePairContrast(p.SolvedFG, p.SuccessBG, p, minContrastNormal)
	p.SuccessBorder = ensurePairContrast(p.SuccessBorder, p.SuccessBG, p, minContrastLarge)

	return p
}
