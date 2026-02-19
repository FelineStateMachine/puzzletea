package theme

import (
	"image/color"

	"charm.land/lipgloss/v2"
)

// DefaultThemeName is the name of the built-in earth-tone palette that ships
// with PuzzleTea. When no theme is configured, this palette is active.
const DefaultThemeName = "PuzzleTea Default"

// defaultPalette returns the hand-tuned palette that ships with PuzzleTea.
// Base colors are chosen so that every critical FG/BG pairing meets WCAG AA
// after [enforcePaletteContrast] runs. The contrast enforcement step
// auto-selects cursor text color (light or dark) based on accent luminance.
func defaultPalette() Palette {
	bg := lipgloss.Color("#262626")      // 235 — main background
	fg := lipgloss.Color("#d0d0d0")      // 252 — primary text
	surface := lipgloss.Color("#3a3a3a") // 237 — subtle elevation

	// FG-role tokens
	accent := lipgloss.Color("#cc8844")        // 3  yellow — primary accent
	accentSoft := lipgloss.Color("#cc9966")    // 11 brightYellow
	accentText := lipgloss.Color("#e4e4e4")    // 7  white — text on AccentBG
	success := lipgloss.Color("#55aa55")       // 2  green — success FG
	successBorder := lipgloss.Color("#99cc33") // 10 brightGreen
	solvedFG := lipgloss.Color("#eeeeee")      // 15 brightWhite
	errorFG := lipgloss.Color("#cc3333")       // 1  red — error FG
	info := lipgloss.Color("#bb8844")          // 4  blue — info
	given := lipgloss.Color("#993399")         // 5  purple — given
	linked := lipgloss.Color("#cc9933")        // 6  cyan — linked
	highlight := lipgloss.Color("#99cccc")     // 14 brightCyan — highlight FG
	secondary := lipgloss.Color("#cccc99")     // 12 brightBlue — secondary
	tertiary := lipgloss.Color("#99cc99")      // 13 brightPurple — tertiary

	p := Palette{
		BG:      bg,
		FG:      fg,
		Surface: surface,

		Accent:     accent,
		AccentSoft: accentSoft,
		AccentText: accentText,
		TextDim:    lipgloss.Color("#949494"), // ~246 — muted text
		Border:     lipgloss.Color("#585858"), // ~240 — borders

		Success:       success,
		SuccessBorder: successBorder,
		SolvedFG:      solvedFG,
		Error:         errorFG,
		Info:          info,
		Given:         given,
		Linked:        linked,
		Highlight:     highlight,
		Secondary:     secondary,
		Tertiary:      tertiary,

		// Derived BGs (blended toward BG/Surface)
		AccentBG:    Blend(bg, accent, blendAccentBG),
		SuccessBG:   Blend(bg, success, blendSuccessBG),
		ErrorBG:     Blend(bg, errorFG, blendErrorBG),
		SelectionBG: Blend(surface, highlight, blendSelectionBG),
		HighlightBG: Blend(surface, linked, blendHighlightBG),

		ANSI: [16]color.Color{
			lipgloss.Color("#262626"), // 0  black — border
			lipgloss.Color("#cc3333"), // 1  red — error
			lipgloss.Color("#55aa55"), // 2  green — success
			lipgloss.Color("#cc8844"), // 3  yellow — accent
			lipgloss.Color("#bb8844"), // 4  blue — info
			lipgloss.Color("#993399"), // 5  purple — given
			lipgloss.Color("#cc9933"), // 6  cyan — linked
			lipgloss.Color("#d0d0d0"), // 7  white — accentText
			lipgloss.Color("#949494"), // 8  brightBlack — surface
			lipgloss.Color("#cc6666"), // 9  brightRed — (unassigned)
			lipgloss.Color("#99cc33"), // 10 brightGreen — successBorder
			lipgloss.Color("#cc9966"), // 11 brightYellow — accentSoft
			lipgloss.Color("#cccc99"), // 12 brightBlue — secondary
			lipgloss.Color("#99cc99"), // 13 brightPurple — tertiary
			lipgloss.Color("#99cccc"), // 14 brightCyan — highlight
			lipgloss.Color("#eeeeee"), // 15 brightWhite — solvedFG
		},
	}
	return enforcePaletteContrast(p)
}
