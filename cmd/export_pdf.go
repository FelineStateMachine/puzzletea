package cmd

import (
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/FelineStateMachine/puzzletea/builtinprint"
	"github.com/FelineStateMachine/puzzletea/pdfexport"
	"github.com/FelineStateMachine/puzzletea/puzzle"
	"github.com/FelineStateMachine/puzzletea/registry"

	"github.com/spf13/cobra"
)

var (
	flagPDFOutput      string
	flagPDFTitle       string
	flagPDFHeader      string
	flagPDFVolume      int
	flagPDFAdvert      string
	flagPDFShuffleSeed string
	flagPDFSheetLayout string
)

var exportPDFCmd = &cobra.Command{
	Use:   "export-pdf <input.jsonl> [more.jsonl ...]",
	Short: "Convert one or more PuzzleTea JSONL exports into a printable PDF",
	Long:  "Parse one or more JSONL export files, order puzzles by progressive difficulty with seeded mixing, and render either a half-letter PDF or an imposed duplex-booklet PDF with a title page, one puzzle per logical half-letter page, and automatic page-count padding to a multiple of 4 for booklet printing. The duplex-booklet layout includes the 4-page black-ink cover block automatically.",
	Args:  cobra.MinimumNArgs(1),
	RunE:  runExportPDF,
}

func init() {
	exportPDFCmd.Flags().StringVarP(&flagPDFOutput, "output", "o", "", "write output PDF path (defaults to <first-input>-print.pdf)")
	exportPDFCmd.Flags().StringVar(&flagPDFTitle, "title", "", "subtitle shown on the title page (and on the outside cover when enabled)")
	exportPDFCmd.Flags().StringVar(&flagPDFHeader, "header", "", "optional intro paragraph shown on the title page under 'PuzzleTea Puzzle Pack'")
	exportPDFCmd.Flags().IntVar(&flagPDFVolume, "volume", 1, "volume number shown on the title page (and on the outside cover when enabled) (must be >= 1)")
	exportPDFCmd.Flags().StringVar(&flagPDFAdvert, "advert", "Find more puzzles at github.com/FelineStateMachine/puzzletea", "advert text shown on the title page")
	exportPDFCmd.Flags().StringVar(&flagPDFShuffleSeed, "shuffle-seed", "", "seed for deterministic within-band difficulty mixing")
	exportPDFCmd.Flags().StringVar(&flagPDFSheetLayout, "sheet-layout", "half-letter", "physical PDF layout: half-letter or duplex-booklet (landscape US Letter with two half-letter pages per sheet side; print duplex on short edge). duplex-booklet automatically includes the 4-page cover block")
}

func runExportPDF(cmd *cobra.Command, args []string) error {
	builtinprint.Register()

	docs, err := pdfexport.ParseJSONLFiles(args)
	if err != nil {
		return err
	}

	puzzles := flattenPuzzles(docs)
	if len(puzzles) == 0 {
		return nil
	}

	lookup := buildModeDifficultyLookup(registry.Definitions())
	annotateDifficulty(puzzles, lookup)

	shuffleSeed := strings.TrimSpace(flagPDFShuffleSeed)
	if shuffleSeed == "" {
		shuffleSeed = time.Now().Format(time.RFC3339Nano)
	}
	ordered := pdfexport.OrderPuzzlesForPrint(puzzles, shuffleSeed)

	output := strings.TrimSpace(flagPDFOutput)
	if output == "" {
		base := filepath.Base(args[0])
		output = strings.TrimSuffix(base, filepath.Ext(base)) + "-print.pdf"
	}
	if !strings.EqualFold(filepath.Ext(output), ".pdf") {
		return fmt.Errorf("--output must use a .pdf extension")
	}

	cfg, err := buildRenderConfigForPDF(docs, shuffleSeed, time.Now())
	if err != nil {
		return err
	}
	if err := pdfexport.WritePDF(output, docs, ordered, cfg); err != nil {
		return err
	}

	_, err = fmt.Fprintf(cmd.OutOrStdout(), "wrote %s with %d puzzles\n", output, len(ordered))
	return err
}

func flattenPuzzles(docs []pdfexport.PackDocument) []pdfexport.Puzzle {
	puzzles := []pdfexport.Puzzle{}
	for _, doc := range docs {
		puzzles = append(puzzles, doc.Puzzles...)
	}
	return puzzles
}

