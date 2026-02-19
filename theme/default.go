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
	p := Palette{
		// Surfaces
		BG:            lipgloss.Color("#262626"), // 235 — main background
		FG:            lipgloss.Color("#d0d0d0"), // 252 — primary text (9.8:1 on BG)
		Surface:       lipgloss.Color("#3a3a3a"), // 237 — subtle elevation
		SurfaceBright: lipgloss.Color("#e4e4e4"), // 254 — bright surface / cursor text

		// Accent — earth-tone orange
		Accent:     lipgloss.Color("#cc8844"), // ~180 — primary accent (5.2:1 on BG)
		AccentSoft: lipgloss.Color("#cc9966"), // 180  — secondary highlight
		AccentText: lipgloss.Color("#e4e4e4"), // 254  — text on accent background

		// Text
		TextDim: lipgloss.Color("#949494"), // ~246 — muted text (5.0:1 on BG)

		// Chrome
		Border: lipgloss.Color("#585858"), // ~240 — borders (2.1:1 on BG, 1.6:1 on Surface)

		// Game semantics
		Success:       lipgloss.Color("#145530"), // deeper green
		SuccessBorder: lipgloss.Color("#99cc33"), // 149 — solved border text
		Error:         lipgloss.Color("#cc3333"), // 167 — conflict foreground
		ErrorBG:       lipgloss.Color("#330000"), // 52  — conflict background
		Info:          lipgloss.Color("#bb8844"), // ~180 — hint/info text (4.8:1 on BG)
		Highlight:     lipgloss.Color("#cc9933"), // 179 — selection highlight
		Warm:          lipgloss.Color("#b86b00"), // ~166 — warm cursor BG
		WarmText:      lipgloss.Color("#e4e4e4"), // 254  — text on warm background

		ANSI: [16]color.Color{
			lipgloss.Color("#262626"), // 0  black
			lipgloss.Color("#cc3333"), // 1  red
			lipgloss.Color("#145530"), // 2  green  — success
			lipgloss.Color("#cc8844"), // 3  yellow — primary accent
			lipgloss.Color("#bb8844"), // 4  blue   — info/hint
			lipgloss.Color("#993399"), // 5  purple
			lipgloss.Color("#cc9933"), // 6  cyan   — highlight
			lipgloss.Color("#d0d0d0"), // 7  white
			lipgloss.Color("#949494"), // 8  brightBlack — textDim
			lipgloss.Color("#330000"), // 9  brightRed   — error bg
			lipgloss.Color("#99cc33"), // 10 brightGreen — success border
			lipgloss.Color("#cc9966"), // 11 brightYellow — accent soft
			lipgloss.Color("#cccc99"), // 12 brightBlue
			lipgloss.Color("#99cc99"), // 13 brightPurple
			lipgloss.Color("#ffcc99"), // 14 brightCyan
			lipgloss.Color("#eeeeee"), // 15 brightWhite
		},
	}
	return enforcePaletteContrast(p)
}
