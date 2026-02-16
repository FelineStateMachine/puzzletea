package ui

import (
	"charm.land/bubbles/v2/list"
	"charm.land/lipgloss/v2"
	"charm.land/lipgloss/v2/compat"
)

// InitList creates a styled list widget with the standard puzzletea theme.
func InitList(items []list.Item, title string) list.Model {
	d := list.NewDefaultDelegate()
	d.Styles.SelectedTitle = d.Styles.SelectedTitle.
		Foreground(MenuAccent).
		BorderLeftForeground(MenuAccent)
	d.Styles.SelectedDesc = d.Styles.SelectedDesc.
		Foreground(MenuAccentLight).
		BorderLeftForeground(MenuAccent)
	d.Styles.NormalTitle = d.Styles.NormalTitle.
		Foreground(MenuText)
	d.Styles.NormalDesc = d.Styles.NormalDesc.
		Foreground(MenuTextDim)

	l := list.New(items, d, 64, 64)
	l.Title = title
	l.Styles.Title = lipgloss.NewStyle().
		Bold(true).
		Foreground(compat.AdaptiveColor{Light: lipgloss.Color("255"), Dark: lipgloss.Color("255")}).
		Background(MenuAccent).
		Padding(0, 1)
	l.DisableQuitKeybindings()
	l.SetFilteringEnabled(false)
	l.SetShowHelp(false)
	l.SetShowStatusBar(false)
	l.SetShowPagination(false)
	return l
}

// ListHeight returns the minimum height needed to display all items in a
// single page, avoiding pagination. The bubbles list component has a
// circular dependency between pagination visibility and PerPage calculation
// that causes rendering glitches when pages change, so we size the list to
// prevent pagination entirely when the terminal is tall enough.
func ListHeight(l list.Model) int {
	n := len(l.Items())
	// title bar (title line + bottom padding) = 2 lines
	const titleHeight = 2
	// Default delegate: height=2, spacing=1 → 3 lines per item.
	const itemSlot = 3
	return titleHeight + itemSlot*n
}
