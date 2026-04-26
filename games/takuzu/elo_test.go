package takuzu

import (
	"context"
	"fmt"
	"reflect"
	"testing"

	"github.com/FelineStateMachine/puzzletea/difficulty"
)

func TestSpawnEloRejectsInvalidElo(t *testing.T) {
	mode := NewMode("Elo", "test", 8, 0.4)

	tests := []difficulty.Elo{
		difficulty.MinElo - 1,
		difficulty.SoftCapElo + 1,
	}
	for _, elo := range tests {
		t.Run(fmt.Sprint(elo), func(t *testing.T) {
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
		})
	}
}

func TestSpawnEloContextCanceled(t *testing.T) {
	mode := NewMode("Elo", "test", 8, 0.4)
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
	mode := NewMode("Elo", "test", 8, 0.4)
	const seed = "takuzu-deterministic"
	const elo = difficulty.Elo(1700)

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
		t.Fatalf("report mismatch for same seed/Elo:\n%#v\n\n%#v", reportA, reportB)
	}
}

func TestSpawnEloPopulatesDifficultyReport(t *testing.T) {
	mode := NewMode("Elo", "test", 8, 0.4)
	const target = difficulty.Elo(2200)

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

	requiredMetrics := map[string]float64{
		"size":             1,
		"cells":            1,
		"clue_count":       1,
		"empty_count":      1,
		"clue_density":     0,
		"target_prefilled": 0,
		"solution_count":   1,
		"solver_nodes":     1,
		"branches":         0,
		"max_depth":        0,
	}
	for key, min := range requiredMetrics {
		got, ok := report.Metrics[key]
		if !ok {
			t.Fatalf("Metrics[%q] missing from %#v", key, report.Metrics)
		}
		if got < min {
			t.Fatalf("Metrics[%q] = %f, want >= %f", key, got, min)
		}
	}
}
