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

	// LogoStyle renders the ASCII art brand logo.
	LogoStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(MenuText).
			Background(MenuAccent).
			Padding(0, 1).
			Margin(1, 0)

	// MainMenuFrame wraps the main menu in a heavy double border.
	MainMenuFrame = lipgloss.NewStyle().
			Border(lipgloss.DoubleBorder()).
			BorderForeground(MenuAccent).
			Padding(2, 4)

	// PanelFrame wraps sub-menus in a lighter rounded border.
	PanelFrame = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(MenuDim).
			Padding(1, 2)

	// PanelTitle styles the title line inside a panel.
	PanelTitle = lipgloss.NewStyle().
			Bold(true).
			Foreground(MenuAccent)

	// FooterHint styles the navigation hint line at the bottom of panels.
	FooterHint = lipgloss.NewStyle().
			Foreground(MenuTextDim)

	// CursorStyle styles the active cursor indicator.
	CursorStyle = lipgloss.NewStyle().
			Foreground(MenuAccent).
			Bold(true)

	// SelectedItemStyle styles the currently highlighted menu item title.
	SelectedItemStyle = lipgloss.NewStyle().
				Foreground(MenuAccent).
				Bold(true)

	// NormalItemStyle styles non-selected menu item titles.
	NormalItemStyle = lipgloss.NewStyle().
			Foreground(MenuText)

	// DimItemStyle styles descriptions and secondary text.
	DimItemStyle = lipgloss.NewStyle().
			Foreground(MenuTextDim)

	// GeneratingFrame wraps the generating spinner in a small box.
	GeneratingFrame = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(MenuAccent).
			Padding(1, 3)
)

// CenterView wraps content in centered placement within the available area.
func CenterView(width, height int, content string) string {
	return lipgloss.Place(width, height, lipgloss.Center, lipgloss.Center, content)
}
