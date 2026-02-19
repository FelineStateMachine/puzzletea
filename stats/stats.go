package stats

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

	"charm.land/bubbles/v2/list"
	"charm.land/lipgloss/v2"
)

// ModeKey identifies a specific mode within a game category.
type ModeKey struct {
	GameType string
	Mode     string
}

// ModeXP maps each mode to its base XP award per completion.
var ModeXP map[ModeKey]int

// InitModeXP builds the ModeXP map from the provided game categories.
// Must be called once at startup before any stats operations.
func InitModeXP(categories []list.Item) {
	ModeXP = make(map[ModeKey]int, 64)
	for _, item := range categories {
		cat := item.(game.Category)
		count := len(cat.Modes)
		for i, m := range cat.Modes {
			mode := m.(game.Mode)
			xp := int(math.Round(float64(i) / float64(count) * 10))
			if xp < 1 {
				xp = 1
			}
			ModeXP[ModeKey{cat.Name, mode.Title()}] = xp
		}
	}
}

// LevelFromXP returns the level for the given total XP.
// Uses the inverse of XPForLevel: level = floor((xp / 5) ^ (1/1.6))
func LevelFromXP(xp int) int {
	if xp <= 0 {
		return 0
	}
	return int(math.Floor(math.Pow(float64(xp)/5.0, 1.0/1.6)))
}

// XPForLevel returns the total XP required to reach the given level.
// Formula: ceil(5 * level^1.6)
func XPForLevel(level int) int {
	if level <= 0 {
		return 0
	}
	return int(math.Ceil(5.0 * math.Pow(float64(level), 1.6)))
}

// ComputeCategoryXP calculates total XP for a game category by summing
// XP across all modes. Daily victories earn 2x XP.
func ComputeCategoryXP(gameType string, modeStats []store.ModeStat) int {
	total := 0
	for _, ms := range modeStats {
		if ms.GameType != gameType {
			continue
		}
		baseXP := ModeXP[ModeKey{ms.GameType, ms.Mode}]
		if baseXP == 0 {
			baseXP = 1
		}
		normalVictories := ms.Victories - ms.DailyVictories
		total += normalVictories * baseXP
		total += ms.DailyVictories * baseXP * 2
	}
	return total
}

// ComputeDailyStreak calculates the length of the current daily completion
// streak. The streak is "alive" if the most recent completion is today or
// yesterday (giving the player until end of day to extend it).
func ComputeDailyStreak(dates []time.Time, now time.Time) int {
	if len(dates) == 0 {
		return 0
	}

	today := TruncateToDate(now)
	yesterday := today.AddDate(0, 0, -1)

	most := TruncateToDate(dates[0])
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
		dt := TruncateToDate(d)
		if dt.Equal(expected) {
			streak++
			expected = expected.AddDate(0, 0, -1)
		} else if dt.Before(expected) {
			break
		}
	}
	return streak
}

// TruncateToDate strips the time component, returning midnight local time.
func TruncateToDate(t time.Time) time.Time {
	y, m, d := t.Date()
	return time.Date(y, m, d, 0, 0, 0, 0, t.Location())
}

// --- Stats card ---

// Card holds the computed stats for a single game category.
type Card struct {
	GameType      string
	Level         int
	PreferredMode string
	Victories     int
	Attempts      int
	DailyPlayed   int
	CurrentXP     int
	NextLevelXP   int
}

// BuildCards constructs a Card for each category the player has
// interacted with. Categories with zero attempts are excluded.
func BuildCards(catStats []store.CategoryStat, modeStats []store.ModeStat) []Card {
	var cards []Card
	for _, cs := range catStats {
		if cs.TotalAttempts == 0 {
			continue
		}
		totalXP := ComputeCategoryXP(cs.GameType, modeStats)
		lvl := LevelFromXP(totalXP)
		nextXP := XPForLevel(lvl + 1)

		preferred := cs.PreferredMode
		if preferred == "" {
			preferred = "\u2014" // em dash
		}

		cards = append(cards, Card{
			GameType:      cs.GameType,
			Level:         lvl,
			PreferredMode: preferred,
			Victories:     cs.TotalVictories,
			Attempts:      cs.TotalAttempts,
			DailyPlayed:   cs.TimesDaily,
			CurrentXP:     totalXP,
			NextLevelXP:   nextXP,
		})
	}

	// Sort by level descending, then by XP descending for ties.
	slices.SortStableFunc(cards, func(a, b Card) int {
		if a.Level != b.Level {
			return b.Level - a.Level
		}
		return b.CurrentXP - a.CurrentXP
	})

	return cards
}

