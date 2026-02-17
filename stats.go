package main

import (
	"fmt"
	"math"
	"slices"
	"strings"
	"time"

	"github.com/FelineStateMachine/puzzletea/daily"
	"github.com/FelineStateMachine/puzzletea/game"
	"github.com/FelineStateMachine/puzzletea/store"
	"github.com/FelineStateMachine/puzzletea/ui"

	"charm.land/lipgloss/v2"
)

// modeKey identifies a specific mode within a game category.
type modeKey struct {
	gameType string
	mode     string
}

// modeXP maps each mode to its base XP award per completion.
var modeXP map[modeKey]int

func init() {
	modeXP = make(map[modeKey]int, 64)
	for _, item := range GameCategories {
		cat := item.(game.Category)
		count := len(cat.Modes)
		for i, m := range cat.Modes {
			mode := m.(game.Mode)
			xp := int(math.Round(float64(i) / float64(count) * 10))
			if xp < 1 {
				xp = 1
			}
			modeXP[modeKey{cat.Name, mode.Title()}] = xp
		}
	}
}

// levelFromXP returns the level for the given total XP.
// Uses the inverse of xpForLevel: level = floor((xp / 5) ^ (1/1.6))
func levelFromXP(xp int) int {
	if xp <= 0 {
		return 0
	}
	return int(math.Floor(math.Pow(float64(xp)/5.0, 1.0/1.6)))
}

// xpForLevel returns the total XP required to reach the given level.
// Formula: ceil(5 * level^1.6)
func xpForLevel(level int) int {
	if level <= 0 {
		return 0
	}
	return int(math.Ceil(5.0 * math.Pow(float64(level), 1.6)))
}

// computeCategoryXP calculates total XP for a game category by summing
// XP across all modes. Daily victories earn 2x XP.
func computeCategoryXP(gameType string, modeStats []store.ModeStat) int {
	total := 0
	for _, ms := range modeStats {
		if ms.GameType != gameType {
			continue
		}
		baseXP := modeXP[modeKey{ms.GameType, ms.Mode}]
		if baseXP == 0 {
			baseXP = 1
		}
		normalVictories := ms.Victories - ms.DailyVictories
		total += normalVictories * baseXP
		total += ms.DailyVictories * baseXP * 2
	}
	return total
}

// computeDailyStreak calculates the length of the current daily completion
// streak. The streak is "alive" if the most recent completion is today or
// yesterday (giving the player until end of day to extend it).
func computeDailyStreak(dates []time.Time, now time.Time) int {
	if len(dates) == 0 {
		return 0
	}

	today := truncateToDate(now)
	yesterday := today.AddDate(0, 0, -1)

	most := truncateToDate(dates[0])
	var start time.Time
	switch {
	case most.Equal(today):
		start = today
	case most.Equal(yesterday):
		start = yesterday
	default:
		return 0
	}

	streak := 0
	expected := start
	for _, d := range dates {
		dt := truncateToDate(d)
		if dt.Equal(expected) {
			streak++
			expected = expected.AddDate(0, 0, -1)
		} else if dt.Before(expected) {
			break
		}
	}
	return streak
}

// truncateToDate strips the time component, returning midnight local time.
func truncateToDate(t time.Time) time.Time {
	y, m, d := t.Date()
	return time.Date(y, m, d, 0, 0, 0, 0, t.Location())
}

// --- Stats card ---

type statsCard struct {
	gameType      string
	level         int
	preferredMode string
	victories     int
	attempts      int
	dailyPlayed   int
	currentXP     int
	nextLevelXP   int
}

// buildStatsCards constructs a statsCard for each category the player has
// interacted with. Categories with zero attempts are excluded.
func buildStatsCards(catStats []store.CategoryStat, modeStats []store.ModeStat) []statsCard {
	var cards []statsCard
	for _, cs := range catStats {
		if cs.TotalAttempts == 0 {
			continue
		}
		totalXP := computeCategoryXP(cs.GameType, modeStats)
		lvl := levelFromXP(totalXP)
		nextXP := xpForLevel(lvl + 1)

		preferred := cs.PreferredMode
		if preferred == "" {
			preferred = "\u2014" // em dash
		}

		cards = append(cards, statsCard{
			gameType:      cs.GameType,
			level:         lvl,
			preferredMode: preferred,
			victories:     cs.TotalVictories,
			attempts:      cs.TotalAttempts,
			dailyPlayed:   cs.TimesDaily,
			currentXP:     totalXP,
			nextLevelXP:   nextXP,
		})
	}

	// Sort by level descending, then by XP descending for ties.
	slices.SortStableFunc(cards, func(a, b statsCard) int {
		if a.level != b.level {
			return b.level - a.level
		}
		return b.currentXP - a.currentXP
	})

	return cards
}

