package pdfexport

import (
	"fmt"
	"strings"

	"codeberg.org/go-pdf/fpdf"
)

func splitCoverTextLines(pdf *fpdf.Fpdf, text string, maxW float64) []string {
	trimmed := strings.TrimSpace(text)
	if trimmed == "" {
		return []string{"PuzzleTea Collection"}
	}
	chunks := pdf.SplitLines([]byte(trimmed), maxW)
	if len(chunks) == 0 {
		return []string{trimmed}
	}
	lines := make([]string, 0, len(chunks))
	for _, c := range chunks {
		lines = append(lines, string(c))
	}
	return lines
}

func splitClampedTextLines(pdf *fpdf.Fpdf, text string, maxW float64, maxLines int) []string {
	if maxLines < 1 {
		maxLines = 1
	}

	lines := splitCoverTextLines(pdf, text, maxW)
	if len(lines) <= maxLines {
		return lines
	}

	out := make([]string, 0, maxLines)
	out = append(out, lines[:maxLines-1]...)
	out = append(out, strings.Join(lines[maxLines-1:], " "))
	return out
}

func splitCoverSubtitleLines(pdf *fpdf.Fpdf, subtitle string, maxW float64, maxLines int) []string {
	return splitClampedTextLines(pdf, subtitle, maxW, maxLines)
}

func renderCoverPage(pdf *fpdf.Fpdf, _ []Puzzle, cfg RenderConfig, coverColor RGB) {
	ink := RGB{R: 8, G: 8, B: 8}

	pdf.AddPage()
	pageW, pageH := pdf.GetPageSize()

	pdf.SetFillColor(int(coverColor.R), int(coverColor.G), int(coverColor.B))
	pdf.Rect(0, 0, pageW, pageH, "F")

	frameInset := 7.5
	drawCoverFrame(pdf, frameInset, pageW, pageH, ink)

	subtitle := strings.TrimSpace(cfg.CoverSubtitle)
	if subtitle == "" {
		subtitle = "PuzzleTea Collection"
	}

	labelW := pageW - 2*(frameInset+6.0)
	fontSize := 20.0
	for fontSize >= 13.0 {
		pdf.SetFont(coverFontFamily, "", fontSize)
		if len(splitCoverTextLines(pdf, subtitle, labelW)) <= 2 {
			break
		}
		fontSize -= 1.0
	}

	pdf.SetFont(sansFontFamily, "B", 9.8)
	pdf.SetTextColor(int(ink.R), int(ink.G), int(ink.B))
	pdf.SetXY(frameInset+6.0, frameInset+2.8)
	pdf.CellFormat(labelW, 5.0, fmt.Sprintf("VOL. %02d", cfg.VolumeNumber), "", 0, "L", false, 0, "")

	pdf.SetFont(coverFontFamily, "", fontSize)
	titleLines := splitCoverSubtitleLines(pdf, subtitle, labelW, 2)
	lineH := fontSize * 0.45
	y := frameInset + 12.0
	for _, line := range titleLines {
		pdf.SetXY(frameInset+6.0, y)
		pdf.CellFormat(labelW, lineH, line, "", 0, "L", false, 0, "")
		y += lineH
	}

	pdf.SetFont(sansFontFamily, "B", 7.8)
	pdf.SetXY(frameInset+6.0, pageH-frameInset-7.0)
	pdf.CellFormat(labelW, 4.0, "PuzzleTea", "", 0, "L", false, 0, "")
}

func drawCoverFrame(pdf *fpdf.Fpdf, inset, pageW, pageH float64, ink RGB) {
	pdf.SetDrawColor(int(ink.R), int(ink.G), int(ink.B))
	pdf.SetLineWidth(1.1)
	pdf.Rect(inset, inset, pageW-2*inset, pageH-2*inset, "D")

	pdf.SetLineWidth(0.28)
	inner := inset + 1.7
	pdf.Rect(inner, inner, pageW-2*inner, pageH-2*inner, "D")
}

func renderBackCoverPage(pdf *fpdf.Fpdf, cfg RenderConfig, coverColor RGB) {
	ink := RGB{R: 8, G: 8, B: 8}

	pdf.AddPage()
	pageW, pageH := pdf.GetPageSize()
	pdf.SetFillColor(int(coverColor.R), int(coverColor.G), int(coverColor.B))
	pdf.Rect(0, 0, pageW, pageH, "F")

	frameInset := 7.5
	drawCoverFrame(pdf, frameInset, pageW, pageH, ink)

	labelW := pageW - 2*(frameInset+6.0)
	pdf.SetTextColor(int(ink.R), int(ink.G), int(ink.B))
	pdf.SetFont(sansFontFamily, "B", 8.4)
	pdf.SetXY(frameInset+6.0, pageH-frameInset-23.0)
	pdf.CellFormat(labelW, 4.2, "PuzzleTea", "", 0, "L", false, 0, "")

	pdf.SetFont(sansFontFamily, "", 8.0)
	advertLines := splitCoverTextLines(pdf, cfg.AdvertText, labelW)
	advertLineH := 4.2
	advertY := pageH - frameInset - 17.0 - float64(len(advertLines)-1)*advertLineH
	for _, line := range advertLines {
		pdf.SetXY(frameInset+6.0, advertY)
		pdf.CellFormat(labelW, advertLineH, line, "", 0, "L", false, 0, "")
		advertY += advertLineH
	}

	pdf.SetFont(sansFontFamily, "B", 8.2)
	pdf.SetXY(frameInset+6.0, pageH-frameInset-10.5)
	pdf.CellFormat(labelW, 4.2, fmt.Sprintf("VOL. %02d", cfg.VolumeNumber), "", 0, "L", false, 0, "")
}
