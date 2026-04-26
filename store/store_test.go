package store

import (
	"database/sql"
	"path/filepath"
	"testing"
	"time"

	"github.com/FelineStateMachine/puzzletea/weekly"
)

func openTestStore(t *testing.T) *Store {
	t.Helper()
	s, err := Open(filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { s.Close() })
	return s
}

func newTestRecord(name string) *GameRecord {
	return &GameRecord{
		Name:         name,
		GameType:     "Sudoku",
		Mode:         "Easy",
		InitialState: `{"grid":"..."}`,
		SaveState:    `{"grid":"..."}`,
		Status:       StatusNew,
		RunKind:      RunKindNormal,
	}
}

func newDailyTestRecord(name string, date time.Time) *GameRecord {
	rec := newTestRecord(name)
	day := dayOnly(date)
	rec.RunKind = RunKindDaily
	rec.RunDate = &day
	return rec
}

func newWeeklyTestRecord(name string, year, weekNumber, index int) *GameRecord {
	rec := newTestRecord(name)
	rec.RunKind = RunKindWeekly
	rec.WeekYear = year
	rec.WeekNumber = weekNumber
	rec.WeekIndex = index
	return rec
}

func TestOpen(t *testing.T) {
	t.Run("creates directory and database", func(t *testing.T) {
		dir := t.TempDir()
		dbPath := filepath.Join(dir, "sub", "nested", "test.db")
		s, err := Open(dbPath)
		if err != nil {
			t.Fatal(err)
		}
		defer s.Close()
	})

	t.Run("opens existing database", func(t *testing.T) {
		dir := t.TempDir()
		dbPath := filepath.Join(dir, "test.db")

		s1, err := Open(dbPath)
		if err != nil {
			t.Fatal(err)
		}
		rec := newTestRecord("alpha")
		if err := s1.CreateGame(rec); err != nil {
			t.Fatal(err)
		}
		s1.Close()

		s2, err := Open(dbPath)
		if err != nil {
			t.Fatal(err)
		}
		defer s2.Close()

		got, err := s2.GetGameByName("alpha")
		if err != nil {
			t.Fatal(err)
		}
		if got == nil {
			t.Fatal("expected record to persist across reopen")
		}
	})

	t.Run("records current schema version", func(t *testing.T) {
		s := openTestStore(t)

		version, err := schemaVersion(s.db)
		if err != nil {
			t.Fatal(err)
		}
		if got, want := version, currentSchemaVersion; got != want {
			t.Fatalf("schema version = %d, want %d", got, want)
		}
	})

	t.Run("upgrades databases missing run metadata and backfills it", func(t *testing.T) {
		dbPath := filepath.Join(t.TempDir(), "legacy.db")

		raw, err := sql.Open("sqlite", dbPath)
		if err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() { raw.Close() })

		_, err = raw.Exec(`
CREATE TABLE games (
    id            INTEGER PRIMARY KEY AUTOINCREMENT,
    name          TEXT    NOT NULL UNIQUE,
    game_type     TEXT    NOT NULL,
    mode          TEXT    NOT NULL,
    initial_state TEXT    NOT NULL,
    save_state    TEXT,
    status        TEXT    NOT NULL DEFAULT 'new'
                  CHECK(status IN ('new','in_progress','completed','abandoned')),
    created_at    DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at    DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    completed_at  DATETIME
);`)
		if err != nil {
			t.Fatal(err)
		}

		_, err = raw.Exec(`
INSERT INTO games (name, game_type, mode, initial_state, save_state, status)
VALUES (?, ?, ?, ?, ?, ?)`,
			"Daily Feb 16 26 - amber-fox",
			"Sudoku",
			"Easy",
			`{"grid":"initial"}`,
			`{"grid":"save"}`,
			string(StatusNew),
		)
		if err != nil {
			t.Fatal(err)
		}
		if err := raw.Close(); err != nil {
			t.Fatal(err)
		}

		s, err := Open(dbPath)
		if err != nil {
			t.Fatal(err)
		}
		defer s.Close()

		version, err := schemaVersion(s.db)
		if err != nil {
			t.Fatal(err)
		}
		if got, want := version, currentSchemaVersion; got != want {
			t.Fatalf("schema version = %d, want %d", got, want)
		}

		rec, err := s.GetGameByName("Daily Feb 16 26 - amber-fox")
		if err != nil {
			t.Fatal(err)
		}
		if rec == nil {
			t.Fatal("expected migrated record")
		}
		if got, want := rec.GameID, "sudoku"; got != want {
			t.Fatalf("GameID = %q, want %q", got, want)
		}
		if got, want := rec.ModeID, "easy"; got != want {
			t.Fatalf("ModeID = %q, want %q", got, want)
		}
		if got, want := rec.VariantID, "easy"; got != want {
			t.Fatalf("VariantID = %q, want %q", got, want)
		}
		if got, want := rec.Variant, "Easy"; got != want {
			t.Fatalf("Variant = %q, want %q", got, want)
		}
		if got, want := rec.RunKind, RunKindDaily; got != want {
			t.Fatalf("RunKind = %q, want %q", got, want)
		}
		if rec.RunDate == nil {
			t.Fatal("expected RunDate to be backfilled")
		}

		if _, err := s.GetCategoryStats(); err != nil {
			t.Fatalf("GetCategoryStats() error after migration: %v", err)
		}
	})
}

