package theme

import (
	"image/color"
	"math"
)

// relativeLuminance returns the WCAG 2.x relative luminance of a color.
// See https://www.w3.org/TR/WCAG21/#dfn-relative-luminance
func relativeLuminance(c color.Color) float64 {
	r, g, b, _ := c.RGBA()
	// Convert from 16-bit pre-multiplied to 0..1 sRGB.
	sr := linearize(float64(r) / 0xffff)
	sg := linearize(float64(g) / 0xffff)
	sb := linearize(float64(b) / 0xffff)
	return 0.2126*sr + 0.7152*sg + 0.0722*sb
}

// linearize converts an sRGB channel value (0..1) to linear light.
func linearize(v float64) float64 {
	if v <= 0.04045 {
		return v / 12.92
	}
	return math.Pow((v+0.055)/1.055, 2.4)
}

// contrastRatio returns the WCAG 2.x contrast ratio between two colors.
// The result is in the range [1, 21].
func contrastRatio(a, b color.Color) float64 {
	la := relativeLuminance(a)
	lb := relativeLuminance(b)
	if la < lb {
		la, lb = lb, la
	}
	return (la + 0.05) / (lb + 0.05)
}

// minContrastNormal is the WCAG AA minimum for normal text (4.5:1).
const minContrastNormal = 4.5

// minContrastLarge is the WCAG AA minimum for large/bold text (3:1).
const minContrastLarge = 3.0

// ensurePairContrast ensures that the foreground/background pair meets the
// minimum contrast ratio. When the original fg falls short, it scans the
// full 16-color ANSI palette for the candidate that meets the ratio and
// has the closest relative luminance to the original fg — preserving the
// intended lightness rather than jumping to an extreme.
//
// If no ANSI color qualifies, the closest-luminance ANSI candidate is
// nudged toward white or black just enough to satisfy the ratio, keeping
// the result close to the palette's tonal range.
func ensurePairContrast(fg, bg color.Color, p Palette, minRatio float64) color.Color {
	if contrastRatio(fg, bg) >= minRatio {
		return fg
	}

	targetLum := relativeLuminance(fg)

	// Find the closest-luminance ANSI color that meets contrast, and
	// separately track the closest-luminance ANSI color overall as a
	// nudge candidate if nothing qualifies outright.
	var bestOK, bestAny color.Color
	bestOKDist := math.Inf(1)
	bestAnyDist := math.Inf(1)

	for _, c := range p.ANSI {
		dist := math.Abs(relativeLuminance(c) - targetLum)
		if dist < bestAnyDist {
			bestAny = c
			bestAnyDist = dist
		}
		if contrastRatio(c, bg) >= minRatio && dist < bestOKDist {
			bestOK = c
			bestOKDist = dist
		}
	}

	if bestOK != nil {
		return bestOK
	}

	// No ANSI color meets the ratio — nudge the closest candidate.
	return nudgeContrast(bestAny, bg, minRatio)
}

// nudgeContrast blends fg toward white or black until the contrast ratio
// against bg meets minRatio. The direction (lighter or darker) is chosen
// by moving away from the background luminance, so the result stays as
// close to the original fg as possible.
func nudgeContrast(fg, bg color.Color, minRatio float64) color.Color {
	fgLum := relativeLuminance(fg)
	bgLum := relativeLuminance(bg)

	type endpoint struct{ r, g, b uint8 }
	light := endpoint{0xff, 0xff, 0xff}
	dark := endpoint{0x00, 0x00, 0x00}

	// Move away from the background luminance.
	target := light
	if fgLum <= bgLum {
		target = dark
	}

	// If the chosen extreme can't reach minRatio, flip.
	tColor := color.NRGBA{R: target.r, G: target.g, B: target.b, A: 0xff}
	if contrastRatio(tColor, bg) < minRatio {
		if target == light {
			target = dark
		} else {
			target = light
		}
	}

	fr, fg2, fb, _ := fg.RGBA()
	fr8, fg8, fb8 := uint8(fr>>8), uint8(fg2>>8), uint8(fb>>8)

	// Binary search for the minimum blend factor.
	lo, hi := 0.0, 1.0
	for range 32 {
		mid := (lo + hi) / 2
		c := blendRGB(fr8, fg8, fb8, target.r, target.g, target.b, mid)
		if contrastRatio(c, bg) >= minRatio {
			hi = mid
		} else {
			lo = mid
		}
	}
	return blendRGB(fr8, fg8, fb8, target.r, target.g, target.b, hi)
}

// blendRGB linearly interpolates between two RGB colors at factor t ∈ [0,1].
func blendRGB(r1, g1, b1, r2, g2, b2 uint8, t float64) color.Color {
	lerp := func(a, b uint8) uint8 {
		return uint8(float64(a)*(1-t) + float64(b)*t + 0.5)
	}
	return color.NRGBA{R: lerp(r1, r2), G: lerp(g1, g2), B: lerp(b1, b2), A: 0xff}
}

// TextOnBG returns the current palette's FG or BG color — whichever has
// better contrast against bg. This is useful for choosing readable text
// over an arbitrary background (e.g. colored rectangles in shikaku).
func TextOnBG(bg color.Color) color.Color {
	p := Current()
	if contrastRatio(p.FG, bg) >= contrastRatio(p.BG, bg) {
		return p.FG
	}
	return p.BG
}

// Blend returns a color that is t fraction of the way from a toward b.
// t=0 returns a, t=1 returns b. Used to derive muted background colors
// from foreground accent colors.
func Blend(a, b color.Color, t float64) color.Color {
	ar, ag, ab, _ := a.RGBA()
	br, bg, bb, _ := b.RGBA()
	return blendRGB(uint8(ar>>8), uint8(ag>>8), uint8(ab>>8),
		uint8(br>>8), uint8(bg>>8), uint8(bb>>8), t)
}

// MidTone returns a color perceptually halfway between a and b by
// blending their sRGB channels. This produces a muted tone suitable for
// dim/secondary text that is guaranteed to sit between the two anchors.
func MidTone(a, b color.Color) color.Color {
	ar, ag, ab, _ := a.RGBA()
	br, bg, bb, _ := b.RGBA()
	return color.NRGBA{
		R: uint8(((ar >> 8) + (br >> 8)) / 2),
		G: uint8(((ag >> 8) + (bg >> 8)) / 2),
		B: uint8(((ab >> 8) + (bb >> 8)) / 2),
		A: 0xff,
	}
}
