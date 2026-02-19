package theme

import (
	"image/color"
	"strings"

	"charm.land/lipgloss/v2"
)

const (
	swatchW  = 6
	previewW = 28
)

// PreviewPanel renders a visual color preview for the named theme.
// It shows labeled swatches for the 16 ANSI colors, background/foreground,
// and the semantic roles.
func PreviewPanel(themeName string, height int) string {
	var p Palette
	if themeName == "" || strings.EqualFold(themeName, DefaultThemeName) {
		p = defaultPalette()
	} else if t := LookupTheme(themeName); t != nil {
		p = t.Palette()
	} else {
		p = defaultPalette()
	}

	title := lipgloss.NewStyle().Bold(true).Foreground(p.Accent)
	dim := lipgloss.NewStyle().Foreground(p.TextDim)
	label := lipgloss.NewStyle().Foreground(p.FG)

	var lines []string

	lines = append(lines, title.Render(themeName))
	lines = append(lines, "")

	// Background + foreground sample text.
	sample := lipgloss.NewStyle().
		Background(p.BG).Foreground(p.FG).
		Width(previewW).Padding(0, 1).
		Render("The quick brown fox")
	lines = append(lines, sample)
	lines = append(lines, "")

	// Semantic color swatches: two per row to keep labels readable.
	lines = append(lines, dim.Render("Semantic"))
	lines = append(lines, swatchPair(p, "Accent", p.Accent, "Soft", p.AccentSoft))
	lines = append(lines, swatchPair(p, "Info", p.Info, "Warm", p.Warm))
	lines = append(lines, swatchPair(p, "OK", p.Success, "Error", p.Error))
	lines = append(lines, swatchPair(p, "FG", p.FG, "Dim", p.TextDim))
	lines = append(lines, "")

	// ANSI 16-color grid: 8 per row, compact 3-char swatches.
	lines = append(lines, dim.Render("ANSI Palette"))
	lines = append(lines, ansiRow(p.ANSI[0:8]))
	lines = append(lines, ansiRow(p.ANSI[8:16]))
	lines = append(lines, "")

	// Game state previews.
	lines = append(lines, dim.Render("Game States"))
	lines = append(lines, lipgloss.NewStyle().
		Background(p.Accent).Foreground(p.AccentText).Bold(true).
		Width(previewW).Padding(0, 1).
		Render("Cursor"))
	lines = append(lines, lipgloss.NewStyle().
		Background(p.Success).Foreground(p.SuccessBorder).Bold(true).
		Width(previewW).Padding(0, 1).
		Render("Solved"))
	lines = append(lines, lipgloss.NewStyle().
		Background(p.ErrorBG).Foreground(p.Error).Bold(true).
		Width(previewW).Padding(0, 1).
		Render("Conflict"))

	_ = label // reserved for future use

	// Truncate to available height.
	if height > 0 && len(lines) > height {
		lines = lines[:height]
	}

	return lipgloss.JoinVertical(lipgloss.Left, lines...)
}

// swatchPair renders two labeled color swatches side by side on two lines.
func swatchPair(p Palette, lbl1 string, c1 color.Color, lbl2 string, c2 color.Color) string {
	sw := func(c color.Color) string {
		return lipgloss.NewStyle().Background(c).Width(swatchW).Render(strings.Repeat(" ", swatchW))
	}
	lb := func(s string) string {
		return lipgloss.NewStyle().Foreground(p.TextDim).Width(swatchW).Render(s)
	}
	gap := "  "
	return sw(c1) + gap + sw(c2) + "\n" + lb(lbl1) + gap + lb(lbl2)
}

// ansiRow renders 8 color swatches in a single compact row.
func ansiRow(colors []color.Color) string {
	var cells []string
	for _, c := range colors {
		cells = append(cells, lipgloss.NewStyle().Background(c).Render("   "))
	}
	return strings.Join(cells, "")
}
