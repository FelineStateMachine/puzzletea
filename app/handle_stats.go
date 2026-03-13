package app

import (
	"log"
	"time"

	"github.com/FelineStateMachine/puzzletea/daily"
	"github.com/FelineStateMachine/puzzletea/registry"
	"github.com/FelineStateMachine/puzzletea/stats"
	"github.com/FelineStateMachine/puzzletea/ui"

	"charm.land/bubbles/v2/viewport"
	tea "charm.land/bubbletea/v2"
)

func (m model) handleStatsEnter() (tea.Model, tea.Cmd) {
	catStats, err := m.store.GetCategoryStats()
	if err != nil {
		log.Printf("failed to get category stats: %v", err)
		return m, nil
	}
	modeStats, err := m.store.GetModeStats()
	if err != nil {
		log.Printf("failed to get mode stats: %v", err)
		return m, nil
	}
	streakDates, err := m.store.GetDailyStreakDates()
	if err != nil {
		log.Printf("failed to get daily streak dates: %v", err)
		return m, nil
	}
	weekliesCompleted, err := m.store.GetCompletedWeeklyGauntlets()
	if err != nil {
		log.Printf("failed to get weekly gauntlet completions: %v", err)
		return m, nil
	}
	now := time.Now()
	currentYear, currentWeek := now.ISOWeek()
	thisWeekHighestIndex, err := m.store.GetCurrentWeeklyHighestCompletedIndex(currentYear, currentWeek)
	if err != nil {
		log.Printf("failed to get current weekly progress: %v", err)
		return m, nil
	}
	currentDaily := false
	rec, err := m.store.GetDailyGame(daily.Name(now))
	if err != nil {
		log.Printf("failed to check current daily: %v", err)
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
	return m, nil
}
