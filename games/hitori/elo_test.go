package hitori

import (
	"reflect"
	"testing"

	"github.com/FelineStateMachine/puzzletea/difficulty"
)

func TestSpawnEloRejectsInvalidElo(t *testing.T) {
	mode := NewMode("Test", "Test mode.", 5, 0.32)

	gamer, report, err := mode.SpawnElo("seed", difficulty.SoftCapElo+1)
	if err == nil {
		t.Fatal("SpawnElo returned nil error for invalid Elo")
	}
	if gamer != nil {
		t.Fatalf("SpawnElo gamer = %#v, want nil", gamer)
	}
	if !reflect.DeepEqual(report, difficulty.Report{}) {
		t.Fatalf("SpawnElo report = %#v, want zero report", report)
	}
}

func TestSpawnEloDeterministicForSameSeedAndElo(t *testing.T) {
	mode := NewMode("Test", "Test mode.", 5, 0.32)
	target := difficulty.Elo(1500)

	gamerA, reportA, err := mode.SpawnElo("same-seed", target)
	if err != nil {
		t.Fatalf("first SpawnElo returned error: %v", err)
	}
	gamerB, reportB, err := mode.SpawnElo("same-seed", target)
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
		t.Fatalf("SpawnElo reports differ for same seed and Elo:\n%#v\n%#v", reportA, reportB)
	}
}

func TestSpawnEloPopulatesDifficultyReport(t *testing.T) {
	mode := NewMode("Test", "Test mode.", 5, 0.32)
	target := difficulty.Elo(2200)

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
	if report.ActualElo < difficulty.MinElo || report.ActualElo > difficulty.SoftCapElo {
		t.Fatalf("ActualElo = %d, want in supported range", report.ActualElo)
	}
	if report.Confidence != difficulty.ConfidenceMedium {
		t.Fatalf("Confidence = %q, want %q", report.Confidence, difficulty.ConfidenceMedium)
	}

	requiredMetrics := []string{
		"size",
		"cells",
		"target_black_pct",
		"duplicate_cells",
		"duplicate_groups",
		"duplicate_pct",
		"solution_count",
	}
	for _, name := range requiredMetrics {
		if _, ok := report.Metrics[name]; !ok {
			t.Fatalf("report metric %q missing from %#v", name, report.Metrics)
		}
	}
	if report.Metrics["solution_count"] != 1 {
		t.Fatalf("solution_count = %v, want 1", report.Metrics["solution_count"])
	}
}
