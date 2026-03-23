package app

import (
	"github.com/FelineStateMachine/puzzletea/theme"
	"github.com/FelineStateMachine/puzzletea/ui"

	"charm.land/bubbles/v2/list"
	tea "charm.land/bubbletea/v2"
)

func (m model) handleThemeEnter() (tea.Model, tea.Cmd) {
	m.theme.previous = m.cfg.Theme

	names := theme.ThemeNames()
	items := make([]list.Item, len(names))
	for i, n := range names {
		desc := "dark theme"
		if n == theme.DefaultThemeName {
			desc = "built-in earth-tone palette"
		} else if t := theme.LookupTheme(n); t != nil && !t.Meta.IsDark {
			desc = "light theme"
		}
		items[i] = ui.MenuItem{ItemTitle: n, Desc: desc}
	}

	const maxVisibleItems = 8
	listH := min(m.height, maxVisibleItems*3)
	listW := min(m.width, theme.MaxNameLen+4)

	m.theme.list = ui.InitThemeList(items, listW, listH)
	for i, item := range items {
		if mi, ok := item.(ui.MenuItem); ok && mi.ItemTitle == m.theme.previous {
			m.theme.list.Select(i)
			break
		}
	}
	if m.theme.previous == "" {
		m.theme.list.Select(0)
	}

	m.state = themeSelectView
	m = m.clearNotice()
	return m, nil
}

func (m model) handleThemeConfirm() (tea.Model, tea.Cmd) {
	item, ok := m.theme.list.SelectedItem().(ui.MenuItem)
	if !ok {
		return m, nil
	}

	themeName := item.ItemTitle
	if themeName == theme.DefaultThemeName {
		themeName = ""
	}

	_ = theme.Apply(item.ItemTitle)
	m.cfg.Theme = themeName
	if err := m.cfg.Save(m.configPath); err != nil {
		m = m.setErrorf("Theme changed, but saving config failed: %v", err)
	}

	m.state = mainMenuView
	return m, nil
}
