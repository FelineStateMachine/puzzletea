package ui

import (
	"strings"

	"charm.land/lipgloss/v2"
)

// logo is the pre-rendered "puzzletea" ASCII art (FIGlet "small" font).
// Stored as a raw string to avoid a runtime dependency on go-figure.
const logo = "" +
	"                          _         _\n" +
	"  _ __   _  _   ___  ___ | |  ___  | |_   ___   __ _\n" +
	" | '_ \\ | || | |_ / |_ / | | / -_) |  _| / -_) / _` |\n" +
	" | .__/  \\_,_| /__| /__| |_| \\___|  \\__| \\___| \\__,_|\n" +
	" |_|"

// MainMenu is a custom main menu component with an ASCII art logo
// and styled item list. It replaces the generic list.Model for the
// root menu to give the application a game-themed identity.
type MainMenu struct {
	items  []MenuItem
	cursor int
}

// NewMainMenu creates a main menu with an ASCII art logo and the given items.
func NewMainMenu(items []MenuItem) MainMenu {
	return MainMenu{
		items: items,
	}
}

// CursorUp moves the selection cursor up, wrapping around.
func (m *MainMenu) CursorUp() {
	m.cursor--
	if m.cursor < 0 {
		m.cursor = len(m.items) - 1
	}
}

// CursorDown moves the selection cursor down, wrapping around.
func (m *MainMenu) CursorDown() {
	m.cursor++
	if m.cursor >= len(m.items) {
		m.cursor = 0
	}
}

// Selected returns the currently highlighted menu item.
func (m MainMenu) Selected() MenuItem {
	return m.items[m.cursor]
}

// View renders the main menu as a framed panel with logo and items.
func (m MainMenu) View() string {
	styledLogo := LogoStyle.Render(logo)

	var itemLines []string
	for i, item := range m.items {
		var line string
		if i == m.cursor {
			cursor := CursorStyle.Render("▸ ")
			title := SelectedItemStyle.Render(item.ItemTitle)
			desc := DimItemStyle.Render("  " + item.Desc)
			line = cursor + title + desc
		} else {
			title := NormalItemStyle.Render(item.ItemTitle)
			desc := DimItemStyle.Render("  " + item.Desc)
			line = "  " + title + desc
		}
		itemLines = append(itemLines, line)
	}
	items := strings.Join(itemLines, "\n")

	footer := FooterHint.Render("↑/↓ navigate • enter select • ctrl+c quit")

	inner := lipgloss.JoinVertical(lipgloss.Left,
		styledLogo,
		"",
		items,
		"",
		footer,
	)

	return HeavyPanel(inner)
}
