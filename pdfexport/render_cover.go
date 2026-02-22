package pdfexport

import (
	"fmt"
	"hash/fnv"
	"math"
	"math/rand"
	"strings"

	"github.com/go-pdf/fpdf"
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
	drawWoodcutScene(pdf, scene, coverColor, ink, cfg.ShuffleSeed)

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

func drawWoodcutScene(pdf *fpdf.Fpdf, scene rectMM, bg, ink RGB, seed string) {
	drawStippleTexture(pdf, scene, ink, seed)
	drawHatchingBands(pdf, scene, ink)
	drawPineForestMass(pdf, scene, ink)
	drawIsometricRuins(pdf, scene, bg, ink)
	drawSkullMotifs(pdf, scene, bg, ink)
}

func drawStippleTexture(pdf *fpdf.Fpdf, scene rectMM, ink RGB, seed string) {
	h := fnv.New64a()
	h.Write([]byte(strings.TrimSpace(seed) + "-stipple"))
	rng := rand.New(rand.NewSource(int64(h.Sum64())))

	pdf.SetFillColor(int(ink.R), int(ink.G), int(ink.B))
	for x := scene.x + 1.0; x < scene.x+scene.w-1.0; x += 3.1 {
		for y := scene.y + 2.0; y < scene.y+scene.h*0.58; y += 3.2 {
			if rng.Float64() < 0.22 {
				jitterX := (rng.Float64() - 0.5) * 0.9
				jitterY := (rng.Float64() - 0.5) * 0.9
				pdf.Circle(x+jitterX, y+jitterY, 0.11, "F")
			}
		}
	}
}

func drawHatchingBands(pdf *fpdf.Fpdf, scene rectMM, ink RGB) {
	pdf.SetDrawColor(int(ink.R), int(ink.G), int(ink.B))
	pdf.SetLineWidth(0.28)

	baseY := scene.y + scene.h*0.64
	for i := -8; i < 34; i++ {
		x0 := scene.x + float64(i)*4.4
		y0 := baseY
		x1 := x0 + scene.h*0.55
		y1 := baseY + scene.h*0.30
		pdf.Line(x0, y0, x1, y1)
	}

	for i := -8; i < 30; i++ {
		x0 := scene.x + float64(i)*5.0
		y0 := baseY + 7.5
		x1 := x0 + scene.h*0.48
		y1 := baseY + scene.h*0.33
		pdf.Line(x0, y0, x1, y1)
	}
}

func drawPineForestMass(pdf *fpdf.Fpdf, scene rectMM, ink RGB) {
	pdf.SetFillColor(int(ink.R), int(ink.G), int(ink.B))
	pdf.SetDrawColor(int(ink.R), int(ink.G), int(ink.B))

	layers := []struct {
		y      float64
		count  int
		w      float64
		h      float64
		offset float64
	}{
		{y: scene.y + scene.h*0.45, count: 22, w: 4.4, h: 13.0, offset: 0.8},
		{y: scene.y + scene.h*0.53, count: 19, w: 5.4, h: 16.0, offset: 1.4},
		{y: scene.y + scene.h*0.62, count: 16, w: 6.2, h: 18.4, offset: 2.0},
	}

	for _, layer := range layers {
		step := scene.w / float64(layer.count)
		for i := 0; i < layer.count; i++ {
			cx := scene.x + float64(i)*step + layer.offset
			pts := []fpdf.PointType{
				{X: cx, Y: layer.y - layer.h},
				{X: cx - layer.w, Y: layer.y},
				{X: cx + layer.w, Y: layer.y},
			}
			pdf.Polygon(pts, "F")
			pdf.SetLineWidth(0.22)
			pdf.Line(cx, layer.y, cx, layer.y+2.4)
		}
	}
}

func drawIsometricRuins(pdf *fpdf.Fpdf, scene rectMM, bg, ink RGB) {
	pdf.SetDrawColor(int(ink.R), int(ink.G), int(ink.B))
	pdf.SetFillColor(int(ink.R), int(ink.G), int(ink.B))

	centerX := scene.x + scene.w*0.52
	baseY := scene.y + scene.h*0.80
	drawIsometricGrid(pdf, centerX-44, baseY-17, 88, 34, ink)

	drawStoneArch(pdf, centerX-22, baseY-35, 25, 27, 8, bg, ink)
	drawStoneArch(pdf, centerX+8, baseY-48, 20, 23, 7, bg, ink)

	drawRubbleBlock(pdf, centerX-39, baseY-6, 9, 5, 5, ink)
	drawRubbleBlock(pdf, centerX+31, baseY-3, 10, 5, 6, ink)
	drawRubbleBlock(pdf, centerX+12, baseY+5, 8, 4, 4, ink)
}

func drawIsometricGrid(pdf *fpdf.Fpdf, x, y, w, h float64, ink RGB) {
	pdf.SetDrawColor(int(ink.R), int(ink.G), int(ink.B))
	pdf.SetLineWidth(0.26)

	for i := -8; i < 14; i++ {
		sx := x + float64(i)*4
		pdf.Line(sx, y, sx+w*0.5, y+h*0.5)
	}
	for i := -8; i < 14; i++ {
		sx := x + w + float64(i)*4
		pdf.Line(sx, y, sx-w*0.5, y+h*0.5)
	}
}

func drawStoneArch(pdf *fpdf.Fpdf, x, y, w, h, depth float64, bg, ink RGB) {
	frontLeft := []fpdf.PointType{{X: x, Y: y}, {X: x + w*0.18, Y: y - depth}, {X: x + w*0.18, Y: y + h - depth}, {X: x, Y: y + h}}
	frontRight := []fpdf.PointType{{X: x + w*0.82, Y: y - depth}, {X: x + w, Y: y}, {X: x + w, Y: y + h}, {X: x + w*0.82, Y: y + h - depth}}

	pdf.SetFillColor(int(ink.R), int(ink.G), int(ink.B))
	pdf.Polygon(frontLeft, "F")
	pdf.Polygon(frontRight, "F")

	outerR := w * 0.50
	innerR := w * 0.33
	cx := x + w*0.5
	cy := y + h*0.42
	drawArcPolyline(pdf, cx, cy, outerR, math.Pi, 2*math.Pi, 24)

	pdf.SetDrawColor(int(bg.R), int(bg.G), int(bg.B))
	pdf.SetLineWidth(2.0)
	drawArcPolyline(pdf, cx, cy, innerR, math.Pi, 2*math.Pi, 24)
	pdf.Line(x+w*0.33, y+h, x+w*0.33, cy)
	pdf.Line(x+w*0.67, y+h-depth*0.2, x+w*0.67, cy)

	pdf.SetDrawColor(int(ink.R), int(ink.G), int(ink.B))
	pdf.SetLineWidth(0.28)
	for i := 0; i < 6; i++ {
		sy := y + 3 + float64(i)*4.2
		pdf.Line(x+1.5, sy, x+w*0.16, sy-0.6)
		pdf.Line(x+w*0.84, sy-0.5, x+w-1.5, sy+0.4)
	}
}

func drawArcPolyline(pdf *fpdf.Fpdf, cx, cy, r, start, end float64, segments int) {
	if segments < 2 {
		segments = 2
	}
	step := (end - start) / float64(segments)
	px := cx + math.Cos(start)*r
	py := cy + math.Sin(start)*r
	for i := 1; i <= segments; i++ {
		a := start + float64(i)*step
		x := cx + math.Cos(a)*r
		y := cy + math.Sin(a)*r
		pdf.Line(px, py, x, y)
		px, py = x, y
	}
}

func drawRubbleBlock(pdf *fpdf.Fpdf, x, y, w, h, skew float64, ink RGB) {
	pdf.SetFillColor(int(ink.R), int(ink.G), int(ink.B))
	pts := []fpdf.PointType{
		{X: x, Y: y},
		{X: x + w, Y: y - skew*0.2},
		{X: x + w + skew, Y: y + h},
		{X: x + skew, Y: y + h + skew*0.2},
	}
	pdf.Polygon(pts, "F")
}

func drawSkullMotifs(pdf *fpdf.Fpdf, scene rectMM, bg, ink RGB) {
	centers := []struct {
		x float64
		y float64
		r float64
	}{
		{x: scene.x + scene.w*0.30, y: scene.y + scene.h*0.84, r: 2.9},
		{x: scene.x + scene.w*0.69, y: scene.y + scene.h*0.87, r: 2.4},
	}

	for _, c := range centers {
		pdf.SetDrawColor(int(ink.R), int(ink.G), int(ink.B))
		pdf.SetFillColor(int(ink.R), int(ink.G), int(ink.B))
		pdf.Circle(c.x, c.y, c.r, "F")

		pdf.SetFillColor(int(bg.R), int(bg.G), int(bg.B))
		pdf.Circle(c.x-c.r*0.35, c.y-c.r*0.1, c.r*0.22, "F")
		pdf.Circle(c.x+c.r*0.35, c.y-c.r*0.08, c.r*0.20, "F")
		jaw := []fpdf.PointType{{X: c.x - c.r*0.48, Y: c.y + c.r*0.2}, {X: c.x + c.r*0.48, Y: c.y + c.r*0.2}, {X: c.x, Y: c.y + c.r*0.72}}
		pdf.Polygon(jaw, "F")

		pdf.SetDrawColor(int(ink.R), int(ink.G), int(ink.B))
		pdf.SetLineWidth(0.22)
		pdf.Line(c.x-c.r*0.6, c.y-c.r*0.85, c.x-c.r*1.15, c.y-c.r*1.45)
		pdf.Line(c.x+c.r*0.5, c.y-c.r*0.7, c.x+c.r*1.0, c.y-c.r*1.3)
	}
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
	drawBackMotif(pdf, motif, coverColor, ink)

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

func drawBackMotif(pdf *fpdf.Fpdf, scene rectMM, bg, ink RGB) {
	drawStippleTexture(pdf, scene, ink, "back-cover")
	drawHatchingBands(pdf, scene, ink)
	drawPineForestMass(pdf, scene, ink)
	drawStoneArch(pdf, scene.x+scene.w*0.58, scene.y+scene.h*0.56, 22, 24, 7, bg, ink)
	drawRubbleBlock(pdf, scene.x+scene.w*0.28, scene.y+scene.h*0.72, 8, 4, 4, ink)
}
