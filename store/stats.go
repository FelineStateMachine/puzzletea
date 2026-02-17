package store

import (
	"database/sql"
	"fmt"
	"time"
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
	return stats, rows.Err()
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
