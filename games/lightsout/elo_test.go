package lightsout

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/FelineStateMachine/puzzletea/difficulty"
)

func TestSpawnEloRejectsInvalidElo(t *testing.T) {
	mode := NewMode("Elo", "test", 5, 5)

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

func TestSpawnEloDeterministicForSameSeedAndElo(t *testing.T) {
	mode := NewMode("Elo", "test", 5, 5)
	const seed = "lightsout-deterministic"
	const elo = difficulty.Elo(1700)

	gamerA, reportA, err := mode.SpawnElo(seed, elo)
	if err != nil {
		t.Fatalf("first SpawnElo returned error: %v", err)
	}
	gamerB, reportB, err := mode.SpawnElo(seed, elo)
	if err != nil {
		t.Fatalf("second SpawnElo returned error: %v", err)
	}

	saveA, err := gamerA.(Model).GetSave()
	if err != nil {
		t.Fatalf("first GetSave returned error: %v", err)
	}
	saveB, err := gamerB.(Model).GetSave()
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
	mode := NewMode("Elo", "test", 5, 5)
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
		"width":              1,
		"height":             1,
		"cells":              1,
		"lights_on":          1,
		"light_density":      0,
		"min_solution_moves": 1,
		"move_density":       0,
		"solvable":           1,
	}
	for key, min := range requiredMetrics {
		got, ok := report.Metrics[key]
		if !ok {
			t.Fatalf("Metrics[%q] missing from %#v", key, report.Metrics)
		}
		if got < min {
			t.Fatalf("Metrics[%q] = %v, want >= %v", key, got, min)
		}
	}
}

func TestSpawnEloPreservesSelectedDimensions(t *testing.T) {
	mode := NewMode("Easy", "3x3 grid", 3, 3)
	for _, elo := range []difficulty.Elo{100, 2800} {
		_, report, err := mode.SpawnElo("fixed-size", elo)
		if err != nil {
			t.Fatalf("SpawnElo(%d) returned error: %v", elo, err)
		}
		if got, want := report.Metrics["width"], 3.0; got != want {
			t.Fatalf("SpawnElo(%d) width metric = %.0f, want %.0f", elo, got, want)
		}
		if got, want := report.Metrics["height"], 3.0; got != want {
			t.Fatalf("SpawnElo(%d) height metric = %.0f, want %.0f", elo, got, want)
		}
	}
}
