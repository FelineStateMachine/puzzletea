package pdfexport

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"codeberg.org/go-pdf/fpdf"
	"github.com/FelineStateMachine/puzzletea/game"
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

	pdf := fpdf.NewCustom(&fpdf.InitType{
		OrientationStr: "P",
		UnitStr:        "mm",
		Size: fpdf.SizeType{
			Wd: halfLetterWidthMM,
			Ht: halfLetterHeightMM,
		},
	})
	if err := registerPDFFonts(pdf); err != nil {
		return err
	}
	pdf.SetAutoPageBreak(false, 0)
	pdf.SetCreator("PuzzleTea", true)
	pdf.SetAuthor("PuzzleTea", true)
	pdf.SetTitle(cfg.Title, true)
	footerExcludedPages := map[int]struct{}{}
	pdf.SetFooterFunc(func() {
		pageNo := pdf.PageNo()
		if pageNo <= 2 {
			return
		}
		if _, skip := footerExcludedPages[pageNo]; skip {
			return
		}
		pdf.SetY(-8)
		pdf.SetFont(sansFontFamily, "", 8)
		pdf.SetTextColor(footerTextGray, footerTextGray, footerTextGray)
		pdf.CellFormat(0, 4, strconv.Itoa(pageNo), "", 0, "C", false, 0, "")
	})

	coverColor := resolveCoverColor(cfg)
	renderCoverPage(pdf, printablePuzzles, cfg)
	renderTitlePage(pdf, docs, printablePuzzles, cfg)
	for _, puzzle := range printablePuzzles {
		if err := renderPuzzlePage(pdf, puzzle); err != nil {
			return fmt.Errorf("render puzzle %q (%s #%d): %w", puzzle.Name, puzzle.Category, puzzle.Index, err)
		}
	}

	totalPagesWithoutPadding := pdf.PageNo() + 1 // include upcoming back cover
	for range saddleStitchPadCount(totalPagesWithoutPadding) {
		renderPadPage(pdf)
		footerExcludedPages[pdf.PageNo()] = struct{}{}
	}

	renderBackCoverPage(pdf, cfg, coverColor)
	footerExcludedPages[pdf.PageNo()] = struct{}{}

	dir := filepath.Dir(outputPath)
	if dir != "" && dir != "." {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return fmt.Errorf("create output directory: %w", err)
		}
	}

	if err := pdf.OutputFileAndClose(outputPath); err != nil {
		return fmt.Errorf("write pdf file: %w", err)
	}
	return nil
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
		if game.IsNilPrintPayload(puzzle.PrintPayload) {
			continue
		}
		if _, ok := game.LookupPrintAdapter(puzzle.Category); !ok {
			continue
		}
		printable = append(printable, puzzle)
	}
	return printable
}

func renderPadPage(pdf *fpdf.Fpdf) {
	pdf.AddPage()
}

func renderPuzzlePage(pdf *fpdf.Fpdf, puzzle Puzzle) error {
	if game.IsNilPrintPayload(puzzle.PrintPayload) {
		return nil
	}
	adapter, ok := game.LookupPrintAdapter(puzzle.Category)
	if !ok {
		return nil
	}

	pdf.AddPage()
	pageW, _ := pdf.GetPageSize()

	setPuzzleTitleStyle(pdf)
	pdf.SetXY(0, 10)
	title := fmt.Sprintf("%s %d: %s", puzzle.Category, puzzle.Index, puzzle.Name)
	pdf.CellFormat(pageW, 7, title, "", 0, "C", false, 0, "")

	setPuzzleSubtitleStyle(pdf)
	pdf.SetXY(0, 17)
	subtitleParts := []string{fmt.Sprintf("Difficulty Score: %d/10", difficultyScoreOutOfTen(puzzle.DifficultyScore))}
	if !isMixedModes(puzzle.ModeSelection) {
		subtitleParts = append([]string{fmt.Sprintf("Mode: %s", puzzle.ModeSelection)}, subtitleParts...)
	}
	subtitle := strings.Join(subtitleParts, " | ")
	pdf.CellFormat(pageW, 5, subtitle, "", 0, "C", false, 0, "")
	if err := adapter.RenderPDFBody(pdf, puzzle.PrintPayload); err != nil {
		return err
	}
	return nil
}
