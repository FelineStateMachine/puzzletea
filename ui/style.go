package ui

import (
	"charm.land/lipgloss/v2"
	"github.com/FelineStateMachine/puzzletea/theme"
)

// MenuAccent returns the primary accent color from the active theme.
func MenuAccent() lipgloss.Style { return lipgloss.NewStyle().Foreground(theme.Current().Accent) }

// DebugStyle returns the style for the debug overlay border.
func DebugStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder(), true).
		BorderForeground(theme.Current().Error)
}

// LogoStyle returns the style for the ASCII art brand logo.
func LogoStyle() lipgloss.Style {
	p := theme.Current()
	return lipgloss.NewStyle().
		Bold(true).
		Foreground(theme.TextOnBG(p.Accent)).
		Background(p.Accent).
		Padding(0, 1).
		Margin(1, 0)
}

// MainMenuFrame returns the double-border frame for the main menu.
func MainMenuFrame() lipgloss.Style {
	return lipgloss.NewStyle().
		Border(lipgloss.DoubleBorder()).
		BorderForeground(theme.Current().Accent).
		Padding(2, 4)
}

// PanelFrame returns the rounded-border frame for sub-menus.
func PanelFrame() lipgloss.Style {
	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(theme.Current().Border).
		Padding(1, 2)
}

// PanelTitle returns the style for panel title text.
func PanelTitle() lipgloss.Style {
	return lipgloss.NewStyle().
		Bold(true).
		Foreground(theme.Current().Accent)
}

// FooterHint returns the style for navigation hint text.
func FooterHint() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(theme.Current().TextDim)
}

// CursorStyle returns the style for the active cursor indicator in menus.
func CursorStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(theme.Current().Accent).
		Bold(true)
}

// SelectedItemStyle returns the style for the highlighted menu item title.
func SelectedItemStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(theme.Current().Accent).
		Bold(true)
}

// NormalItemStyle returns the style for non-selected menu item titles.
func NormalItemStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(theme.Current().FG)
}

// DimItemStyle returns the style for descriptions and secondary text.
func DimItemStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(theme.Current().TextDim)
}

// GeneratingFrame returns the frame style for the generating spinner box.
func GeneratingFrame() lipgloss.Style {
	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(theme.Current().Accent).
		Padding(1, 3)
}

// CenterView wraps content in centered placement within the available area.
func CenterView(width, height int, content string) string {
	return lipgloss.Place(width, height, lipgloss.Center, lipgloss.Center, content)
}
