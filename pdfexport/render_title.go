package pdfexport

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"codeberg.org/go-pdf/fpdf"
)

func renderTitlePage(pdf *fpdf.Fpdf, docs []PackDocument, puzzles []Puzzle, cfg RenderConfig) {
	pageW, pageH := pdf.GetPageSize()
	margin := 12.0
	contentWidth := pageW - 2*margin
	categoryTotals := summarizeCategoryTotals(puzzles)

	pdf.SetTextColor(20, 20, 20)
	pdf.SetFont(sansFontFamily, "B", 22+pdfFontSizeDelta)
	pdf.SetXY(0, 24)
	pdf.CellFormat(pageW, 10, fmt.Sprintf("PuzzleTea Volume %02d", cfg.VolumeNumber), "", 0, "C", false, 0, "")

	pdf.SetTextColor(50, 50, 50)
	pdf.SetFont(coverFontFamily, "", 16+pdfFontSizeDelta)
	pdf.SetXY(0, 35)
	pdf.CellFormat(pageW, 8, cfg.CoverSubtitle, "", 0, "C", false, 0, "")

	bodyY := 49.0
	if header := strings.TrimSpace(cfg.HeaderText); header != "" {
		pdf.SetFont(sansFontFamily, "", 9.2+pdfFontSizeDelta)
		pdf.SetTextColor(74, 74, 74)
		wrappedHeader := pdf.SplitLines([]byte(header), contentWidth-20)
		if len(wrappedHeader) == 0 {
			wrappedHeader = [][]byte{[]byte(header)}
		}
		for _, line := range wrappedHeader {
			pdf.SetXY(margin, bodyY)
			pdf.CellFormat(contentWidth, 4.6, string(line), "", 0, "C", false, 0, "")
			bodyY += 4.6
		}
		bodyY += 2.0
	}

	introLines := splitCoverTextLines(pdf, cfg.AdvertText, contentWidth)
	if len(introLines) > 0 {
		pdf.SetTextColor(58, 58, 58)
		pdf.SetFont(sansFontFamily, "", 9.3+pdfFontSizeDelta)
		for _, line := range introLines {
			pdf.SetXY(margin, bodyY)
			pdf.CellFormat(contentWidth, 4.8, line, "", 0, "C", false, 0, "")
			bodyY += 4.8
		}
		bodyY += 3.8
	}

	pdf.SetTextColor(40, 40, 40)
	pdf.SetFont(sansFontFamily, "B", 9.6+pdfFontSizeDelta)
	pdf.SetXY(margin, bodyY)
	pdf.CellFormat(contentWidth, 5.4, fmt.Sprintf("%d puzzles across %d categories", len(puzzles), len(categoryTotals)), "", 0, "C", false, 0, "")
	bodyY += 7.6

	pdf.SetFont(sansFontFamily, "B", 10+pdfFontSizeDelta)
	pdf.SetTextColor(45, 45, 45)
	pdf.SetXY(margin, bodyY)
	pdf.CellFormat(contentWidth, 6, "Inside This Volume", "", 0, "C", false, 0, "")
	bodyY += 7.0

	renderCategoryOverview(pdf, categoryTotals, margin, bodyY, contentWidth, pageH-32)

	pdf.SetTextColor(50, 50, 50)
	pdf.SetFont(sansFontFamily, "B", 12+pdfFontSizeDelta)
	footerTitleY := pageH - 31.0
	pdf.SetXY(margin, footerTitleY)
	pdf.CellFormat(contentWidth, 7, "Made with PuzzleTea", "", 0, "C", false, 0, "")

	colophon := titlePageColophon(docs, cfg.GeneratedAt)
	if colophon != "" {
		pdf.SetFont(sansFontFamily, "", 7.8+pdfFontSizeDelta)
		pdf.SetTextColor(112, 112, 112)
		pdf.SetXY(margin, footerTitleY+8.0)
		pdf.CellFormat(contentWidth, 4.0, colophon, "", 0, "C", false, 0, "")
	}
}

