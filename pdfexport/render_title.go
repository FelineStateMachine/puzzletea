package pdfexport

import (
	"fmt"
	"math"
	"sort"
	"strconv"
	"strings"

	"codeberg.org/go-pdf/fpdf"
)

func renderTitlePage(pdf *fpdf.Fpdf, docs []PackDocument, puzzles []Puzzle, cfg RenderConfig) {
	pdf.AddPage()
	pageW, pageH := pdf.GetPageSize()
	margin := 12.0
	contentWidth := pageW - 2*margin

	pdf.SetTextColor(20, 20, 20)
	pdf.SetFont(sansFontFamily, "B", 22)
	pdf.SetXY(0, 24)
	pdf.CellFormat(pageW, 10, fmt.Sprintf("PuzzleTea Volume %02d", cfg.VolumeNumber), "", 0, "C", false, 0, "")

	pdf.SetTextColor(50, 50, 50)
	pdf.SetFont(coverFontFamily, "", 16)
	pdf.SetXY(0, 35)
	pdf.CellFormat(pageW, 8, cfg.CoverSubtitle, "", 0, "C", false, 0, "")

	pdf.SetFont(sansFontFamily, "", 11)
	pdf.SetTextColor(70, 70, 70)
	pdf.SetXY(0, 44)
	pdf.CellFormat(pageW, 6, "PuzzleTea Puzzle Pack", "", 0, "C", false, 0, "")

	versionLine := fmt.Sprintf("PuzzleTea Version: %s", strings.Join(summarizeVersions(docs), ", "))
	pdf.SetFont(sansFontFamily, "", 10)
	wrappedVersions := pdf.SplitLines([]byte(versionLine), contentWidth)
	if len(wrappedVersions) == 0 {
		wrappedVersions = [][]byte{[]byte(versionLine)}
	}

	headerLineH := 4.8
	headerGap := 1.2
	headerStartY := 54.8
	metaY := 56.0
	if header := strings.TrimSpace(cfg.HeaderText); header != "" {
		pdf.SetFont(sansFontFamily, "", 9.2)
		pdf.SetTextColor(74, 74, 74)
		wrappedHeader := pdf.SplitLines([]byte(header), contentWidth)
		if len(wrappedHeader) == 0 {
			wrappedHeader = [][]byte{[]byte(header)}
		}

		metaY = headerStartY + float64(len(wrappedHeader))*headerLineH + headerGap
		sourceStartY := titlePageSourceTableStartY(metaY, len(wrappedVersions))
		sourceMaxY := pageH - 45
		if spare := titlePageSourceTableWhitespace(sourceMaxY, sourceStartY, len(docs)); spare > 0 {
			headerGap += spare
		}

		headerY := headerStartY
		for _, line := range wrappedHeader {
			pdf.SetXY(margin, headerY)
			pdf.CellFormat(contentWidth, headerLineH, string(line), "", 0, "C", false, 0, "")
			headerY += headerLineH
		}
		metaY = headerY + headerGap
	}

	pdf.SetTextColor(25, 25, 25)
	pdf.SetFont(sansFontFamily, "", 10)
	pdf.SetXY(margin, metaY)
	pdf.CellFormat(contentWidth, 6, fmt.Sprintf("Generated: %s", cfg.GeneratedAt.Format("January 2, 2006")), "", 0, "L", false, 0, "")
	metaY += 6
	for _, line := range wrappedVersions {
		pdf.SetXY(margin, metaY)
		pdf.CellFormat(contentWidth, 5.2, string(line), "", 0, "L", false, 0, "")
		metaY += 5.2
	}
	metaY += 0.8

	pdf.SetXY(margin, metaY)
	pdf.CellFormat(contentWidth, 6, fmt.Sprintf("Puzzles: %d", len(puzzles)), "", 0, "L", false, 0, "")
	metaY += 6
	metaY += 1.8

	pdf.SetFont(sansFontFamily, "B", 10)
	pdf.SetTextColor(45, 45, 45)
	pdf.SetXY(margin, metaY)
	pdf.CellFormat(contentWidth, 6, "Source Exports", "", 0, "L", false, 0, "")
	metaY += 7

	renderSourceExportsTable(pdf, docs, margin, metaY, contentWidth, pageH-45)

	pdf.SetTextColor(50, 50, 50)
	pdf.SetFont(sansFontFamily, "B", 12)
	pdf.SetXY(margin, pageH-30)
	pdf.CellFormat(contentWidth, 7, "Made with PuzzleTea", "", 0, "C", false, 0, "")

	pdf.SetFont(sansFontFamily, "", 10)
	pdf.SetXY(margin, pageH-22)
	pdf.CellFormat(contentWidth, 6, cfg.AdvertText, "", 0, "C", false, 0, "")
}