// --- Profile banner ---

// ProfileBanner holds the summary data shown above the card grid.
type ProfileBanner struct {
	ProfileLevel int
	DailyStreak  int
	TotalDailies int
	CurrentDaily bool
}

// BuildProfileBanner constructs the summary banner shown above the card grid.
func BuildProfileBanner(
	catStats []store.CategoryStat,
	modeStats []store.ModeStat,
	streakDates []time.Time,
	s *store.Store,
) ProfileBanner {
	profileLevel := 0
	totalDailies := 0
	for _, cs := range catStats {
		xp := ComputeCategoryXP(cs.GameType, modeStats)
		profileLevel += LevelFromXP(xp)
		totalDailies += cs.TimesDaily
	}

	now := time.Now()
	streak := ComputeDailyStreak(streakDates, now)

	// Check if today's daily exists in the DB.
	todayName := daily.Name(now)
	rec, _ := s.GetDailyGame(todayName)
	currentDaily := rec != nil

	return ProfileBanner{
		ProfileLevel: profileLevel,
		DailyStreak:  streak,
		TotalDailies: totalDailies,
		CurrentDaily: currentDaily,
	}
}

// --- Rendering ---

const (
	CardInnerWidth = 30
	// CardHeight is the rendered height of a single stats card in lines.
	// 7 inner lines (title + 4 stats + blank + XP bar) + 2 border lines = 9.
	CardHeight = 9
)

// Panel chrome: border top (1) + padding top (1) + title (1) + blank (1)
// ... content ...
// blank (1) + footer (1) + padding bottom (1) + border bottom (1) = 8 total.
const panelChrome = 8

// bannerHeight is the rendered banner (4 lines) plus a blank separator line.
const bannerHeight = 5

// StaticHeight returns the total lines consumed by non-scrollable parts
// of the stats view: panel chrome + banner (when cards exist).
func StaticHeight(cards []Card) int {
	if len(cards) == 0 {
		return panelChrome
	}
	return panelChrome + bannerHeight
}

// CardFullWidth is the rendered width of a single card including border + padding.
const CardFullWidth = CardInnerWidth + 4 // 34

// ContentWidth returns the inner content width for the stats panel.
// In 2-column mode the content is two cards wide; in 1-column mode it is
// one card wide. The result is clamped so it never exceeds the available
// terminal width minus the panel chrome (border + padding = 6 horizontal chars).
func ContentWidth(termWidth int) int {
	twoCol := CardFullWidth * 2 // 68
	available := termWidth - 6  // panel border (1+1) + padding (2+2)
	if available >= twoCol {
		return twoCol
	}
	return max(CardFullWidth, min(available, CardFullWidth))
}

// ViewportHeight computes the viewport height for the card grid.
// Shows at most 2.5 card rows and at least 1 full card row, clamped
// to the available terminal height minus the static chrome.
func ViewportHeight(availableHeight int) int {
	maxRows := CardHeight*2 + CardHeight/2 // ~2.5 cards = 22
	minRows := CardHeight                  // 1 full card = 9
	h := min(availableHeight, maxRows)
	return max(h, minRows)
}

var (
	cardFrame = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(ui.MenuDim).
			Padding(0, 1).
			Width(CardInnerWidth + 4) // +4 for border + padding

	cardTitle = lipgloss.NewStyle().
			Bold(true).
			Foreground(ui.MenuAccent)

	statLabel = lipgloss.NewStyle().
			Foreground(ui.MenuTextDim)

	statValue = lipgloss.NewStyle().
			Foreground(ui.MenuAccent).
			Bold(true)

	xpBarFilled = lipgloss.NewStyle().
			Foreground(ui.MenuAccent)

	xpBarEmpty = lipgloss.NewStyle().
			Foreground(ui.MenuDim)

	xpFraction = lipgloss.NewStyle().
			Foreground(ui.MenuText)

	bannerRule = lipgloss.NewStyle().
			Foreground(ui.MenuDim)

	bannerLabel = lipgloss.NewStyle().
			Foreground(ui.MenuText)

	bannerValue = lipgloss.NewStyle().
			Foreground(ui.MenuAccent).
			Bold(true)
)

