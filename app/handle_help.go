package app

import (
	tea "charm.land/bubbletea/v2"
)

func (m model) handleHelpSelectEnter() (model, tea.Cmd) {
	if m.help.category.Definition.Name == "" {
		return m, nil
	}
	m = m.updateHelpDetailViewport()
	m.state = helpDetailView
	m = m.initScreen(helpDetailView)
	return m, nil
}
