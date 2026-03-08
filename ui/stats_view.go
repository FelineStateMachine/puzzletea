package ui

import (
	"fmt"
	"image/color"
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/FelineStateMachine/puzzletea/stats"
	"github.com/FelineStateMachine/puzzletea/theme"
)

func statsCardFrameStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(theme.Current().Border).
		Padding(0, 1).
		Width(stats.CardInnerWidth + 4)
}

func statsCardTitleStyle(fg color.Color) lipgloss.Style {
	return lipgloss.NewStyle().Bold(true).Foreground(fg)
}

func statsLabelStyle() lipgloss.Style {
	return lipgloss.NewStyle().Foreground(theme.Current().TextDim)
}

func statsValueStyle() lipgloss.Style {
	return lipgloss.NewStyle().Foreground(theme.Current().Accent).Bold(true)
}

func statsXPBarFilledStyle() lipgloss.Style {
	return lipgloss.NewStyle().Foreground(theme.Current().Accent)
}

func statsXPBarEmptyStyle() lipgloss.Style {
	return lipgloss.NewStyle().Foreground(theme.Current().Border)
}

func statsXPFractionStyle() lipgloss.Style {
	return lipgloss.NewStyle().Foreground(theme.Current().FG)
}

func statsBannerRuleStyle() lipgloss.Style {
	return lipgloss.NewStyle().Foreground(theme.Current().Border)
}

func statsBannerLabelStyle() lipgloss.Style {
	return lipgloss.NewStyle().Foreground(theme.Current().FG)
}

func statsBannerValueStyle() lipgloss.Style {
	return lipgloss.NewStyle().Foreground(theme.Current().Accent).Bold(true)
}

func RenderStatsCard(c stats.Card, titleColor color.Color) string {
	title := statsCardTitleStyle(titleColor).Render(
		fmt.Sprintf("%s \u2014 LVL %d", strings.ToUpper(c.GameType), c.Level),
	)

	lines := []string{
		title,
		renderStatsStatLine("Preferred Mode:", c.PreferredMode),
		renderStatsStatLine("Total Victories:", fmt.Sprintf("%d", c.Victories)),
		renderStatsStatLine("Total Attempts:", fmt.Sprintf("%d", c.Attempts)),
		renderStatsStatLine("Daily Played:", fmt.Sprintf("%d", c.DailyPlayed)),
		"",
		renderStatsXPBar(c.CurrentXP, c.Level),
	}

	return statsCardFrameStyle().Render(strings.Join(lines, "\n"))
}

func renderStatsStatLine(label, value string) string {
	l := statsLabelStyle().Render(label)
	v := statsValueStyle().Render(value)
	gap := stats.CardInnerWidth - lipgloss.Width(l) - lipgloss.Width(v)
	gap = max(gap, 1)
	return l + strings.Repeat(" ", gap) + v
}

func renderStatsXPBar(currentXP, level int) string {
	currentLevelXP := stats.XPForLevel(level)
	nextLevelXP := stats.XPForLevel(level + 1)
	xpIntoLevel := currentXP - currentLevelXP
	xpNeeded := nextLevelXP - currentLevelXP
	if xpNeeded <= 0 {
		xpNeeded = 1
	}

	const barWidth = 12
	filled := 0
	if xpNeeded > 0 {
		filled = int(float64(xpIntoLevel) / float64(xpNeeded) * barWidth)
	}
	filled = max(0, min(filled, barWidth))

	bar := statsXPBarFilledStyle().Render(strings.Repeat("\u2588", filled)) +
		statsXPBarEmptyStyle().Render(strings.Repeat("\u2591", barWidth-filled))
	frac := statsXPFractionStyle().Render(fmt.Sprintf(" %d/%d", xpIntoLevel, xpNeeded))
	return statsLabelStyle().Foreground(theme.Current().FG).Render("XP ") + bar + frac
}

func RenderStatsBanner(b stats.ProfileBanner, width int) string {
	if width < 20 {
		width = 40
	}

	ruleStr := statsBannerRuleStyle().Render(
		"\u2500\u2500\u2500 Profile " + strings.Repeat("\u2500", max(width-12, 1)),
	)

	streakStr := fmt.Sprintf("%d days", b.DailyStreak)
	if b.DailyStreak == 1 {
		streakStr = "1 day"
	}
	streakSuffix := "( )"
	if b.CurrentDaily {
		streakSuffix = "(\u2714)"
	}
	streakStr += " " + streakSuffix

	col1 := statsBannerLabelStyle().Render("Profile Level: ") + statsBannerValueStyle().Render(fmt.Sprintf("%d", b.ProfileLevel))
	col2 := statsBannerLabelStyle().Render("Daily Streak: ") + statsBannerValueStyle().Render(streakStr)
	col3 := statsBannerLabelStyle().Render("Total Dailies: ") + statsBannerValueStyle().Render(fmt.Sprintf("%d", b.TotalDailies))
	col4 := statsBannerLabelStyle().Render("Weeklies Completed: ") +
		statsBannerValueStyle().Render(fmt.Sprintf("%d total (%2d/99 this week)", b.WeekliesCompleted, b.ThisWeekHighestIndex))

	gap := width - lipgloss.Width(col1) - lipgloss.Width(col2)
	gap = max(gap, 2)
	row1 := col1 + strings.Repeat(" ", gap) + col2

	gap2 := width - lipgloss.Width(col3) - lipgloss.Width(col4)
	gap2 = max(gap2, 2)
	row2 := col3 + strings.Repeat(" ", gap2) + col4

	bottomRule := statsBannerRuleStyle().Render(strings.Repeat("\u2500", max(width, 1)))
	return strings.Join([]string{ruleStr, row1, row2, bottomRule}, "\n")
}

func RenderStatsCardGrid(cards []stats.Card, width int) string {
	if len(cards) == 0 {
		return lipgloss.NewStyle().
			Foreground(theme.Current().TextDim).
			Render("No stats yet \u2014 play some puzzles!")
	}

	cols := 1
	if width > stats.CardFullWidth {
		cols = max((width+2)/(stats.CardFullWidth+2), 1)
	}
	palette := theme.Current().ThemeColors()
	rendered := make([]string, len(cards))
	for i, c := range cards {
		rendered[i] = RenderStatsCard(c, palette[i%len(palette)])
	}

	rows := make([]string, 0, (len(rendered)+cols-1)/cols)
	for i := 0; i < len(rendered); i += cols {
		end := min(i+cols, len(rendered))
		row := rendered[i]
		for j := i + 1; j < end; j++ {
			row = lipgloss.JoinHorizontal(lipgloss.Top, row, strings.Repeat(" ", 2), rendered[j])
		}
		rows = append(rows, row)
	}
	return strings.Join(rows, strings.Repeat("\n", 3))
}
