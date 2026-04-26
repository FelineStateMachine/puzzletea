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
	DifficultyElo  *int
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
		`SELECT game_type, mode, victories, daily_victories, difficulty_elo
		 FROM mode_stats`,
	)
	if err != nil {
		return nil, fmt.Errorf("querying mode_stats: %w", err)
	}
	defer rows.Close()

	var stats []ModeStat
	for rows.Next() {
		var ms ModeStat
		var difficultyElo sql.NullInt64
		if err := rows.Scan(
			&ms.GameType, &ms.Mode, &ms.Victories, &ms.DailyVictories,
			&difficultyElo,
		); err != nil {
			return nil, fmt.Errorf("scanning mode_stats row: %w", err)
		}
		if difficultyElo.Valid {
			v := int(difficultyElo.Int64)
			ms.DifficultyElo = &v
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
		`SELECT DISTINCT run_date
		 FROM games
		 WHERE run_kind = ?
		   AND status = 'completed'
		   AND run_date IS NOT NULL
		 ORDER BY run_date DESC`,
		string(RunKindDaily),
	)
	if err != nil {
		return nil, fmt.Errorf("querying daily streak dates: %w", err)
	}
	defer rows.Close()

	var dates []time.Time
	for rows.Next() {
		var dateValue sql.NullTime
		if err := rows.Scan(&dateValue); err != nil {
			return nil, fmt.Errorf("scanning daily streak date: %w", err)
		}
		if !dateValue.Valid {
			continue
		}
		t := dateValue.Time.In(time.Local)
		dates = append(dates, time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.Local))
	}
	return dates, rows.Err()
}

// GetCompletedWeeklyGauntlets returns the number of ISO weeks whose highest
// completed weekly index is 99.
func (s *Store) GetCompletedWeeklyGauntlets() (int, error) {
	rows, err := s.db.Query(
		`SELECT week_year, week_number, MAX(week_index)
		 FROM games
		 WHERE status = 'completed'
		   AND run_kind = ?
		   AND week_year IS NOT NULL
		   AND week_number IS NOT NULL
		 GROUP BY week_year, week_number`,
		string(RunKindWeekly),
	)
	if err != nil {
		return 0, fmt.Errorf("querying completed weekly gauntlets: %w", err)
	}
	defer rows.Close()
	completed := 0
	for rows.Next() {
		var year, week, highest int
		if err := rows.Scan(&year, &week, &highest); err != nil {
			return 0, fmt.Errorf("scanning weekly gauntlet row: %w", err)
		}
		if highest >= 99 {
			completed++
		}
	}
	return completed, rows.Err()
}

// GetCurrentWeeklyHighestCompletedIndex returns the highest completed weekly
// index for the provided ISO week-year.
func (s *Store) GetCurrentWeeklyHighestCompletedIndex(year, week int) (int, error) {
	var highest sql.NullInt64
	err := s.db.QueryRow(
		`SELECT MAX(week_index)
		 FROM games
		 WHERE status = 'completed'
		   AND run_kind = ?
		   AND week_year = ?
		   AND week_number = ?`,
		string(RunKindWeekly), year, week,
	).Scan(&highest)
	if err != nil {
		return 0, fmt.Errorf("querying current weekly highest index: %w", err)
	}
	if !highest.Valid {
		return 0, nil
	}
	return int(highest.Int64), nil
}

func (s *Store) applyWeeklyBonusXP(stats []ModeStat) ([]ModeStat, error) {
	rows, err := s.db.Query(
		`SELECT week_index, game_type, mode
		 FROM games
		 WHERE status = 'completed'
		   AND run_kind = ?`,
		string(RunKindWeekly),
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
		var weekIndex int
		var gameType string
		var mode string
		if err := rows.Scan(&weekIndex, &gameType, &mode); err != nil {
			return nil, fmt.Errorf("scanning completed weekly game: %w", err)
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
		if weekIndex >= 10 {
			stats[i].WeeklyBonusXP += weekIndex / 10
		}
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("reading completed weekly games: %w", err)
	}
	return stats, nil
}