// --- Profile banner ---

type profileBanner struct {
	profileLevel int
	dailyStreak  int
	totalDailies int
	currentDaily bool
}

// buildProfileBanner constructs the summary banner shown above the card grid.
func buildProfileBanner(
	catStats []store.CategoryStat,
	modeStats []store.ModeStat,
	streakDates []time.Time,
	s *store.Store,
) profileBanner {
	profileLevel := 0
	totalDailies := 0
	for _, cs := range catStats {
		xp := computeCategoryXP(cs.GameType, modeStats)
		profileLevel += levelFromXP(xp)
		totalDailies += cs.TimesDaily
	}

	now := time.Now()
	streak := computeDailyStreak(streakDates, now)

	// Check if today's daily exists in the DB.
	rng := daily.RNG(now)
	todayName := daily.Name(now, rng)
	rec, _ := s.GetDailyGame(todayName)
	currentDaily := rec != nil

	return profileBanner{
		profileLevel: profileLevel,
		dailyStreak:  streak,
		totalDailies: totalDailies,
		currentDaily: currentDaily,
	}
}

// --- Rendering ---

const (
	cardInnerWidth = 30
	// cardHeight is the rendered height of a single stats card in lines.
	// 7 inner lines (title + 4 stats + blank + XP bar) + 2 border lines = 9.
	cardHeight = 9
)

// Panel chrome: border top (1) + padding top (1) + title (1) + blank (1)
// ... content ...
// blank (1) + footer (1) + padding bottom (1) + border bottom (1) = 8 total.
const panelChrome = 8

// bannerHeight is the rendered banner (4 lines) plus a blank separator line.
const bannerHeight = 5

// statsStaticHeight returns the total lines consumed by non-scrollable parts
// of the stats view: panel chrome + banner (when cards exist).
func statsStaticHeight(cards []statsCard) int {
	if len(cards) == 0 {
		return panelChrome
	}
	return panelChrome + bannerHeight
}

// cardFullWidth is the rendered width of a single card including border + padding.
const cardFullWidth = cardInnerWidth + 4 // 34

// statsContentWidth returns the inner content width for the stats panel.
// In 2-column mode the content is two cards wide; in 1-column mode it is
// one card wide. The result is clamped so it never exceeds the available
// terminal width minus the panel chrome (border + padding = 6 horizontal chars).
func statsContentWidth(termWidth int) int {
	twoCol := cardFullWidth * 2 // 68
	available := termWidth - 6  // panel border (1+1) + padding (2+2)
	if available >= twoCol {
		return twoCol
	}
	return max(cardFullWidth, min(available, cardFullWidth))
}

// statsViewportHeight computes the viewport height for the card grid.
// Shows at most 2.5 card rows and at least 1 full card row, clamped
// to the available terminal height minus the static chrome.
func statsViewportHeight(availableHeight int) int {
	maxRows := cardHeight*2 + cardHeight/2 // ~2.5 cards = 22
	minRows := cardHeight                  // 1 full card = 9
	h := min(availableHeight, maxRows)
	return max(h, minRows)
}

var (
	cardFrame = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(ui.MenuDim).
			Padding(0, 1).
			Width(cardInnerWidth + 4) // +4 for border + padding

	cardTitle = lipgloss.NewStyle().
			Bold(true).
			Foreground(ui.MenuAccent)

	statLabel = lipgloss.NewStyle().
			Foreground(ui.MenuText)

	statValue = lipgloss.NewStyle().
			Foreground(ui.MenuAccent).
			Bold(true)

	xpBarFilled = lipgloss.NewStyle().
			Foreground(ui.MenuAccent)

	xpBarEmpty = lipgloss.NewStyle().
			Foreground(ui.MenuDim)

	xpFraction = lipgloss.NewStyle().
			Foreground(ui.MenuTextDim)

	bannerRule = lipgloss.NewStyle().
			Foreground(ui.MenuDim)

	bannerLabel = lipgloss.NewStyle().
			Foreground(ui.MenuText)

	bannerValue = lipgloss.NewStyle().
			Foreground(ui.MenuAccent).
			Bold(true)
)

func renderCard(c statsCard) string {
	title := cardTitle.Render(
		fmt.Sprintf("%s \u2014 LVL %d", strings.ToUpper(c.gameType), c.level),
	)

	lines := []string{
		title,
		renderStatLine("Preferred Mode:", c.preferredMode),
		renderStatLine("Total Victories:", fmt.Sprintf("%d", c.victories)),
		renderStatLine("Total Attempts:", fmt.Sprintf("%d", c.attempts)),
		renderStatLine("Daily Played:", fmt.Sprintf("%d", c.dailyPlayed)),
		"",
		renderXPBar(c.currentXP, c.level),
	}

	inner := strings.Join(lines, "\n")
	return cardFrame.Render(inner)
}

