package fillomino

import (
	"reflect"
	"testing"

	"github.com/FelineStateMachine/puzzletea/difficulty"
)

func TestSpawnEloRejectsInvalidElo(t *testing.T) {
	mode := NewMode("Elo", "test", 6, 6, 0.6)
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
	mode := NewMode("Elo", "test", 6, 6, 0.6)
	gameA, reportA, err := mode.SpawnElo("same-seed", 1300)
	if err != nil {
		t.Fatal(err)
	}
	gameB, reportB, err := mode.SpawnElo("same-seed", 1300)
	if err != nil {
		t.Fatal(err)
	}
	saveA, err := gameA.GetSave()
	if err != nil {
		t.Fatal(err)
	}
	saveB, err := gameB.GetSave()
	if err != nil {
		t.Fatal(err)
	}
	if string(saveA) != string(saveB) {
		t.Fatalf("SpawnElo saves differ for same seed and Elo:\n%s\n%s", saveA, saveB)
	}
	if !reflect.DeepEqual(reportA, reportB) {
		t.Fatalf("reports differ:\n%#v\n%#v", reportA, reportB)
	}
}

func TestSpawnEloReportPopulated(t *testing.T) {
	mode := NewMode("Elo", "test", 6, 6, 0.6)
	_, report, err := mode.SpawnElo("report-seed", 1800)
	if err != nil {
		t.Fatal(err)
	}
	if report.TargetElo != 1800 {
		t.Fatalf("TargetElo = %d, want 1800", report.TargetElo)
	}
	if err := difficulty.ValidateElo(report.ActualElo); err != nil {
		t.Fatalf("ActualElo = %d is invalid: %v", report.ActualElo, err)
	}
	if report.Confidence == "" {
		t.Fatal("Confidence is empty")
	}
	for _, key := range []string{"width", "height", "cells", "given_count", "unknown_count", "given_ratio", "target_ratio", "max_region", "solution_count"} {
		if _, ok := report.Metrics[key]; !ok {
			t.Fatalf("metric %q missing from %+v", key, report.Metrics)
		}
	}
}