func summarizeVersions(docs []PackDocument) []string {
	set := map[string]struct{}{}
	for _, doc := range docs {
		version := strings.TrimSpace(doc.Metadata.Version)
		if version == "" {
			continue
		}
		set[version] = struct{}{}
	}

	versions := make([]string, 0, len(set))
	for version := range set {
		versions = append(versions, version)
	}
	sort.Strings(versions)
	if len(versions) == 0 {
		return []string{"Unknown"}
	}
	return versions
}

func defaultTitle(docs []PackDocument) string {
	if len(docs) == 1 {
		category := strings.TrimSpace(docs[0].Metadata.Category)
		if category != "" {
			return fmt.Sprintf("%s Puzzle Pack", category)
		}
	}
	return "PuzzleTea Mixed Puzzle Pack"
}

type namedCount struct {
	Name  string
	Count int
}

func summarizeCategoryTotals(puzzles []Puzzle) []namedCount {
	m := make(map[string]*namedCount)
	for _, puzzle := range puzzles {
		category := strings.TrimSpace(puzzle.Category)
		if category == "" {
			category = "Unknown"
		}
		key := normalizeToken(category)
		if entry, ok := m[key]; ok {
			entry.Count++
			continue
		}
		m[key] = &namedCount{Name: category, Count: 1}
	}
	return sortedNamedCounts(m)
}

func sortedNamedCounts(m map[string]*namedCount) []namedCount {
	counts := make([]namedCount, 0, len(m))
	for _, entry := range m {
		counts = append(counts, *entry)
	}
	sort.SliceStable(counts, func(i, j int) bool {
		return strings.Compare(normalizeToken(counts[i].Name), normalizeToken(counts[j].Name)) < 0
	})
	return counts
}

func renderCategoryOverview(
	pdf *fpdf.Fpdf,
	categoryTotals []namedCount,
	x, y, width, maxY float64,
) float64 {
	const (
		lineHeight  = 4.8
		categoryGap = 1.2
		columnGap   = 6.0
		maxWidth    = 96.0
		widthScale  = 0.82
	)

	if width <= 0 || y >= maxY || len(categoryTotals) == 0 {
		return y
	}

	pdf.SetTextColor(ruleTextGray, ruleTextGray, ruleTextGray)
	pdf.SetFont(sansFontFamily, "", 9.0+pdfFontSizeDelta)

	containerWidth := min(width, min(maxWidth, width*widthScale))
	containerX := x + (width-containerWidth)/2
	colWidth := (containerWidth - columnGap) / 2
	leftCount := (len(categoryTotals) + 1) / 2
	maxRows := max(leftCount, len(categoryTotals)-leftCount)
	curY := y
	for row := 0; row < maxRows; row++ {
		rowY := curY + float64(row)*(lineHeight+categoryGap)
		if rowY+lineHeight > maxY {
			return rowY
		}
		if row < leftCount {
			pdf.SetXY(containerX, rowY)
			pdf.CellFormat(colWidth, lineHeight, formatCountLabel(categoryTotals[row]), "", 0, "R", false, 0, "")
		}
		rightIndex := leftCount + row
		if rightIndex < len(categoryTotals) {
			pdf.SetXY(containerX+colWidth+columnGap, rowY)
			pdf.CellFormat(colWidth, lineHeight, formatCountLabel(categoryTotals[rightIndex]), "", 0, "L", false, 0, "")
		}
	}

	return curY + float64(maxRows)*(lineHeight+categoryGap)
}

func formatCountLabel(item namedCount) string {
	return fmt.Sprintf("%s x%d", item.Name, item.Count)
}

func titlePageColophon(docs []PackDocument, generatedAt time.Time) string {
	parts := []string{}
	versions := summarizeVersions(docs)
	if len(versions) > 0 && versions[0] != "Unknown" {
		parts = append(parts, fmt.Sprintf("PuzzleTea %s", strings.Join(versions, ", ")))
	}

	if !generatedAt.IsZero() {
		parts = append(parts, generatedAt.Format("January 2, 2006"))
	}

	if len(parts) == 0 {
		return ""
	}
	return strings.Join(parts, " | ")
}
