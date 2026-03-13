package app

import (
	"fmt"

	"github.com/FelineStateMachine/puzzletea/ui"

	"charm.land/lipgloss/v2"
)

func (m model) viewContent() string {
	switch m.state {
	case mainMenuView:
		return ui.CenterView(m.width, m.height, m.nav.mainMenu.View())
	case playMenuView:
		return ui.CenterView(m.width, m.height, m.nav.playMenu.ViewAsPanel("Play"))
	case optionsMenuView:
		items := m.nav.optionsMenu.RenderItems() + "\n\n" + ui.DimItemStyle().Render("- Dami")
		panel := ui.Panel("Options", items, "↑/↓ navigate • enter select • esc back")
		return ui.CenterView(m.width, m.height, panel)
	case seedInputView:
		panel := ui.Panel(
			"Enter Seed",
			m.seedInputBody(),
			"↑/↓ change field • ←/→ game • enter confirm • esc back",
		)
		return ui.CenterView(m.width, m.height, panel)
	case gameSelectView:
		return m.gameSelectViewContent()
	case modeSelectView:
		panel := ui.Panel(
			m.nav.selectedCategory.Definition.Name+" — Select Mode",
			m.nav.modeSelectList.View(),
			"↑/↓ navigate • enter select • esc back",
		)
		return ui.CenterView(m.width, m.height, panel)
	case generatingView:
		s := m.spinner.View() + " Generating puzzle..."
		box := ui.GeneratingFrame().Render(s)
		return ui.CenterView(m.width, m.height, box)
	case continueView:
		return m.renderContinueView()
	case weeklyView:
		return ui.CenterView(m.width, m.height, m.weeklyViewContent())
	case gameView:
		return m.renderGameView()
	case helpSelectView:
		panel := ui.Panel(
			"How to Play",
			m.nav.helpSelectList.View(),
			"↑/↓ navigate • enter select • esc back",
		)
		return ui.CenterView(m.width, m.height, panel)
	case helpDetailView:
		panel := ui.Panel(
			m.help.category.Definition.Name+" — Guide",
			m.help.viewport.View(),
			"↑/↓ scroll • esc back",
		)
		return ui.CenterView(m.width, m.height, panel)
	case themeSelectView:
		return m.themeSelectViewContent()
	case statsView:
		return m.renderStatsView()
	default:
		return fmt.Sprintf("unknown state: %d", m.state)
	}
}

func (m model) renderContinueView() string {
	if len(m.nav.continueGames) == 0 {
		panel := ui.Panel("Saved Games", "No saved games yet.", "esc back")
		return ui.CenterView(m.width, m.height, panel)
	}

	footer := "↑/↓ navigate • enter resume • esc back"
	if pg := ui.TablePagination(m.nav.continueTable); pg != "" {
		footer = pg + "  " + footer
	}
	panel := ui.Panel("Saved Games", m.nav.continueTable.View(), footer)
	return ui.CenterView(m.width, m.height, panel)
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
	return ui.CenterView(m.width, m.height, centered)
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
	panel := ui.Panel("Stats", statsBody, "↑/↓ scroll • esc back")
	return ui.CenterView(m.width, m.height, panel)
}
