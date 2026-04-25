package pdfexport

import (
	"math"
	"testing"
)

func TestThinGridLineFloor(t *testing.T) {
	if thinGridLineMM < 0.22 {
		t.Fatalf("thinGridLineMM = %.2f, want >= 0.22", thinGridLineMM)
	}
}

func TestCenteredOriginKeepsBoardCentered(t *testing.T) {
	pageArea := puzzleBoardRect(halfLetterWidthMM, halfLetterHeightMM, 2, 1)
	tests := []struct {
		name   string
		cols   int
		rows   int
		family boardFamily
	}{
		{name: "compact", cols: 8, rows: 8, family: boardFamilyCompact},
		{name: "sudoku", cols: 9, rows: 9, family: boardFamilySudoku},
		{name: "hashi-span", cols: 6, rows: 6, family: boardFamilyHashi},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cell := fitBoardCellSize(tt.cols, tt.rows, pageArea, tt.family)
			if cell <= 0 {
				t.Fatal("expected positive cell size")
			}
			x, y := centeredOrigin(pageArea, tt.cols, tt.rows, cell)

			centerX := x + float64(tt.cols)*cell/2
			centerY := y + float64(tt.rows)*cell/2
			wantX := pageArea.x + pageArea.w/2
			wantY := pageArea.y + pageArea.h/2
			if diff := math.Abs(centerX - wantX); diff > 0.01 {
				t.Fatalf("centerX diff = %.4f, want <= 0.01", diff)
			}
			if diff := math.Abs(centerY - wantY); diff > 0.01 {
				t.Fatalf("centerY diff = %.4f, want <= 0.01", diff)
			}
		})
	}
}

func TestPuzzleBodyRectMirrorsGutterByParity(t *testing.T) {
	even := puzzleBodyRect(halfLetterWidthMM, halfLetterHeightMM, 2)
	odd := puzzleBodyRect(halfLetterWidthMM, halfLetterHeightMM, 3)

	if even.w <= 0 || odd.w <= 0 {
		t.Fatal("expected positive body widths")
	}
	if math.Abs(even.w-odd.w) > 0.001 {
		t.Fatalf("body widths differ: even=%.3f odd=%.3f", even.w, odd.w)
	}

	evenCenterX := even.x + even.w/2
	oddCenterX := odd.x + odd.w/2
	pageCenterX := halfLetterWidthMM / 2

	if evenCenterX >= pageCenterX {
		t.Fatalf("even page center = %.3f, want left of %.3f", evenCenterX, pageCenterX)
	}
	if oddCenterX <= pageCenterX {
		t.Fatalf("odd page center = %.3f, want right of %.3f", oddCenterX, pageCenterX)
	}
}

func TestSaddleStitchPadCount(t *testing.T) {
	tests := []struct {
		name  string
		pages int
		want  int
	}{
		{name: "already multiple of four", pages: 36, want: 0},
		{name: "remainder one", pages: 35, want: 1},
		{name: "remainder two", pages: 34, want: 2},
		{name: "remainder three", pages: 33, want: 3},
		{name: "non-positive", pages: 0, want: 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := saddleStitchPadCount(tt.pages)
			if got != tt.want {
				t.Fatalf("saddleStitchPadCount(%d) = %d, want %d", tt.pages, got, tt.want)
			}
		})
	}
}

func TestSaddleStitchPadCountForStandardPackLayout(t *testing.T) {
	tests := []struct {
		name       string
		puzzleRows int
		wantPad    int
	}{
		{name: "single puzzle", puzzleRows: 1, wantPad: 2},
		{name: "two puzzles", puzzleRows: 2, wantPad: 1},
		{name: "thirty-two puzzles", puzzleRows: 32, wantPad: 3},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			totalWithoutPad := tt.puzzleRows + 5 // outside front + inside front + title + inside back + outside back
			got := saddleStitchPadCount(totalWithoutPad)
			if got != tt.wantPad {
				t.Fatalf("pad pages = %d, want %d", got, tt.wantPad)
			}
			if (totalWithoutPad+got)%4 != 0 {
				t.Fatalf("total pages = %d, want multiple of 4", totalWithoutPad+got)
			}
		})
	}
}

