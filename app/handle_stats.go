package app

import (
	"time"

	"github.com/FelineStateMachine/puzzletea/daily"
	"github.com/FelineStateMachine/puzzletea/registry"
	"github.com/FelineStateMachine/puzzletea/stats"
	"github.com/FelineStateMachine/puzzletea/ui"

	"charm.land/bubbles/v2/viewport"
	tea "charm.land/bubbletea/v2"
)

func (m model) handleStatsEnter() (model, tea.Cmd) {
	m = m.clearNotice()

	catStats, err := m.store.GetCategoryStats()
	if err != nil {
		return m.setErrorf("Could not load category stats: %v", err), nil
	}
	modeStats, err := m.store.GetModeStats()
	if err != nil {
		return m.setErrorf("Could not load mode stats: %v", err), nil
	}
	streakDates, err := m.store.GetDailyStreakDates()
	if err != nil {
		return m.setErrorf("Could not load streak data: %v", err), nil
	}
	weekliesCompleted, err := m.store.GetCompletedWeeklyGauntlets()
	if err != nil {
		return m.setErrorf("Could not load weekly progress: %v", err), nil
	}
	now := time.Now()
	currentYear, currentWeek := now.ISOWeek()
	thisWeekHighestIndex, err := m.store.GetCurrentWeeklyHighestCompletedIndex(currentYear, currentWeek)
	if err != nil {
		return m.setErrorf("Could not load this week’s progress: %v", err), nil
	}
	currentDaily := false
	rec, err := m.store.GetDailyGame(daily.Name(now))
	if err != nil {
		m = m.setErrorf("Could not check today’s daily puzzle: %v", err)
	} else {
		currentDaily = rec != nil
	}

	weights := stats.WeightsFromDefinitions(registry.Definitions())
	m.stats.cards = stats.BuildCards(weights, catStats, modeStats)
	m.stats.profile = stats.BuildProfileBanner(
		catStats,
		modeStats,
		weights,
		streakDates,
		currentDaily,
		weekliesCompleted,
		thisWeekHighestIndex,
	)

	statsWidth, statsHeight := statsViewportSize(m.width, m.height, m.stats.cards)
	m.stats.viewport = viewport.New(
		viewport.WithWidth(statsWidth),
		viewport.WithHeight(statsHeight),
	)
	m.stats.viewport.SetContent(ui.RenderStatsCardGrid(m.stats.cards, statsWidth))
	m.state = statsView
	m = m.initScreen(statsView)
	return m, nil
}
