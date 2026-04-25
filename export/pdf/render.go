package pdfexport

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"codeberg.org/go-pdf/fpdf"
	"codeberg.org/go-pdf/fpdf/contrib/gofpdi"
)

func WritePDF(outputPath string, docs []PackDocument, puzzles []Puzzle, cfg RenderConfig) error {
	if strings.TrimSpace(outputPath) == "" {
		return fmt.Errorf("output path is required")
	}

	printablePuzzles := filterPrintablePuzzles(puzzles)
	if len(printablePuzzles) == 0 {
		return nil
	}

	if cfg.GeneratedAt.IsZero() {
		cfg.GeneratedAt = time.Now()
	}
	if cfg.VolumeNumber < 1 {
		cfg.VolumeNumber = 1
	}
	if strings.TrimSpace(cfg.CoverSubtitle) == "" {
		cfg.CoverSubtitle = defaultTitle(docs)
	}
	if strings.TrimSpace(cfg.Title) == "" {
		cfg.Title = fmt.Sprintf("PuzzleTea Volume %02d", cfg.VolumeNumber)
	}
	if strings.TrimSpace(cfg.AdvertText) == "" {
		cfg.AdvertText = "Find more puzzles at github.com/FelineStateMachine/puzzletea"
	}

	logicalPages := buildLogicalPages(len(printablePuzzles), sheetLayoutIncludesCover(cfg.SheetLayout))

	dir := filepath.Dir(outputPath)
	if dir != "" && dir != "." {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return fmt.Errorf("create output directory: %w", err)
		}
	}

	switch cfg.SheetLayout {
	case SheetLayoutHalfLetter:
		pdf, err := newConfiguredPDF(SheetLayoutHalfLetter, cfg.Title, true)
		if err != nil {
			return err
		}
		if err := renderLogicalPages(pdf, logicalPages, docs, printablePuzzles, cfg); err != nil {
			return err
		}
		if got, want := pdf.PageNo(), len(logicalPages); got != want {
			return fmt.Errorf("rendered %d physical pages, want %d", got, want)
		}
		if err := pdf.OutputFileAndClose(outputPath); err != nil {
			return fmt.Errorf("write pdf file: %w", err)
		}
	case SheetLayoutDuplexBooklet:
		if err := writeDuplexBookletPDF(outputPath, logicalPages, docs, printablePuzzles, cfg); err != nil {
			return err
		}
	default:
		return fmt.Errorf("unsupported sheet layout %d", cfg.SheetLayout)
	}
	return nil
}

func newRenderPDF(layout SheetLayout) *fpdf.Fpdf {
	init := &fpdf.InitType{
		OrientationStr: "P",
		UnitStr:        "mm",
		Size: fpdf.SizeType{
			Wd: halfLetterWidthMM,
			Ht: halfLetterHeightMM,
		},
	}
	if layout == SheetLayoutDuplexBooklet {
		init.Size = fpdf.SizeType{
			Wd: letterWidthMM,
			Ht: letterHeightMM,
		}
	}
	return fpdf.NewCustom(init)
}

func newConfiguredPDF(layout SheetLayout, title string, registerFonts bool) (*fpdf.Fpdf, error) {
	pdf := newRenderPDF(layout)
	if registerFonts {
		if err := registerPDFFonts(pdf); err != nil {
			return nil, err
		}
	}
	pdf.SetAutoPageBreak(false, 0)
	pdf.SetCreator("PuzzleTea", true)
	pdf.SetAuthor("PuzzleTea", true)
	pdf.SetTitle(title, true)
	return pdf, nil
}

func sheetLayoutIncludesCover(layout SheetLayout) bool {
	return layout == SheetLayoutDuplexBooklet
}

func saddleStitchPadCount(totalPages int) int {
	if totalPages <= 0 {
		return 0
	}

	remainder := totalPages % 4
	if remainder == 0 {
		return 0
	}
	return 4 - remainder
}

func filterPrintablePuzzles(puzzles []Puzzle) []Puzzle {
	printable := make([]Puzzle, 0, len(puzzles))
	for _, puzzle := range puzzles {
		if IsNilPrintPayload(puzzle.PrintPayload) {
			continue
		}
		if _, ok := LookupPrintAdapter(puzzle.Category); !ok {
			continue
		}
		printable = append(printable, puzzle)
	}
	return printable
}

func renderLogicalPages(
	pdf *fpdf.Fpdf,
	pages []logicalPage,
	docs []PackDocument,
	puzzles []Puzzle,
	cfg RenderConfig,
) error {
	for _, page := range pages {
		pdf.AddPage()
		if err := drawLogicalPage(pdf, page, docs, puzzles, cfg); err != nil {
			return err
		}
	}
	return nil
}