func TestCreateGame(t *testing.T) {
	t.Run("sets ID", func(t *testing.T) {
		s := openTestStore(t)
		rec := newTestRecord("alpha")
		if err := s.CreateGame(rec); err != nil {
			t.Fatal(err)
		}
		if rec.ID <= 0 {
			t.Fatalf("expected positive ID, got %d", rec.ID)
		}
	})

	t.Run("sequential IDs", func(t *testing.T) {
		s := openTestStore(t)
		r1 := newTestRecord("alpha")
		r2 := newTestRecord("beta")
		if err := s.CreateGame(r1); err != nil {
			t.Fatal(err)
		}
		if err := s.CreateGame(r2); err != nil {
			t.Fatal(err)
		}
		if r2.ID <= r1.ID {
			t.Fatalf("expected r2.ID (%d) > r1.ID (%d)", r2.ID, r1.ID)
		}
	})

	t.Run("preserves fields", func(t *testing.T) {
		s := openTestStore(t)
		rec := &GameRecord{
			Name:         "gamma",
			GameType:     "Nonogram",
			Mode:         "Hard 10x10",
			InitialState: `{"init":true}`,
			SaveState:    `{"save":true}`,
			Status:       StatusNew,
			RunKind:      RunKindNormal,
		}
		targetElo := 1200
		actualElo := 1185
		rec.TargetDifficultyElo = &targetElo
		rec.ActualDifficultyElo = &actualElo
		rec.DifficultyConfidence = "high"
		if err := s.CreateGame(rec); err != nil {
			t.Fatal(err)
		}

		got, err := s.GetGameByName("gamma")
		if err != nil {
			t.Fatal(err)
		}
		if got == nil {
			t.Fatal("expected record")
		}
		if got.Name != "gamma" {
			t.Errorf("Name = %q, want %q", got.Name, "gamma")
		}
		if got.GameType != "Nonogram" {
			t.Errorf("GameType = %q, want %q", got.GameType, "Nonogram")
		}
		if got.Mode != "Hard 10x10" {
			t.Errorf("Mode = %q, want %q", got.Mode, "Hard 10x10")
		}
		if got.Variant != "Hard 10x10" {
			t.Errorf("Variant = %q, want %q", got.Variant, "Hard 10x10")
		}
		if got.VariantID != "hard 10x10" {
			t.Errorf("VariantID = %q, want %q", got.VariantID, "hard 10x10")
		}
		if got.InitialState != `{"init":true}` {
			t.Errorf("InitialState = %q, want %q", got.InitialState, `{"init":true}`)
		}
		if got.SaveState != `{"save":true}` {
			t.Errorf("SaveState = %q, want %q", got.SaveState, `{"save":true}`)
		}
		if got.TargetDifficultyElo == nil || *got.TargetDifficultyElo != targetElo {
			t.Fatalf("TargetDifficultyElo = %v, want %d", got.TargetDifficultyElo, targetElo)
		}
		if got.ActualDifficultyElo == nil || *got.ActualDifficultyElo != actualElo {
			t.Fatalf("ActualDifficultyElo = %v, want %d", got.ActualDifficultyElo, actualElo)
		}
		if got.DifficultyConfidence != "high" {
			t.Fatalf("DifficultyConfidence = %q, want high", got.DifficultyConfidence)
		}
		if got.Status != StatusNew {
			t.Errorf("Status = %q, want %q", got.Status, StatusNew)
		}
	})

	t.Run("rejects duplicate name", func(t *testing.T) {
		s := openTestStore(t)
		r1 := newTestRecord("dup")
		if err := s.CreateGame(r1); err != nil {
			t.Fatal(err)
		}
		r2 := newTestRecord("dup")
		if err := s.CreateGame(r2); err == nil {
			t.Fatal("expected error for duplicate name")
		}
	})

	t.Run("does not infer deterministic metadata from name", func(t *testing.T) {
		s := openTestStore(t)
		rec := newTestRecord("Daily Feb 16 26 - amber-fox")
		if err := s.CreateGame(rec); err != nil {
			t.Fatal(err)
		}

		got, err := s.GetGameByName(rec.Name)
		if err != nil {
			t.Fatal(err)
		}
		if got == nil {
			t.Fatal("expected record")
		}
		if got.RunKind != RunKindNormal {
			t.Fatalf("RunKind = %q, want %q", got.RunKind, RunKindNormal)
		}
		if got.RunDate != nil {
			t.Fatalf("RunDate = %v, want nil", got.RunDate)
		}
		if got.SeedText != "" {
			t.Fatalf("SeedText = %q, want empty", got.SeedText)
		}
		if got.WeekYear != 0 || got.WeekNumber != 0 || got.WeekIndex != 0 {
			t.Fatalf("weekly metadata = (%d, %d, %d), want zero values", got.WeekYear, got.WeekNumber, got.WeekIndex)
		}
	})
}

