package cmd

import (
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/FelineStateMachine/puzzletea/app"
	"github.com/FelineStateMachine/puzzletea/game"
	"github.com/FelineStateMachine/puzzletea/pdfexport"

	"github.com/spf13/cobra"
)

var (
	flagPDFOutput      string
	flagPDFTitle       string
	flagPDFAdvert      string
	flagPDFShuffleSeed string
)

var exportPDFCmd = &cobra.Command{
	Use:   "export-pdf <input.jsonl> [more.jsonl ...]",
	Short: "Convert one or more PuzzleTea JSONL exports into a half-letter printable PDF",
	Long:  "Parse one or more JSONL export files, order puzzles by progressive difficulty with seeded mixing, and render a half-letter PDF with a title page and one puzzle per page.",
	Args:  cobra.MinimumNArgs(1),
	RunE:  runExportPDF,
}

func init() {
	exportPDFCmd.Flags().StringVarP(&flagPDFOutput, "output", "o", "", "write output PDF path (defaults to <first-input>-print.pdf)")
	exportPDFCmd.Flags().StringVar(&flagPDFTitle, "title", "", "title shown on the generated title page")
	exportPDFCmd.Flags().StringVar(&flagPDFAdvert, "advert", "Find more puzzles at github.com/FelineStateMachine/puzzletea", "advert text shown on the title page")
	exportPDFCmd.Flags().StringVar(&flagPDFShuffleSeed, "shuffle-seed", "", "seed for deterministic within-band difficulty mixing")
}

func runExportPDF(cmd *cobra.Command, args []string) error {
	docs, err := pdfexport.ParseJSONLFiles(args)
	if err != nil {
		return err
	}

	puzzles := flattenPuzzles(docs)
	if len(puzzles) == 0 {
		return fmt.Errorf("no puzzles found in input jsonl files")
	}

	lookup := buildModeDifficultyLookup(app.Categories)
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

	title := strings.TrimSpace(flagPDFTitle)
	if title == "" {
		title = defaultPDFTitle(docs)
	}

	cfg := pdfexport.RenderConfig{
		Title:       title,
		AdvertText:  flagPDFAdvert,
		GeneratedAt: time.Now(),
		ShuffleSeed: shuffleSeed,
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

func buildModeDifficultyLookup(categories []game.Category) map[string]map[string]float64 {
	lookup := make(map[string]map[string]float64, len(categories))

	for _, cat := range categories {
		titles := []string{}
		for _, item := range cat.Modes {
			mode, ok := item.(game.Mode)
			if !ok {
				continue
			}
			title := strings.TrimSpace(mode.Title())
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

		lookup[normalizeDifficultyToken(cat.Name)] = scores
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
	s = strings.ToLower(strings.TrimSpace(s))
	s = strings.ReplaceAll(s, "-", " ")
	s = strings.ReplaceAll(s, "_", " ")
	return strings.Join(strings.Fields(s), " ")
}
