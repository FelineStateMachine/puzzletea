package nurikabe

import (
	"reflect"
	"testing"

	"github.com/FelineStateMachine/puzzletea/difficulty"
)

func TestSpawnEloRejectsInvalidElo(t *testing.T) {
	mode := NewMode("Elo", "test", 5, 5, 0.28, 5)
	gamer, report, err := mode.SpawnElo("seed", difficulty.SoftCapElo+1)
	if err == nil {
		t.Fatal("SpawnElo returned nil error")
	}
	if gamer != nil {
		t.Fatalf("gamer = %#v, want nil", gamer)
	}
	if !reflect.DeepEqual(report, difficulty.Report{}) {
		t.Fatalf("report = %#v, want zero report", report)
	}
}

func TestSpawnEloDeterministicForSameSeedAndElo(t *testing.T) {
	mode := NewMode("Elo", "test", 5, 5, 0.28, 5)
	const target = difficulty.Elo(600)

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
		t.Fatalf("first save returned error: %v", err)
	}
	saveB, err := gameB.GetSave()
	if err != nil {
		t.Fatalf("second save returned error: %v", err)
	}
	if string(saveA) != string(saveB) {
		t.Fatalf("SpawnElo saves differ for same seed and Elo:\n%s\n%s", saveA, saveB)
	}
	if !reflect.DeepEqual(reportA, reportB) {
		t.Fatalf("SpawnElo reports differ for same seed and Elo:\n%#v\n%#v", reportA, reportB)
	}
}

func TestSpawnEloReportPopulated(t *testing.T) {
	mode := NewMode("Elo", "test", 5, 5, 0.28, 5)
	const target = difficulty.Elo(900)

	gamer, report, err := mode.SpawnElo("report-seed", target)
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
		"width",
		"height",
		"cells",
		"clue_count",
		"unknown_count",
		"solution_count",
		"solver_nodes",
		"branches",
		"max_depth",
	} {
		if _, ok := report.Metrics[key]; !ok {
			t.Fatalf("metric %q missing from %+v", key, report.Metrics)
		}
	}
}
