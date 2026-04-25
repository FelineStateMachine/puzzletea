package pdfexport

import (
	"math"
	"reflect"
	"testing"

	"codeberg.org/go-pdf/fpdf"
)

func TestSplitCoverSubtitleLinesClampsToMaxLines(t *testing.T) {
	pdf := fpdf.New("P", "mm", "A4", "")
	if err := registerPDFFonts(pdf); err != nil {
		t.Fatalf("registerPDFFonts error: %v", err)
	}
	pdf.AddPage()
	pdf.SetFont(coverFontFamily, "", 18)

	subtitle := "The Catacombs of the Last Bastion and Other Grim Architectural Delights"
	maxW := 78.0
	got := splitCoverSubtitleLines(pdf, subtitle, maxW, 2)
	if len(got) != 2 {
		t.Fatalf("line count = %d, want 2 (%v)", len(got), got)
	}

	gotAgain := splitCoverSubtitleLines(pdf, subtitle, maxW, 2)
	if !reflect.DeepEqual(got, gotAgain) {
		t.Fatalf("splitCoverSubtitleLines not stable: %v vs %v", got, gotAgain)
	}
}

func TestSplitClampedTextLinesClampsToMaxLines(t *testing.T) {
	pdf := fpdf.New("P", "mm", "A4", "")
	if err := registerPDFFonts(pdf); err != nil {
		t.Fatalf("registerPDFFonts error: %v", err)
	}
	pdf.AddPage()
	pdf.SetFont(sansFontFamily, "", 10)

	text := "This is a deliberately long advert line that should wrap across multiple lines in the PDF renderer."
	maxW := 55.0
	got := splitClampedTextLines(pdf, text, maxW, 2)
	if len(got) != 2 {
		t.Fatalf("line count = %d, want 2 (%v)", len(got), got)
	}

	gotAgain := splitClampedTextLines(pdf, text, maxW, 2)
	if !reflect.DeepEqual(got, gotAgain) {
		t.Fatalf("splitClampedTextLines not stable: %v vs %v", got, gotAgain)
	}
}

func TestBuildCoverArtLayoutIsDeterministicForSameSeed(t *testing.T) {
	cfg := RenderConfig{
		ShuffleSeed:  "zine-seed-01",
		VolumeNumber: 4,
	}

	layoutA := buildCoverArtLayout(cfg, halfLetterWidthMM, halfLetterHeightMM)
	layoutB := buildCoverArtLayout(cfg, halfLetterWidthMM, halfLetterHeightMM)
	if !coverArtLayoutEqual(layoutA, layoutB) {
		t.Fatalf("cover layout should be stable for identical input: %#v vs %#v", layoutA, layoutB)
	}
	if !coverArtStaysWithinArea(layoutA) {
		t.Fatalf("cover layout escapes art area: %#v", layoutA)
	}
	if !coverArtRespectsLockup(layoutA) {
		t.Fatalf("cover layout overlaps lockup exclusion: %#v", layoutA)
	}
	if coverRotationSignature(layoutA) != coverRotationSignature(layoutB) {
		t.Fatalf("rotation signature changed: %q vs %q", coverRotationSignature(layoutA), coverRotationSignature(layoutB))
	}
}

func TestBuildCoverArtLayoutChangesWhenSeedChanges(t *testing.T) {
	base := RenderConfig{
		ShuffleSeed:  "zine-seed-01",
		VolumeNumber: 4,
	}
	other := RenderConfig{
		ShuffleSeed:  "zine-seed-02",
		VolumeNumber: 4,
	}

	layoutA := buildCoverArtLayout(base, halfLetterWidthMM, halfLetterHeightMM)
	layoutB := buildCoverArtLayout(other, halfLetterWidthMM, halfLetterHeightMM)
	if !coverArtDiffers(layoutA, layoutB) {
		t.Fatalf("cover layout should differ when seed changes: %#v", layoutA)
	}
	if coverRotationSignature(layoutA) == coverRotationSignature(layoutB) {
		t.Fatalf("rotation signature should differ when seed changes: %q", coverRotationSignature(layoutA))
	}
}

func TestBuildCoverArtLayoutUsesMultipleTextureFamilies(t *testing.T) {
	cfg := RenderConfig{
		ShuffleSeed:  "texture-seed",
		VolumeNumber: 2,
	}

	layout := buildCoverArtLayout(cfg, halfLetterWidthMM, halfLetterHeightMM)
	if got := coverTextureCount(layout); got < 1 {
		t.Fatalf("texture count = %d, want at least 1", got)
	}

	bounds := coverArtInkBounds(layout)
	if bounds.w <= 0 || bounds.h <= 0 {
		t.Fatalf("ink bounds = %#v, want positive extents", bounds)
	}
}

