package theme

import (
	"image/color"

	"charm.land/lipgloss/v2"
)

// Palette holds the semantic color tokens derived from a terminal color scheme.
// Every style in the application reads from the active palette via [Current].
type Palette struct {
	// Surfaces
	BG            color.Color // main background
	FG            color.Color // primary text / foreground
	Surface       color.Color // slightly elevated surface (crosshair, panels)
	SurfaceBright color.Color // bright surface (light-mode backgrounds)

	// Brand / accent
	Accent     color.Color // primary accent (menu highlights, cursor BG)
	AccentSoft color.Color // muted accent (secondary highlights)
	AccentText color.Color // text on accent background (cursor text, title bars)

	// Text
	TextDim color.Color // secondary / muted text

	// Chrome
	Border color.Color // borders, separators, dividers

	// Game semantics
	Success       color.Color // solved state
	SuccessBorder color.Color // solved border
	Error         color.Color // conflict / error foreground
	ErrorBG       color.Color // conflict / error background
	Info          color.Color // hints, informational text
	Highlight     color.Color // selection, active highlight
	Warm          color.Color // warm accent (cursor variant)
	WarmText      color.Color // text on warm background

	// Full 16-color ANSI palette for puzzle-specific use (e.g. shikaku rects).
	// Order: black, red, green, yellow, blue, purple, cyan, white,
	//        brightBlack..brightWhite.
	ANSI [16]color.Color
}

// CardColors returns a curated slice of distinct, colorful palette colors
// suitable for differentiating UI cards. The colors are drawn from the
// chromatic ANSI slots (skipping black/white/gray) and the semantic accent
// colors, giving up to 10 visually distinct options per theme.
func (p Palette) CardColors() []color.Color {
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

// derivePalette maps a Theme's 16 terminal colors onto semantic tokens and
// enforces minimum WCAG contrast ratios for every critical FG/BG pairing.
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

	// Pick distinct slots for Surface / TextDim / Border so they are not
	// all identical.  Surface stays at brightBlack (8); TextDim uses white
	// (7) or brightBlack (8) depending on which has better contrast against
	// BG; Border uses black (0).
	surface := ansi[8]         // brightBlack — subtle elevation
	textDim := midTone(bg, fg) // halfway between BG and FG
	border := ansi[0]          // black — distinct from surface

	// If border is too close to BG, swap to brightBlack.
	if contrastRatio(border, bg) < 1.5 {
		border = surface
	}

	p := Palette{
		BG:            bg,
		FG:            fg,
		Surface:       surface,
		SurfaceBright: ansi[7],  // white
		Accent:        ansi[3],  // yellow
		AccentSoft:    ansi[11], // brightYellow
		AccentText:    ansi[7],  // white — text on accent background
		TextDim:       textDim,
		Border:        border,
		Success:       ansi[2],  // green
		SuccessBorder: ansi[10], // brightGreen
		Error:         ansi[1],  // red
		ErrorBG:       ansi[9],  // brightRed
		Info:          ansi[4],  // blue
		Highlight:     ansi[6],  // cyan
		Warm:          ansi[11], // brightYellow
		WarmText:      ansi[7],  // white — text on warm background
		ANSI:          ansi,
	}

	return enforcePaletteContrast(p)
}

// enforcePaletteContrast ensures every critical FG-on-BG pairing in p meets
// WCAG AA minimums. It nudges foreground colors toward the palette's brightest
// or darkest ANSI slot when the original contrast is insufficient.
//
// Ordering matters: background-role colors (Accent, Success, ErrorBG) are
// finalized first, then foreground-role colors that sit on top of them.
func enforcePaletteContrast(p Palette) Palette {
	// --- Phase 1: finalize colors used as backgrounds ---

	// Accent on BG (normal text for menu items → 4.5:1).
	p.Accent = ensurePairContrast(p.Accent, p.BG, p, minContrastNormal)

	// --- Phase 2: finalize foreground colors against their backgrounds ---

	// AccentText on Accent (cursor text, title bars — large/bold text → 3:1).
	p.AccentText = ensurePairContrast(p.AccentText, p.Accent, p, minContrastLarge)

	// WarmText on Warm (warm cursor text — large/bold text → 3:1).
	p.WarmText = ensurePairContrast(p.WarmText, p.Warm, p, minContrastLarge)

	// Solved: SuccessBorder on Success (large/bold text → 3:1).
	p.SuccessBorder = ensurePairContrast(p.SuccessBorder, p.Success, p, minContrastLarge)

	// Conflict: Error on ErrorBG (large/bold text → 3:1).
	p.Error = ensurePairContrast(p.Error, p.ErrorBG, p, minContrastLarge)

	// TextDim on BG (normal text → 4.5:1).
	p.TextDim = ensurePairContrast(p.TextDim, p.BG, p, minContrastNormal)

	// FG on BG (normal text → 4.5:1).
	p.FG = ensurePairContrast(p.FG, p.BG, p, minContrastNormal)

	// Info on BG (normal text → 4.5:1).
	p.Info = ensurePairContrast(p.Info, p.BG, p, minContrastNormal)

	return p
}
