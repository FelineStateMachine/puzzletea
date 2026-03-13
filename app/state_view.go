package app

import (
	"github.com/FelineStateMachine/puzzletea/ui"

	"charm.land/lipgloss/v2"
)

func (m model) renderContinueView() string {
	if len(m.cont.games) == 0 {
		return m.renderPanel("Saved Games", "No saved games yet.", "esc back")
	}

	footer := "↑/↓ navigate • enter resume • esc back"
	if pg := ui.TablePagination(m.cont.table); pg != "" {
		footer = pg + "  " + footer
	}
	return m.renderPanel("Saved Games", m.cont.table.View(), footer)
}

func (m model) renderGameView() string {
	if m.session.game == nil {
		return ""
	}

	gameView := lipgloss.NewStyle().MaxWidth(m.width).Render(m.session.game.View())
	centered := gameView
	if m.debug.enabled {
		debugInfo := lipgloss.NewStyle().MaxWidth(m.width).Render(
			ui.DebugStyle().Render(m.debug.info),
		)
		centered = lipgloss.JoinVertical(lipgloss.Center, gameView, debugInfo)
	}
	return ui.CenterView(m.width, m.height, m.appendNotice(centered))
}

func (m model) renderStatsView() string {
	statsWidth, _ := statsViewportSize(m.width, m.height, m.stats.cards)
	var statsBody string
	if len(m.stats.cards) == 0 {
		statsBody = m.stats.viewport.View()
	} else {
		banner := ui.RenderStatsBanner(m.stats.profile, statsWidth)
		statsBody = lipgloss.JoinVertical(lipgloss.Left,
			banner,
			"",
			m.stats.viewport.View(),
		)
	}
	statsBody = lipgloss.NewStyle().Width(statsWidth).Render(statsBody)
	return m.renderPanel("Stats", statsBody, "↑/↓ scroll • esc back")
}
