package ui

import "github.com/charmbracelet/lipgloss"

// Earth-tone palette for menus — ANSI 256 colors with light/dark adaptivity.
var (
	MenuAccent      = lipgloss.AdaptiveColor{Light: "130", Dark: "173"}
	MenuAccentLight = lipgloss.AdaptiveColor{Light: "137", Dark: "180"}
	MenuText        = lipgloss.AdaptiveColor{Light: "235", Dark: "252"}
	MenuTextDim     = lipgloss.AdaptiveColor{Light: "243", Dark: "243"}
	MenuDim         = lipgloss.AdaptiveColor{Light: "250", Dark: "238"}
	MenuTableHeader = lipgloss.AdaptiveColor{Light: "130", Dark: "180"}

	RootStyle  = lipgloss.NewStyle().Margin(1, 2)
	DebugStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder(), true).
			BorderForeground(lipgloss.AdaptiveColor{Light: "124", Dark: "124"})
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
