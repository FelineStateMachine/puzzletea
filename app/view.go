package app

import (
	"fmt"
	"strings"

	"github.com/FelineStateMachine/puzzletea/game"
	"github.com/FelineStateMachine/puzzletea/stats"
	"github.com/FelineStateMachine/puzzletea/theme"
	"github.com/FelineStateMachine/puzzletea/ui"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

func (m model) View() tea.View {
	var content string

	switch m.state {
	case mainMenuView:
		content = ui.CenterView(m.width, m.height, m.nav.mainMenu.View())
	case playMenuView:
		content = ui.CenterView(m.width, m.height, m.nav.playMenu.ViewAsPanel("Play"))
	case optionsMenuView:
		items := m.nav.optionsMenu.RenderItems() + "\n\n" + ui.DimItemStyle().Render("- Dami")
		panel := ui.Panel("Options", items, "↑/↓ navigate • enter select • esc back")
		content = ui.CenterView(m.width, m.height, panel)
	case seedInputView:
		panel := ui.Panel(
			"Enter Seed",
			m.seedInputBody(),
			"↑/↓ change field • ←/→ game • enter confirm • esc back",
		)
		content = ui.CenterView(m.width, m.height, panel)
	case gameSelectView:
		content = m.gameSelectViewContent()
	case modeSelectView:
		panel := ui.Panel(
			m.nav.selectedCategory.Name+" — Select Mode",
			m.nav.modeSelectList.View(),
			"↑/↓ navigate • enter select • esc back",
		)
		content = ui.CenterView(m.width, m.height, panel)
	case generatingView:
		s := m.spinner.View() + " Generating puzzle..."
		box := ui.GeneratingFrame().Render(s)
		content = ui.CenterView(m.width, m.height, box)
	case continueView:
		var s string
		if len(m.nav.continueGames) == 0 {
			s = ui.Panel(
				"Saved Games",
				"No saved games yet.",
				"esc back",
			)
		} else {
			footer := "↑/↓ navigate • enter resume • esc back"
			if pg := ui.TablePagination(m.nav.continueTable); pg != "" {
				footer = pg + "  " + footer
			}
			s = ui.Panel(
				"Saved Games",
				m.nav.continueTable.View(),
				footer,
			)
		}
		content = ui.CenterView(m.width, m.height, s)
	case gameView:
		if m.session.game == nil {
			content = ""
		} else {
			gameView := lipgloss.NewStyle().MaxWidth(m.width).Render(m.session.game.View())
			centered := gameView
			if m.debug.enabled {
				debugInfo := lipgloss.NewStyle().MaxWidth(m.width).Render(
					ui.DebugStyle().Render(m.debug.info),
				)
				centered = lipgloss.JoinVertical(lipgloss.Center, gameView, debugInfo)
			}
			content = ui.CenterView(m.width, m.height, centered)
		}
	case helpSelectView:
		panel := ui.Panel(
			"How to Play",
			m.nav.helpSelectList.View(),
			"↑/↓ navigate • enter select • esc back",
		)
		content = ui.CenterView(m.width, m.height, panel)
	case helpDetailView:
		panel := ui.Panel(
			m.help.category.Name+" — Guide",
			m.help.viewport.View(),
			"↑/↓ scroll • esc back",
		)
		content = ui.CenterView(m.width, m.height, panel)
	case themeSelectView:
		content = m.themeSelectViewContent()
	case statsView:
		statsWidth, _ := statsViewportSize(m.width, m.height, m.stats.cards)
		var statsBody string
		if len(m.stats.cards) == 0 {
			statsBody = m.stats.viewport.View()
		} else {
			banner := stats.RenderBanner(m.stats.profile, statsWidth)
			statsBody = lipgloss.JoinVertical(lipgloss.Left,
				banner,
				"",
				m.stats.viewport.View(),
			)
		}
		statsBody = lipgloss.NewStyle().Width(statsWidth).Render(statsBody)
		panel := ui.Panel(
			"Stats",
			statsBody,
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

func (m model) gameSelectViewContent() string {
	metrics := categoryPickerSize(m.width, m.height)

	listView := lipgloss.NewStyle().
		Width(metrics.listWidth).
		Height(metrics.listHeight).
		Render(m.nav.gameSelectList.View())

	detailBox := lipgloss.NewStyle().
		Width(metrics.detailWidth).
		Height(metrics.detailHeight).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(theme.Current().Border).
		Padding(1, 2).
		Render(m.nav.categoryDetail.View())

	body := lipgloss.JoinHorizontal(lipgloss.Top, listView, strings.Repeat(" ", categoryGapWidth), detailBox)
	if metrics.stacked {
		body = lipgloss.JoinVertical(lipgloss.Left, listView, "", detailBox)
	}
	body = lipgloss.NewStyle().
		Width(metrics.bodyWidth).
		Height(metrics.bodyHeight).
		Render(body)

	panel := ui.Panel(
		"Select Category",
		body,
		"↑/↓ navigate • pgup/pgdn details • / filter • enter select • esc back",
	)
	return ui.CenterView(m.width, m.height, panel)
}

func categoryPickerListHeight() int {
	const (
		maxVisibleItems = 8
		filterChrome    = 2
	)

	return maxVisibleItems + filterChrome
}

func renderCategoryDetailContent(cat game.Category, width int) string {
	p := theme.Current()
	desc := lipgloss.NewStyle().
		Foreground(p.FG).
		Width(width).
		Render(cat.Desc)

	meta := ui.DimItemStyle().Render(fmt.Sprintf("%d modes available", len(cat.Modes)))
	content := lipgloss.JoinVertical(lipgloss.Left,
		ui.PanelTitle().Render(cat.Name),
		meta,
		"",
		desc,
		"",
		ui.SelectedItemStyle().Render("Modes"),
		renderModeList(cat, width),
	)
	return content
}

func renderModeList(cat game.Category, width int) string {
	if len(cat.Modes) == 0 {
		return ui.DimItemStyle().Render("No modes available.")
	}

	lines := make([]string, 0, len(cat.Modes))
	for _, item := range cat.Modes {
		mode, ok := item.(game.Mode)
		if !ok {
			continue
		}
		line := lipgloss.NewStyle().Width(width).Render("• " + mode.Title())
		lines = append(lines, line)
	}
	if len(lines) == 0 {
		return ui.DimItemStyle().Render("No modes available.")
	}
	return strings.Join(lines, "\n")
}

// themeSelectViewContent renders the theme picker as a side-by-side layout:
// theme list on the left, color preview panel on the right.
func (m model) themeSelectViewContent() string {
	p := theme.Current()

	// Determine selected theme name for the preview.
	themeName := theme.DefaultThemeName
	if item, ok := m.theme.list.SelectedItem().(ui.MenuItem); ok {
		themeName = item.ItemTitle
	}

	// Compute available inner height for the panel content.
	// Panel chrome: border (2) + padding (2) + title (1) + blank (1) + footer (1) + blank (1) = 8
	const panelChrome = 8
	innerH := m.height - panelChrome
	if innerH < 10 {
		innerH = 10
	}

	// Left side: theme list.
	listView := m.theme.list.View()

	// Right side: color preview.
	previewBorder := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(p.Border).
		Padding(1, 2)

	preview := theme.PreviewPanel(themeName, innerH-4)
	previewBox := previewBorder.Render(preview)

	// Join side by side with a gap.
	spacer := "  "
	body := lipgloss.JoinHorizontal(lipgloss.Top, listView, spacer, previewBox)

	panel := ui.Panel(
		"Select Theme",
		body,
		"↑/↓ navigate • / filter • enter select • esc back",
	)
	return ui.CenterView(m.width, m.height, panel)
}
