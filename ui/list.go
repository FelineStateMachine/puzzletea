package ui

import (
	"charm.land/bubbles/v2/list"
)

// InitList creates a styled list widget with the standard puzzletea theme.
// The list title is hidden because lists are rendered inside Panel frames
// that provide their own styled title.
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
	l.SetShowTitle(false)
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
	// Default delegate: height=2, spacing=1 → 3 lines per item.
	// Title is hidden (rendered by Panel), so no extra title height needed.
	const itemSlot = 3
	return itemSlot * n
}
