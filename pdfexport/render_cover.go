package pdfexport

import (
	"fmt"
	"math"
	"reflect"
	"sort"
	"strings"

	"codeberg.org/go-pdf/fpdf"
)

type coverOutsideSlice int

const (
	coverOutsideBack coverOutsideSlice = iota
	coverOutsideFront
)

type coverCompositionFamily int

const (
	coverFamilyFrame coverCompositionFamily = iota
	coverFamilyStack
	coverFamilyHinge
	coverFamilyIsland
)

type coverDirection int

const (
	coverDirectionVertical coverDirection = iota
	coverDirectionHorizontal
	coverDirectionDiagonal
	coverDirectionClustered
)

type coverAspectBucket int

const (
	coverAspectTall coverAspectBucket = iota
	coverAspectWide
	coverAspectCompact
	coverAspectOffset
)

type coverTextureKind int

const (
	coverTextureNone coverTextureKind = iota
	coverTextureStripes
	coverTextureHatch
	coverTextureDots
	coverTextureChecker
	coverTextureLattice
)

type coverFillMode int

const (
	coverFillSolid coverFillMode = iota
	coverFillSlits
	coverFillOpenGrid
	coverFillPunctured
	coverFillBanded
)

type coverCutoutKind int

const (
	coverCutoutRect coverCutoutKind = iota
	coverCutoutCircle
)

type coverShapeKind int

const (
	coverShapeRect coverShapeKind = iota
	coverShapeCircle
	coverShapeCroppedCircle
	coverShapeStepped
)

type coverShapeEdge int

const (
	coverEdgeLeft coverShapeEdge = iota
	coverEdgeRight
	coverEdgeTop
	coverEdgeBottom
)

type coverCutout struct {
	Kind coverCutoutKind
	X    float64
	Y    float64
	W    float64
	H    float64
	R    float64
}

type coverShape struct {
	Kind    coverShapeKind
	Edge    coverShapeEdge
	Bounds  rectMM
	PivotX  float64
	PivotY  float64
	Rotate  float64
	Locked  bool
	Fill    coverFillMode
	Texture coverTextureKind
	Step    float64
	Weight  float64
	Cutouts []coverCutout
}

type coverArtLayout struct {
	ArtArea         rectMM
	LockupExclusion rectMM
	Family          coverCompositionFamily
	Direction       coverDirection
	Shapes          []coverShape
}