func TestUpdateSaveState(t *testing.T) {
	t.Run("updates save state", func(t *testing.T) {
		s := openTestStore(t)
		rec := newTestRecord("alpha")
		if err := s.CreateGame(rec); err != nil {
			t.Fatal(err)
		}

		if err := s.UpdateSaveState(rec.ID, `{"new":"state"}`); err != nil {
			t.Fatal(err)
		}

		got, err := s.GetGameByName("alpha")
		if err != nil {
			t.Fatal(err)
		}
		if got.SaveState != `{"new":"state"}` {
			t.Errorf("SaveState = %q, want %q", got.SaveState, `{"new":"state"}`)
		}
	})

	t.Run("advances updated_at", func(t *testing.T) {
		s := openTestStore(t)
		rec := newTestRecord("alpha")
		if err := s.CreateGame(rec); err != nil {
			t.Fatal(err)
		}

		before, _ := s.GetGameByName("alpha")
		time.Sleep(10 * time.Millisecond)

		if err := s.UpdateSaveState(rec.ID, `{"v2":true}`); err != nil {
			t.Fatal(err)
		}

		after, _ := s.GetGameByName("alpha")
		if !after.UpdatedAt.After(before.UpdatedAt) {
			t.Errorf("UpdatedAt not advanced: before=%v, after=%v", before.UpdatedAt, after.UpdatedAt)
		}
	})
}