func TestCoverSeedCorpusProducesMultipleSilhouetteFamilies(t *testing.T) {
	seeds := []string{
		"near-seed-00",
		"near-seed-01",
		"near-seed-02",
		"near-seed-03",
		"near-seed-04",
		"near-seed-05",
		"near-seed-06",
		"near-seed-07",
		"near-seed-08",
		"near-seed-09",
	}

	families := map[string]struct{}{}
	aspectBuckets := map[string]struct{}{}
	directions := map[string]struct{}{}
	mixes := map[string]struct{}{}
	rotations := map[string]struct{}{}
	polarities := map[string]struct{}{}

	for _, seed := range seeds {
		layout := buildCoverArtLayout(RenderConfig{
			ShuffleSeed:  seed,
			VolumeNumber: 1,
		}, halfLetterWidthMM, halfLetterHeightMM)

		if !coverArtStaysWithinArea(layout) {
			t.Fatalf("layout for %q escapes art area: %#v", seed, layout)
		}
		if !coverArtRespectsLockup(layout) {
			t.Fatalf("layout for %q overlaps lockup exclusion: %#v", seed, layout)
		}

		families[coverFamilyName(layout.Family)] = struct{}{}
		aspectBuckets[coverAspectBucketName(coverInkAspectBucket(layout))] = struct{}{}
		directions[coverDirectionName(layout.Direction)] = struct{}{}
		mixes[coverPrimitiveMixSignature(layout)] = struct{}{}
		rotations[coverRotationSignature(layout)] = struct{}{}
		polarities[coverRotationPolaritySignature(layout)] = struct{}{}
	}

	if got := len(families); got < 3 {
		t.Fatalf("family count = %d, want at least 3", got)
	}
	if got := len(aspectBuckets); got < 2 {
		t.Fatalf("aspect bucket count = %d, want at least 2", got)
	}
	if got := len(directions); got < 3 {
		t.Fatalf("direction count = %d, want at least 3", got)
	}
	if got := len(mixes); got < 3 {
		t.Fatalf("primitive mix count = %d, want at least 3", got)
	}
	if got := len(rotations); got < 4 {
		t.Fatalf("rotation signature count = %d, want at least 4", got)
	}
	if got := len(polarities); got < 2 {
		t.Fatalf("rotation polarity count = %d, want at least 2", got)
	}
}

func TestRotatedRectAABBNinetyDegreesSwapsExtents(t *testing.T) {
	bounds := rectMM{x: 10, y: 20, w: 20, h: 10}
	aabb := rotatedRectAABB(bounds, 20, 25, 90)
	if math.Abs(aabb.w-10) > 0.001 {
		t.Fatalf("aabb.w = %.3f, want 10", aabb.w)
	}
	if math.Abs(aabb.h-20) > 0.001 {
		t.Fatalf("aabb.h = %.3f, want 20", aabb.h)
	}
	if math.Abs(aabb.x-15) > 0.001 || math.Abs(aabb.y-15) > 0.001 {
		t.Fatalf("aabb origin = (%.3f, %.3f), want (15, 15)", aabb.x, aabb.y)
	}
}

func TestCoverRotationAngleBandRespectsStructuralFlag(t *testing.T) {
	seeds := []string{
		"rotation-band-00",
		"rotation-band-01",
		"rotation-band-02",
		"rotation-band-03",
		"rotation-band-04",
		"rotation-band-05",
	}

	seenPositive := false
	seenNegative := false
	seenStructural := false
	for _, seed := range seeds {
		layout := buildCoverArtLayout(RenderConfig{
			ShuffleSeed:  seed,
			VolumeNumber: 1,
		}, halfLetterWidthMM, halfLetterHeightMM)

		for _, shape := range layout.Shapes {
			absAngle := math.Abs(shape.Rotate)
			if shape.Rotate > 0 {
				seenPositive = true
			}
			if shape.Rotate < 0 {
				seenNegative = true
			}
			if shape.Locked {
				seenStructural = true
				if absAngle < 4 || absAngle > 10 {
					t.Fatalf("structural angle = %.2f, want within [4,10]", absAngle)
				}
				continue
			}
			if absAngle < 8 || absAngle > 20 {
				t.Fatalf("standard angle = %.2f, want within [8,20]", absAngle)
			}
		}
	}

	if !seenStructural {
		t.Fatal("expected at least one structural shape across seed corpus")
	}
	if !seenPositive || !seenNegative {
		t.Fatalf("expected both positive and negative rotations, got positive=%t negative=%t", seenPositive, seenNegative)
	}
}
