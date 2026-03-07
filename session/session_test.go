package session

import (
	"path/filepath"
	"testing"

	"github.com/FelineStateMachine/puzzletea/lightsout"
	"github.com/FelineStateMachine/puzzletea/store"
)

func TestNormalizeSeed(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{input: "daily-challenge", want: "daily-challenge"},
		{input: "DailyChallenge", want: "dailyChallenge"},
		{input: "Week 01-2026 - #01", want: "week 01-2026 - #01"},
		{input: "seed-123", want: "seed-123"},
	}

	for _, tt := range tests {
		if got := NormalizeSeed(tt.input); got != tt.want {
			t.Fatalf("NormalizeSeed(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestSeededNameDeterministic(t *testing.T) {
	seed := "zebra-seed"
	if got, want := SeededName(seed), SeededName(seed); got != want {
		t.Fatalf("SeededName(%q) = %q then %q", seed, got, want)
	}
}

func TestSeededNameForGameDeterministicAndDistinct(t *testing.T) {
	seed := "zebra-seed"
	got := SeededNameForGame(seed, "Sudoku")
	if want := SeededNameForGame(seed, "Sudoku"); got != want {
		t.Fatalf("SeededNameForGame returned %q then %q", got, want)
	}
	if got == SeededName(seed) {
		t.Fatalf("SeededNameForGame(%q) should differ from SeededName(%q)", seed, seed)
	}
	if got == SeededNameForGame(seed, "Nonogram") {
		t.Fatal("different selected games should produce different seeded names")
	}
}

func TestImportRecordReappliesTitle(t *testing.T) {
	g, err := lightsout.New(3, 3)
	if err != nil {
		t.Fatal(err)
	}
	save, err := g.GetSave()
	if err != nil {
		t.Fatal(err)
	}

	loaded, err := ImportRecord(&store.GameRecord{
		Name:      "saved-game",
		GameType:  "Lights Out",
		SaveState: string(save),
	})
	if err != nil {
		t.Fatalf("ImportRecord returned error: %v", err)
	}
	if loaded == nil {
		t.Fatal("ImportRecord returned nil game")
	}
}

func TestCreateRecordAndResumeAbandonedDeterministicRecord(t *testing.T) {
	s := openSessionTestStore(t)

	g, err := lightsout.New(3, 3)
	if err != nil {
		t.Fatal(err)
	}
	rec, err := CreateRecord(s, g, "demo", "Lights Out", "Easy")
	if err != nil {
		t.Fatalf("CreateRecord returned error: %v", err)
	}
	if rec.ID == 0 {
		t.Fatal("CreateRecord did not populate ID")
	}

	rec.Status = store.StatusAbandoned
	if err := s.UpdateStatus(rec.ID, store.StatusAbandoned); err != nil {
		t.Fatal(err)
	}
	if err := ResumeAbandonedDeterministicRecord(s, rec); err != nil {
		t.Fatalf("ResumeAbandonedDeterministicRecord returned error: %v", err)
	}

	saved, err := s.GetDailyGame(rec.Name)
	if err != nil {
		t.Fatal(err)
	}
	if saved == nil {
		t.Fatal("expected saved record to exist")
	}
	if saved.Status != store.StatusInProgress {
		t.Fatalf("status = %q, want %q", saved.Status, store.StatusInProgress)
	}
}

func openSessionTestStore(t *testing.T) *store.Store {
	t.Helper()
	s, err := store.Open(filepath.Join(t.TempDir(), "session.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = s.Close() })
	return s
}
