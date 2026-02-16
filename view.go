package main

import (
	"fmt"

	"github.com/FelineStateMachine/puzzletea/ui"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"charm.land/lipgloss/v2/compat"
)

func (m model) View() tea.View {
	var content string

	switch m.state {
	case mainMenuView:
		content = ui.CenterView(m.width, m.height, m.mainMenuList.View())
	case gameSelectView:
		content = ui.CenterView(m.width, m.height, m.gameSelectList.View())
	case modeSelectView:
		content = ui.CenterView(m.width, m.height, m.modeSelectList.View())
	case generatingView:
		s := m.spinner.View() + " Generating puzzle..."
		content = ui.CenterView(m.width, m.height, s)
	case continueView:
		var s string
		if len(m.continueGames) == 0 {
			s = "No saved games yet.\n\nPress Escape to return."
		} else {
			title := lipgloss.NewStyle().
				Bold(true).
				Foreground(compat.AdaptiveColor{Light: lipgloss.Color("255"), Dark: lipgloss.Color("255")}).
				Background(ui.MenuAccent).
				Padding(0, 1).
				Render("Saved Games")
			s = lipgloss.JoinVertical(lipgloss.Left, title, "", m.continueTable.View())
		}
		content = ui.CenterView(m.width, m.height, s)
	case gameView:
		if m.game == nil {
			content = ""
		} else {
			gameView := lipgloss.NewStyle().MaxWidth(m.width).Render(m.game.View())
			centered := gameView
			if m.debug {
				debugInfo := lipgloss.NewStyle().MaxWidth(m.width).Render(
					ui.DebugStyle.Render(m.debuginfo),
				)
				centered = lipgloss.JoinVertical(lipgloss.Center, gameView, debugInfo)
			}
			content = ui.CenterView(m.width, m.height, centered)
		}
	case helpSelectView:
		content = ui.CenterView(m.width, m.height, m.helpSelectList.View())
	case helpDetailView:
		footer := lipgloss.NewStyle().
			Foreground(ui.MenuTextDim).
			Render("↑/↓ scroll • esc back")
		s := lipgloss.JoinVertical(lipgloss.Left,
			m.helpViewport.View(),
			footer,
		)
		content = ui.CenterView(m.width, m.height, s)
	default:
		content = fmt.Sprintf("unknown state: %d", m.state)
	}

	v := tea.NewView(content)
	v.AltScreen = true
	if m.state == gameView {
		v.MouseMode = tea.MouseModeCellMotion
		v.KeyboardEnhancements.ReportEventTypes = true
	}
	return v
}
