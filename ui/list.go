package ui

import (
	"charm.land/bubbles/v2/list"
	"github.com/FelineStateMachine/puzzletea/theme"
)

func applyThemeListStyles(d *list.DefaultDelegate, p theme.Palette, width int) {
	// Title styles intentionally avoid fixed Width: bubbles uses them for
	// rune-level filter highlighting and Width can force wrapped segments.
	d.Styles.SelectedTitle = d.Styles.SelectedTitle.
		Foreground(p.Accent).
		BorderLeftForeground(p.Accent)
	d.Styles.NormalTitle = d.Styles.NormalTitle.
		Foreground(p.FG)

	// Description lines carry the fixed width so the list column remains stable.
	d.Styles.SelectedDesc = d.Styles.SelectedDesc.
		Foreground(p.AccentSoft).
		BorderLeftForeground(p.Accent).
		Width(width)
	d.Styles.NormalDesc = d.Styles.NormalDesc.
		Foreground(p.TextDim).
		Width(width)
}

// InitList creates a styled list widget with the active theme's colors.
// The list title is hidden because lists are rendered inside Panel frames
// that provide their own styled title.
func InitList(items []list.Item, title string) list.Model {
	p := theme.Current()
	d := list.NewDefaultDelegate()
	d.Styles.SelectedTitle = d.Styles.SelectedTitle.
		Foreground(p.Accent).
		BorderLeftForeground(p.Accent)
	d.Styles.SelectedDesc = d.Styles.SelectedDesc.
		Foreground(p.AccentSoft).
		BorderLeftForeground(p.Accent)
	d.Styles.NormalTitle = d.Styles.NormalTitle.
		Foreground(p.FG)
	d.Styles.NormalDesc = d.Styles.NormalDesc.
		Foreground(p.TextDim)

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

// InitThemeList creates a styled list widget for the theme picker with
// a fixed-width delegate so the list column doesn't shift as the cursor
// moves across names of different lengths.
func InitThemeList(items []list.Item, width, height int) list.Model {
	p := theme.Current()
	d := list.NewDefaultDelegate()
	applyThemeListStyles(&d, p, width)

	l := list.New(items, d, width, height)
	l.SetShowTitle(false)
	l.DisableQuitKeybindings()
	l.SetFilteringEnabled(true)
	l.SetShowHelp(false)
	l.SetShowStatusBar(false)
	return l
}

// UpdateThemeListStyles refreshes the delegate colors on an existing theme
// list so that live-preview theme changes are reflected immediately.
func UpdateThemeListStyles(l *list.Model) {
	p := theme.Current()
	d := list.NewDefaultDelegate()

	// Preserve the width that was set when the list was created.
	w := l.Width()
	applyThemeListStyles(&d, p, w)

	l.SetDelegate(d)
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
