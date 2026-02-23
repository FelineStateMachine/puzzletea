package pdfexport

import (
	"fmt"
	"hash/fnv"
	"math/rand"
	"strings"

	"codeberg.org/go-pdf/fpdf"
)

var natureTonePalette = []RGB{
	{250, 76, 56},  // riso red
	{255, 112, 0},  // fluorescent orange
	{255, 183, 3},  // sunflower
	{53, 169, 255}, // sky blue
	{0, 204, 160},  // seafoam
	{152, 226, 68}, // neon moss
	{255, 82, 133}, // hot pink
	{104, 83, 255}, // ultramarine
}

func resolveCoverColor(cfg RenderConfig) RGB {
	if cfg.CoverColor != nil {
		return *cfg.CoverColor
	}
	seed := strings.TrimSpace(cfg.ShuffleSeed)
	var idx int
	if seed != "" {
		h := fnv.New64a()
		h.Write([]byte(seed))
		idx = int(h.Sum64() % uint64(len(natureTonePalette)))
	} else {
		idx = rand.Intn(len(natureTonePalette))
	}
	return natureTonePalette[idx]
}

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

func splitCoverSubtitleLines(pdf *fpdf.Fpdf, subtitle string, maxW float64, maxLines int) []string {
	if maxLines < 1 {
		maxLines = 1
	}
	lines := splitCoverTextLines(pdf, subtitle, maxW)
	if len(lines) <= maxLines {
		return lines
	}

	out := make([]string, 0, maxLines)
	out = append(out, lines[:maxLines-1]...)
	out = append(out, strings.Join(lines[maxLines-1:], " "))
	return out
}

func renderCoverPage(pdf *fpdf.Fpdf, _ []Puzzle, cfg RenderConfig) {
	coverColor := resolveCoverColor(cfg)
	ink := RGB{R: 8, G: 8, B: 8}

	pdf.AddPage()
	pageW, pageH := pdf.GetPageSize()

	pdf.SetFillColor(int(coverColor.R), int(coverColor.G), int(coverColor.B))
	pdf.Rect(0, 0, pageW, pageH, "F")

	frameInset := 7.5
	drawCoverFrame(pdf, frameInset, pageW, pageH, ink)

	scene := rectMM{x: frameInset + 4.0, y: frameInset + 10.0, w: pageW - (frameInset+4.0)*2, h: 132.0}
	drawCoverArtwork(pdf, scene, cfg.ShuffleSeed, coverColor, ink)

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
	y := scene.y + scene.h + 8.5
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

func drawCoverArtwork(pdf *fpdf.Fpdf, scene rectMM, seed string, bg, ink RGB) {
	drawCoverArtworkImage(pdf, scene, seed, "front", bg)

	pdf.SetDrawColor(int(ink.R), int(ink.G), int(ink.B))
	pdf.SetLineWidth(0.40)
	pdf.Rect(scene.x, scene.y, scene.w, scene.h, "D")

	pdf.SetLineWidth(0.20)
	inset := 1.8
	pdf.Rect(scene.x+inset, scene.y+inset, scene.w-2*inset, scene.h-2*inset, "D")
}

func renderBackCoverPage(pdf *fpdf.Fpdf, cfg RenderConfig, coverColor RGB) {
	ink := RGB{R: 8, G: 8, B: 8}

	pdf.AddPage()
	pageW, pageH := pdf.GetPageSize()
	pdf.SetFillColor(int(coverColor.R), int(coverColor.G), int(coverColor.B))
	pdf.Rect(0, 0, pageW, pageH, "F")

	frameInset := 7.5
	drawCoverFrame(pdf, frameInset, pageW, pageH, ink)

	motif := rectMM{x: frameInset + 5.5, y: frameInset + 14.0, w: pageW - 2*(frameInset+5.5), h: 96.0}
	drawBackMotif(pdf, motif, cfg.ShuffleSeed, coverColor, ink)

	labelW := pageW - 2*(frameInset+6.0)
	pdf.SetTextColor(int(ink.R), int(ink.G), int(ink.B))
	pdf.SetFont(sansFontFamily, "B", 8.4)
	pdf.SetXY(frameInset+6.0, pageH-frameInset-23.0)
	pdf.CellFormat(labelW, 4.2, "PuzzleTea", "", 0, "L", false, 0, "")

	pdf.SetFont(sansFontFamily, "", 8.0)
	pdf.SetXY(frameInset+6.0, pageH-frameInset-17.0)
	pdf.CellFormat(labelW, 4.2, cfg.AdvertText, "", 0, "L", false, 0, "")

	pdf.SetFont(sansFontFamily, "B", 8.2)
	pdf.SetXY(frameInset+6.0, pageH-frameInset-10.5)
	pdf.CellFormat(labelW, 4.2, fmt.Sprintf("VOL. %02d", cfg.VolumeNumber), "", 0, "L", false, 0, "")
}

func drawBackMotif(pdf *fpdf.Fpdf, scene rectMM, seed string, bg, ink RGB) {
	drawCoverArtworkImage(pdf, scene, seed, "back", bg)

	pdf.SetDrawColor(int(ink.R), int(ink.G), int(ink.B))
	pdf.SetLineWidth(0.34)
	pdf.Rect(scene.x, scene.y, scene.w, scene.h, "D")
}