func drawLogicalPage(
	pdf *fpdf.Fpdf,
	page logicalPage,
	docs []PackDocument,
	puzzles []Puzzle,
	cfg RenderConfig,
) error {
	return withLogicalPageNumber(page.Number, func() error {
		switch page.Kind {
		case logicalPageCoverOutside:
			renderOutsideCoverPage(pdf, cfg, page.OutsideSlice)
		case logicalPageCoverBlank:
			renderCoverBlankPage(pdf)
		case logicalPageTitle:
			renderTitlePage(pdf, docs, puzzles, cfg)
		case logicalPagePuzzle:
			if page.PuzzleIndex < 0 || page.PuzzleIndex >= len(puzzles) {
				return fmt.Errorf("logical page %d references puzzle index %d", page.Number, page.PuzzleIndex)
			}
			puzzle := puzzles[page.PuzzleIndex]
			if err := renderPuzzlePage(pdf, puzzle); err != nil {
				return fmt.Errorf("render puzzle %q (%s #%d): %w", puzzle.Name, puzzle.Category, puzzle.Index, err)
			}
		case logicalPagePad:
			renderPadPage(pdf)
		default:
			return fmt.Errorf("unsupported logical page kind %d", page.Kind)
		}

		if page.ShowFooter {
			drawPageFooter(pdf, page.Number)
		}
		return nil
	})
}

func renderPadPage(_ *fpdf.Fpdf) {}

func writeDuplexBookletPDF(
	outputPath string,
	logicalPages []logicalPage,
	docs []PackDocument,
	puzzles []Puzzle,
	cfg RenderConfig,
) error {
	tempFile, err := os.CreateTemp("", "puzzletea-booklet-*.pdf")
	if err != nil {
		return fmt.Errorf("create temporary booklet pdf: %w", err)
	}
	tempPath := tempFile.Name()
	if err := tempFile.Close(); err != nil {
		return fmt.Errorf("close temporary booklet pdf handle: %w", err)
	}
	defer os.Remove(tempPath)

	sourcePDF, err := newConfiguredPDF(SheetLayoutHalfLetter, cfg.Title, true)
	if err != nil {
		return err
	}
	if err := renderLogicalPages(sourcePDF, logicalPages, docs, puzzles, cfg); err != nil {
		return err
	}
	if got, want := sourcePDF.PageNo(), len(logicalPages); got != want {
		return fmt.Errorf("rendered %d logical pages, want %d", got, want)
	}
	if err := sourcePDF.OutputFileAndClose(tempPath); err != nil {
		return fmt.Errorf("write temporary booklet pdf: %w", err)
	}

	imposedPDF, err := newConfiguredPDF(SheetLayoutDuplexBooklet, cfg.Title, false)
	if err != nil {
		return err
	}
	importer := gofpdi.NewImporter()
	for _, sheet := range duplexBookletSheets(len(logicalPages)) {
		imposedPDF.AddPage()
		drawImportedPage(imposedPDF, importer, tempPath, sheet.Front.LeftPage, 0, 0)
		drawImportedPage(imposedPDF, importer, tempPath, sheet.Front.RightPage, halfLetterWidthMM, 0)

		imposedPDF.AddPage()
		drawImportedPage(imposedPDF, importer, tempPath, sheet.Back.LeftPage, 0, 0)
		drawImportedPage(imposedPDF, importer, tempPath, sheet.Back.RightPage, halfLetterWidthMM, 0)
	}
	if got, want := imposedPDF.PageNo(), len(logicalPages)/2; got != want {
		return fmt.Errorf("rendered %d physical pages, want %d", got, want)
	}
	if err := imposedPDF.OutputFileAndClose(outputPath); err != nil {
		return fmt.Errorf("write pdf file: %w", err)
	}
	return nil
}

func drawImportedPage(
	pdf *fpdf.Fpdf,
	importer *gofpdi.Importer,
	sourcePath string,
	pageNumber int,
	x, y float64,
) {
	templateID := importer.ImportPage(pdf, sourcePath, pageNumber, "/MediaBox")
	importer.UseImportedTemplate(pdf, templateID, x, y, halfLetterWidthMM, halfLetterHeightMM)
}

func drawPageFooter(pdf *fpdf.Fpdf, pageNo int) {
	pdf.SetY(-8)
	pdf.SetFont(sansFontFamily, "", 8+pdfFontSizeDelta)
	pdf.SetTextColor(footerTextGray, footerTextGray, footerTextGray)
	pdf.CellFormat(0, 4, fmt.Sprintf("%d", pageNo), "", 0, "C", false, 0, "")
}

func renderPuzzlePage(pdf *fpdf.Fpdf, puzzle Puzzle) error {
	if IsNilPrintPayload(puzzle.PrintPayload) {
		return nil
	}
	adapter, ok := LookupPrintAdapter(puzzle.Category)
	if !ok {
		return nil
	}

	pageW, _ := pdf.GetPageSize()

	setPuzzleTitleStyle(pdf)
	pdf.SetXY(0, 10)
	title := strings.TrimSpace(puzzle.Category)
	if name := strings.TrimSpace(puzzle.Name); name != "" {
		title = fmt.Sprintf("%s: %s", title, name)
	}
	pdf.CellFormat(pageW, 7, title, "", 0, "C", false, 0, "")

	setPuzzleSubtitleStyle(pdf)
	renderPuzzleDifficultySubtitle(pdf, pageW, 17, puzzle)
	if err := adapter.RenderPDFBody(pdf, puzzle.PrintPayload); err != nil {
		return err
	}
	return nil
}
