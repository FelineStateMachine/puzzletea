package app

import (
	"sort"
	"strconv"
	"time"

	"charm.land/bubbles/v2/table"
	tea "charm.land/bubbletea/v2"
	sessionflow "github.com/FelineStateMachine/puzzletea/session"
	"github.com/FelineStateMachine/puzzletea/store"
	"github.com/FelineStateMachine/puzzletea/ui"
	"github.com/FelineStateMachine/puzzletea/weekly"
)

const weeklyEntryCount = 99

type weeklyRow struct {
	Index    int
	Name     string
	GameType string
	Mode     string
	BonusXP  int
	Status   store.GameStatus
	Record   *store.GameRecord
	Playable bool
	ReadOnly bool
}

func (r weeklyRow) tableRow() table.Row {
	return table.Row{
		formatWeeklyIndex(r.Index),
		r.GameType,
		r.Mode,
		formatWeeklyBonus(r.BonusXP),
		ui.FormatStatus(r.Status),
	}
}

func (m model) enterWeeklyView() (model, tea.Cmd) {
	current := weekly.Current(time.Now())
	m.weekly.cursor = weekly.StartOfWeek(current.Year, current.Week, time.Local)
	m = m.refreshWeeklyBrowser()
	m.state = weeklyView
	m = m.initScreen(weeklyView)
	return m, nil
}

func (m model) refreshWeeklyBrowser() model {
	year, weekNumber := m.selectedWeek()
	games, err := m.store.ListWeeklyGames(year, weekNumber)
	if err != nil {
		m = m.setErrorf("Could not load weekly puzzles: %v", err)
		games = nil
	} else {
		m = m.clearNotice()
	}

	if m.isCurrentWeeklySelection() {
		m.weekly.rows = buildCurrentWeeklyRows(year, weekNumber, games)
	} else {
		m.weekly.rows = buildReviewWeeklyRows(games)
	}

	rows := make([]table.Row, 0, len(m.weekly.rows))
	for _, row := range m.weekly.rows {
		rows = append(rows, row.tableRow())
	}
	m.weekly.table = ui.InitWeeklyTable(rows, m.height)
	return m
}

func (m model) isCurrentWeeklySelection() bool {
	return isCurrentWeeklySelection(m.weekly.cursor)
}

func isCurrentWeeklySelection(cursor time.Time) bool {
	current := weekly.Current(time.Now())
	year, weekNumber := resolvedWeekCursor(cursor).ISOWeek()
	return current.Year == year && current.Week == weekNumber
}

func weeklyPanelTitle(cursor time.Time) string {
	year, weekNumber := resolvedWeekCursor(cursor).ISOWeek()
	return "Week " + formatTwoDigits(weekNumber) + "-" + strconv.Itoa(year)
}

func resolvedWeekCursor(cursor time.Time) time.Time {
	if cursor.IsZero() {
		current := weekly.Current(time.Now())
		return weekly.StartOfWeek(current.Year, current.Week, time.Local)
	}
	return cursor
}

func (m model) selectedWeek() (int, int) {
	return resolvedWeekCursor(m.weekly.cursor).ISOWeek()
}

func (m model) moveWeeklyWeek(delta int) model {
	if m.weekly.cursor.IsZero() {
		current := weekly.Current(time.Now())
		m.weekly.cursor = weekly.StartOfWeek(current.Year, current.Week, time.Local)
	}

	nextCursor := weekly.AddWeeks(m.weekly.cursor, delta)
	current := weekly.Current(time.Now())
	currentCursor := weekly.StartOfWeek(current.Year, current.Week, time.Local)
	if nextCursor.After(currentCursor) {
		nextCursor = currentCursor
	}

	m.weekly.cursor = nextCursor
	return m.refreshWeeklyBrowser()
}

func (m model) handleWeeklyEnter() (model, tea.Cmd) {
	idx := m.weekly.table.Cursor()
	if idx < 0 || idx >= len(m.weekly.rows) {
		return m, nil
	}

	return m.openWeeklyRow(m.weekly.rows[idx])
}

func (m model) openWeeklyRow(row weeklyRow) (model, tea.Cmd) {
	info, ok := weekly.ParseName(row.Name)
	if !ok {
		return m, nil
	}

	options := gameOpenOptions{
		returnState: weeklyView,
	}
	if row.Playable && !row.ReadOnly {
		infoCopy := info
		options.weeklyInfo = &infoCopy
	}

	if row.Record != nil {
		if row.ReadOnly {
			m, _ = m.importAndActivateRecordWithOptions(*row.Record, gameOpenOptions{
				readOnly:    true,
				returnState: weeklyView,
			})
			return m, nil
		}

		var resumed bool
		m, resumed = m.importAndActivateRecordWithOptions(*row.Record, options)
		if resumed {
			if err := sessionflow.ResumeAbandonedDeterministicRecord(m.store, row.Record); err != nil {
				m = m.setErrorf("%v", err)
			}
		}
		return m, nil
	}

	if !row.Playable {
		return m, nil
	}

	spawner, gameType, modeTitle := weekly.Mode(info.Year, info.Week, info.Index)
	if spawner == nil {
		return m.setErrorf("No weekly puzzle is configured for %s", row.Name), nil
	}

	rng := weekly.RNG(info.Year, info.Week, info.Index)
	infoCopy := info
	cmd := newSessionController(&m).startSeededSpawn(spawner, rng, spawnRequest{
		source:      spawnSourceWeekly,
		name:        row.Name,
		gameType:    gameType,
		modeTitle:   modeTitle,
		run:         store.WeeklyRunMetadata(info.Year, info.Week, info.Index),
		returnState: weeklyView,
		exitState:   weeklyView,
		weeklyInfo:  &infoCopy,
	})
	return m, cmd
}

