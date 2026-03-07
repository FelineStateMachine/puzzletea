package store

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/FelineStateMachine/puzzletea/weekly"
)

// CategoryStat holds aggregate statistics for a single game category.
type CategoryStat struct {
	GameType       string
	TotalAttempts  int
	TotalVictories int
	TotalAbandoned int
	TimesDaily     int
	DailyVictories int
	PreferredMode  string
}

// ModeStat holds per-mode victory counts for XP computation.
type ModeStat struct {
	GameType       string
	Mode           string
	Victories      int
	DailyVictories int
	WeeklyBonusXP  int
}

// GetCategoryStats returns aggregate stats per game category from the
// category_stats view.
func (s *Store) GetCategoryStats() ([]CategoryStat, error) {
	rows, err := s.db.Query(
		`SELECT game_type, total_attempts, total_victories, total_abandoned,
		        times_daily, daily_victories, preferred_mode
		 FROM category_stats`,
	)
	if err != nil {
		return nil, fmt.Errorf("querying category_stats: %w", err)
	}
	defer rows.Close()

	var stats []CategoryStat
	for rows.Next() {
		var cs CategoryStat
		var preferredMode sql.NullString
		if err := rows.Scan(
			&cs.GameType, &cs.TotalAttempts, &cs.TotalVictories,
			&cs.TotalAbandoned, &cs.TimesDaily, &cs.DailyVictories,
			&preferredMode,
		); err != nil {
			return nil, fmt.Errorf("scanning category_stats row: %w", err)
		}
		if preferredMode.Valid {
			cs.PreferredMode = preferredMode.String
		}
		stats = append(stats, cs)
	}
	return stats, rows.Err()
}

// GetModeStats returns per-mode victory counts from the mode_stats view.
func (s *Store) GetModeStats() ([]ModeStat, error) {
	rows, err := s.db.Query(
		`SELECT game_type, mode, victories, daily_victories
		 FROM mode_stats`,
	)
	if err != nil {
		return nil, fmt.Errorf("querying mode_stats: %w", err)
	}
	defer rows.Close()

	var stats []ModeStat
	for rows.Next() {
		var ms ModeStat
		if err := rows.Scan(
			&ms.GameType, &ms.Mode, &ms.Victories, &ms.DailyVictories,
		); err != nil {
			return nil, fmt.Errorf("scanning mode_stats row: %w", err)
		}
		stats = append(stats, ms)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	stats, err = s.applyWeeklyBonusXP(stats)
	if err != nil {
		return nil, err
	}
	return stats, nil
}

// GetDailyStreakDates returns distinct completion dates for daily games,
// ordered most recent first. Dates are in local time with time set to midnight.
func (s *Store) GetDailyStreakDates() ([]time.Time, error) {
	rows, err := s.db.Query(
		`SELECT DISTINCT DATE(REPLACE(SUBSTR(completed_at, 1, 19), 'T', ' '), 'localtime') AS d
		 FROM games
		 WHERE name LIKE 'Daily %'
		   AND status = 'completed'
		   AND completed_at IS NOT NULL
		 ORDER BY d DESC`,
	)
	if err != nil {
		return nil, fmt.Errorf("querying daily streak dates: %w", err)
	}
	defer rows.Close()

	var dates []time.Time
	for rows.Next() {
		var dateStr string
		if err := rows.Scan(&dateStr); err != nil {
			return nil, fmt.Errorf("scanning daily streak date: %w", err)
		}
		t, err := time.ParseInLocation("2006-01-02", dateStr, time.Local)
		if err != nil {
			return nil, fmt.Errorf("parsing date %q: %w", dateStr, err)
		}
		dates = append(dates, t)
	}
	return dates, rows.Err()
}

// GetCompletedWeeklyGauntlets returns the number of ISO weeks whose highest
// completed weekly index is 99.
func (s *Store) GetCompletedWeeklyGauntlets() (int, error) {
	highestByWeek, err := s.highestCompletedWeeklyIndexByWeek()
	if err != nil {
		return 0, err
	}

	completed := 0
	for _, highest := range highestByWeek {
		if highest >= 99 {
			completed++
		}
	}
	return completed, nil
}

// GetCurrentWeeklyHighestCompletedIndex returns the highest completed weekly
// index for the provided ISO week-year.
func (s *Store) GetCurrentWeeklyHighestCompletedIndex(year, week int) (int, error) {
	highestByWeek, err := s.highestCompletedWeeklyIndexByWeek()
	if err != nil {
		return 0, err
	}
	return highestByWeek[weeklyKey{year: year, week: week}], nil
}

type weeklyKey struct {
	year int
	week int
}

func (s *Store) highestCompletedWeeklyIndexByWeek() (map[weeklyKey]int, error) {
	rows, err := s.db.Query(
		`SELECT name
		 FROM games
		 WHERE status = 'completed'
		   AND name LIKE 'Week %'`,
	)
	if err != nil {
		return nil, fmt.Errorf("querying completed weekly names: %w", err)
	}
	defer rows.Close()

	highestByWeek := make(map[weeklyKey]int)
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, fmt.Errorf("scanning completed weekly name: %w", err)
		}

		info, ok := weekly.ParseName(name)
		if !ok {
			continue
		}

		key := weeklyKey{year: info.Year, week: info.Week}
		if info.Index > highestByWeek[key] {
			highestByWeek[key] = info.Index
		}
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("reading completed weekly names: %w", err)
	}
	return highestByWeek, nil
}

func (s *Store) applyWeeklyBonusXP(stats []ModeStat) ([]ModeStat, error) {
	rows, err := s.db.Query(
		`SELECT name, game_type, mode
		 FROM games
		 WHERE status = 'completed'
		   AND name LIKE 'Week %'`,
	)
	if err != nil {
		return nil, fmt.Errorf("querying completed weekly games: %w", err)
	}
	defer rows.Close()

	type key struct {
		gameType string
		mode     string
	}

	indexByKey := make(map[key]int, len(stats))
	for i := range stats {
		indexByKey[key{gameType: stats[i].GameType, mode: stats[i].Mode}] = i
	}

	for rows.Next() {
		var name string
		var gameType string
		var mode string
		if err := rows.Scan(&name, &gameType, &mode); err != nil {
			return nil, fmt.Errorf("scanning completed weekly game: %w", err)
		}

		info, ok := weekly.ParseName(name)
		if !ok {
			continue
		}

		k := key{gameType: gameType, mode: mode}
		i, ok := indexByKey[k]
		if !ok {
			stats = append(stats, ModeStat{
				GameType:  gameType,
				Mode:      mode,
				Victories: 1,
			})
			i = len(stats) - 1
			indexByKey[k] = i
		}
		stats[i].WeeklyBonusXP += weekly.BonusXP(info.Index)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("reading completed weekly games: %w", err)
	}
	return stats, nil
}
