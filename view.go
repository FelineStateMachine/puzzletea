package main

import (
	"fmt"

	"github.com/FelineStateMachine/puzzletea/ui"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

func (m model) View() tea.View {
	var content string

	switch m.state {
	case mainMenuView:
		content = ui.CenterView(m.width, m.height, m.mainMenu.View())
	case gameSelectView:
		panel := ui.Panel(
			"Select Category",
			m.gameSelectList.View(),
			"↑/↓ navigate • enter select • esc back",
		)
		content = ui.CenterView(m.width, m.height, panel)
	case modeSelectView:
		panel := ui.Panel(
			m.selectedCategory.Name+" — Select Mode",
			m.modeSelectList.View(),
			"↑/↓ navigate • enter select • esc back",
		)
		content = ui.CenterView(m.width, m.height, panel)
	case generatingView:
		s := m.spinner.View() + " Generating puzzle..."
		box := ui.GeneratingFrame.Render(s)
		content = ui.CenterView(m.width, m.height, box)
	case continueView:
		var s string
		if len(m.continueGames) == 0 {
			s = ui.Panel(
				"Saved Games",
				"No saved games yet.",
				"esc back",
			)
		} else {
			footer := "↑/↓ navigate • enter resume • esc back"
			if pg := ui.TablePagination(m.continueTable); pg != "" {
				footer = pg + "  " + footer
			}
			s = ui.Panel(
				"Saved Games",
				m.continueTable.View(),
				footer,
			)
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
		panel := ui.Panel(
			"How to Play",
			m.helpSelectList.View(),
			"↑/↓ navigate • enter select • esc back",
		)
		content = ui.CenterView(m.width, m.height, panel)
	case helpDetailView:
		panel := ui.Panel(
			m.helpCategory.Name+" — Guide",
			m.helpViewport.View(),
			"↑/↓ scroll • esc back",
		)
		content = ui.CenterView(m.width, m.height, panel)
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
