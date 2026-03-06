package pdfexport

import "testing"

func TestParseNonogramPrintData(t *testing.T) {
	save := []byte(`{"state":" .-\n.. ","row-hints":[[1],[2]],"col-hints":[[2],[1],[1]]}`)

	data, err := ParseNonogramPrintData(save)
	if err != nil {
		t.Fatal(err)
	}
	if data == nil {
		t.Fatal("expected nonogram print data")
	}
	if got, want := data.Width, 3; got != want {
		t.Fatalf("width = %d, want %d", got, want)
	}
	if got, want := data.Height, 2; got != want {
		t.Fatalf("height = %d, want %d", got, want)
	}
	if got, want := data.RowHints[0][0], 1; got != want {
		t.Fatalf("row hint[0][0] = %d, want %d", got, want)
	}
	if got, want := data.ColHints[0][0], 2; got != want {
		t.Fatalf("col hint[0][0] = %d, want %d", got, want)
	}
	if got, want := data.Grid[0][0], " "; got != want {
		t.Fatalf("grid[0][0] = %q, want %q", got, want)
	}
	if got, want := data.Grid[0][1], "."; got != want {
		t.Fatalf("grid[0][1] = %q, want %q", got, want)
	}
	if got, want := data.Grid[0][2], "-"; got != want {
		t.Fatalf("grid[0][2] = %q, want %q", got, want)
	}
}

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

func TestParseFillominoPrintData(t *testing.T) {
	save := []byte(`{"width":3,"height":2,"state":"1 . 2\n. 3 .","provided":"#.#\n.#."}`)

	data, err := ParseFillominoPrintData(save)
	if err != nil {
		t.Fatal(err)
	}
	if data == nil {
		t.Fatal("expected fillomino print data")
	}
	if got, want := data.Width, 3; got != want {
		t.Fatalf("width = %d, want %d", got, want)
	}
	if got, want := data.Height, 2; got != want {
		t.Fatalf("height = %d, want %d", got, want)
	}
	if got, want := data.Givens[0][0], 1; got != want {
		t.Fatalf("givens[0][0] = %d, want %d", got, want)
	}
	if got := data.Givens[0][1]; got != 0 {
		t.Fatalf("givens[0][1] = %d, want 0", got)
	}
	if got, want := data.Givens[1][1], 3; got != want {
		t.Fatalf("givens[1][1] = %d, want %d", got, want)
	}
}

func TestParseRippleEffectPrintData(t *testing.T) {
	save := []byte(`{"width":3,"height":3,"givens":"1 . 1\n. 1 3\n3 . .","cages":[{"size":2,"cells":[{"x":0,"y":0},{"x":1,"y":0}]},{"size":1,"cells":[{"x":2,"y":0}]},{"size":3,"cells":[{"x":0,"y":1},{"x":0,"y":2},{"x":1,"y":2}]},{"size":3,"cells":[{"x":1,"y":1},{"x":2,"y":1},{"x":2,"y":2}]}]}`)

	data, err := ParseRippleEffectPrintData(save)
	if err != nil {
		t.Fatal(err)
	}
	if data == nil {
		t.Fatal("expected ripple effect print data")
	}
	if got, want := data.Width, 3; got != want {
		t.Fatalf("width = %d, want %d", got, want)
	}
	if got, want := len(data.Cages), 4; got != want {
		t.Fatalf("cage count = %d, want %d", got, want)
	}
	if got, want := data.Givens[0][0], 1; got != want {
		t.Fatalf("givens[0][0] = %d, want %d", got, want)
	}
	if got := data.Givens[2][1]; got != 0 {
		t.Fatalf("givens[2][1] = %d, want 0", got)
	}
}
