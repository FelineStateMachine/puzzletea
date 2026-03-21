package pdfexport

type bookletRenderPlan struct {
	puzzlePages  int
	includeCover bool
	padPages     int
}

type logicalPageKind int

const (
	logicalPageCoverOutside logicalPageKind = iota
	logicalPageCoverBlank
	logicalPageTitle
	logicalPagePuzzle
	logicalPagePad
)

type logicalPage struct {
	Number       int
	Kind         logicalPageKind
	PuzzleIndex  int
	OutsideSlice coverOutsideSlice
	ShowFooter   bool
}

type duplexBookletSpread struct {
	LeftPage  int
	RightPage int
}

type duplexBookletSheet struct {
	Front duplexBookletSpread
	Back  duplexBookletSpread
}

func newBookletRenderPlan(puzzlePages int, includeCover bool) bookletRenderPlan {
	extraPages := 1 // title page
	if includeCover {
		extraPages += 4
	}

	return bookletRenderPlan{
		puzzlePages:  puzzlePages,
		includeCover: includeCover,
		padPages:     saddleStitchPadCount(puzzlePages + extraPages),
	}
}

func (p bookletRenderPlan) totalPagesWithoutPadding() int {
	total := p.puzzlePages + 1 // title page
	if p.includeCover {
		total += 4
	}
	return total
}

func (p bookletRenderPlan) totalPages() int {
	return p.totalPagesWithoutPadding() + p.padPages
}

func (p bookletRenderPlan) titlePageNumber() int {
	if p.includeCover {
		return 3
	}
	return 1
}

func (p bookletRenderPlan) puzzleStartPage() int {
	return p.titlePageNumber() + 1
}

func (p bookletRenderPlan) firstPadPage() int {
	return p.puzzleStartPage() + p.puzzlePages
}

func (p bookletRenderPlan) coverPageNumbers() (frontOutside, frontInside, backInside, backOutside int, ok bool) {
	if !p.includeCover {
		return 0, 0, 0, 0, false
	}

	total := p.totalPages()
	return 1, 2, total - 1, total, true
}

func (p bookletRenderPlan) footerExcludedPages() map[int]struct{} {
	excluded := map[int]struct{}{
		p.titlePageNumber(): {},
	}

	for page := p.firstPadPage(); page < p.firstPadPage()+p.padPages; page++ {
		excluded[page] = struct{}{}
	}

	frontOutside, frontInside, backInside, backOutside, ok := p.coverPageNumbers()
	if ok {
		excluded[frontOutside] = struct{}{}
		excluded[frontInside] = struct{}{}
		excluded[backInside] = struct{}{}
		excluded[backOutside] = struct{}{}
	}

	return excluded
}

func buildLogicalPages(puzzlePages int, includeCover bool) []logicalPage {
	plan := newBookletRenderPlan(puzzlePages, includeCover)
	excluded := plan.footerExcludedPages()
	pages := make([]logicalPage, 0, plan.totalPages())

	appendPage := func(kind logicalPageKind, puzzleIndex int, slice coverOutsideSlice) {
		pageNo := len(pages) + 1
		_, skipFooter := excluded[pageNo]
		pages = append(pages, logicalPage{
			Number:       pageNo,
			Kind:         kind,
			PuzzleIndex:  puzzleIndex,
			OutsideSlice: slice,
			ShowFooter:   !skipFooter,
		})
	}

	if includeCover {
		appendPage(logicalPageCoverOutside, -1, coverOutsideFront)
		appendPage(logicalPageCoverBlank, -1, coverOutsideFront)
	}

	appendPage(logicalPageTitle, -1, coverOutsideFront)

	for i := range puzzlePages {
		appendPage(logicalPagePuzzle, i, coverOutsideFront)
	}
	for range plan.padPages {
		appendPage(logicalPagePad, -1, coverOutsideFront)
	}

	if includeCover {
		appendPage(logicalPageCoverBlank, -1, coverOutsideBack)
		appendPage(logicalPageCoverOutside, -1, coverOutsideBack)
	}

	return pages
}

func duplexBookletSheets(totalPages int) []duplexBookletSheet {
	if totalPages <= 0 || totalPages%4 != 0 {
		return nil
	}

	sheets := make([]duplexBookletSheet, 0, totalPages/4)
	for sheetIndex := 0; sheetIndex < totalPages/4; sheetIndex++ {
		sheets = append(sheets, duplexBookletSheet{
			Front: duplexBookletSpread{
				LeftPage:  totalPages - 2*sheetIndex,
				RightPage: 1 + 2*sheetIndex,
			},
			Back: duplexBookletSpread{
				LeftPage:  2 + 2*sheetIndex,
				RightPage: totalPages - 1 - 2*sheetIndex,
			},
		})
	}
	return sheets
}
