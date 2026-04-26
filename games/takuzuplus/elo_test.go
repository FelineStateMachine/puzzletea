package takuzuplus

import (
	"context"
	"reflect"
	"testing"

	"github.com/FelineStateMachine/puzzletea/difficulty"
)

func TestSpawnEloRejectsInvalidElo(t *testing.T) {
	mode := NewMode("Elo", "test", 8, 0.4, modeProfiles[1])
	tests := []difficulty.Elo{
		difficulty.MinElo - 1,
		difficulty.SoftCapElo + 1,
	}

	for _, elo := range tests {
		gamer, report, err := mode.SpawnElo("seed", elo)
		if err == nil {
			t.Fatalf("SpawnElo(%d) error = nil, want invalid Elo error", elo)
		}
		if gamer != nil {
			t.Fatalf("SpawnElo(%d) gamer = %#v, want nil", elo, gamer)
		}
		if !reflect.DeepEqual(report, difficulty.Report{}) {
			t.Fatalf("SpawnElo(%d) report = %#v, want zero report", elo, report)
		}
	}
}

func TestSpawnEloContextCanceled(t *testing.T) {
	mode := NewMode("Elo", "test", 8, 0.4, modeProfiles[1])
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	gamer, report, err := mode.SpawnEloContext(ctx, "seed", 1200)
	if err == nil {
		t.Fatal("SpawnEloContext returned nil error for canceled context")
	}
	if gamer != nil {
		t.Fatalf("gamer = %#v, want nil", gamer)
	}
	if !reflect.DeepEqual(report, difficulty.Report{}) {
		t.Fatalf("report = %#v, want zero report", report)
	}
}

func TestSpawnEloDeterministicForSameSeedAndElo(t *testing.T) {
	mode := NewMode("Elo", "test", 8, 0.4, modeProfiles[1])
	const elo = difficulty.Elo(900)
	const seed = "same-seed"

	gamerA, reportA, err := mode.SpawnElo(seed, elo)
	if err != nil {
		t.Fatalf("first SpawnElo returned error: %v", err)
	}
	gamerB, reportB, err := mode.SpawnElo(seed, elo)
	if err != nil {
		t.Fatalf("second SpawnElo returned error: %v", err)
	}

	saveA, err := gamerA.GetSave()
	if err != nil {
		t.Fatalf("first GetSave returned error: %v", err)
	}
	saveB, err := gamerB.GetSave()
	if err != nil {
		t.Fatalf("second GetSave returned error: %v", err)
	}
	if string(saveA) != string(saveB) {
		t.Fatalf("SpawnElo saves differ for same seed and Elo:\n%s\n%s", saveA, saveB)
	}
	if !reflect.DeepEqual(reportA, reportB) {
		t.Fatalf("reports differ for same seed and Elo:\n%#v\n%#v", reportA, reportB)
	}
}

func TestSpawnEloPopulatesDifficultyReport(t *testing.T) {
	mode := NewMode("Elo", "test", 8, 0.4, modeProfiles[1])
	const target = difficulty.Elo(1200)

	gamer, report, err := mode.SpawnElo("report-fields", target)
	if err != nil {
		t.Fatalf("SpawnElo returned error: %v", err)
	}
	if gamer == nil {
		t.Fatal("SpawnElo returned nil gamer")
	}
	if report.TargetElo != target {
		t.Fatalf("TargetElo = %d, want %d", report.TargetElo, target)
	}
	if err := difficulty.ValidateElo(report.ActualElo); err != nil {
		t.Fatalf("ActualElo = %d is invalid: %v", report.ActualElo, err)
	}
	if report.Confidence == "" {
		t.Fatal("Confidence is empty")
	}

	for _, key := range []string{
		"size",
		"clue_density",
		"solution_count",
		"solver_nodes",
		"max_depth",
		"relation_count",
		"same_relations",
		"diff_relations",
		"occupied_regions",
	} {
		if _, ok := report.Metrics[key]; !ok {
			t.Fatalf("Metrics missing %q: %#v", key, report.Metrics)
		}
	}
	if report.Metrics["relation_count"] == 0 {
		t.Fatalf("relation_count = 0, want relation clues in report: %#v", report.Metrics)
	}
}
