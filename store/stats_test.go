package store

import (
	"testing"
	"time"

	"github.com/FelineStateMachine/puzzletea/weekly"
)

// --- GetCategoryStats (P1) ---

func TestGetCategoryStats(t *testing.T) {
	t.Run("empty database", func(t *testing.T) {
		s := openTestStore(t)
		stats, err := s.GetCategoryStats()
		if err != nil {
			t.Fatal(err)
		}
		if len(stats) != 0 {
			t.Errorf("expected 0 stats, got %d", len(stats))
		}
	})

	t.Run("counts by category", func(t *testing.T) {
		s := openTestStore(t)
		games := []*GameRecord{
			{Name: "a", GameType: "Sudoku", Mode: "Easy", InitialState: "{}", SaveState: "{}", Status: StatusNew},
			{Name: "b", GameType: "Sudoku", Mode: "Medium", InitialState: "{}", SaveState: "{}", Status: StatusNew},
			{Name: "c", GameType: "Nonogram", Mode: "Mini", InitialState: "{}", SaveState: "{}", Status: StatusNew},
		}
		for _, g := range games {
			if err := s.CreateGame(g); err != nil {
				t.Fatal(err)
			}
		}
		// Complete one Sudoku.
		if err := s.UpdateStatus(games[0].ID, StatusCompleted); err != nil {
			t.Fatal(err)
		}
		// Abandon one Sudoku.
		if err := s.UpdateStatus(games[1].ID, StatusAbandoned); err != nil {
			t.Fatal(err)
		}

		stats, err := s.GetCategoryStats()
		if err != nil {
			t.Fatal(err)
		}

		byType := make(map[string]CategoryStat)
		for _, cs := range stats {
			byType[cs.GameType] = cs
		}

		sudoku := byType["Sudoku"]
		if sudoku.TotalAttempts != 2 {
			t.Errorf("Sudoku TotalAttempts = %d, want 2", sudoku.TotalAttempts)
		}
		if sudoku.TotalVictories != 1 {
			t.Errorf("Sudoku TotalVictories = %d, want 1", sudoku.TotalVictories)
		}
		if sudoku.TotalAbandoned != 1 {
			t.Errorf("Sudoku TotalAbandoned = %d, want 1", sudoku.TotalAbandoned)
		}
		if sudoku.PreferredMode != "Easy" {
			t.Errorf("Sudoku PreferredMode = %q, want %q", sudoku.PreferredMode, "Easy")
		}

		nonogram := byType["Nonogram"]
		if nonogram.TotalAttempts != 1 {
			t.Errorf("Nonogram TotalAttempts = %d, want 1", nonogram.TotalAttempts)
		}
		if nonogram.TotalVictories != 0 {
			t.Errorf("Nonogram TotalVictories = %d, want 0", nonogram.TotalVictories)
		}
		if nonogram.PreferredMode != "" {
			t.Errorf("Nonogram PreferredMode = %q, want empty", nonogram.PreferredMode)
		}
	})

	t.Run("daily games counted", func(t *testing.T) {
		s := openTestStore(t)
		games := []*GameRecord{
			newDailyTestRecord("Daily Feb 16 26 - amber-fox", time.Date(2026, time.February, 16, 12, 0, 0, 0, time.Local)),
			newDailyTestRecord("Daily Feb 15 26 - blue-cat", time.Date(2026, time.February, 15, 12, 0, 0, 0, time.Local)),
			{Name: "regular-game", GameType: "Sudoku", Mode: "Easy", InitialState: "{}", SaveState: "{}", Status: StatusNew},
		}
		games[0].GameType, games[0].Mode = "Sudoku", "Easy"
		games[0].InitialState, games[0].SaveState, games[0].Status = "{}", "{}", StatusNew
		games[1].GameType, games[1].Mode = "Sudoku", "Medium"
		games[1].InitialState, games[1].SaveState, games[1].Status = "{}", "{}", StatusNew
		for _, g := range games {
			if err := s.CreateGame(g); err != nil {
				t.Fatal(err)
			}
		}
		if err := s.UpdateStatus(games[0].ID, StatusCompleted); err != nil {
			t.Fatal(err)
		}

		stats, err := s.GetCategoryStats()
		if err != nil {
			t.Fatal(err)
		}
		if len(stats) != 1 {
			t.Fatalf("expected 1 category, got %d", len(stats))
		}
		cs := stats[0]
		if cs.TimesDaily != 2 {
			t.Errorf("TimesDaily = %d, want 2", cs.TimesDaily)
		}
		if cs.DailyVictories != 1 {
			t.Errorf("DailyVictories = %d, want 1", cs.DailyVictories)
		}
	})
}

// --- GetModeStats (P1) ---

