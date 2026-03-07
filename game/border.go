package game

import (
	"image/color"

	"charm.land/lipgloss/v2"
	"github.com/FelineStateMachine/puzzletea/theme"
)

// GridBorderColors holds the color configuration for a grid border with
// crosshair and solved-state highlighting.
type GridBorderColors struct {
	BorderFG       color.Color
	BackgroundBG   color.Color
	CrosshairBG    color.Color
	SolvedBorderFG color.Color
	SolvedBG       color.Color
}

func borderColors(colors GridBorderColors, solved bool, bg color.Color) (color.Color, color.Color) {
	fg := colors.BorderFG
	background := bg

	if solved {
		fg = colors.SolvedBorderFG
	}

	if background == nil {
		background = colors.BackgroundBG
		if solved {
			background = colors.SolvedBG
		}
		return fg, background
	}

	// Tinted bridge/crosshair backgrounds need a contrast-aware foreground.
	return theme.TextOnBG(background), background
}

// DefaultBorderColors returns border colors from the active theme.
func DefaultBorderColors() GridBorderColors {
	p := theme.Current()
	return GridBorderColors{
		BorderFG:       p.Border,
		BackgroundBG:   p.BG,
		CrosshairBG:    p.Surface,
		SolvedBorderFG: p.SuccessBorder,
		SolvedBG:       p.SuccessBG,
	}
}

// BorderChar renders a single border character with optional crosshair highlighting.
func BorderChar(ch string, colors GridBorderColors, solved, highlight bool) string {
	bg := color.Color(nil)
	if highlight && !solved {
		bg = colors.CrosshairBG
	}
	fg, bg := borderColors(colors, solved, bg)
	return lipgloss.NewStyle().Foreground(fg).Background(bg).Render(ch)
}

// HBorderRow builds a horizontal border row (top or bottom) with per-column
// crosshair highlighting. cellWidth is the visual width of each cell.
func HBorderRow(w, cursorX, cellWidth int, left, right string, colors GridBorderColors, solved bool) string {
	var parts []string
	parts = append(parts, BorderChar(left, colors, solved, false))
	segment := ""
	for range cellWidth {
		segment += "─"
	}
	for x := range w {
		highlight := !solved && x == cursorX
		bg := color.Color(nil)
		if highlight {
			bg = colors.CrosshairBG
		}
		fg, bg := borderColors(colors, solved, bg)
		s := lipgloss.NewStyle().Foreground(fg).Background(bg)
		parts = append(parts, s.Render(segment))
	}
	parts = append(parts, BorderChar(right, colors, solved, false))
	return lipgloss.JoinHorizontal(lipgloss.Top, parts...)
}