func TestUpdateStatus(t *testing.T) {
	t.Run("new to in_progress", func(t *testing.T) {
		s := openTestStore(t)
		rec := newTestRecord("alpha")
		if err := s.CreateGame(rec); err != nil {
			t.Fatal(err)
		}
		if err := s.UpdateStatus(rec.ID, StatusInProgress); err != nil {
			t.Fatal(err)
		}

		got, _ := s.GetGameByName("alpha")
		if got.Status != StatusInProgress {
			t.Errorf("Status = %q, want %q", got.Status, StatusInProgress)
		}
		if got.CompletedAt != nil {
			t.Error("CompletedAt should be nil for in_progress")
		}
	})

	t.Run("in_progress to completed sets completed_at", func(t *testing.T) {
		s := openTestStore(t)
		rec := newTestRecord("alpha")
		if err := s.CreateGame(rec); err != nil {
			t.Fatal(err)
		}
		if err := s.UpdateStatus(rec.ID, StatusCompleted); err != nil {
			t.Fatal(err)
		}

		got, _ := s.GetGameByName("alpha")
		if got.Status != StatusCompleted {
			t.Errorf("Status = %q, want %q", got.Status, StatusCompleted)
		}
		if got.CompletedAt == nil {
			t.Fatal("CompletedAt should be set for completed")
		}
		if time.Since(*got.CompletedAt) > 5*time.Second {
			t.Errorf("CompletedAt too old: %v", *got.CompletedAt)
		}
	})

	t.Run("in_progress to abandoned", func(t *testing.T) {
		s := openTestStore(t)
		rec := newTestRecord("alpha")
		if err := s.CreateGame(rec); err != nil {
			t.Fatal(err)
		}
		if err := s.UpdateStatus(rec.ID, StatusAbandoned); err != nil {
			t.Fatal(err)
		}

		// GetGameByName excludes abandoned, so query directly.
		all, _ := s.ListAllGames()
		if len(all) != 1 {
			t.Fatalf("expected 1 game, got %d", len(all))
		}
		if all[0].Status != StatusAbandoned {
			t.Errorf("Status = %q, want %q", all[0].Status, StatusAbandoned)
		}
		if all[0].CompletedAt != nil {
			t.Error("CompletedAt should be nil for abandoned")
		}
	})
}

func TestNameExists(t *testing.T) {
	t.Run("returns true for existing name", func(t *testing.T) {
		s := openTestStore(t)
		rec := newTestRecord("alpha")
		if err := s.CreateGame(rec); err != nil {
			t.Fatal(err)
		}

		exists, err := s.NameExists("alpha")
		if err != nil {
			t.Fatal(err)
		}
		if !exists {
			t.Error("expected true for existing name")
		}
	})

	t.Run("returns false for missing name", func(t *testing.T) {
		s := openTestStore(t)

		exists, err := s.NameExists("nope")
		if err != nil {
			t.Fatal(err)
		}
		if exists {
			t.Error("expected false for missing name")
		}
	})

	t.Run("includes abandoned games", func(t *testing.T) {
		s := openTestStore(t)
		rec := newTestRecord("alpha")
		if err := s.CreateGame(rec); err != nil {
			t.Fatal(err)
		}
		if err := s.UpdateStatus(rec.ID, StatusAbandoned); err != nil {
			t.Fatal(err)
		}

		exists, err := s.NameExists("alpha")
		if err != nil {
			t.Fatal(err)
		}
		if !exists {
			t.Error("NameExists should include abandoned games")
		}
	})

	t.Run("empty database", func(t *testing.T) {
		s := openTestStore(t)

		exists, err := s.NameExists("anything")
		if err != nil {
			t.Fatal(err)
		}
		if exists {
			t.Error("expected false for empty database")
		}
	})
}