func defaultPDFTitle(docs []pdfexport.PackDocument) string {
	if len(docs) == 1 {
		category := strings.TrimSpace(docs[0].Metadata.Category)
		if category != "" {
			return fmt.Sprintf("%s Puzzle Pack", category)
		}
	}
	return "PuzzleTea Mixed Puzzle Pack"
}

func validatePDFVolume(volume int) error {
	if volume < 1 {
		return fmt.Errorf("--volume must be >= 1")
	}
	return nil
}

func buildRenderConfigForPDF(docs []pdfexport.PackDocument, shuffleSeed string, now time.Time) (pdfexport.RenderConfig, error) {
	if err := validatePDFVolume(flagPDFVolume); err != nil {
		return pdfexport.RenderConfig{}, err
	}
	sheetLayout, err := parsePDFSheetLayout(flagPDFSheetLayout)
	if err != nil {
		return pdfexport.RenderConfig{}, err
	}

	subtitle := strings.TrimSpace(flagPDFTitle)
	if subtitle == "" {
		subtitle = defaultPDFTitle(docs)
	}

	cfg := pdfexport.RenderConfig{
		CoverSubtitle: subtitle,
		HeaderText:    strings.TrimSpace(flagPDFHeader),
		VolumeNumber:  flagPDFVolume,
		AdvertText:    flagPDFAdvert,
		GeneratedAt:   now,
		ShuffleSeed:   shuffleSeed,
		SheetLayout:   sheetLayout,
	}
	return cfg, nil
}

func parsePDFSheetLayout(raw string) (pdfexport.SheetLayout, error) {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "", "half-letter":
		return pdfexport.SheetLayoutHalfLetter, nil
	case "duplex-booklet":
		return pdfexport.SheetLayoutDuplexBooklet, nil
	default:
		return pdfexport.SheetLayoutHalfLetter, fmt.Errorf("--sheet-layout must be half-letter or duplex-booklet")
	}
}

func buildModeDifficultyLookup(definitions []puzzle.Definition) map[string]map[string]float64 {
	lookup := make(map[string]map[string]float64, len(definitions))

	for _, def := range definitions {
		titles := []string{}
		for _, mode := range def.Modes {
			title := strings.TrimSpace(mode.Title)
			if title == "" {
				continue
			}
			titles = append(titles, title)
		}

		if len(titles) == 0 {
			continue
		}

		scores := make(map[string]float64, len(titles))
		if len(titles) == 1 {
			scores[normalizeDifficultyToken(titles[0])] = 0.5
		} else {
			for i, title := range titles {
				scores[normalizeDifficultyToken(title)] = float64(i) / float64(len(titles)-1)
			}
		}

		lookup[normalizeDifficultyToken(def.Name)] = scores
	}

	return lookup
}

func annotateDifficulty(puzzles []pdfexport.Puzzle, lookup map[string]map[string]float64) {
	for i := range puzzles {
		mode := normalizeDifficultyToken(puzzles[i].ModeSelection)
		if mode == "" || strings.Contains(mode, "mixed modes") {
			puzzles[i].DifficultyScore = 0.5
			puzzles[i].DifficultyConfidence = pdfexport.DifficultyConfidenceMedium
			puzzles[i].DifficultySource = "mixed-mode fallback"
			continue
		}

		category := normalizeDifficultyToken(puzzles[i].Category)
		modes, ok := lookup[category]
		if !ok {
			puzzles[i].DifficultyScore = 0.5
			puzzles[i].DifficultyConfidence = pdfexport.DifficultyConfidenceMedium
			puzzles[i].DifficultySource = "category lookup fallback"
			continue
		}

		score, ok := modes[mode]
		if !ok {
			puzzles[i].DifficultyScore = 0.5
			puzzles[i].DifficultyConfidence = pdfexport.DifficultyConfidenceMedium
			puzzles[i].DifficultySource = "mode lookup fallback"
			continue
		}

		puzzles[i].DifficultyScore = score
		puzzles[i].DifficultyConfidence = pdfexport.DifficultyConfidenceHigh
		puzzles[i].DifficultySource = "mode-order"
	}
}

func normalizeDifficultyToken(s string) string {
	return puzzle.NormalizeName(s)
}
