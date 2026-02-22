package pdfexport

import "testing"

func TestParseNurikabePrintData(t *testing.T) {
	save := []byte(`{"width":3,"height":2,"clues":"1,0,2\n0,3,0","marks":"???\n???"}`)

	data, err := ParseNurikabePrintData(save)
	if err != nil {
		t.Fatal(err)
	}
	if data == nil {
		t.Fatal("expected nurikabe print data")
	}
	if got, want := data.Width, 3; got != want {
		t.Fatalf("width = %d, want %d", got, want)
	}
	if got, want := data.Height, 2; got != want {
		t.Fatalf("height = %d, want %d", got, want)
	}
	if got, want := data.Clues[0][0], 1; got != want {
		t.Fatalf("row 0 col 0 = %d, want %d", got, want)
	}
	if got, want := data.Clues[0][2], 2; got != want {
		t.Fatalf("row 0 col 2 = %d, want %d", got, want)
	}
	if got, want := data.Clues[1][1], 3; got != want {
		t.Fatalf("row 1 col 1 = %d, want %d", got, want)
	}
}

func TestParseHashiPrintData(t *testing.T) {
	save := []byte(`{"width":7,"height":7,"islands":[{"x":0,"y":0,"required":3},{"x":6,"y":6,"required":2},{"x":9,"y":9,"required":5}]}`)

	data, err := ParseHashiPrintData(save)
	if err != nil {
		t.Fatal(err)
	}
	if data == nil {
		t.Fatal("expected hashi print data")
	}
	if got, want := data.Width, 7; got != want {
		t.Fatalf("width = %d, want %d", got, want)
	}
	if got, want := data.Height, 7; got != want {
		t.Fatalf("height = %d, want %d", got, want)
	}
	if got, want := len(data.Islands), 2; got != want {
		t.Fatalf("island count = %d, want %d", got, want)
	}
	if got, want := data.Islands[0].Required, 3; got != want {
		t.Fatalf("island[0].required = %d, want %d", got, want)
	}
}

func TestParseShikakuPrintDataDuplicateClueCoordinatesUseLatest(t *testing.T) {
	save := []byte(`{"width":3,"height":3,"clues":[{"x":1,"y":1,"value":2},{"x":1,"y":1,"value":5},{"x":2,"y":0,"value":4}]}`)

	data, err := ParseShikakuPrintData(save)
	if err != nil {
		t.Fatal(err)
	}
	if data == nil {
		t.Fatal("expected shikaku print data")
	}
	if got, want := data.Clues[1][1], 5; got != want {
		t.Fatalf("row 1 col 1 = %d, want %d", got, want)
	}
	if got, want := data.Clues[0][2], 4; got != want {
		t.Fatalf("row 0 col 2 = %d, want %d", got, want)
	}
}

func TestParseHitoriPrintData(t *testing.T) {
	save := []byte(`{"size":4,"numbers":"1 2 10\n4 . .\n7\n"}`)

	data, err := ParseHitoriPrintData(save)
	if err != nil {
		t.Fatal(err)
	}
	if data == nil {
		t.Fatal("expected hitori print data")
	}
	if got, want := data.Size, 4; got != want {
		t.Fatalf("size = %d, want %d", got, want)
	}

	if got, want := data.Numbers[0][2], "10"; got != want {
		t.Fatalf("row 0 col 2 = %q, want %q", got, want)
	}
	if got, want := data.Numbers[1][1], ""; got != want {
		t.Fatalf("row 1 col 1 = %q, want empty", got)
	}
	if got, want := data.Numbers[2][0], "7"; got != want {
		t.Fatalf("row 2 col 0 = %q, want %q", got, want)
	}
	if got, want := data.Numbers[3][3], ""; got != want {
		t.Fatalf("row 3 col 3 = %q, want empty", got)
	}
}

func TestParseHitoriPrintDataCompactRuneEncoding(t *testing.T) {
	save := []byte(`{"size":3,"numbers":"12:\n4..\n789"}`)

	data, err := ParseHitoriPrintData(save)
	if err != nil {
		t.Fatal(err)
	}
	if data == nil {
		t.Fatal("expected hitori print data")
	}

	if got, want := data.Numbers[0][2], "10"; got != want {
		t.Fatalf("row 0 col 2 = %q, want %q", got, want)
	}
	if got, want := data.Numbers[1][1], ""; got != want {
		t.Fatalf("row 1 col 1 = %q, want empty", got)
	}
}