func TestListGames(t *testing.T) {
	t.Run("excludes abandoned", func(t *testing.T) {
		s := openTestStore(t)
		for _, name := range []string{"a", "b", "c"} {
			rec := newTestRecord(name)
			if err := s.CreateGame(rec); err != nil {
				t.Fatal(err)
			}
			if name == "b" {
				if err := s.UpdateStatus(rec.ID, StatusAbandoned); err != nil {
					t.Fatal(err)
				}
			}
		}

		games, err := s.ListGames()
		if err != nil {
			t.Fatal(err)
		}
		if len(games) != 2 {
			t.Fatalf("expected 2 games, got %d", len(games))
		}
		for _, g := range games {
			if g.Status == StatusAbandoned {
				t.Error("ListGames should not include abandoned games")
			}
		}
	})

	t.Run("ordered by updated_at DESC", func(t *testing.T) {
		s := openTestStore(t)
		for _, name := range []string{"first", "second", "third"} {
			rec := newTestRecord(name)
			if err := s.CreateGame(rec); err != nil {
				t.Fatal(err)
			}
			time.Sleep(20 * time.Millisecond)
		}
		// Update "first" to make it most recent.
		first, _ := s.GetGameByName("first")
		time.Sleep(20 * time.Millisecond)
		if err := s.UpdateSaveState(first.ID, `{"updated":true}`); err != nil {
			t.Fatal(err)
		}

		games, err := s.ListGames()
		if err != nil {
			t.Fatal(err)
		}
		if len(games) < 3 {
			t.Fatalf("expected 3 games, got %d", len(games))
		}
		if games[0].Name != "first" {
			t.Errorf("expected first game to be 'first', got %q", games[0].Name)
		}
	})

	t.Run("empty database returns empty slice", func(t *testing.T) {
		s := openTestStore(t)

		games, err := s.ListGames()
		if err != nil {
			t.Fatal(err)
		}
		if len(games) != 0 {
			t.Errorf("expected empty slice, got %d games", len(games))
		}
	})

	t.Run("includes all non-abandoned statuses", func(t *testing.T) {
		s := openTestStore(t)
		for _, name := range []string{"new-game", "progress-game", "done-game"} {
			rec := newTestRecord(name)
			if err := s.CreateGame(rec); err != nil {
				t.Fatal(err)
			}
		}
		pg, _ := s.GetGameByName("progress-game")
		if err := s.UpdateStatus(pg.ID, StatusInProgress); err != nil {
			t.Fatal(err)
		}
		dg, _ := s.GetGameByName("done-game")
		if err := s.UpdateStatus(dg.ID, StatusCompleted); err != nil {
			t.Fatal(err)
		}

		games, err := s.ListGames()
		if err != nil {
			t.Fatal(err)
		}
		if len(games) != 3 {
			t.Fatalf("expected 3 games, got %d", len(games))
		}
	})
}

func TestListAllGames(t *testing.T) {
	t.Run("includes abandoned", func(t *testing.T) {
		s := openTestStore(t)
		for _, name := range []string{"a", "b", "c"} {
			rec := newTestRecord(name)
			if err := s.CreateGame(rec); err != nil {
				t.Fatal(err)
			}
			if name == "b" {
				if err := s.UpdateStatus(rec.ID, StatusAbandoned); err != nil {
					t.Fatal(err)
				}
			}
		}

		games, err := s.ListAllGames()
		if err != nil {
			t.Fatal(err)
		}
		if len(games) != 3 {
			t.Fatalf("expected 3 games, got %d", len(games))
		}
	})
}

func TestGetGameByName(t *testing.T) {
	t.Run("finds by name", func(t *testing.T) {
		s := openTestStore(t)
		rec := newTestRecord("alpha")
		if err := s.CreateGame(rec); err != nil {
			t.Fatal(err)
		}

		got, err := s.GetGameByName("alpha")
		if err != nil {
			t.Fatal(err)
		}
		if got == nil {
			t.Fatal("expected record")
		}
		if got.Name != "alpha" {
			t.Errorf("Name = %q, want %q", got.Name, "alpha")
		}
	})

	t.Run("excludes abandoned", func(t *testing.T) {
		s := openTestStore(t)
		rec := newTestRecord("alpha")
		if err := s.CreateGame(rec); err != nil {
			t.Fatal(err)
		}
		if err := s.UpdateStatus(rec.ID, StatusAbandoned); err != nil {
			t.Fatal(err)
		}

		got, err := s.GetGameByName("alpha")
		if err != nil {
			t.Fatal(err)
		}
		if got != nil {
			t.Error("expected nil for abandoned game")
		}
	})

	t.Run("returns nil for missing name", func(t *testing.T) {
		s := openTestStore(t)

		got, err := s.GetGameByName("nope")
		if err != nil {
			t.Fatal(err)
		}
		if got != nil {
			t.Error("expected nil for missing name")
		}
	})

	t.Run("preserves completed_at", func(t *testing.T) {
		s := openTestStore(t)
		rec := newTestRecord("alpha")
		if err := s.CreateGame(rec); err != nil {
			t.Fatal(err)
		}
		if err := s.UpdateStatus(rec.ID, StatusCompleted); err != nil {
			t.Fatal(err)
		}

		got, _ := s.GetGameByName("alpha")
		if got.CompletedAt == nil {
			t.Fatal("expected CompletedAt to be set")
		}
	})
}

