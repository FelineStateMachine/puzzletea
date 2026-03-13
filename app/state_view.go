package app

import (
	"fmt"

	"github.com/FelineStateMachine/puzzletea/ui"

	"charm.land/lipgloss/v2"
)

func (m model) viewContent() string {
	switch m.state {
	case mainMenuView:
		return ui.CenterView(m.width, m.height, m.appendNotice(m.nav.mainMenu.View()))
	case playMenuView:
		return ui.CenterView(m.width, m.height, m.appendNotice(m.nav.playMenu.ViewAsPanel("Play")))
	case optionsMenuView:
		items := m.nav.optionsMenu.RenderItems() + "\n\n" + ui.DimItemStyle().Render("- Dami")
		return m.renderPanel("Options", items, "↑/↓ navigate • enter select • esc back")
	case seedInputView:
		return m.renderPanel(
			"Enter Seed",
			m.seedInputBody(),
			"↑/↓ change field • ←/→ game • enter confirm • esc back",
		)
	case gameSelectView:
		return m.gameSelectViewContent()
	case modeSelectView:
		return m.renderPanel(
			m.nav.selectedCategory.Definition.Name+" — Select Mode",
			m.nav.modeSelectList.View(),
			"↑/↓ navigate • enter select • esc back",
		)
	case generatingView:
		s := m.spinner.View() + " Generating puzzle..."
		box := ui.GeneratingFrame().Render(m.appendNotice(s))
		return ui.CenterView(m.width, m.height, box)
	case continueView:
		return m.renderContinueView()
	case weeklyView:
		return ui.CenterView(m.width, m.height, m.appendNotice(m.weeklyViewContent()))
	case gameView:
		return m.renderGameView()
	case helpSelectView:
		return m.renderPanel(
			"How to Play",
			m.help.selectList.View(),
			"↑/↓ navigate • enter select • esc back",
		)
	case helpDetailView:
		return m.renderPanel(
			m.help.category.Definition.Name+" — Guide",
			m.help.viewport.View(),
			"↑/↓ scroll • esc back",
		)
	case themeSelectView:
		return m.themeSelectViewContent()
	case statsView:
		return m.renderStatsView()
	default:
		return fmt.Sprintf("unknown state: %d", m.state)
	}
}

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