// RenderCard renders a single stats card.
func RenderCard(c Card) string {
	title := cardTitle.Render(
		fmt.Sprintf("%s \u2014 LVL %d", strings.ToUpper(c.GameType), c.Level),
	)

	lines := []string{
		title,
		renderStatLine("Preferred Mode:", c.PreferredMode),
		renderStatLine("Total Victories:", fmt.Sprintf("%d", c.Victories)),
		renderStatLine("Total Attempts:", fmt.Sprintf("%d", c.Attempts)),
		renderStatLine("Daily Played:", fmt.Sprintf("%d", c.DailyPlayed)),
		"",
		renderXPBar(c.CurrentXP, c.Level),
	}

	inner := strings.Join(lines, "\n")
	return cardFrame.Render(inner)
}

func renderStatLine(label, value string) string {
	l := statLabel.Render(label)
	v := statValue.Render(value)
	// Right-align value within card width.
	gap := CardInnerWidth - lipgloss.Width(l) - lipgloss.Width(v)
	if gap < 1 {
		gap = 1
	}
	return l + strings.Repeat(" ", gap) + v
}

func renderXPBar(currentXP, level int) string {
	currentLevelXP := XPForLevel(level)
	nextLevelXP := XPForLevel(level + 1)
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

	return statLabel.Foreground(ui.MenuText).Render("XP ") + bar + frac
}

// RenderBanner renders the profile summary banner.
func RenderBanner(b ProfileBanner, width int) string {
	if width < 20 {
		width = 40
	}

	ruleStr := bannerRule.Render(
		"\u2500\u2500\u2500 Profile " + strings.Repeat("\u2500", max(width-12, 1)),
	)

	check := bannerValue.Render("\u2714")
	if !b.CurrentDaily {
		check = lipgloss.NewStyle().Foreground(ui.MenuTextDim).Render("\u2014")
	}

	streakStr := fmt.Sprintf("%d days", b.DailyStreak)
	if b.DailyStreak == 1 {
		streakStr = "1 day"
	}

	col1 := bannerLabel.Render("Profile Level: ") + bannerValue.Render(fmt.Sprintf("%d", b.ProfileLevel))
	col2 := bannerLabel.Render("Daily Streak: ") + bannerValue.Render(streakStr)
	col3 := bannerLabel.Render("Total Dailies: ") + bannerValue.Render(fmt.Sprintf("%d", b.TotalDailies))
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

// RenderCardGrid renders just the card grid (no banner). This is the
// content placed inside the scrollable viewport.
func RenderCardGrid(cards []Card, width int) string {
	if len(cards) == 0 {
		return lipgloss.NewStyle().
			Foreground(ui.MenuTextDim).
			Render("No stats yet \u2014 play some puzzles!")
	}

	if width >= CardFullWidth*2 {
		return renderCardColumns(cards, 2)
	}
	return renderCardColumns(cards, 1)
}

// RenderView renders the full stats view (banner + card grid) as a single
// string. Used only by tests and the empty-state path.
func RenderView(banner ProfileBanner, cards []Card, width int) string {
	if len(cards) == 0 {
		return lipgloss.NewStyle().
			Foreground(ui.MenuTextDim).
			Render("No stats yet \u2014 play some puzzles!")
	}

	bannerWidth := width
	if bannerWidth > 70 {
		bannerWidth = 70
	}
	bannerStr := RenderBanner(banner, bannerWidth)
	cardGrid := RenderCardGrid(cards, width)

	return lipgloss.JoinVertical(lipgloss.Left,
		bannerStr,
		"",
		cardGrid,
	)
}

func renderCardColumns(cards []Card, cols int) string {
	if cols < 1 {
		cols = 1
	}

	columns := make([][]string, cols)
	for i, c := range cards {
		col := i % cols
		columns[col] = append(columns[col], RenderCard(c))
	}

	rendered := make([]string, cols)
	for i, col := range columns {
		rendered[i] = lipgloss.JoinVertical(lipgloss.Left, col...)
	}

	return lipgloss.JoinHorizontal(lipgloss.Top, rendered...)
}