func titlePageSourceTableStartY(metaY float64, versionLineCount int) float64 {
	y := metaY
	y += 6
	y += float64(max(versionLineCount, 1)) * 5.2
	y += 0.8
	y += 6
	y += 1.8
	y += 7
	return y
}

func titlePageSourceTableWhitespace(maxY, sourceStartY float64, docCount int) float64 {
	const (
		headerHeight = 5.2
		rowHeight    = 4.8
	)

	availableRowsHeight := maxY - sourceStartY - headerHeight
	if availableRowsHeight < rowHeight {
		return 0
	}

	maxRows := int(math.Floor(availableRowsHeight / rowHeight))
	if maxRows < 1 {
		return 0
	}

	rowCount := min(docCount, maxRows)
	used := headerHeight + float64(rowCount)*rowHeight
	if spare := maxY - sourceStartY - used; spare > 0 {
		return spare
	}
	return 0
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

func renderSourceExportsTable(
	pdf *fpdf.Fpdf,
	docs []PackDocument,
	x, y, width, maxY float64,
) float64 {
	if width <= 0 || y >= maxY {
		return y
	}

	headers := []string{"Source", "Category", "Mode", "Count", "Seed"}
	columnRatios := []float64{0.33, 0.20, 0.22, 0.10, 0.15}
	columnWidths := make([]float64, len(columnRatios))
	usedWidth := 0.0
	for i := 0; i < len(columnRatios)-1; i++ {
		columnWidths[i] = width * columnRatios[i]
		usedWidth += columnWidths[i]
	}
	columnWidths[len(columnWidths)-1] = width - usedWidth

	headerHeight := 5.2
	rowHeight := 4.8
	availableRowsHeight := maxY - y - headerHeight
	if availableRowsHeight < rowHeight {
		return y
	}
	maxRows := int(math.Floor(availableRowsHeight / rowHeight))
	if maxRows < 1 {
		return y
	}

	rowCount := len(docs)
	rowCount = min(rowCount, maxRows)

	pdf.SetDrawColor(125, 125, 125)
	pdf.SetLineWidth(thinGridLineMM)
	pdf.SetFillColor(245, 245, 245)
	pdf.SetTextColor(45, 45, 45)
	pdf.SetFont(sansFontFamily, "B", 8.9)

	curX := x
	for i, header := range headers {
		pdf.SetXY(curX, y)
		pdf.CellFormat(columnWidths[i], headerHeight, header, "1", 0, "C", true, 0, "")
		curX += columnWidths[i]
	}

	pdf.SetFont(sansFontFamily, "", 8.6)
	pdf.SetTextColor(ruleTextGray, ruleTextGray, ruleTextGray)
	for i := 0; i < rowCount; i++ {
		rowY := y + headerHeight + float64(i)*rowHeight
		mode := ""
		if !isMixedModes(docs[i].Metadata.ModeSelection) {
			mode = docs[i].Metadata.ModeSelection
		}

		values := []string{
			docs[i].Metadata.SourceFileName,
			docs[i].Metadata.Category,
			mode,
			strconv.Itoa(docs[i].Metadata.Count),
			emptyAs(docs[i].Metadata.Seed, "none"),
		}

		curX = x
		for col := range values {
			cellText := fitTableCellText(pdf, values[col], columnWidths[col]-1.6)
			align := "L"
			if col == 3 {
				align = "C"
			}
			pdf.SetXY(curX, rowY)
			pdf.CellFormat(columnWidths[col], rowHeight, cellText, "1", 0, align, false, 0, "")
			curX += columnWidths[col]
		}
	}

	return y + headerHeight + float64(rowCount)*rowHeight
}

func fitTableCellText(pdf *fpdf.Fpdf, text string, maxWidth float64) string {
	text = strings.TrimSpace(text)
	if text == "" {
		return ""
	}
	if maxWidth <= 0 {
		return ""
	}
	if pdf.GetStringWidth(text) <= maxWidth {
		return text
	}

	ellipsis := "..."
	if pdf.GetStringWidth(ellipsis) > maxWidth {
		return ellipsis
	}

	runes := []rune(text)
	for len(runes) > 0 {
		candidate := string(runes) + ellipsis
		if pdf.GetStringWidth(candidate) <= maxWidth {
			return candidate
		}
		runes = runes[:len(runes)-1]
	}
	return ellipsis
}

func emptyAs(v, fallback string) string {
	if strings.TrimSpace(v) == "" {
		return fallback
	}
	return v
}
