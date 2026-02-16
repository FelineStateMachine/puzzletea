package ui

import (
	"charm.land/lipgloss/v2"
	"charm.land/lipgloss/v2/compat"
)

// Earth-tone palette for menus — ANSI 256 colors with light/dark adaptivity.
var (
	MenuAccent      = compat.AdaptiveColor{Light: lipgloss.Color("130"), Dark: lipgloss.Color("173")}
	MenuAccentLight = compat.AdaptiveColor{Light: lipgloss.Color("137"), Dark: lipgloss.Color("180")}
	MenuText        = compat.AdaptiveColor{Light: lipgloss.Color("235"), Dark: lipgloss.Color("252")}
	MenuTextDim     = compat.AdaptiveColor{Light: lipgloss.Color("243"), Dark: lipgloss.Color("243")}
	MenuDim         = compat.AdaptiveColor{Light: lipgloss.Color("250"), Dark: lipgloss.Color("238")}
	MenuTableHeader = compat.AdaptiveColor{Light: lipgloss.Color("130"), Dark: lipgloss.Color("180")}

	DebugStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder(), true).
			BorderForeground(compat.AdaptiveColor{Light: lipgloss.Color("124"), Dark: lipgloss.Color("124")})
)

// CenterView wraps content in centered placement within the available area.
func CenterView(width, height int, content string) string {
	return lipgloss.Place(width, height, lipgloss.Center, lipgloss.Center, content)
}
