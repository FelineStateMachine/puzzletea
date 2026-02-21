package app

import (
	"fmt"

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
		content = ui.CenterView(m.width, m.height, m.mainMenu.View())
	case playMenuView:
		content = ui.CenterView(m.width, m.height, m.playMenu.ViewAsPanel("Play"))
	case optionsMenuView:
		items := m.optionsMenu.RenderItems() + "\n\n" + ui.DimItemStyle().Render("- Dami")
		panel := ui.Panel("Options", items, "↑/↓ navigate • enter select • esc back")
		content = ui.CenterView(m.width, m.height, panel)
	case seedInputView:
		panel := ui.Panel(
			"Enter Seed",
			m.seedInput.View(),
			"enter confirm • esc back",
		)
		content = ui.CenterView(m.width, m.height, panel)
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
		box := ui.GeneratingFrame().Render(s)
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
					ui.DebugStyle().Render(m.debuginfo),
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
	case themeSelectView:
		content = m.themeSelectViewContent()
	case statsView:
		statsWidth, _ := statsViewportSize(m.width, m.height, m.statsCards)
		var statsBody string
		if len(m.statsCards) == 0 {
			statsBody = m.statsViewport.View()
		} else {
			banner := stats.RenderBanner(m.statsProfile, statsWidth)
			statsBody = lipgloss.JoinVertical(lipgloss.Left,
				banner,
				"",
				m.statsViewport.View(),
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

// themeSelectViewContent renders the theme picker as a side-by-side layout:
// theme list on the left, color preview panel on the right.
func (m model) themeSelectViewContent() string {
	p := theme.Current()

	// Determine selected theme name for the preview.
	themeName := theme.DefaultThemeName
	if item, ok := m.themeList.SelectedItem().(ui.MenuItem); ok {
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
	listView := m.themeList.View()

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