func TestGetModeStats(t *testing.T) {
	t.Run("empty database", func(t *testing.T) {
		s := openTestStore(t)
		stats, err := s.GetModeStats()
		if err != nil {
			t.Fatal(err)
		}
		if len(stats) != 0 {
			t.Errorf("expected 0 stats, got %d", len(stats))
		}
	})

	t.Run("victory counts per mode", func(t *testing.T) {
		s := openTestStore(t)
		games := []*GameRecord{
			{Name: "a", GameType: "Sudoku", Mode: "Easy", InitialState: "{}", SaveState: "{}", Status: StatusNew},
			{Name: "b", GameType: "Sudoku", Mode: "Easy", InitialState: "{}", SaveState: "{}", Status: StatusNew},
			{Name: "c", GameType: "Sudoku", Mode: "Medium", InitialState: "{}", SaveState: "{}", Status: StatusNew},
			newDailyTestRecord("Daily Feb 16 26 - x", time.Date(2026, time.February, 16, 12, 0, 0, 0, time.Local)),
		}
		games[3].GameType, games[3].Mode = "Sudoku", "Easy"
		games[3].InitialState, games[3].SaveState, games[3].Status = "{}", "{}", StatusNew
		for _, g := range games {
			if err := s.CreateGame(g); err != nil {
				t.Fatal(err)
			}
		}
		// Complete two Easy (one daily) and one Medium.
		for _, id := range []int64{games[0].ID, games[2].ID, games[3].ID} {
			if err := s.UpdateStatus(id, StatusCompleted); err != nil {
				t.Fatal(err)
			}
		}

		stats, err := s.GetModeStats()
		if err != nil {
			t.Fatal(err)
		}

		type key struct{ gt, mode string }
		byKey := make(map[key]ModeStat)
		for _, ms := range stats {
			byKey[key{ms.GameType, ms.Mode}] = ms
		}

		easy := byKey[key{"Sudoku", "Easy"}]
		if easy.Victories != 2 {
			t.Errorf("Easy Victories = %d, want 2", easy.Victories)
		}
		if easy.DailyVictories != 1 {
			t.Errorf("Easy DailyVictories = %d, want 1", easy.DailyVictories)
		}

		medium := byKey[key{"Sudoku", "Medium"}]
		if medium.Victories != 1 {
			t.Errorf("Medium Victories = %d, want 1", medium.Victories)
		}
		if medium.DailyVictories != 0 {
			t.Errorf("Medium DailyVictories = %d, want 0", medium.DailyVictories)
		}
	})

	t.Run("weekly bonus xp is aggregated by parsed name", func(t *testing.T) {
		s := openTestStore(t)
		games := []*GameRecord{
			newWeeklyTestRecord(weekly.Name(2026, 1, 10), 2026, 1, 10),
			newWeeklyTestRecord(weekly.Name(2026, 1, 25), 2026, 1, 25),
		}
		for _, g := range games {
			g.GameType = "Sudoku"
			g.Mode = "Easy"
			g.InitialState = "{}"
			g.SaveState = "{}"
			g.Status = StatusNew
		}
		for _, g := range games {
			if err := s.CreateGame(g); err != nil {
				t.Fatal(err)
			}
			if err := s.UpdateStatus(g.ID, StatusCompleted); err != nil {
				t.Fatal(err)
			}
		}

		stats, err := s.GetModeStats()
		if err != nil {
			t.Fatal(err)
		}

		found := false
		for _, ms := range stats {
			if ms.GameType != "Sudoku" || ms.Mode != "Easy" {
				continue
			}
			found = true
			want := weekly.BonusXP(10) + weekly.BonusXP(25)
			if ms.WeeklyBonusXP != want {
				t.Fatalf("WeeklyBonusXP = %d, want %d", ms.WeeklyBonusXP, want)
			}
		}
		if !found {
			t.Fatal("expected Sudoku/Easy mode stats")
		}
	})
}

// --- GetDailyStreakDates (P1) ---

