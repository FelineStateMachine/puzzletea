package pdfexport

import (
	"reflect"
	"testing"
	"time"
)

func TestSummarizeCategoryTotalsAggregatesAcrossSources(t *testing.T) {
	puzzles := []Puzzle{
		{SourceFileName: "sudoku-a.jsonl", Category: "Sudoku"},
		{SourceFileName: "sudoku-b.jsonl", Category: "Sudoku"},
		{SourceFileName: "takuzu-a.jsonl", Category: "Takuzu"},
	}

	got := summarizeCategoryTotals(puzzles)
	want := []namedCount{
		{Name: "Sudoku", Count: 2},
		{Name: "Takuzu", Count: 1},
	}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("summarizeCategoryTotals() = %+v, want %+v", got, want)
	}
}

func TestSummarizeCategoryTotalsUsesUnknownFallback(t *testing.T) {
	puzzles := []Puzzle{
		{Category: ""},
		{Category: "  "},
	}

	got := summarizeCategoryTotals(puzzles)
	want := []namedCount{{Name: "Unknown", Count: 2}}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("summarizeCategoryTotals() = %+v, want %+v", got, want)
	}
}

func TestTitlePageColophonIncludesVersionAndDate(t *testing.T) {
	docs := []PackDocument{
		{Metadata: PackMetadata{Version: "v1.2.3"}},
	}
	ts := time.Date(2026, 3, 10, 12, 0, 0, 0, time.UTC)

	got := titlePageColophon(docs, ts)
	want := "PuzzleTea v1.2.3 | March 10, 2026"
	if got != want {
		t.Fatalf("titlePageColophon() = %q, want %q", got, want)
	}
}
