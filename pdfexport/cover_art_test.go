package pdfexport

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"math"
	"testing"
)

// --- Cover art seeding (P1) ---

func TestRenderCoverArtworkPNGDeterministic(t *testing.T) {
	base := RGB{R: 86, G: 124, B: 149}
	seed := "quiet-fjord"

	first := renderCoverArtworkPNG(seed, "front", base)
	second := renderCoverArtworkPNG(seed, "front", base)

	if len(first) == 0 || len(second) == 0 {
		t.Fatalf("renderCoverArtworkPNG returned empty output")
	}
	if !bytes.Equal(first, second) {
		t.Fatalf("expected identical PNG bytes for identical seed and variant")
	}
}

func TestRenderCoverArtworkPNGVariesBySeedAndVariant(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping cover art variation check in short mode")
	}

	base := RGB{R: 90, G: 128, B: 154}
	frontA := renderCoverArtworkPNG("quiet-fjord", "front", base)
	frontB := renderCoverArtworkPNG("quiet-fjord-alt", "front", base)
	backA := renderCoverArtworkPNG("quiet-fjord", "back", base)

	assertArtworkDiff(t, frontA, frontB, 0.03)
	assertArtworkDiff(t, frontA, backA, 0.03)
}

func TestCoverArtworkBatchSaturationFloor(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping batch saturation check in short mode")
	}

	base := RGB{R: 90, G: 128, B: 154}
	seeds := coverSeedSamples(30)
	var total float64
	for _, seed := range seeds {
		img := decodePNG(t, renderCoverArtworkPNG(seed, "front", base))
		total += sampledMeanSaturation(img, 14)
	}
	meanSaturation := total / float64(len(seeds))
	if meanSaturation < 0.24 {
		t.Fatalf("mean saturation too low: got %.4f want >= 0.2400", meanSaturation)
	}
	if meanSaturation > 0.36 {
		t.Fatalf("mean saturation unexpectedly high: got %.4f want <= 0.3600", meanSaturation)
	}
}

func TestCoverArtworkBatchDiversityFloor(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping batch diversity check in short mode")
	}

	base := RGB{R: 90, G: 128, B: 154}
	seeds := coverSeedSamples(30)
	images := make([]image.Image, 0, len(seeds))
	for _, seed := range seeds {
		images = append(images, decodePNG(t, renderCoverArtworkPNG(seed, "front", base)))
	}

	minDiff := math.MaxFloat64
	var total float64
	var pairs int
	for i := range images {
		for j := i + 1; j < len(images); j++ {
			diff := sampledMeanChannelDiff(images[i], images[j], 12)
			if diff < minDiff {
				minDiff = diff
			}
			total += diff
			pairs++
		}
	}
	meanDiff := total / float64(pairs)
	if meanDiff < 0.090 {
		t.Fatalf("mean pairwise diff too low: got %.4f want >= 0.0900", meanDiff)
	}
	if minDiff < 0.040 {
		t.Fatalf("minimum pairwise diff too low: got %.4f want >= 0.0400", minDiff)
	}
}

func TestCoverArtworkArchetypeCoverage(t *testing.T) {
	seen := make(map[coverArchetype]bool, int(coverArchetypeCount))
	for i := 1; i <= 120; i++ {
		seed := fmt.Sprintf("coverage-%03d|front", i)
		seen[coverArchetypeForSeed(coverSeedHash(seed))] = true
	}

	if len(seen) != int(coverArchetypeCount) {
		t.Fatalf("archetype coverage incomplete: got %d want %d", len(seen), coverArchetypeCount)
	}
}

func TestRenderCoverArtworkPNGHandlesLowAndHighChromaBase(t *testing.T) {
	lowChroma := RGB{R: 128, G: 128, B: 128}
	highChroma := RGB{R: 245, G: 62, B: 58}

	low := renderCoverArtworkPNG("chroma-check", "front", lowChroma)
	high := renderCoverArtworkPNG("chroma-check", "front", highChroma)

	if len(low) == 0 || len(high) == 0 {
		t.Fatalf("renderCoverArtworkPNG returned empty output for chroma check")
	}
	assertArtworkDiff(t, low, high, 0.03)
}

func assertArtworkDiff(t *testing.T, left, right []byte, minMeanDiff float64) {
	t.Helper()

	if bytes.Equal(left, right) {
		t.Fatalf("unexpected byte-identical images for different inputs")
	}

	imgA := decodePNG(t, left)
	imgB := decodePNG(t, right)
	diff := sampledMeanChannelDiff(imgA, imgB, 12)
	if diff < minMeanDiff {
		t.Fatalf("mean sampled channel diff too low: got %.4f want >= %.4f", diff, minMeanDiff)
	}
}

func decodePNG(t *testing.T, data []byte) image.Image {
	t.Helper()
	img, err := png.Decode(bytes.NewReader(data))
	if err != nil {
		t.Fatalf("png.Decode error: %v", err)
	}
	return img
}

func sampledMeanChannelDiff(a, b image.Image, step int) float64 {
	bounds := a.Bounds()
	if step < 1 {
		step = 1
	}

	var total float64
	var samples int
	for y := bounds.Min.Y; y < bounds.Max.Y; y += step {
		for x := bounds.Min.X; x < bounds.Max.X; x += step {
			ar, ag, ab, _ := a.At(x, y).RGBA()
			br, bg, bb, _ := b.At(x, y).RGBA()
			total += math.Abs(float64(ar)-float64(br)) / 65535
			total += math.Abs(float64(ag)-float64(bg)) / 65535
			total += math.Abs(float64(ab)-float64(bb)) / 65535
			samples += 3
		}
	}

	if samples == 0 {
		return 0
	}
	return total / float64(samples)
}

func sampledMeanSaturation(img image.Image, step int) float64 {
	bounds := img.Bounds()
	if step < 1 {
		step = 1
	}

	var total float64
	var samples int
	for y := bounds.Min.Y; y < bounds.Max.Y; y += step {
		for x := bounds.Min.X; x < bounds.Max.X; x += step {
			r, g, b, _ := img.At(x, y).RGBA()
			c := color.RGBA{R: uint8(r >> 8), G: uint8(g >> 8), B: uint8(b >> 8), A: 255}
			_, sat, _ := rgbToHSV(c)
			total += sat
			samples++
		}
	}

	if samples == 0 {
		return 0
	}
	return total / float64(samples)
}

func coverSeedSamples(n int) []string {
	out := make([]string, 0, n)
	for i := 1; i <= n; i++ {
		out = append(out, fmt.Sprintf("test-cover-%02d", i))
	}
	return out
}
