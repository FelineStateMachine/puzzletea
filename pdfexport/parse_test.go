package pdfexport

import (
	"os"
	"path/filepath"
	"strconv"
	"testing"
)

func TestParseMarkdownNonogram(t *testing.T) {
	doc, err := ParseMarkdown("sample.md", sampleNonogramDoc("Standard", 1))
	if err != nil {
		t.Fatal(err)
	}

	if got, want := doc.Metadata.Category, "Nonogram"; got != want {
		t.Fatalf("category = %q, want %q", got, want)
	}
	if got, want := len(doc.Puzzles), 1; got != want {
		t.Fatalf("puzzles = %d, want %d", got, want)
	}

	p := doc.Puzzles[0]
	if p.Nonogram == nil {
		t.Fatal("expected parsed nonogram data")
	}
	if p.Nonogram.Width != 2 || p.Nonogram.Height != 2 {
		t.Fatalf("nonogram size = %dx%d, want 2x2", p.Nonogram.Width, p.Nonogram.Height)
	}

	if got, want := p.Nonogram.RowHints[0][0], 1; got != want {
		t.Fatalf("first row first hint = %d, want %d", got, want)
	}
	if got, want := p.Nonogram.ColHints[0][0], 1; got != want {
		t.Fatalf("first col first hint = %d, want %d", got, want)
	}

	if got, want := p.Nonogram.Grid[0][0], " "; got != want {
		t.Fatalf("grid dot replacement = %q, want %q", got, want)
	}
}

func TestParseFilesMultipleInputs(t *testing.T) {
	temp := t.TempDir()

	fileA := filepath.Join(temp, "pack-a.md")
	if err := os.WriteFile(fileA, []byte(sampleNonogramDoc("Standard", 1)), 0o644); err != nil {
		t.Fatal(err)
	}

	fileB := filepath.Join(temp, "pack-b.md")
	if err := os.WriteFile(fileB, []byte(sampleNonogramDoc("Classic", 2)), 0o644); err != nil {
		t.Fatal(err)
	}

	docs, err := ParseFiles([]string{fileA, fileB})
	if err != nil {
		t.Fatal(err)
	}
	if got, want := len(docs), 2; got != want {
		t.Fatalf("docs = %d, want %d", got, want)
	}

	if got, want := docs[0].Metadata.SourceFileName, "pack-a.md"; got != want {
		t.Fatalf("first source file = %q, want %q", got, want)
	}
	if got, want := docs[1].Puzzles[0].ModeSelection, "Classic"; got != want {
		t.Fatalf("second mode selection = %q, want %q", got, want)
	}
}

func TestParseMarkdownTakuzuTable(t *testing.T) {
	doc, err := ParseMarkdown("takuzu.md", sampleTakuzuDoc())
	if err != nil {
		t.Fatal(err)
	}

	if got, want := doc.Metadata.Category, "Takuzu"; got != want {
		t.Fatalf("category = %q, want %q", got, want)
	}

	p := doc.Puzzles[0]
	if p.Nonogram != nil {
		t.Fatal("expected nonogram data to be nil for takuzu")
	}
	if p.Table == nil {
		t.Fatal("expected parsed grid table for takuzu")
	}
	if !p.Table.HasHeaderRow {
		t.Fatal("expected takuzu table to detect a header row")
	}
	if !p.Table.HasHeaderCol {
		t.Fatal("expected takuzu table to detect a header column")
	}

	if got, want := p.Table.Rows[1][1], "."; got != want {
		t.Fatalf("table cell = %q, want %q", got, want)
	}
}

func sampleNonogramDoc(mode string, idx int) string {
	return "# PuzzleTea Export\n\n" +
		"- Generated: 2026-02-21T20:42:05-07:00\n" +
		"- Version: v1.6.0\n" +
		"- Category: Nonogram\n" +
		"- Mode Selection: " + mode + "\n" +
		"- Count: 1\n" +
		"- Seed: zine\n\n" +
		"## ember-newt - " + strconv.Itoa(idx) + "\n\n" +
		"### Puzzle Grid with Integrated Hints\n\n" +
		"| R1 | R2 | C1 | C2 |\n" +
		"| --- | --- | --- | --- |\n" +
		"| . | . | 1 | 2 |\n" +
		"| . | . | 3 | 4 |\n" +
		"| 1 | 1 | . | . |\n" +
		"| . | 2 | . | . |\n\n" +
		"Row hints are right-aligned beside each row.\n"
}

func sampleTakuzuDoc() string {
	return "# PuzzleTea Export\n\n" +
		"- Generated: 2026-02-21T20:42:05-07:00\n" +
		"- Version: v1.6.0\n" +
		"- Category: Takuzu\n" +
		"- Mode Selection: Beginner\n" +
		"- Count: 1\n" +
		"- Seed: zine\n\n" +
		"## scarlet-lichen - 1\n\n" +
		"### Given Grid\n\n" +
		"|   | 1 | 2 | 3 |\n" +
		"| --- | --- | --- | --- |\n" +
		"| 1 | . | 0 | . |\n" +
		"| 2 | 1 | . | 0 |\n\n" +
		"Goal: fill with 0/1.\n"
}