type coverRand interface {
	Float64() float64
	IntN(int) int
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

func renderOutsideCoverPage(pdf *fpdf.Fpdf, cfg RenderConfig, slice coverOutsideSlice) {
	const frameInset = 7.5

	pageW, pageH := pdf.GetPageSize()
	drawOutsideSpreadFrameSlice(pdf, slice, frameInset, pageW, pageH)

	if slice == coverOutsideFront {
		layout := buildCoverArtLayout(cfg, pageW, pageH)
		drawCoverArtLayout(pdf, layout)
		renderFrontCoverLockup(pdf, cfg, pageW, frameInset)
		return
	}

	renderBackCoverImprint(pdf, cfg, pageW, pageH, frameInset)
}

func renderCoverBlankPage(_ *fpdf.Fpdf) {}

func drawOutsideSpreadFrameSlice(pdf *fpdf.Fpdf, slice coverOutsideSlice, inset, pageW, pageH float64) {
	drawCoverSpreadBorder(pdf, slice, inset, 1.1, pageW, pageH)
	drawCoverSpreadBorder(pdf, slice, inset+1.7, 0.28, pageW, pageH)
}

func drawCoverSpreadBorder(pdf *fpdf.Fpdf, slice coverOutsideSlice, inset, lineW, pageW, pageH float64) {
	pdf.SetDrawColor(8, 8, 8)
	pdf.SetLineWidth(lineW)

	leftX, rightX := coverSliceHorizontalSpan(slice, inset, pageW)
	pdf.Line(leftX, inset, rightX, inset)
	pdf.Line(leftX, pageH-inset, rightX, pageH-inset)

	if slice == coverOutsideBack {
		pdf.Line(inset, inset, inset, pageH-inset)
		return
	}
	pdf.Line(pageW-inset, inset, pageW-inset, pageH-inset)
}

func coverSliceHorizontalSpan(slice coverOutsideSlice, inset, pageW float64) (left, right float64) {
	if slice == coverOutsideBack {
		return inset, pageW
	}
	return 0, pageW - inset
}

func renderFrontCoverLockup(pdf *fpdf.Fpdf, cfg RenderConfig, pageW, frameInset float64) {
	labelW := min(pageW*0.54, 74.0)
	labelX := pageW - frameInset - 7.0 - labelW

	subtitle := strings.TrimSpace(cfg.CoverSubtitle)
	if subtitle == "" {
		subtitle = "PuzzleTea Collection"
	}

	fontSize := 18.0 + pdfFontSizeDelta
	for fontSize >= 13.5+pdfFontSizeDelta {
		pdf.SetFont(coverFontFamily, "", fontSize)
		if len(splitCoverSubtitleLines(pdf, subtitle, labelW, 3)) <= 3 {
			break
		}
		fontSize -= 0.8
	}

	pdf.SetTextColor(8, 8, 8)
	pdf.SetFont(sansFontFamily, "B", 9.6+pdfFontSizeDelta)
	pdf.SetXY(labelX, frameInset+5.0)
	pdf.CellFormat(labelW, 4.5, fmt.Sprintf("VOL. %02d", cfg.VolumeNumber), "", 0, "R", false, 0, "")

	pdf.SetFont(coverFontFamily, "", fontSize)
	lineH := fontSize * 0.44
	y := frameInset + 13.2
	for _, line := range splitCoverSubtitleLines(pdf, subtitle, labelW, 3) {
		pdf.SetXY(labelX, y)
		pdf.CellFormat(labelW, lineH, line, "", 0, "R", false, 0, "")
		y += lineH
	}
}

func renderBackCoverImprint(pdf *fpdf.Fpdf, cfg RenderConfig, pageW, pageH, frameInset float64) {
	labelX := frameInset + 6.0
	labelW := min(pageW*0.42, 52.0)

	pdf.SetTextColor(8, 8, 8)
	pdf.SetFont(sansFontFamily, "B", 8.2+pdfFontSizeDelta)
	pdf.SetXY(labelX, pageH-frameInset-13.4)
	pdf.CellFormat(labelW, 4.0, "PuzzleTea", "", 0, "L", false, 0, "")

	pdf.SetFont(sansFontFamily, "", 7.3+pdfFontSizeDelta)
	pdf.SetXY(labelX, pageH-frameInset-8.5)
	pdf.CellFormat(labelW, 3.8, fmt.Sprintf("Vol. %02d", cfg.VolumeNumber), "", 0, "L", false, 0, "")
}

func buildCoverArtLayout(cfg RenderConfig, pageW, pageH float64) coverArtLayout {
	rng := seededRand(fmt.Sprintf("cover:%s:%d", cfg.ShuffleSeed, cfg.VolumeNumber))
	layout := coverArtLayout{
		ArtArea:         coverArtArea(pageW, pageH),
		LockupExclusion: coverLockupExclusion(pageW),
		Family:          coverCompositionFamily(rng.IntN(4)),
	}

	switch layout.Family {
	case coverFamilyFrame:
		layout.Direction = buildFrameFamily(&layout, rng)
	case coverFamilyStack:
		layout.Direction = buildStackFamily(&layout, rng)
	case coverFamilyHinge:
		layout.Direction = buildHingeFamily(&layout, rng)
	default:
		layout.Direction = buildIslandFamily(&layout, rng)
	}

	return layout
}

func coverArtArea(pageW, pageH float64) rectMM {
	return rectMM{
		x: 12.5,
		y: 37.0,
		w: pageW - 25.0,
		h: pageH - 56.0,
	}
}

func coverLockupExclusion(pageW float64) rectMM {
	return rectMM{
		x: pageW - 89.0,
		y: 8.0,
		w: 78.0,
		h: 32.0,
	}
}

func buildFrameFamily(layout *coverArtLayout, rng coverRand) coverDirection {
	outer := rectMM{
		x: layout.ArtArea.x + layout.ArtArea.w*(0.06+rng.Float64()*0.14),
		y: layout.ArtArea.y + layout.ArtArea.h*(0.08+rng.Float64()*0.10),
		w: layout.ArtArea.w * (0.42 + rng.Float64()*0.20),
		h: layout.ArtArea.h * (0.48 + rng.Float64()*0.20),
	}
	thickness := min(outer.w, outer.h) * (0.16 + rng.Float64()*0.07)
	openSide := coverShapeEdge(rng.IntN(4))
	direction := coverDirectionVertical
	if outer.w > outer.h {
		direction = coverDirectionHorizontal
	}

	add := func(kind coverShapeKind, edge coverShapeEdge, bounds rectMM) {
		layout.Shapes = append(layout.Shapes, decoratedCoverShape(layout, kind, edge, bounds, coverFamilyFrame, true, rng))
	}

	if openSide != coverEdgeTop {
		add(shapeFamilyKind(rng, coverShapeRect, coverShapeStepped), coverEdgeTop, rectMM{
			x: outer.x,
			y: outer.y,
			w: outer.w,
			h: thickness,
		})
	}
	if openSide != coverEdgeLeft {
		add(shapeFamilyKind(rng, coverShapeRect, coverShapeStepped), coverEdgeLeft, rectMM{
			x: outer.x,
			y: outer.y + thickness*0.55,
			w: thickness * (0.9 + rng.Float64()*0.35),
			h: outer.h - thickness*0.55,
		})
	}
	if openSide != coverEdgeRight {
		add(shapeFamilyKind(rng, coverShapeRect, coverShapeStepped), coverEdgeRight, rectMM{
			x: outer.x + outer.w - thickness*(0.9+rng.Float64()*0.35),
			y: outer.y + thickness*0.35,
			w: thickness * (0.9 + rng.Float64()*0.35),
			h: outer.h - thickness*0.35,
		})
	}
	if openSide != coverEdgeBottom {
		add(shapeFamilyKind(rng, coverShapeRect, coverShapeStepped), coverEdgeBottom, rectMM{
			x: outer.x + outer.w*(0.04+rng.Float64()*0.10),
			y: outer.y + outer.h - thickness*(0.95+rng.Float64()*0.25),
			w: outer.w * (0.56 + rng.Float64()*0.20),
			h: thickness * (0.95 + rng.Float64()*0.25),
		})
	}

	if rng.IntN(2) == 0 {
		layout.Shapes = append(layout.Shapes, decoratedCoverShape(layout, coverShapeCroppedCircle, oppositeEdge(openSide), rectMM{
			x: outer.x + outer.w*(0.52+rng.Float64()*0.16),
			y: outer.y + outer.h*(0.48+rng.Float64()*0.16),
			w: outer.w * (0.16 + rng.Float64()*0.12),
			h: outer.w * (0.16 + rng.Float64()*0.12),
		}, coverFamilyFrame, false, rng))
	}

	return direction
}

func buildStackFamily(layout *coverArtLayout, rng coverRand) coverDirection {
	vertical := rng.IntN(2) == 0
	if vertical {
		x := layout.ArtArea.x + layout.ArtArea.w*(0.06+rng.Float64()*0.18)
		y := layout.ArtArea.y + layout.ArtArea.h*(0.08+rng.Float64()*0.10)
		widths := []float64{
			layout.ArtArea.w * (0.26 + rng.Float64()*0.10),
			layout.ArtArea.w * (0.48 + rng.Float64()*0.16),
			layout.ArtArea.w * (0.34 + rng.Float64()*0.16),
		}
		heights := []float64{
			layout.ArtArea.h * (0.18 + rng.Float64()*0.08),
			layout.ArtArea.h * (0.22 + rng.Float64()*0.10),
			layout.ArtArea.h * (0.16 + rng.Float64()*0.08),
		}
		gap := 4.0 + rng.Float64()*4.0
		curY := y
		for i := range widths {
			bounds := rectMM{x: x, y: curY, w: widths[i], h: heights[i]}
			kind := coverShapeRect
			if i == 0 && rng.IntN(2) == 0 {
				kind = coverShapeStepped
			}
			layout.Shapes = append(layout.Shapes, decoratedCoverShape(layout, kind, coverEdgeLeft, bounds, coverFamilyStack, false, rng))
			curY += heights[i] + gap
		}
		layout.Shapes = append(layout.Shapes, decoratedCoverShape(layout, coverShapeCircle, coverEdgeRight, rectMM{
			x: x + layout.ArtArea.w*(0.44+rng.Float64()*0.12),
			y: y + layout.ArtArea.h*(0.18+rng.Float64()*0.18),
			w: layout.ArtArea.w * (0.16 + rng.Float64()*0.10),
			h: layout.ArtArea.w * (0.16 + rng.Float64()*0.10),
		}, coverFamilyStack, false, rng))
		return coverDirectionVertical
	}

	x := layout.ArtArea.x + layout.ArtArea.w*(0.06+rng.Float64()*0.10)
	y := layout.ArtArea.y + layout.ArtArea.h*(0.14+rng.Float64()*0.16)
	widths := []float64{
		layout.ArtArea.w * (0.18 + rng.Float64()*0.10),
		layout.ArtArea.w * (0.30 + rng.Float64()*0.12),
		layout.ArtArea.w * (0.20 + rng.Float64()*0.08),
	}
	heights := []float64{
		layout.ArtArea.h * (0.28 + rng.Float64()*0.12),
		layout.ArtArea.h * (0.18 + rng.Float64()*0.08),
		layout.ArtArea.h * (0.32 + rng.Float64()*0.10),
	}
	gap := 4.0 + rng.Float64()*4.0
	curX := x
	for i := range widths {
		bounds := rectMM{x: curX, y: y, w: widths[i], h: heights[i]}
		kind := coverShapeRect
		if i == 1 && rng.IntN(2) == 0 {
			kind = coverShapeStepped
		}
		layout.Shapes = append(layout.Shapes, decoratedCoverShape(layout, kind, coverEdgeBottom, bounds, coverFamilyStack, false, rng))
		curX += widths[i] + gap
	}
	layout.Shapes = append(layout.Shapes, decoratedCoverShape(layout, coverShapeCroppedCircle, coverEdgeBottom, rectMM{
		x: x + layout.ArtArea.w*(0.30+rng.Float64()*0.18),
		y: y + layout.ArtArea.h*(0.42+rng.Float64()*0.10),
		w: layout.ArtArea.w * (0.18 + rng.Float64()*0.09),
		h: layout.ArtArea.w * (0.18 + rng.Float64()*0.09),
	}, coverFamilyStack, false, rng))
	return coverDirectionHorizontal
}

func buildHingeFamily(layout *coverArtLayout, rng coverRand) coverDirection {
	baseX := layout.ArtArea.x + layout.ArtArea.w*(0.06+rng.Float64()*0.08)
	baseY := layout.ArtArea.y + layout.ArtArea.h*(0.42+rng.Float64()*0.12)
	upperX := layout.ArtArea.x + layout.ArtArea.w*(0.48+rng.Float64()*0.10)
	upperY := layout.ArtArea.y + layout.ArtArea.h*(0.08+rng.Float64()*0.10)

	layout.Shapes = append(layout.Shapes,
		decoratedCoverShape(layout, shapeFamilyKind(rng, coverShapeRect, coverShapeStepped), coverEdgeBottom, rectMM{
			x: baseX,
			y: baseY,
			w: layout.ArtArea.w * (0.30 + rng.Float64()*0.10),
			h: layout.ArtArea.h * (0.26 + rng.Float64()*0.10),
		}, coverFamilyHinge, false, rng),
	)
	layout.Shapes = append(layout.Shapes,
		decoratedCoverShape(layout, shapeFamilyKind(rng, coverShapeRect, coverShapeCroppedCircle), coverEdgeTop, rectMM{
			x: upperX,
			y: upperY,
			w: layout.ArtArea.w * (0.24 + rng.Float64()*0.12),
			h: layout.ArtArea.h * (0.26 + rng.Float64()*0.10),
		}, coverFamilyHinge, false, rng),
	)
	layout.Shapes = append(layout.Shapes,
		decoratedCoverShape(layout, coverShapeRect, coverEdgeRight, rectMM{
			x: layout.ArtArea.x + layout.ArtArea.w*(0.38+rng.Float64()*0.08),
			y: layout.ArtArea.y + layout.ArtArea.h*(0.28+rng.Float64()*0.10),
			w: layout.ArtArea.w * (0.09 + rng.Float64()*0.04),
			h: layout.ArtArea.h * (0.34 + rng.Float64()*0.12),
		}, coverFamilyHinge, true, rng),
	)
	layout.Shapes = append(layout.Shapes,
		decoratedCoverShape(layout, coverShapeCircle, coverEdgeRight, rectMM{
			x: layout.ArtArea.x + layout.ArtArea.w*(0.34+rng.Float64()*0.10),
			y: layout.ArtArea.y + layout.ArtArea.h*(0.22+rng.Float64()*0.12),
			w: layout.ArtArea.w * (0.12 + rng.Float64()*0.08),
			h: layout.ArtArea.w * (0.12 + rng.Float64()*0.08),
		}, coverFamilyHinge, false, rng),
	)
	return coverDirectionDiagonal
}

func buildIslandFamily(layout *coverArtLayout, rng coverRand) coverDirection {
	clustered := rng.IntN(2) == 0
	layout.Shapes = append(layout.Shapes,
		decoratedCoverShape(layout, shapeFamilyKind(rng, coverShapeStepped, coverShapeCircle), coverEdgeLeft, rectMM{
			x: layout.ArtArea.x + layout.ArtArea.w*(0.06+rng.Float64()*0.14),
			y: layout.ArtArea.y + layout.ArtArea.h*(0.34+rng.Float64()*0.16),
			w: layout.ArtArea.w * (0.30 + rng.Float64()*0.14),
			h: layout.ArtArea.h * (0.24 + rng.Float64()*0.14),
		}, coverFamilyIsland, false, rng),
	)

	if clustered {
		layout.Shapes = append(layout.Shapes,
			decoratedCoverShape(layout, coverShapeCircle, coverEdgeTop, rectMM{
				x: layout.ArtArea.x + layout.ArtArea.w*(0.40+rng.Float64()*0.14),
				y: layout.ArtArea.y + layout.ArtArea.h*(0.18+rng.Float64()*0.14),
				w: layout.ArtArea.w * (0.14 + rng.Float64()*0.10),
				h: layout.ArtArea.w * (0.14 + rng.Float64()*0.10),
			}, coverFamilyIsland, false, rng),
			decoratedCoverShape(layout, coverShapeRect, coverEdgeBottom, rectMM{
				x: layout.ArtArea.x + layout.ArtArea.w*(0.46+rng.Float64()*0.14),
				y: layout.ArtArea.y + layout.ArtArea.h*(0.48+rng.Float64()*0.12),
				w: layout.ArtArea.w * (0.18 + rng.Float64()*0.10),
				h: layout.ArtArea.h * (0.18 + rng.Float64()*0.10),
			}, coverFamilyIsland, false, rng),
			decoratedCoverShape(layout, coverShapeCroppedCircle, coverEdgeRight, rectMM{
				x: layout.ArtArea.x + layout.ArtArea.w*(0.28+rng.Float64()*0.18),
				y: layout.ArtArea.y + layout.ArtArea.h*(0.06+rng.Float64()*0.10),
				w: layout.ArtArea.w * (0.12 + rng.Float64()*0.08),
				h: layout.ArtArea.w * (0.12 + rng.Float64()*0.08),
			}, coverFamilyIsland, false, rng),
		)
		return coverDirectionClustered
	}

	layout.Shapes = append(layout.Shapes,
		decoratedCoverShape(layout, coverShapeCircle, coverEdgeTop, rectMM{
			x: layout.ArtArea.x + layout.ArtArea.w*(0.34+rng.Float64()*0.10),
			y: layout.ArtArea.y + layout.ArtArea.h*(0.18+rng.Float64()*0.08),
			w: layout.ArtArea.w * (0.16 + rng.Float64()*0.10),
			h: layout.ArtArea.w * (0.16 + rng.Float64()*0.10),
		}, coverFamilyIsland, false, rng),
		decoratedCoverShape(layout, coverShapeRect, coverEdgeBottom, rectMM{
			x: layout.ArtArea.x + layout.ArtArea.w*(0.56+rng.Float64()*0.10),
			y: layout.ArtArea.y + layout.ArtArea.h*(0.48+rng.Float64()*0.08),
			w: layout.ArtArea.w * (0.18 + rng.Float64()*0.10),
			h: layout.ArtArea.h * (0.20 + rng.Float64()*0.10),
		}, coverFamilyIsland, false, rng),
	)
	return coverDirectionDiagonal
}

func shapeFamilyKind(rng coverRand, a, b coverShapeKind) coverShapeKind {
	if rng.IntN(2) == 0 {
		return a
	}
	return b
}

func oppositeEdge(edge coverShapeEdge) coverShapeEdge {
	switch edge {
	case coverEdgeLeft:
		return coverEdgeRight
	case coverEdgeRight:
		return coverEdgeLeft
	case coverEdgeTop:
		return coverEdgeBottom
	default:
		return coverEdgeTop
	}
}

func decoratedCoverShape(
	layout *coverArtLayout,
	kind coverShapeKind,
	edge coverShapeEdge,
	bounds rectMM,
	family coverCompositionFamily,
	locked bool,
	rng coverRand,
) coverShape {
	shape := coverShape{
		Kind:    kind,
		Edge:    edge,
		Bounds:  bounds,
		Locked:  locked,
		Fill:    pickCoverFillMode(family, rng),
		Texture: pickCoverTexture(family, rng),
		Step:    4.0 + rng.Float64()*3.8,
		Weight:  0.7 + rng.Float64()*0.8,
	}
	shape.Rotate = pickCoverRotationDeg(locked, rng)
	shape = sanitizeCoverShape(shape, layout.ArtArea, layout.LockupExclusion)
	shape.Cutouts = buildFamilyCutouts(family, shape, rng)
	return shape
}

func pickCoverRotationDeg(locked bool, rng coverRand) float64 {
	minDeg := 8.0
	maxDeg := 20.0
	if locked {
		minDeg = 4.0
		maxDeg = 10.0
	}
	angle := minDeg + rng.Float64()*(maxDeg-minDeg)
	if rng.IntN(2) == 0 {
		return -angle
	}
	return angle
}

func sanitizeCoverShape(shape coverShape, artArea, exclusion rectMM) coverShape {
	if shape.Bounds.w < 8 {
		shape.Bounds.w = 8
	}
	if shape.Bounds.h < 8 {
		shape.Bounds.h = 8
	}
	if shape.Bounds.w > artArea.w {
		shape.Bounds.w = artArea.w
	}
	if shape.Bounds.h > artArea.h {
		shape.Bounds.h = artArea.h
	}

	if shape.Bounds.x < artArea.x {
		shape.Bounds.x = artArea.x
	}
	if shape.Bounds.y < artArea.y {
		shape.Bounds.y = artArea.y
	}
	if shape.Bounds.x+shape.Bounds.w > artArea.x+artArea.w {
		shape.Bounds.x = artArea.x + artArea.w - shape.Bounds.w
	}
	if shape.Bounds.y+shape.Bounds.h > artArea.y+artArea.h {
		shape.Bounds.y = artArea.y + artArea.h - shape.Bounds.h
	}

	shape = withCoverShapePivot(shape)

	for range 4 {
		env := coverShapeEnvelope(shape)
		dx := 0.0
		dy := 0.0

		if env.x < artArea.x {
			dx += artArea.x - env.x
		}
		if env.x+env.w > artArea.x+artArea.w {
			dx -= env.x + env.w - (artArea.x + artArea.w)
		}
		if env.y < artArea.y {
			dy += artArea.y - env.y
		}
		if env.y+env.h > artArea.y+artArea.h {
			dy -= env.y + env.h - (artArea.y + artArea.h)
		}

		if dx != 0 || dy != 0 {
			shape.Bounds.x += dx
			shape.Bounds.y += dy
			shape = withCoverShapePivot(shape)
			continue
		}

		if rectsIntersect(env, exclusion) {
			clearBelow := exclusion.y + exclusion.h + 5.0
			if shape.Bounds.y < clearBelow {
				shape.Bounds.y += clearBelow - shape.Bounds.y
			}
			shape = withCoverShapePivot(shape)
			env = coverShapeEnvelope(shape)
			if rectsIntersect(env, exclusion) {
				targetRight := exclusion.x - 5.0
				shape.Bounds.x -= env.x + env.w - targetRight
				shape = withCoverShapePivot(shape)
			}
			continue
		}

		break
	}

	return shape
}

func withCoverShapePivot(shape coverShape) coverShape {
	shape.PivotX = shape.Bounds.x + shape.Bounds.w*0.5
	shape.PivotY = shape.Bounds.y + shape.Bounds.h*0.5
	return shape
}

func pickCoverFillMode(family coverCompositionFamily, rng coverRand) coverFillMode {
	switch family {
	case coverFamilyFrame:
		return []coverFillMode{coverFillSolid, coverFillSlits, coverFillOpenGrid, coverFillBanded}[rng.IntN(4)]
	case coverFamilyStack:
		return []coverFillMode{coverFillSolid, coverFillBanded, coverFillSlits, coverFillOpenGrid}[rng.IntN(4)]
	case coverFamilyHinge:
		return []coverFillMode{coverFillSolid, coverFillBanded, coverFillPunctured, coverFillSlits}[rng.IntN(4)]
	default:
		return []coverFillMode{coverFillPunctured, coverFillOpenGrid, coverFillSolid, coverFillBanded}[rng.IntN(4)]
	}
}

func pickCoverTexture(family coverCompositionFamily, rng coverRand) coverTextureKind {
	switch family {
	case coverFamilyFrame:
		return []coverTextureKind{coverTextureNone, coverTextureStripes, coverTextureChecker, coverTextureDots, coverTextureLattice}[rng.IntN(5)]
	case coverFamilyStack:
		return []coverTextureKind{coverTextureNone, coverTextureStripes, coverTextureHatch, coverTextureChecker, coverTextureLattice}[rng.IntN(5)]
	case coverFamilyHinge:
		return []coverTextureKind{coverTextureNone, coverTextureHatch, coverTextureStripes, coverTextureDots, coverTextureLattice}[rng.IntN(5)]
	default:
		return []coverTextureKind{coverTextureNone, coverTextureDots, coverTextureChecker, coverTextureHatch, coverTextureStripes}[rng.IntN(5)]
	}
}

func buildFamilyCutouts(family coverCompositionFamily, shape coverShape, rng coverRand) []coverCutout {
	switch family {
	case coverFamilyFrame:
		return buildFrameCutouts(shape, rng)
	case coverFamilyStack:
		return buildStackCutouts(shape, rng)
	case coverFamilyHinge:
		return buildHingeCutouts(shape, rng)
	default:
		return buildIslandCutouts(shape, rng)
	}
}

func buildFrameCutouts(shape coverShape, rng coverRand) []coverCutout {
	cutouts := []coverCutout{
		{
			Kind: coverCutoutRect,
			X:    shape.Bounds.x + shape.Bounds.w*(0.14+rng.Float64()*0.32),
			Y:    shape.Bounds.y + shape.Bounds.h*(0.18+rng.Float64()*0.40),
			W:    shape.Bounds.w * (0.12 + rng.Float64()*0.18),
			H:    shape.Bounds.h * (0.12 + rng.Float64()*0.22),
		},
	}
	if rng.IntN(2) == 0 {
		cutouts = append(cutouts, coverCutout{
			Kind: coverCutoutCircle,
			X:    shape.Bounds.x + shape.Bounds.w*(0.26+rng.Float64()*0.48),
			Y:    shape.Bounds.y + shape.Bounds.h*(0.22+rng.Float64()*0.48),
			R:    min(shape.Bounds.w, shape.Bounds.h) * (0.08 + rng.Float64()*0.10),
		})
	}
	return cutouts
}

func buildStackCutouts(shape coverShape, rng coverRand) []coverCutout {
	if rng.IntN(2) == 0 {
		return []coverCutout{
			{
				Kind: coverCutoutRect,
				X:    shape.Bounds.x + shape.Bounds.w*(0.10+rng.Float64()*0.12),
				Y:    shape.Bounds.y + shape.Bounds.h*(0.36+rng.Float64()*0.18),
				W:    shape.Bounds.w * (0.42 + rng.Float64()*0.22),
				H:    shape.Bounds.h * (0.10 + rng.Float64()*0.10),
			},
		}
	}
	return []coverCutout{
		{
			Kind: coverCutoutRect,
			X:    shape.Bounds.x + shape.Bounds.w*(0.36+rng.Float64()*0.18),
			Y:    shape.Bounds.y + shape.Bounds.h*(0.08+rng.Float64()*0.12),
			W:    shape.Bounds.w * (0.12 + rng.Float64()*0.12),
			H:    shape.Bounds.h * (0.34 + rng.Float64()*0.20),
		},
	}
}

func buildHingeCutouts(shape coverShape, rng coverRand) []coverCutout {
	cutouts := []coverCutout{
		{
			Kind: coverCutoutCircle,
			X:    shape.Bounds.x + shape.Bounds.w*(0.34+rng.Float64()*0.24),
			Y:    shape.Bounds.y + shape.Bounds.h*(0.30+rng.Float64()*0.26),
			R:    min(shape.Bounds.w, shape.Bounds.h) * (0.08 + rng.Float64()*0.08),
		},
	}
	if rng.IntN(2) == 0 {
		cutouts = append(cutouts, coverCutout{
			Kind: coverCutoutRect,
			X:    shape.Bounds.x + shape.Bounds.w*(0.58+rng.Float64()*0.10),
			Y:    shape.Bounds.y + shape.Bounds.h*(0.56+rng.Float64()*0.10),
			W:    shape.Bounds.w * (0.18 + rng.Float64()*0.12),
			H:    shape.Bounds.h * (0.14 + rng.Float64()*0.10),
		})
	}
	return cutouts
}

func buildIslandCutouts(shape coverShape, rng coverRand) []coverCutout {
	return []coverCutout{
		{
			Kind: coverCutoutCircle,
			X:    shape.Bounds.x + shape.Bounds.w*(0.18+rng.Float64()*0.54),
			Y:    shape.Bounds.y + shape.Bounds.h*(0.18+rng.Float64()*0.54),
			R:    min(shape.Bounds.w, shape.Bounds.h) * (0.10 + rng.Float64()*0.10),
		},
	}
}

func drawCoverArtLayout(pdf *fpdf.Fpdf, layout coverArtLayout) {
	for _, shape := range layout.Shapes {
		drawCoverShape(pdf, shape)
	}
}

func drawCoverShape(pdf *fpdf.Fpdf, shape coverShape) {
	pdf.TransformBegin()
	pdf.TransformRotate(shape.Rotate, shape.PivotX, shape.PivotY)
	switch shape.Kind {
	case coverShapeCircle:
		drawCircleCoverShape(pdf, shape)
	case coverShapeCroppedCircle:
		drawCroppedCircleCoverShape(pdf, shape)
	case coverShapeStepped:
		drawSteppedCoverShape(pdf, shape)
	default:
		drawRectCoverShape(pdf, shape)
	}
	pdf.TransformEnd()
}

func drawRectCoverShape(pdf *fpdf.Fpdf, shape coverShape) {
	pdf.SetFillColor(8, 8, 8)
	pdf.Rect(shape.Bounds.x, shape.Bounds.y, shape.Bounds.w, shape.Bounds.h, "F")
	pdf.ClipRect(shape.Bounds.x, shape.Bounds.y, shape.Bounds.w, shape.Bounds.h, false)
	drawCoverFill(pdf, shape)
	drawCoverTexture(pdf, shape)
	pdf.ClipEnd()
	drawCoverCutouts(pdf, shape.Cutouts)
}

func drawCircleCoverShape(pdf *fpdf.Fpdf, shape coverShape) {
	cx, cy, r := coverCircleMetrics(shape.Bounds)
	pdf.SetFillColor(8, 8, 8)
	pdf.Circle(cx, cy, r, "F")
	pdf.ClipCircle(cx, cy, r, false)
	drawCoverFill(pdf, shape)
	drawCoverTexture(pdf, shape)
	pdf.ClipEnd()
	drawCoverCutouts(pdf, shape.Cutouts)
}

func drawCroppedCircleCoverShape(pdf *fpdf.Fpdf, shape coverShape) {
	cx, cy, r := croppedCircleMetrics(shape)
	pdf.SetFillColor(8, 8, 8)
	pdf.ClipRect(shape.Bounds.x, shape.Bounds.y, shape.Bounds.w, shape.Bounds.h, false)
	pdf.Circle(cx, cy, r, "F")
	pdf.ClipCircle(cx, cy, r, false)
	drawCoverFill(pdf, shape)
	drawCoverTexture(pdf, shape)
	pdf.ClipEnd()
	pdf.ClipEnd()
	drawCoverCutouts(pdf, shape.Cutouts)
}

func drawSteppedCoverShape(pdf *fpdf.Fpdf, shape coverShape) {
	points := steppedShapePoints(shape)
	pdf.SetFillColor(8, 8, 8)
	pdf.Polygon(points, "F")
	pdf.ClipPolygon(points, false)
	drawCoverFill(pdf, shape)
	drawCoverTexture(pdf, shape)
	pdf.ClipEnd()
	drawCoverCutouts(pdf, shape.Cutouts)
}

func coverCircleMetrics(bounds rectMM) (cx, cy, r float64) {
	r = min(bounds.w, bounds.h) * 0.5
	cx = bounds.x + bounds.w*0.5
	cy = bounds.y + bounds.h*0.5
	return cx, cy, r
}

func croppedCircleMetrics(shape coverShape) (cx, cy, r float64) {
	cx, cy, _ = coverCircleMetrics(shape.Bounds)
	r = max(shape.Bounds.w, shape.Bounds.h) * 0.56
	switch shape.Edge {
	case coverEdgeLeft:
		cx -= shape.Bounds.w * 0.18
	case coverEdgeRight:
		cx += shape.Bounds.w * 0.18
	case coverEdgeTop:
		cy -= shape.Bounds.h * 0.18
	default:
		cy += shape.Bounds.h * 0.18
	}
	return cx, cy, r
}

func steppedShapePoints(shape coverShape) []fpdf.PointType {
	b := shape.Bounds
	notch := min(b.w, b.h) * 0.18
	step1 := min(b.w, b.h) * 0.10
	step2 := min(b.w, b.h) * 0.20

	switch shape.Edge {
	case coverEdgeLeft:
		return []fpdf.PointType{
			{X: b.x + notch, Y: b.y},
			{X: b.x + b.w, Y: b.y},
			{X: b.x + b.w, Y: b.y + b.h},
			{X: b.x + notch, Y: b.y + b.h},
			{X: b.x + notch, Y: b.y + b.h*0.66},
			{X: b.x, Y: b.y + b.h*0.66 - step2},
			{X: b.x, Y: b.y + b.h*0.34 + step2},
			{X: b.x + notch, Y: b.y + b.h*0.34},
		}
	case coverEdgeRight:
		return []fpdf.PointType{
			{X: b.x, Y: b.y},
			{X: b.x + b.w - notch, Y: b.y},
			{X: b.x + b.w - notch, Y: b.y + b.h*0.34},
			{X: b.x + b.w, Y: b.y + b.h*0.34 + step1},
			{X: b.x + b.w, Y: b.y + b.h*0.66 - step1},
			{X: b.x + b.w - notch, Y: b.y + b.h*0.66},
			{X: b.x + b.w - notch, Y: b.y + b.h},
			{X: b.x, Y: b.y + b.h},
		}
	case coverEdgeTop:
		return []fpdf.PointType{
			{X: b.x, Y: b.y + notch},
			{X: b.x + b.w*0.34, Y: b.y + notch},
			{X: b.x + b.w*0.34 + step1, Y: b.y},
			{X: b.x + b.w*0.66 - step1, Y: b.y},
			{X: b.x + b.w*0.66, Y: b.y + notch},
			{X: b.x + b.w, Y: b.y + notch},
			{X: b.x + b.w, Y: b.y + b.h},
			{X: b.x, Y: b.y + b.h},
		}
	default:
		return []fpdf.PointType{
			{X: b.x, Y: b.y},
			{X: b.x + b.w, Y: b.y},
			{X: b.x + b.w, Y: b.y + b.h - notch},
			{X: b.x + b.w*0.66, Y: b.y + b.h - notch},
			{X: b.x + b.w*0.66 - step1, Y: b.y + b.h},
			{X: b.x + b.w*0.34 + step1, Y: b.y + b.h},
			{X: b.x + b.w*0.34, Y: b.y + b.h - notch},
			{X: b.x, Y: b.y + b.h - notch},
		}
	}
}

func drawCoverFill(pdf *fpdf.Fpdf, shape coverShape) {
	pdf.SetFillColor(255, 255, 255)

	switch shape.Fill {
	case coverFillSlits:
		drawCoverSlits(pdf, shape.Bounds)
	case coverFillOpenGrid:
		drawCoverOpenGrid(pdf, shape.Bounds)
	case coverFillPunctured:
		drawCoverPunctures(pdf, shape.Bounds)
	case coverFillBanded:
		drawCoverBands(pdf, shape.Bounds)
	default:
		return
	}
}

func drawCoverSlits(pdf *fpdf.Fpdf, bounds rectMM) {
	if bounds.w >= bounds.h {
		slitW := max(1.6, bounds.w*0.05)
		for x := bounds.x + bounds.w*0.12; x < bounds.x+bounds.w*0.88; x += slitW * 1.9 {
			pdf.Rect(x, bounds.y-1, slitW, bounds.h+2, "F")
		}
		return
	}

	slitH := max(1.6, bounds.h*0.05)
	for y := bounds.y + bounds.h*0.12; y < bounds.y+bounds.h*0.88; y += slitH * 1.9 {
		pdf.Rect(bounds.x-1, y, bounds.w+2, slitH, "F")
	}
}

func drawCoverOpenGrid(pdf *fpdf.Fpdf, bounds rectMM) {
	cell := max(3.2, min(bounds.w, bounds.h)*0.12)
	for row, y := 0, bounds.y; y < bounds.y+bounds.h; row, y = row+1, y+cell {
		for col, x := 0, bounds.x; x < bounds.x+bounds.w; col, x = col+1, x+cell {
			if (row+col)%2 == 0 {
				pdf.Rect(x, y, cell*0.78, cell*0.78, "F")
			}
		}
	}
}

func drawCoverPunctures(pdf *fpdf.Fpdf, bounds rectMM) {
	step := max(4.4, min(bounds.w, bounds.h)*0.16)
	radius := max(1.0, step*0.18)
	for row, y := 0, bounds.y+radius; y < bounds.y+bounds.h-radius; row, y = row+1, y+step {
		shift := 0.0
		if row%2 == 1 {
			shift = step * 0.45
		}
		for x := bounds.x + radius + shift; x < bounds.x+bounds.w-radius; x += step {
			pdf.Circle(x, y, radius, "F")
		}
	}
}

func drawCoverBands(pdf *fpdf.Fpdf, bounds rectMM) {
	if bounds.w >= bounds.h {
		bandH := max(3.4, bounds.h*0.16)
		for i := 0; i < 2; i++ {
			y := bounds.y + bounds.h*(0.18+float64(i)*0.34)
			pdf.Rect(bounds.x-1, y, bounds.w+2, bandH, "F")
		}
		return
	}

	bandW := max(3.4, bounds.w*0.16)
	for i := 0; i < 2; i++ {
		x := bounds.x + bounds.w*(0.18+float64(i)*0.34)
		pdf.Rect(x, bounds.y-1, bandW, bounds.h+2, "F")
	}
}

func drawCoverTexture(pdf *fpdf.Fpdf, shape coverShape) {
	setCoverRemovalStroke(pdf, shape.Weight)

	switch shape.Texture {
	case coverTextureStripes:
		barW := max(1.2, shape.Step*0.34)
		for x := shape.Bounds.x - shape.Step; x <= shape.Bounds.x+shape.Bounds.w+shape.Step; x += shape.Step {
			pdf.SetFillColor(255, 255, 255)
			pdf.Rect(x, shape.Bounds.y-1, barW, shape.Bounds.h+2, "F")
		}
	case coverTextureHatch:
		for offset := -shape.Bounds.h; offset <= shape.Bounds.w+shape.Bounds.h; offset += shape.Step {
			pdf.Line(shape.Bounds.x+offset, shape.Bounds.y+shape.Bounds.h, shape.Bounds.x+offset+shape.Bounds.h, shape.Bounds.y)
		}
	case coverTextureDots:
		radius := max(0.9, shape.Step*0.16)
		rowStep := shape.Step * 0.94
		for row, y := 0, shape.Bounds.y+radius; y <= shape.Bounds.y+shape.Bounds.h-radius; row, y = row+1, y+rowStep {
			shift := 0.0
			if row%2 == 1 {
				shift = shape.Step * 0.42
			}
			for x := shape.Bounds.x + radius + shift; x <= shape.Bounds.x+shape.Bounds.w-radius; x += shape.Step {
				pdf.SetFillColor(255, 255, 255)
				pdf.Circle(x, y, radius, "F")
			}
		}
	case coverTextureChecker:
		cell := max(2.6, shape.Step*0.68)
		for row, y := 0, shape.Bounds.y; y <= shape.Bounds.y+shape.Bounds.h; row, y = row+1, y+cell {
			for col, x := 0, shape.Bounds.x; x <= shape.Bounds.x+shape.Bounds.w; col, x = col+1, x+cell {
				if (row+col)%2 == 0 {
					pdf.SetFillColor(255, 255, 255)
					pdf.Rect(x, y, cell, cell, "F")
				}
			}
		}
	case coverTextureLattice:
		barW := max(0.95, shape.Step*0.16)
		for x := shape.Bounds.x; x <= shape.Bounds.x+shape.Bounds.w; x += shape.Step {
			pdf.SetFillColor(255, 255, 255)
			pdf.Rect(x, shape.Bounds.y-1, barW, shape.Bounds.h+2, "F")
		}
		for y := shape.Bounds.y; y <= shape.Bounds.y+shape.Bounds.h; y += shape.Step * 0.84 {
			pdf.SetFillColor(255, 255, 255)
			pdf.Rect(shape.Bounds.x-1, y, shape.Bounds.w+2, barW, "F")
		}
	default:
		return
	}
}

func setCoverRemovalStroke(pdf *fpdf.Fpdf, weight float64) {
	pdf.SetDrawColor(255, 255, 255)
	pdf.SetLineWidth(weight)
}

func drawCoverCutouts(pdf *fpdf.Fpdf, cutouts []coverCutout) {
	pdf.SetFillColor(255, 255, 255)
	for _, cutout := range cutouts {
		switch cutout.Kind {
		case coverCutoutCircle:
			pdf.Circle(cutout.X, cutout.Y, cutout.R, "F")
		default:
			pdf.Rect(cutout.X, cutout.Y, cutout.W, cutout.H, "F")
		}
	}
}

func rectsIntersect(a, b rectMM) bool {
	return a.x < b.x+b.w && a.x+a.w > b.x && a.y < b.y+b.h && a.y+a.h > b.y
}

func rotatePoint(x, y, pivotX, pivotY, angleDeg float64) (float64, float64) {
	rad := angleDeg * math.Pi / 180.0
	sinA, cosA := math.Sincos(rad)
	dx := x - pivotX
	dy := y - pivotY
	return pivotX + dx*cosA - dy*sinA, pivotY + dx*sinA + dy*cosA
}

func rotatedRectAABB(bounds rectMM, pivotX, pivotY, angleDeg float64) rectMM {
	if angleDeg == 0 {
		return bounds
	}
	points := [4][2]float64{
		{bounds.x, bounds.y},
		{bounds.x + bounds.w, bounds.y},
		{bounds.x + bounds.w, bounds.y + bounds.h},
		{bounds.x, bounds.y + bounds.h},
	}
	minX := math.MaxFloat64
	minY := math.MaxFloat64
	maxX := -math.MaxFloat64
	maxY := -math.MaxFloat64
	for _, pt := range points {
		x, y := rotatePoint(pt[0], pt[1], pivotX, pivotY, angleDeg)
		minX = min(minX, x)
		minY = min(minY, y)
		maxX = max(maxX, x)
		maxY = max(maxY, y)
	}
	return rectMM{x: minX, y: minY, w: maxX - minX, h: maxY - minY}
}

func coverShapeEnvelope(shape coverShape) rectMM {
	return rotatedRectAABB(shape.Bounds, shape.PivotX, shape.PivotY, shape.Rotate)
}

func coverArtLayoutEqual(a, b coverArtLayout) bool {
	return reflect.DeepEqual(a, b)
}

func coverArtStaysWithinArea(layout coverArtLayout) bool {
	for _, shape := range layout.Shapes {
		env := coverShapeEnvelope(shape)
		if env.x < layout.ArtArea.x || env.y < layout.ArtArea.y {
			return false
		}
		if env.x+env.w > layout.ArtArea.x+layout.ArtArea.w {
			return false
		}
		if env.y+env.h > layout.ArtArea.y+layout.ArtArea.h {
			return false
		}
	}
	return true
}

func coverArtRespectsLockup(layout coverArtLayout) bool {
	for _, shape := range layout.Shapes {
		if rectsIntersect(coverShapeEnvelope(shape), layout.LockupExclusion) {
			return false
		}
	}
	return true
}

func coverArtDiffers(a, b coverArtLayout) bool {
	return !reflect.DeepEqual(a, b)
}

func coverFamilyName(family coverCompositionFamily) string {
	switch family {
	case coverFamilyFrame:
		return "frame"
	case coverFamilyStack:
		return "stack"
	case coverFamilyHinge:
		return "hinge"
	default:
		return "island"
	}
}

func coverDirectionName(direction coverDirection) string {
	switch direction {
	case coverDirectionVertical:
		return "vertical"
	case coverDirectionHorizontal:
		return "horizontal"
	case coverDirectionDiagonal:
		return "diagonal"
	default:
		return "clustered"
	}
}

func coverShapeKindName(kind coverShapeKind) string {
	switch kind {
	case coverShapeRect:
		return "rect"
	case coverShapeCircle:
		return "circle"
	case coverShapeCroppedCircle:
		return "cropped-circle"
	default:
		return "stepped"
	}
}

func coverTextureCount(layout coverArtLayout) int {
	seen := map[coverTextureKind]struct{}{}
	for _, shape := range layout.Shapes {
		if shape.Texture == coverTextureNone {
			continue
		}
		seen[shape.Texture] = struct{}{}
	}
	return len(seen)
}

func coverPrimitiveMixSignature(layout coverArtLayout) string {
	parts := map[string]struct{}{}
	for _, shape := range layout.Shapes {
		parts[coverShapeKindName(shape.Kind)] = struct{}{}
	}

	keys := make([]string, 0, len(parts))
	for key := range parts {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return strings.Join(keys, "+")
}

func coverRotationSignature(layout coverArtLayout) string {
	parts := make([]string, 0, len(layout.Shapes))
	for _, shape := range layout.Shapes {
		parts = append(parts, fmt.Sprintf("%.1f", shape.Rotate))
	}
	sort.Strings(parts)
	return strings.Join(parts, ",")
}

func coverRotationPolaritySignature(layout coverArtLayout) string {
	hasPos := false
	hasNeg := false
	for _, shape := range layout.Shapes {
		if shape.Rotate > 0 {
			hasPos = true
		}
		if shape.Rotate < 0 {
			hasNeg = true
		}
	}
	switch {
	case hasPos && hasNeg:
		return "mixed"
	case hasPos:
		return "positive"
	case hasNeg:
		return "negative"
	default:
		return "flat"
	}
}

func coverArtInkBounds(layout coverArtLayout) rectMM {
	if len(layout.Shapes) == 0 {
		return rectMM{}
	}

	minX := math.MaxFloat64
	minY := math.MaxFloat64
	maxX := 0.0
	maxY := 0.0
	for _, shape := range layout.Shapes {
		env := coverShapeEnvelope(shape)
		minX = min(minX, env.x)
		minY = min(minY, env.y)
		maxX = max(maxX, env.x+env.w)
		maxY = max(maxY, env.y+env.h)
	}
	return rectMM{x: minX, y: minY, w: maxX - minX, h: maxY - minY}
}

func coverInkAspectBucket(layout coverArtLayout) coverAspectBucket {
	bounds := coverArtInkBounds(layout)
	if bounds.w <= 0 || bounds.h <= 0 {
		return coverAspectCompact
	}

	if bounds.h/bounds.w >= 1.28 {
		return coverAspectTall
	}
	if bounds.w/bounds.h >= 1.28 {
		return coverAspectWide
	}

	centerX := bounds.x + bounds.w*0.5
	centerY := bounds.y + bounds.h*0.5
	artCenterX := layout.ArtArea.x + layout.ArtArea.w*0.5
	artCenterY := layout.ArtArea.y + layout.ArtArea.h*0.5
	offset := math.Hypot(centerX-artCenterX, centerY-artCenterY)
	if offset > min(layout.ArtArea.w, layout.ArtArea.h)*0.12 {
		return coverAspectOffset
	}
	return coverAspectCompact
}

func coverAspectBucketName(bucket coverAspectBucket) string {
	switch bucket {
	case coverAspectTall:
		return "tall"
	case coverAspectWide:
		return "wide"
	case coverAspectOffset:
		return "offset"
	default:
		return "compact"
	}
}
