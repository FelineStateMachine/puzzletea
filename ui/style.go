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

	RootStyle  = lipgloss.NewStyle().Margin(1, 2)
	DebugStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder(), true).
			BorderForeground(compat.AdaptiveColor{Light: lipgloss.Color("124"), Dark: lipgloss.Color("124")})
)

// RootFrameSize returns the horizontal and vertical frame size of RootStyle.
// Isolates the lipgloss v1 GetFrameSize() call for easier v2 migration.
func RootFrameSize() (int, int) {
	return RootStyle.GetFrameSize()
}

// CenterView wraps content in centered placement within the root frame.
// Isolates the lipgloss.Place() call for easier v2 migration.
func CenterView(width, height int, content string) string {
	return RootStyle.Render(lipgloss.Place(width, height, lipgloss.Center, lipgloss.Center, content))
}
