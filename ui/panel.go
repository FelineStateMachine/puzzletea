package ui

import (
	"charm.land/lipgloss/v2"
)

// Panel wraps content in a bordered frame with a styled title and footer hint.
// It uses RoundedBorder by default, suitable for sub-menus and secondary views.
func Panel(title, content, footer string) string {
	titleLine := PanelTitle().Render(title)
	footerLine := FooterHint().Render(footer)

	inner := lipgloss.JoinVertical(lipgloss.Left,
		titleLine,
		"",
		content,
		"",
		footerLine,
	)

	return PanelFrame().Render(inner)
}

// HeavyPanel wraps content in a double-border frame, used for the main menu.
func HeavyPanel(content string) string {
	return MainMenuFrame().Render(content)
}
