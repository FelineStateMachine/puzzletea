package app

import (
	tea "charm.land/bubbletea/v2"
)

func (m model) handleHelpSelectEnter() (tea.Model, tea.Cmd) {
	entry, ok := selectedCategoryEntry(m.help.selectList.SelectedItem())
	if !ok {
		return m, nil
	}
	m.help.category = entry
	m = m.updateHelpDetailViewport()
	m.state = helpDetailView
	return m, nil
}
