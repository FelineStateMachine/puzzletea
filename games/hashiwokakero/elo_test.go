package hashiwokakero

import (
	"reflect"
	"testing"

	"github.com/FelineStateMachine/puzzletea/difficulty"
)

func TestSpawnEloRejectsInvalidElo(t *testing.T) {
	mode := NewMode("Elo", "test", 7, 7, 8, 10)

	for _, elo := range []difficulty.Elo{difficulty.MinElo - 1, difficulty.SoftCapElo + 1} {
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

func TestSpawnEloDeterministicForSameSeedAndElo(t *testing.T) {
	mode := NewMode("Elo", "test", 7, 7, 8, 10)
	const target = difficulty.Elo(1700)

	gameA, reportA, err := mode.SpawnElo("same-seed", target)
	if err != nil {
		t.Fatalf("first SpawnElo returned error: %v", err)
	}
	gameB, reportB, err := mode.SpawnElo("same-seed", target)
	if err != nil {
		t.Fatalf("second SpawnElo returned error: %v", err)
	}

	saveA, err := gameA.GetSave()
	if err != nil {
		t.Fatalf("first GetSave returned error: %v", err)
	}
	saveB, err := gameB.GetSave()
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
	mode := NewMode("Elo", "test", 7, 7, 8, 10)
	const target = difficulty.Elo(2100)

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
	if report.Confidence != difficulty.ConfidenceMedium {
		t.Fatalf("Confidence = %q, want %q", report.Confidence, difficulty.ConfidenceMedium)
	}

	requiredMetrics := map[string]float64{
		"width":             1,
		"height":            1,
		"cells":             1,
		"islands":           1,
		"island_density":    0,
		"target_minlands":   1,
		"target_maxlands":   1,
		"total_required":    1,
		"avg_required":      1,
		"max_required":      1,
		"high_required":     0,
		"connectable_pairs": 1,
		"pair_density":      1,
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
	if report.Metrics["island_density"] > 1 {
		t.Fatalf("island_density = %v, want <= 1", report.Metrics["island_density"])
	}
}

func TestSpawnEloPreservesSelectedDimensions(t *testing.T) {
	mode := NewMode("Easy 7x7", "7x7 grid with 8-10 islands.", 7, 7, 8, 10)
	_, report, err := mode.SpawnElo("fixed-size", 2800)
	if err != nil {
		t.Fatalf("SpawnElo returned error: %v", err)
	}
	if got, want := report.Metrics["width"], 7.0; got != want {
		t.Fatalf("width metric = %.0f, want %.0f", got, want)
	}
	if got, want := report.Metrics["height"], 7.0; got != want {
		t.Fatalf("height metric = %.0f, want %.0f", got, want)
	}
}

func TestModeDefinitionsIncludePresetElo(t *testing.T) {
	if len(ModeDefinitions) != len(Modes) {
		t.Fatalf("ModeDefinitions length = %d, want %d", len(ModeDefinitions), len(Modes))
	}
	for i, def := range ModeDefinitions {
		if def.PresetElo == nil {
			t.Fatalf("ModeDefinitions[%d].PresetElo is nil", i)
		}
		if err := difficulty.ValidateElo(*def.PresetElo); err != nil {
			t.Fatalf("ModeDefinitions[%d].PresetElo = %d is invalid: %v", i, *def.PresetElo, err)
		}
	}
}
