package app

import (
	"fmt"
	"strings"

	"github.com/FelineStateMachine/puzzletea/registry"
	"github.com/FelineStateMachine/puzzletea/theme"
	"github.com/FelineStateMachine/puzzletea/ui"

	"charm.land/bubbles/v2/list"
	"charm.land/bubbles/v2/viewport"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

func (m model) View() tea.View {
	v := tea.NewView(m.viewContent())
	v.AltScreen = true
	if m.state == gameView || m.state == createView || m.state == exportView {
		v.MouseMode = tea.MouseModeCellMotion
		v.KeyboardEnhancements.ReportEventTypes = true
	}
	return v
}

func (m model) viewContent() string {
	if m.state == gameView {
		return m.renderGameView()
	}

	screen := m.activeScreen()
	if screen == nil {
		return fmt.Sprintf("unknown state: %d", m.state)
	}
	return screen.View(m.notice)
}

func gameSelectViewContent(width, height int, gameList list.Model, detail viewport.Model, notice noticeState) string {
	metrics := categoryPickerSize(width, height)

	listView := lipgloss.NewStyle().
		Width(metrics.listWidth).
		Height(metrics.listHeight).
		Border(lipgloss.NormalBorder(), false, true, false, false).
		BorderForeground(theme.Current().Border).
		Render(gameList.View())

	detailBox := lipgloss.NewStyle().
		Width(metrics.detailWidth).
		Height(metrics.detailHeight).
		Padding(2, 3).
		Render(detail.View())

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
		appendNoticeContent(width, notice, body),
		"↑/↓ navigate • / filter • enter select • esc back",
	)
	return ui.CenterView(width, height, panel)
}

func weeklyViewContent(w weeklyState) string {
	title := "Weekly Gauntlet — " + weeklyPanelTitle(w.cursor)
	if len(w.rows) == 0 {
		body := "No completed puzzles for this week yet."
		if isCurrentWeeklySelection(w.cursor) {
			body = "No weekly puzzles are available."
		}
		return ui.Panel(
			title,
			body,
			"←/→ week • esc back",
		)
	}

	footer := "←/→ week • enter open • esc back"
	if !isCurrentWeeklySelection(w.cursor) {
		footer = "←/→ week • enter review • esc back"
	}
	if pg := ui.TablePagination(w.table); pg != "" {
		footer = pg + "  " + footer
	}

	description := w.table.View()
	if !isCurrentWeeklySelection(w.cursor) {
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

func themeSelectViewContent(width, height int, thm themeState, notice noticeState) string {
	p := theme.Current()

	themeName := theme.DefaultThemeName
	if item, ok := thm.list.SelectedItem().(ui.MenuItem); ok {
		themeName = item.ItemTitle
	}

	const panelChrome = 8
	innerH := height - panelChrome
	innerH = max(innerH, 10)

	listView := thm.list.View()

	previewBorder := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(p.Border).
		Padding(1, 2)

	preview := theme.PreviewPanel(themeName, innerH-4)
	previewBox := previewBorder.Render(preview)

	body := lipgloss.JoinHorizontal(lipgloss.Top, listView, "  ", previewBox)

	panel := ui.Panel(
		"Select Theme",
		appendNoticeContent(width, notice, body),
		"↑/↓ navigate • / filter • enter select • esc back",
	)
	return ui.CenterView(width, height, panel)
}
