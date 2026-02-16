package main

import (
	"fmt"

	"github.com/FelineStateMachine/puzzletea/ui"

	"github.com/charmbracelet/lipgloss"
)

func (m model) View() string {
	switch m.state {
	case mainMenuView:
		return ui.CenterView(m.width, m.height, m.mainMenuList.View())
	case gameSelectView:
		return ui.CenterView(m.width, m.height, m.gameSelectList.View())
	case modeSelectView:
		return ui.CenterView(m.width, m.height, m.modeSelectList.View())
	case generatingView:
		s := m.spinner.View() + " Generating puzzle..."
		return ui.CenterView(m.width, m.height, s)
	case continueView:
		var s string
		if len(m.continueGames) == 0 {
			s = "No saved games yet.\n\nPress Escape to return."
		} else {
			title := lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.AdaptiveColor{Light: "255", Dark: "255"}).
				Background(ui.MenuAccent).
				Padding(0, 1).
				Render("Saved Games")
			s = lipgloss.JoinVertical(lipgloss.Left, title, "", m.continueTable.View())
		}
		return ui.CenterView(m.width, m.height, s)
	case gameView:
		if m.game == nil {
			return ""
		}
		var debugInfo string
		if m.debug {
			debugInfo = ui.DebugStyle.Render(m.debuginfo)
		}
		s := lipgloss.JoinVertical(lipgloss.Center,
			m.game.View(),
			debugInfo,
		)
		return ui.CenterView(m.width, m.height, s)
	case helpSelectView:
		return ui.CenterView(m.width, m.height, m.helpSelectList.View())
	case helpDetailView:
		footer := lipgloss.NewStyle().
			Foreground(ui.MenuTextDim).
			Render("↑/↓ scroll • esc back")
		s := lipgloss.JoinVertical(lipgloss.Left,
			m.helpViewport.View(),
			footer,
		)
		return ui.CenterView(m.width, m.height, s)
	default:
		return fmt.Sprintf("unknown state: %d", m.state)
	}
}