func renderStatLine(label, value string) string {
	l := statLabel.Render(label)
	v := statValue.Render(value)
	// Right-align value within card width.
	gap := cardInnerWidth - lipgloss.Width(l) - lipgloss.Width(v)
	if gap < 1 {
		gap = 1
	}
	return l + strings.Repeat(" ", gap) + v
}

func renderXPBar(currentXP, level int) string {
	currentLevelXP := xpForLevel(level)
	nextLevelXP := xpForLevel(level + 1)
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
	if filled > barWidth {
		filled = barWidth
	}
	if filled < 0 {
		filled = 0
	}

	bar := xpBarFilled.Render(strings.Repeat("\u2588", filled)) +
		xpBarEmpty.Render(strings.Repeat("\u2591", barWidth-filled))

	frac := xpFraction.Render(fmt.Sprintf(" %d/%d", currentXP, nextLevelXP))

	return statLabel.Render("XP ") + bar + frac
}

func renderBanner(b profileBanner, width int) string {
	if width < 20 {
		width = 40
	}

	ruleStr := bannerRule.Render(
		"\u2500\u2500\u2500 Profile " + strings.Repeat("\u2500", max(width-12, 1)),
	)

	check := bannerValue.Render("\u2714")
	if !b.currentDaily {
		check = lipgloss.NewStyle().Foreground(ui.MenuTextDim).Render("\u2014")
	}

	streakStr := fmt.Sprintf("%d days", b.dailyStreak)
	if b.dailyStreak == 1 {
		streakStr = "1 day"
	}

	col1 := bannerLabel.Render("Profile Level: ") + bannerValue.Render(fmt.Sprintf("%d", b.profileLevel))
	col2 := bannerLabel.Render("Daily Streak: ") + bannerValue.Render(streakStr)
	col3 := bannerLabel.Render("Total Dailies: ") + bannerValue.Render(fmt.Sprintf("%d", b.totalDailies))
	col4 := bannerLabel.Render("Current daily: ") + check

	// Two-column layout: left side + gap + right side.
	gap := width - lipgloss.Width(col1) - lipgloss.Width(col2)
	if gap < 2 {
		gap = 2
	}
	row1 := col1 + strings.Repeat(" ", gap) + col2

	gap2 := width - lipgloss.Width(col3) - lipgloss.Width(col4)
	if gap2 < 2 {
		gap2 = 2
	}
	row2 := col3 + strings.Repeat(" ", gap2) + col4

	bottomRule := bannerRule.Render(strings.Repeat("\u2500", max(width, 1)))

	return strings.Join([]string{ruleStr, row1, row2, bottomRule}, "\n")
}

// renderStatsCardGrid renders just the card grid (no banner). This is the
// content placed inside the scrollable viewport.
func renderStatsCardGrid(cards []statsCard, width int) string {
	if len(cards) == 0 {
		return lipgloss.NewStyle().
			Foreground(ui.MenuTextDim).
			Render("No stats yet \u2014 play some puzzles!")
	}

	if width >= cardFullWidth*2 {
		return renderCardColumns(cards, 2)
	}
	return renderCardColumns(cards, 1)
}

// renderStatsView renders the full stats view (banner + card grid) as a single
// string. Used only by tests and the empty-state path.
func renderStatsView(banner profileBanner, cards []statsCard, width int) string {
	if len(cards) == 0 {
		return lipgloss.NewStyle().
			Foreground(ui.MenuTextDim).
			Render("No stats yet \u2014 play some puzzles!")
	}

	bannerWidth := width
	if bannerWidth > 70 {
		bannerWidth = 70
	}
	bannerStr := renderBanner(banner, bannerWidth)
	cardGrid := renderStatsCardGrid(cards, width)

	return lipgloss.JoinVertical(lipgloss.Left,
		bannerStr,
		"",
		cardGrid,
	)
}

func renderCardColumns(cards []statsCard, cols int) string {
	if cols < 1 {
		cols = 1
	}

	columns := make([][]string, cols)
	for i, c := range cards {
		col := i % cols
		columns[col] = append(columns[col], renderCard(c))
	}

	rendered := make([]string, cols)
	for i, col := range columns {
		rendered[i] = lipgloss.JoinVertical(lipgloss.Left, col...)
	}

	return lipgloss.JoinHorizontal(lipgloss.Top, rendered...)
}
