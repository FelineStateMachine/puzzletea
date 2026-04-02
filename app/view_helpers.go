package app

import (
	"github.com/FelineStateMachine/puzzletea/ui"

	"charm.land/lipgloss/v2"
)

func renderContinueView(cont continueState, width, height int, notice noticeState) string {
	if len(cont.games) == 0 {
		return renderPanelView(width, height, notice, "Saved Games", "No saved games yet.", "esc back")
	}

	footer := "↑/↓ navigate • enter resume • esc back"
	if pg := ui.TablePagination(cont.table); pg != "" {
		footer = pg + "  " + footer
	}
	return renderPanelView(width, height, notice, "Saved Games", cont.table.View(), footer)
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

func renderStatsView(stats statsState, width, height int, notice noticeState) string {
	statsWidth, _ := statsViewportSize(width, height, stats.cards)
	var statsBody string
	if len(stats.cards) == 0 {
		statsBody = stats.viewport.View()
	} else {
		banner := ui.RenderStatsBanner(stats.profile, statsWidth)
		statsBody = lipgloss.JoinVertical(lipgloss.Left,
			banner,
			"",
			stats.viewport.View(),
		)
	}
	statsBody = lipgloss.NewStyle().Width(statsWidth).Render(statsBody)
	return renderPanelView(width, height, notice, "Stats", statsBody, "↑/↓ scroll • esc back")
}
