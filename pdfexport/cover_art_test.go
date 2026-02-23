package pdfexport

import (
	"bytes"
	"image"
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
