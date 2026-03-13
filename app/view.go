package app

import (
	"fmt"
	"strings"

	"github.com/FelineStateMachine/puzzletea/registry"
	"github.com/FelineStateMachine/puzzletea/theme"
	"github.com/FelineStateMachine/puzzletea/ui"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

func (m model) View() tea.View {
	v := tea.NewView(m.viewContent())
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
		Border(lipgloss.NormalBorder(), false, true, false, false).
		BorderForeground(theme.Current().Border).
		Render(m.nav.gameSelectList.View())

	detailBox := lipgloss.NewStyle().
		Width(metrics.detailWidth).
		Height(metrics.detailHeight).
		Padding(2, 3).
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
		m.appendNotice(body),
		"↑/↓ navigate • / filter • enter select • esc back",
	)
	return ui.CenterView(m.width, m.height, panel)
}

func (m model) weeklyViewContent() string {
	title := "Weekly Gauntlet — " + m.weeklyPanelTitle()
	if len(m.nav.weeklyRows) == 0 {
		body := "No completed puzzles for this week yet."
		if m.isCurrentWeeklySelection() {
			body = "No weekly puzzles are available."
		}
		return ui.Panel(
			title,
			body,
			"←/→ week • esc back",
		)
	}

	footer := "←/→ week • enter open • esc back"
	if !m.isCurrentWeeklySelection() {
		footer = "←/→ week • enter review • esc back"
	}
	if pg := ui.TablePagination(m.nav.weeklyTable); pg != "" {
		footer = pg + "  " + footer
	}

	description := m.nav.weeklyTable.View()
	if !m.isCurrentWeeklySelection() {
		description = lipgloss.JoinVertical(
			lipgloss.Left,
			ui.DimItemStyle().Render("Review only: completed puzzles from this week."),
			"",
			description,
		)
	}

	return ui.Panel(title, description, footer)
}

func categoryPickerListHeight() int {
	const (
		maxVisibleItems = 8
		filterChrome    = 2
	)

	return maxVisibleItems + filterChrome
}

func renderCategoryDetailContent(entry registry.Entry, width int) string {
	p := theme.Current()
	desc := lipgloss.NewStyle().
		Foreground(p.FG).
		Width(width).
		Render(entry.Definition.Description)

	meta := ui.DimItemStyle().Render(fmt.Sprintf("%d modes available", len(entry.Modes)))
	content := lipgloss.JoinVertical(lipgloss.Left,
		ui.PanelTitle().Render(entry.Definition.Name),
		meta,
		"",
		desc,
		"",
		ui.SelectedItemStyle().Render("Modes"),
		renderModeList(entry, width),
	)
	return content
}

func renderModeList(entry registry.Entry, width int) string {
	if len(entry.Modes) == 0 {
		return ui.DimItemStyle().Render("No modes available.")
	}

	displayTitles := modeDisplayTitles(entry)
	modeNames := make([]string, 0, len(displayTitles))
	for _, title := range displayTitles {
		modeNames = append(modeNames, "• "+title)
	}
	if len(modeNames) == 0 {
		return ui.DimItemStyle().Render("No modes available.")
	}

	if width < 24 || len(modeNames) == 1 {
		lines := make([]string, 0, len(modeNames))
		for _, name := range modeNames {
			lines = append(lines, lipgloss.NewStyle().Width(width).Render(name))
		}
		return strings.Join(lines, "\n")
	}

	rows := (len(modeNames) + 1) / 2
	leftWidth := max(width/2-1, 1)
	rightWidth := max(width-leftWidth-2, 1)

	left := make([]string, 0, rows)
	right := make([]string, 0, rows)
	for i, name := range modeNames {
		line := lipgloss.NewStyle().Width(leftWidth).Render(name)
		if i < rows {
			left = append(left, line)
			continue
		}
		line = lipgloss.NewStyle().Width(rightWidth).Render(name)
		right = append(right, line)
	}
	for len(right) < len(left) {
		right = append(right, strings.Repeat(" ", rightWidth))
	}

	lines := make([]string, 0, rows)
	for i := range left {
		lines = append(lines, lipgloss.JoinHorizontal(lipgloss.Top, left[i], "  ", right[i]))
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
	innerH = max(innerH, 10)

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
		m.appendNotice(body),
		"↑/↓ navigate • / filter • enter select • esc back",
	)
	return ui.CenterView(m.width, m.height, panel)
}