func TestGetDeterministicGame(t *testing.T) {
	s := openTestStore(t)
	rec := newWeeklyTestRecord("Week 01-2026 - #01", 2026, 1, 1)
	if err := s.CreateGame(rec); err != nil {
		t.Fatal(err)
	}
	if err := s.UpdateStatus(rec.ID, StatusAbandoned); err != nil {
		t.Fatal(err)
	}

	got, err := s.GetDeterministicGame(rec.Name)
	if err != nil {
		t.Fatal(err)
	}
	if got == nil {
		t.Fatal("expected deterministic record")
	}
	if got.Status != StatusAbandoned {
		t.Fatalf("status = %q, want %q", got.Status, StatusAbandoned)
	}
}

func TestWeeklyQueries(t *testing.T) {
	s := openTestStore(t)

	records := []*GameRecord{
		newWeeklyTestRecord(weekly.Name(2026, 1, 1), 2026, 1, 1),
		newWeeklyTestRecord(weekly.Name(2026, 1, 2), 2026, 1, 2),
		newWeeklyTestRecord(weekly.Name(2026, 2, 1), 2026, 2, 1),
		newTestRecord("Week 1-2026 - #01"),
	}
	for _, rec := range records {
		if err := s.CreateGame(rec); err != nil {
			t.Fatal(err)
		}
	}

	got, err := s.GetWeeklyGame(2026, 1, 2)
	if err != nil {
		t.Fatal(err)
	}
	if got == nil || got.Name != weekly.Name(2026, 1, 2) {
		t.Fatalf("GetWeeklyGame() = %#v, want %q", got, weekly.Name(2026, 1, 2))
	}

	weeklyGames, err := s.ListWeeklyGames(2026, 1)
	if err != nil {
		t.Fatal(err)
	}
	if len(weeklyGames) != 2 {
		t.Fatalf("ListWeeklyGames() len = %d, want 2", len(weeklyGames))
	}
	for _, rec := range weeklyGames {
		info, ok := weekly.ParseName(rec.Name)
		if !ok {
			t.Fatalf("ListWeeklyGames() included non-canonical name %q", rec.Name)
		}
		if info.Year != 2026 || info.Week != 1 {
			t.Fatalf("ListWeeklyGames() included wrong week %#v", info)
		}
	}
}

func TestFullLifecycle(t *testing.T) {
	s := openTestStore(t)

	rec := newTestRecord("lifecycle")
	if err := s.CreateGame(rec); err != nil {
		t.Fatal(err)
	}
	if rec.Status != StatusNew {
		t.Errorf("initial status = %q, want %q", rec.Status, StatusNew)
	}

	if err := s.UpdateStatus(rec.ID, StatusInProgress); err != nil {
		t.Fatal(err)
	}

	if err := s.UpdateSaveState(rec.ID, `{"progress":"halfway"}`); err != nil {
		t.Fatal(err)
	}

	mid, _ := s.GetGameByName("lifecycle")
	if mid.Status != StatusInProgress {
		t.Errorf("mid-game status = %q, want %q", mid.Status, StatusInProgress)
	}
	if mid.SaveState != `{"progress":"halfway"}` {
		t.Errorf("mid-game SaveState = %q", mid.SaveState)
	}

	if err := s.UpdateStatus(rec.ID, StatusCompleted); err != nil {
		t.Fatal(err)
	}

	done, _ := s.GetGameByName("lifecycle")
	if done.Status != StatusCompleted {
		t.Errorf("final status = %q, want %q", done.Status, StatusCompleted)
	}
	if done.CompletedAt == nil {
		t.Error("expected CompletedAt to be set")
	}

	games, _ := s.ListGames()
	found := false
	for _, g := range games {
		if g.Name == "lifecycle" {
			found = true
		}
	}
	if !found {
		t.Error("completed game should appear in ListGames")
	}
}
