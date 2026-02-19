package ui

import (
	"strings"

	"charm.land/lipgloss/v2"
)

const logo = "puzzletea"

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

func (m MainMenu) Selected() MenuItem {
	return m.items[m.cursor]
}

// View renders the main menu as a framed panel with logo and items.
func (m MainMenu) View() string {
	styledLogo := LogoStyle().Render(logo)

	items := m.RenderItems()

	footer := FooterHint().Render("↑/↓ navigate • enter select • ctrl+c quit")

	inner := lipgloss.JoinVertical(lipgloss.Left,
		styledLogo,
		"",
		items,
		"",
		footer,
	)

	return HeavyPanel(inner)
}

// ViewAsPanel renders the menu items inside a rounded-border Panel,
// suitable for submenus that should look visually distinct from the
// root main menu.
func (m MainMenu) ViewAsPanel(title string) string {
	return Panel(title, m.RenderItems(), "↑/↓ navigate • enter select • esc back")
}

// RenderItems returns the styled item list shared by View and ViewAsPanel.
func (m MainMenu) RenderItems() string {
	// Pre-compute the widest item line so the block has a stable width
	// regardless of which item is selected.
	maxW := 0
	for _, item := range m.items {
		// Selected cursor prefix ("⦾  ") is the widest variant at 3 chars.
		w := lipgloss.Width("⦾  " + item.ItemTitle)
		if dw := lipgloss.Width("  " + item.Desc); dw > w {
			w = dw
		}
		if w > maxW {
			maxW = w
		}
	}

	var itemLines []string
	for i, item := range m.items {
		var line, cursor, title string
		if i == m.cursor {
			cursor = CursorStyle().Render("⦾  ")
			title = SelectedItemStyle().Render(item.ItemTitle)
		} else {
			cursor = CursorStyle().Render("◦ ")
			title = NormalItemStyle().Render(item.ItemTitle)
		}
		desc := DimItemStyle().Render("\n  " + item.Desc)
		line = cursor + title + desc + "\n"
		itemLines = append(itemLines, line)
	}

	content := strings.Join(itemLines, "\n")
	return lipgloss.NewStyle().Width(maxW).Render(content)
}