func TestParseTakuzuPrintData(t *testing.T) {
	save := []byte(`{"size":4,"state":"01..\n10..\n0011\n1111","provided":"#.\n.##\n####\n#"}`)

	data, err := ParseTakuzuPrintData(save)
	if err != nil {
		t.Fatal(err)
	}
	if data == nil {
		t.Fatal("expected takuzu print data")
	}
	if got, want := data.Size, 4; got != want {
		t.Fatalf("size = %d, want %d", got, want)
	}
	if !data.GroupEveryTwo {
		t.Fatal("expected GroupEveryTwo to be true")
	}

	if got, want := data.Givens[0][0], "0"; got != want {
		t.Fatalf("row 0 col 0 = %q, want %q", got, want)
	}
	if got, want := data.Givens[1][2], ""; got != want {
		t.Fatalf("row 1 col 2 = %q, want empty", got)
	}
	if got, want := data.Givens[2][3], "1"; got != want {
		t.Fatalf("row 2 col 3 = %q, want %q", got, want)
	}
	if got, want := data.Givens[3][1], ""; got != want {
		t.Fatalf("row 3 col 1 = %q, want empty", got)
	}
}

func TestHydratePuzzlePrintDataForTakuzu(t *testing.T) {
	p := Puzzle{
		Category: "Takuzu",
		SaveData: []byte(`{"size":2,"state":"01\n10","provided":"##\n#."}`),
	}

	hydratePuzzlePrintData(&p)
	if p.Takuzu == nil {
		t.Fatal("expected takuzu payload from save hydration")
	}
	if got, want := p.Takuzu.Givens[0][1], "1"; got != want {
		t.Fatalf("row 0 col 1 = %q, want %q", got, want)
	}
	if got, want := p.Takuzu.Givens[1][1], ""; got != want {
		t.Fatalf("row 1 col 1 = %q, want empty", got)
	}
}

func TestHydratePuzzlePrintDataForNurikabeAndShikaku(t *testing.T) {
	nurikabePuzzle := Puzzle{
		Category: "Nurikabe",
		SaveData: []byte(`{"width":2,"height":2,"clues":"1,0\n0,2","marks":"??\n??"}`),
	}
	hydratePuzzlePrintData(&nurikabePuzzle)
	if nurikabePuzzle.Nurikabe == nil {
		t.Fatal("expected nurikabe payload from save hydration")
	}
	if got, want := nurikabePuzzle.Nurikabe.Clues[1][1], 2; got != want {
		t.Fatalf("nurikabe row 1 col 1 = %d, want %d", got, want)
	}

	shikakuPuzzle := Puzzle{
		Category: "Shikaku",
		SaveData: []byte(`{"width":2,"height":2,"clues":[{"x":0,"y":0,"value":1},{"x":1,"y":1,"value":3}]}`),
	}
	hydratePuzzlePrintData(&shikakuPuzzle)
	if shikakuPuzzle.Shikaku == nil {
		t.Fatal("expected shikaku payload from save hydration")
	}
	if got, want := shikakuPuzzle.Shikaku.Clues[1][1], 3; got != want {
		t.Fatalf("shikaku row 1 col 1 = %d, want %d", got, want)
	}
}

func TestHydratePuzzlePrintDataForHashi(t *testing.T) {
	puzzle := Puzzle{
		Category: "Hashiwokakero",
		SaveData: []byte(`{"width":5,"height":5,"islands":[{"x":0,"y":0,"required":2},{"x":4,"y":4,"required":3}],"bridges":[]}`),
	}

	hydratePuzzlePrintData(&puzzle)
	if puzzle.Hashi == nil {
		t.Fatal("expected hashi payload from save hydration")
	}
	if got, want := len(puzzle.Hashi.Islands), 2; got != want {
		t.Fatalf("island count = %d, want %d", got, want)
	}
	if got, want := puzzle.Hashi.Islands[1].Required, 3; got != want {
		t.Fatalf("island[1].required = %d, want %d", got, want)
	}
}