func TestBookletRenderPlanWithCoverUsesFirstTwoAndLastTwoPages(t *testing.T) {
	plan := newBookletRenderPlan(32, true)

	if got, want := plan.titlePageNumber(), 3; got != want {
		t.Fatalf("title page = %d, want %d", got, want)
	}
	if got, want := plan.totalPages(), 40; got != want {
		t.Fatalf("total pages = %d, want %d", got, want)
	}
	if got, want := plan.padPages, 3; got != want {
		t.Fatalf("pad pages = %d, want %d", got, want)
	}

	frontOutside, frontInside, backInside, backOutside, ok := plan.coverPageNumbers()
	if !ok {
		t.Fatal("expected cover page numbers")
	}
	if frontOutside != 1 || frontInside != 2 || backInside != 39 || backOutside != 40 {
		t.Fatalf("cover pages = (%d,%d,%d,%d), want (1,2,39,40)", frontOutside, frontInside, backInside, backOutside)
	}

	excluded := plan.footerExcludedPages()
	for _, page := range []int{1, 2, 3, 36, 37, 38, 39, 40} {
		if _, ok := excluded[page]; !ok {
			t.Fatalf("expected page %d to be footer-excluded", page)
		}
	}
	for _, page := range []int{4, 20, 35} {
		if _, ok := excluded[page]; ok {
			t.Fatalf("did not expect puzzle page %d to be footer-excluded", page)
		}
	}
}

func TestBookletRenderPlanWithoutCoverKeepsTitleFirst(t *testing.T) {
	plan := newBookletRenderPlan(2, false)

	if got, want := plan.titlePageNumber(), 1; got != want {
		t.Fatalf("title page = %d, want %d", got, want)
	}
	if got, want := plan.totalPages(), 4; got != want {
		t.Fatalf("total pages = %d, want %d", got, want)
	}

	excluded := plan.footerExcludedPages()
	for _, page := range []int{1, 4} {
		if _, ok := excluded[page]; !ok {
			t.Fatalf("expected page %d to be footer-excluded", page)
		}
	}
	if _, ok := excluded[2]; ok {
		t.Fatal("did not expect first puzzle page to be footer-excluded")
	}
}

func TestBuildLogicalPagesWithCoverUsesExpectedSequence(t *testing.T) {
	pages := buildLogicalPages(32, true)
	if got, want := len(pages), 40; got != want {
		t.Fatalf("logical pages = %d, want %d", got, want)
	}

	if pages[0].Kind != logicalPageCoverOutside || pages[0].OutsideSlice != coverOutsideFront {
		t.Fatal("page 1 should be front outside cover")
	}
	if pages[1].Kind != logicalPageCoverBlank {
		t.Fatal("page 2 should be inside front blank")
	}
	if pages[2].Kind != logicalPageTitle {
		t.Fatal("page 3 should be title")
	}
	if pages[len(pages)-2].Kind != logicalPageCoverBlank {
		t.Fatal("second-last page should be inside back blank")
	}
	if pages[len(pages)-1].Kind != logicalPageCoverOutside || pages[len(pages)-1].OutsideSlice != coverOutsideBack {
		t.Fatal("last page should be back outside cover")
	}
}

func TestDuplexBookletSheetsForFourPageBooklet(t *testing.T) {
	sheets := duplexBookletSheets(4)
	if got, want := len(sheets), 1; got != want {
		t.Fatalf("sheet count = %d, want %d", got, want)
	}
	if sheets[0].Front.LeftPage != 4 || sheets[0].Front.RightPage != 1 {
		t.Fatalf("front pair = (%d,%d), want (4,1)", sheets[0].Front.LeftPage, sheets[0].Front.RightPage)
	}
	if sheets[0].Back.LeftPage != 2 || sheets[0].Back.RightPage != 3 {
		t.Fatalf("back pair = (%d,%d), want (2,3)", sheets[0].Back.LeftPage, sheets[0].Back.RightPage)
	}
}