func TestGetDailyStreakDates(t *testing.T) {
	t.Run("empty database", func(t *testing.T) {
		s := openTestStore(t)
		dates, err := s.GetDailyStreakDates()
		if err != nil {
			t.Fatal(err)
		}
		if len(dates) != 0 {
			t.Errorf("expected 0 dates, got %d", len(dates))
		}
	})

	t.Run("returns distinct dates descending", func(t *testing.T) {
		s := openTestStore(t)
		now := time.Now().UTC()
		yesterday := now.AddDate(0, 0, -1)

		games := []*GameRecord{
			newDailyTestRecord("Daily Feb 16 26 - a", time.Date(2026, time.February, 16, 12, 0, 0, 0, time.Local)),
			newDailyTestRecord("Daily Feb 15 26 - b", time.Date(2026, time.February, 15, 12, 0, 0, 0, time.Local)),
			{Name: "regular-game", GameType: "Sudoku", Mode: "Easy", InitialState: "{}", SaveState: "{}", Status: StatusNew},
		}
		games[0].GameType, games[0].Mode = "Sudoku", "Easy"
		games[0].InitialState, games[0].SaveState, games[0].Status = "{}", "{}", StatusNew
		games[1].GameType, games[1].Mode = "Nonogram", "Mini"
		games[1].InitialState, games[1].SaveState, games[1].Status = "{}", "{}", StatusNew
		for _, g := range games {
			if err := s.CreateGame(g); err != nil {
				t.Fatal(err)
			}
		}

		// Complete both daily games with specific completed_at times.
		if _, err := s.db.Exec(`UPDATE games SET status = 'completed', completed_at = ? WHERE id = ?`, now, games[0].ID); err != nil {
			t.Fatal(err)
		}
		if _, err := s.db.Exec(`UPDATE games SET status = 'completed', completed_at = ? WHERE id = ?`, yesterday, games[1].ID); err != nil {
			t.Fatal(err)
		}
		// Complete regular game too (should not appear in results).
		if _, err := s.db.Exec(`UPDATE games SET status = 'completed', completed_at = ? WHERE id = ?`, now, games[2].ID); err != nil {
			t.Fatal(err)
		}

		dates, err := s.GetDailyStreakDates()
		if err != nil {
			t.Fatal(err)
		}
		if len(dates) != 2 {
			t.Fatalf("expected 2 dates, got %d", len(dates))
		}
		// Most recent first.
		if !dates[0].After(dates[1]) {
			t.Errorf("expected descending order: %v, %v", dates[0], dates[1])
		}
	})

	t.Run("excludes non-completed dailies", func(t *testing.T) {
		s := openTestStore(t)
		games := []*GameRecord{
			newDailyTestRecord("Daily Feb 16 26 - a", time.Date(2026, time.February, 16, 12, 0, 0, 0, time.Local)),
			newDailyTestRecord("Daily Feb 15 26 - b", time.Date(2026, time.February, 15, 12, 0, 0, 0, time.Local)),
		}
		for _, g := range games {
			g.GameType = "Sudoku"
			g.Mode = "Easy"
			g.InitialState = "{}"
			g.SaveState = "{}"
			g.Status = StatusNew
		}
		for _, g := range games {
			if err := s.CreateGame(g); err != nil {
				t.Fatal(err)
			}
		}
		// Only abandon, don't complete.
		if err := s.UpdateStatus(games[0].ID, StatusAbandoned); err != nil {
			t.Fatal(err)
		}

		dates, err := s.GetDailyStreakDates()
		if err != nil {
			t.Fatal(err)
		}
		if len(dates) != 0 {
			t.Errorf("expected 0 dates for non-completed dailies, got %d", len(dates))
		}
	})
}

func TestGetCompletedWeeklyGauntlets(t *testing.T) {
	s := openTestStore(t)

	games := []*GameRecord{
		newWeeklyTestRecord(weekly.Name(2026, 1, 98), 2026, 1, 98),
		newWeeklyTestRecord(weekly.Name(2026, 1, 99), 2026, 1, 99),
		newWeeklyTestRecord(weekly.Name(2026, 2, 50), 2026, 2, 50),
		newWeeklyTestRecord(weekly.Name(2026, 3, 99), 2026, 3, 99),
		{Name: "Week 3-2026 - #99", GameType: "Sudoku", Mode: "Easy", InitialState: "{}", SaveState: "{}", Status: StatusNew},
	}
	for _, g := range games[:4] {
		g.GameType = "Sudoku"
		g.Mode = "Easy"
		g.InitialState = "{}"
		g.SaveState = "{}"
		g.Status = StatusNew
	}
	for _, g := range games {
		if err := s.CreateGame(g); err != nil {
			t.Fatal(err)
		}
		if err := s.UpdateStatus(g.ID, StatusCompleted); err != nil {
			t.Fatal(err)
		}
	}

	got, err := s.GetCompletedWeeklyGauntlets()
	if err != nil {
		t.Fatal(err)
	}
	if got != 2 {
		t.Fatalf("GetCompletedWeeklyGauntlets() = %d, want 2", got)
	}
}

func TestGetCurrentWeeklyHighestCompletedIndex(t *testing.T) {
	s := openTestStore(t)
	games := []*GameRecord{
		newWeeklyTestRecord(weekly.Name(2026, 10, 1), 2026, 10, 1),
		newWeeklyTestRecord(weekly.Name(2026, 10, 17), 2026, 10, 17),
		newWeeklyTestRecord(weekly.Name(2026, 10, 9), 2026, 10, 9),
		newWeeklyTestRecord(weekly.Name(2026, 11, 30), 2026, 11, 30),
	}
	for _, g := range games {
		g.GameType = "Sudoku"
		g.Mode = "Easy"
		g.InitialState = "{}"
		g.SaveState = "{}"
		g.Status = StatusNew
	}
	for _, g := range games {
		if err := s.CreateGame(g); err != nil {
			t.Fatal(err)
		}
		if err := s.UpdateStatus(g.ID, StatusCompleted); err != nil {
			t.Fatal(err)
		}
	}

	got, err := s.GetCurrentWeeklyHighestCompletedIndex(2026, 10)
	if err != nil {
		t.Fatal(err)
	}
	if got != 17 {
		t.Fatalf("GetCurrentWeeklyHighestCompletedIndex() = %d, want 17", got)
	}
}