func (m model) advanceSolvedWeekly() (model, tea.Cmd, bool) {
	if m.state != gameView || m.session.game == nil || !m.session.game.IsSolved() || m.session.weeklyAdvance == nil {
		return m, nil, false
	}

	info := *m.session.weeklyAdvance
	m = m.persistCompletionIfSolved()
	m.weekly.cursor = weekly.StartOfWeek(info.Year, info.Week, time.Local)
	m = m.refreshWeeklyBrowser()
	m.state = weeklyView
	m = m.initScreen(weeklyView)
	if len(m.weekly.rows) == 0 || !m.weekly.rows[0].Playable {
		return m, nil, true
	}

	next, cmd := m.openWeeklyRow(m.weekly.rows[0])
	return next, cmd, true
}

func buildCurrentWeeklyRows(year, weekNumber int, games []store.GameRecord) []weeklyRow {
	byIndex := make(map[int]store.GameRecord, len(games))
	completed := make(map[int]store.GameRecord, len(games))

	for _, rec := range games {
		info, ok := weekly.ParseName(rec.Name)
		if !ok || info.Year != year || info.Week != weekNumber {
			continue
		}
		byIndex[info.Index] = rec
		if rec.Status == store.StatusCompleted {
			completed[info.Index] = rec
		}
	}

	prefix := 0
	for index := 1; index <= weeklyEntryCount; index++ {
		if _, ok := completed[index]; !ok {
			break
		}
		prefix = index
	}

	rows := make([]weeklyRow, 0, prefix+1)
	if prefix < weeklyEntryCount {
		nextIndex := prefix + 1
		if rec, ok := byIndex[nextIndex]; ok {
			recCopy := rec
			rows = append(rows, weeklyRow{
				Index:    nextIndex,
				Name:     rec.Name,
				GameType: rec.GameType,
				Mode:     rec.Mode,
				BonusXP:  weekly.BonusXP(nextIndex),
				Status:   rec.Status,
				Record:   &recCopy,
				Playable: true,
			})
		} else {
			_, gameType, modeTitle := weekly.Mode(year, weekNumber, nextIndex)
			rows = append(rows, weeklyRow{
				Index:    nextIndex,
				Name:     weekly.Name(year, weekNumber, nextIndex),
				GameType: gameType,
				Mode:     modeTitle,
				BonusXP:  weekly.BonusXP(nextIndex),
				Status:   store.StatusNew,
				Playable: true,
			})
		}
	}

	for index := prefix; index >= 1; index-- {
		rec := completed[index]
		recCopy := rec
		rows = append(rows, weeklyRow{
			Index:    index,
			Name:     rec.Name,
			GameType: rec.GameType,
			Mode:     rec.Mode,
			BonusXP:  weekly.BonusXP(index),
			Status:   rec.Status,
			Record:   &recCopy,
			ReadOnly: true,
		})
	}

	return rows
}

func buildReviewWeeklyRows(games []store.GameRecord) []weeklyRow {
	rows := make([]weeklyRow, 0, len(games))
	for _, rec := range games {
		info, ok := weekly.ParseName(rec.Name)
		if !ok || rec.Status != store.StatusCompleted {
			continue
		}

		recCopy := rec
		rows = append(rows, weeklyRow{
			Index:    info.Index,
			Name:     rec.Name,
			GameType: rec.GameType,
			Mode:     rec.Mode,
			BonusXP:  weekly.BonusXP(info.Index),
			Status:   rec.Status,
			Record:   &recCopy,
			ReadOnly: true,
		})
	}

	sort.Slice(rows, func(i, j int) bool {
		return rows[i].Index > rows[j].Index
	})
	return rows
}

func formatWeeklyIndex(index int) string {
	return "#" + formatTwoDigits(index)
}

func formatWeeklyBonus(bonus int) string {
	return "+" + strconv.Itoa(bonus)
}

func formatWeeklyMenuIndex(index int) string {
	if index < 10 {
		return " " + strconv.Itoa(index)
	}
	return strconv.Itoa(index)
}

func formatTwoDigits(value int) string {
	if value < 10 {
		return "0" + strconv.Itoa(value)
	}
	return strconv.Itoa(value)
}