func TestDuplexBookletSheetsForEightPageBooklet(t *testing.T) {
	sheets := duplexBookletSheets(8)
	if got, want := len(sheets), 2; got != want {
		t.Fatalf("sheet count = %d, want %d", got, want)
	}
	if sheets[0].Front.LeftPage != 8 || sheets[0].Front.RightPage != 1 {
		t.Fatalf("outer front pair = (%d,%d), want (8,1)", sheets[0].Front.LeftPage, sheets[0].Front.RightPage)
	}
	if sheets[1].Back.LeftPage != 4 || sheets[1].Back.RightPage != 5 {
		t.Fatalf("inner back pair = (%d,%d), want (4,5)", sheets[1].Back.LeftPage, sheets[1].Back.RightPage)
	}
}

func TestDuplexBookletSheetsWithCoverOuterSheetPairs(t *testing.T) {
	pages := buildLogicalPages(32, true)
	sheets := duplexBookletSheets(len(pages))
	if got, want := len(sheets), 10; got != want {
		t.Fatalf("sheet count = %d, want %d", got, want)
	}
	if sheets[0].Front.LeftPage != 40 || sheets[0].Front.RightPage != 1 {
		t.Fatalf("outer front pair = (%d,%d), want (40,1)", sheets[0].Front.LeftPage, sheets[0].Front.RightPage)
	}
	if sheets[0].Back.LeftPage != 2 || sheets[0].Back.RightPage != 39 {
		t.Fatalf("outer back pair = (%d,%d), want (2,39)", sheets[0].Back.LeftPage, sheets[0].Back.RightPage)
	}
}

func TestDuplexBookletPhysicalPageCountIsHalfLogicalPages(t *testing.T) {
	for _, total := range []int{4, 8, 40} {
		sheets := duplexBookletSheets(total)
		got := len(sheets) * 2
		want := total / 2
		if got != want {
			t.Fatalf("physical pages for %d logical pages = %d, want %d", total, got, want)
		}
	}
}

func TestNewRenderPDFDuplexBookletUsesLandscapeLetterCanvas(t *testing.T) {
	pdf := newRenderPDF(SheetLayoutDuplexBooklet)
	pageW, pageH := pdf.GetPageSize()

	if pageW <= pageH {
		t.Fatalf("duplex-booklet canvas = %.1fx%.1f, want landscape", pageW, pageH)
	}
	if math.Abs(pageW-letterWidthMM) > 0.01 || math.Abs(pageH-letterHeightMM) > 0.01 {
		t.Fatalf(
			"duplex-booklet canvas = %.1fx%.1f, want %.1fx%.1f",
			pageW,
			pageH,
			letterWidthMM,
			letterHeightMM,
		)
	}
}

func TestSaddleStitchPadCountForTitleOnlyPackLayout(t *testing.T) {
	tests := []struct {
		name       string
		puzzleRows int
		wantPad    int
	}{
		{name: "single puzzle", puzzleRows: 1, wantPad: 2},
		{name: "two puzzles", puzzleRows: 2, wantPad: 1},
		{name: "thirty-two puzzles", puzzleRows: 32, wantPad: 3},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			totalWithoutPad := tt.puzzleRows + 1 // title page + puzzle pages
			got := saddleStitchPadCount(totalWithoutPad)
			if got != tt.wantPad {
				t.Fatalf("pad pages = %d, want %d", got, tt.wantPad)
			}
			if (totalWithoutPad+got)%4 != 0 {
				t.Fatalf("total pages = %d, want multiple of 4", totalWithoutPad+got)
			}
		})
	}
}
