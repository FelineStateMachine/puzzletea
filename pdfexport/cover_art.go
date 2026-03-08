package pdfexport

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"hash/fnv"
	"image"
	"image/color"
	"image/png"
	"math"
	"math/rand"
	"strings"
	"time"

	"codeberg.org/go-pdf/fpdf"
)

type coverVec2 struct {
	x float64
	y float64
}

type coverGlow struct {
	center   coverVec2
	spread   float64
	strength float64
	col      color.RGBA
}

type coverFieldProfile struct {
	freqX     float64
	freqY     float64
	freqMixX  float64
	freqMixY  float64
	freqCurlX float64
	freqCurlY float64
	ampX      float64
	ampY      float64
	ampMix    float64
	ampCurl   float64
	phaseX    float64
	phaseY    float64
	phaseMix  float64
	phaseCurl float64
	swirl     float64
	shear     float64
	pinch     float64
	pivot     coverVec2
}

type coverFlowLayer struct {
	count     int
	steps     int
	step      float64
	alpha     uint8
	radius    float64
	drift     float64
	phase     float64
	waveXFreq float64
	waveYFreq float64
	waveXAmp  float64
	waveYAmp  float64
	a         color.RGBA
	b         color.RGBA
}

type coverLatticeProfile struct {
	enabled    bool
	layout     int
	cols       int
	rows       int
	neighbors  int
	jitterX    float64
	jitterY    float64
	center     coverVec2
	radialWarp float64
	edge       color.RGBA
	nodeOuter  color.RGBA
	nodeInner  color.RGBA
	nodeSize   float64
	coreSize   float64
}

type coverOrbitProfile struct {
	enabled        bool
	center         coverVec2
	count          int
	radiusStart    float64
	radiusStep     float64
	arcCoverage    float64
	segmentsMin    int
	segmentsJitter int
	eccentricity   float64
	wobble         float64
	dotSize        float64
	colorA         color.RGBA
	colorB         color.RGBA
	alphaBase      uint8
	alphaStep      uint8
}

type coverMarkProfile struct {
	enabled    bool
	count      int
	gridW      int
	gridH      int
	jitter     float64
	sizeMin    float64
	sizeMax    float64
	alphaBase  uint8
	alphaRange uint8
	colorA     color.RGBA
	colorB     color.RGBA
	cutout     color.RGBA
}

type coverPalette struct {
	bgTop      color.RGBA
	bgMid      color.RGBA
	bgBottom   color.RGBA
	accentA    color.RGBA
	accentB    color.RGBA
	ink        color.RGBA
	nodeOuter  color.RGBA
	nodeInner  color.RGBA
	markCutout color.RGBA
	grainWarm  color.RGBA
	grainCool  color.RGBA
}

type coverPaletteFamily struct {
	name        string
	bgShift     [3]float64
	accentShift [2]float64
	satBias     float64
	valueBias   float64
}

type coverArchetype uint8

const (
	coverArchetypeConstellation coverArchetype = iota
	coverArchetypeVortex
	coverArchetypeBands
	coverArchetypeDriftField
	coverArchetypeRadialMesh
	coverArchetypeSparseGlyph
	coverArchetypeCount
)

type coverModifier uint8

const (
	coverModifierDenseFlow coverModifier = iota
	coverModifierHighOrbit
	coverModifierQuietLattice
	coverModifierMicroMarks
	coverModifierGrainHeavy
	coverModifierNegativeSpace
)

var coverPaletteFamilies = []coverPaletteFamily{
	{
		name:        "tropical",
		bgShift:     [3]float64{0.10, 0.02, -0.14},
		accentShift: [2]float64{0.33, 0.52},
		satBias:     0.08,
		valueBias:   0.03,
	},
	{
		name:        "sunset",
		bgShift:     [3]float64{0.05, -0.05, -0.17},
		accentShift: [2]float64{0.19, 0.38},
		satBias:     0.10,
		valueBias:   0.04,
	},
	{
		name:        "electric-mineral",
		bgShift:     [3]float64{-0.16, -0.08, 0.01},
		accentShift: [2]float64{0.36, -0.28},
		satBias:     0.13,
		valueBias:   -0.02,
	},
	{
		name:        "aurora",
		bgShift:     [3]float64{0.22, 0.10, -0.09},
		accentShift: [2]float64{0.45, 0.62},
		satBias:     0.11,
		valueBias:   0.02,
	},
	{
		name:        "ember",
		bgShift:     [3]float64{0.00, -0.11, -0.22},
		accentShift: [2]float64{-0.36, 0.11},
		satBias:     0.09,
		valueBias:   -0.03,
	},
	{
		name:        "citrus-marine",
		bgShift:     [3]float64{0.14, 0.03, -0.13},
		accentShift: [2]float64{0.30, -0.22},
		satBias:     0.12,
		valueBias:   0.01,
	},
}

type coverArtDirection struct {
	motif         int
	top           color.RGBA
	mid           color.RGBA
	bottom        color.RGBA
	verticalCurve float64
	glows         []coverGlow
	field         coverFieldProfile
	flowLayers    []coverFlowLayer
	lattice       coverLatticeProfile
	orbit         coverOrbitProfile
	marks         coverMarkProfile
	grainCount    int
	grainWarm     color.RGBA
	grainCool     color.RGBA
}

func drawCoverArtworkImage(
	pdf *fpdf.Fpdf,
	scene rectMM,
	seedText string,
	variant string,
	base RGB,
) {
	seedText = strings.TrimSpace(seedText)
	if seedText == "" {
		seedText = "puzzletea" + time.Now().String()
	}
	if strings.TrimSpace(variant) == "" {
		variant = "front"
	}
	imageName := fmt.Sprintf(
		"puzzletea-cover-artwork-%016x",
		coverSeedHash(seedText+"|"+variant),
	)
	artPNG := renderCoverArtworkPNG(seedText, variant, base)
	options := fpdf.ImageOptions{
		ImageType: "PNG",
		ReadDpi:   true,
	}
	pdf.RegisterImageOptionsReader(imageName, options, bytes.NewReader(artPNG))
	pdf.ImageOptions(imageName, scene.x, scene.y, scene.w, scene.h, false, options, 0, "")
}

func renderCoverArtworkPNG(seedText, variant string, base RGB) []byte {
	const (
		width  = 1200
		height = 1400
	)
	seed := coverSeedHash(seedText + "|" + variant)
	direction := newCoverArtDirection(seed, base)
	img := image.NewRGBA(image.Rect(0, 0, width, height))

	paintCoverBackground(img, width, height, direction)
	drawCoverFlowTrails(img, width, height, direction, coverStreamRNG(seed, "flow"))
	drawCoverPuzzleLattice(img, width, height, direction, coverStreamRNG(seed, "lattice"))
	drawCoverOrbitBands(img, width, height, direction, coverStreamRNG(seed, "orbit"))
	drawCoverSeedMarks(img, width, height, direction, coverStreamRNG(seed, "marks"))
	drawCoverFilmGrain(img, width, height, direction, coverStreamRNG(seed, "grain"))

	var out bytes.Buffer
	if err := png.Encode(&out, img); err != nil {
		return []byte{}
	}
	return out.Bytes()
}

func newCoverArtDirection(seed uint64, base RGB) coverArtDirection {
	paletteRNG := coverStreamRNG(seed, "palette")
	structureRNG := coverStreamRNG(seed, "structure")
	palette := newCoverPalette(base, paletteRNG)
	primary, modifiers := pickCoverComposition(structureRNG)

	glows := make([]coverGlow, 0, 5)
	anchorColor := palette.accentA
	if structureRNG.Intn(2) == 0 {
		anchorColor = palette.accentB
	}
	glows = append(glows, coverGlow{
		center: coverVec2{
			x: 0.12 + structureRNG.Float64()*0.76,
			y: 0.10 + structureRNG.Float64()*0.78,
		},
		spread:   0.24 + structureRNG.Float64()*0.30,
		strength: 0.46 + structureRNG.Float64()*0.28,
		col:      jitterRGBA(anchorColor, structureRNG, 26),
	})

	glowCount := 2 + structureRNG.Intn(3)
	for i := range glowCount {
		col := palette.accentA
		if i%2 == 1 {
			col = palette.accentB
		}
		col = jitterRGBA(col, structureRNG, 28)
		glows = append(glows, coverGlow{
			center: coverVec2{
				x: 0.06 + structureRNG.Float64()*0.88,
				y: 0.08 + structureRNG.Float64()*0.84,
			},
			spread:   0.42 + structureRNG.Float64()*0.70,
			strength: 0.12 + structureRNG.Float64()*0.28,
			col:      col,
		})
	}

	field := coverFieldProfile{
		freqX:     3.8 + structureRNG.Float64()*6.8,
		freqY:     3.6 + structureRNG.Float64()*6.5,
		freqMixX:  0.5 + structureRNG.Float64()*1.8,
		freqMixY:  0.5 + structureRNG.Float64()*1.8,
		freqCurlX: 0.4 + structureRNG.Float64()*2.0,
		freqCurlY: 0.4 + structureRNG.Float64()*2.0,
		ampX:      0.60 + structureRNG.Float64()*1.00,
		ampY:      0.60 + structureRNG.Float64()*1.00,
		ampMix:    0.26 + structureRNG.Float64()*0.92,
		ampCurl:   0.24 + structureRNG.Float64()*0.88,
		phaseX:    0.22 + structureRNG.Float64()*1.45,
		phaseY:    0.22 + structureRNG.Float64()*1.45,
		phaseMix:  0.20 + structureRNG.Float64()*1.20,
		phaseCurl: 0.20 + structureRNG.Float64()*1.20,
		swirl:     (structureRNG.Float64()*2 - 1) * 0.40,
		shear:     (structureRNG.Float64()*2 - 1) * 0.36,
		pinch:     (structureRNG.Float64()*2 - 1) * 0.56,
		pivot: coverVec2{
			x: 0.20 + structureRNG.Float64()*0.60,
			y: 0.18 + structureRNG.Float64()*0.64,
		},
	}

	flowLayers := make([]coverFlowLayer, 0, 4)
	layerCount := 2 + structureRNG.Intn(3)
	for i := range layerCount {
		t := float64(i) / float64(maxInt(1, layerCount-1))
		a := jitterRGBA(lerpRGB(palette.accentA, palette.bgMid, 0.28+t*0.42), structureRNG, 24)
		b := jitterRGBA(lerpRGB(palette.accentB, palette.bgTop, 0.16+t*0.38), structureRNG, 24)
		flowLayers = append(flowLayers, coverFlowLayer{
			count:     220 + structureRNG.Intn(360),
			steps:     96 + structureRNG.Intn(132),
			step:      1.10 + structureRNG.Float64()*1.80,
			alpha:     uint8(22 + structureRNG.Intn(42)),
			radius:    0.78 + structureRNG.Float64()*0.98,
			drift:     0.40 + structureRNG.Float64()*1.72,
			phase:     structureRNG.Float64() * math.Pi * 2,
			waveXFreq: 4.6 + structureRNG.Float64()*7.8,
			waveYFreq: 4.6 + structureRNG.Float64()*7.8,
			waveXAmp:  0.06 + structureRNG.Float64()*0.34,
			waveYAmp:  0.06 + structureRNG.Float64()*0.34,
			a:         a,
			b:         b,
		})
	}

	lattice := coverLatticeProfile{
		enabled:    true,
		layout:     structureRNG.Intn(3),
		cols:       7 + structureRNG.Intn(6),
		rows:       6 + structureRNG.Intn(7),
		neighbors:  2 + structureRNG.Intn(3),
		jitterX:    24 + structureRNG.Float64()*48,
		jitterY:    28 + structureRNG.Float64()*56,
		center:     coverVec2{x: 0.28 + structureRNG.Float64()*0.44, y: 0.24 + structureRNG.Float64()*0.50},
		radialWarp: (structureRNG.Float64()*2 - 1) * 0.40,
		edge:       color.RGBA{R: palette.ink.R, G: palette.ink.G, B: palette.ink.B, A: uint8(52 + structureRNG.Intn(76))},
		nodeOuter:  palette.nodeOuter,
		nodeInner:  palette.nodeInner,
		nodeSize:   3.3 + structureRNG.Float64()*3.8,
		coreSize:   1.4 + structureRNG.Float64()*1.7,
	}

	orbit := coverOrbitProfile{
		enabled:        true,
		center:         coverVec2{x: 0.30 + structureRNG.Float64()*0.38, y: 0.26 + structureRNG.Float64()*0.38},
		count:          10 + structureRNG.Intn(18),
		radiusStart:    36 + structureRNG.Float64()*52,
		radiusStep:     13 + structureRNG.Float64()*20,
		arcCoverage:    0.48 + structureRNG.Float64()*0.48,
		segmentsMin:    34 + structureRNG.Intn(40),
		segmentsJitter: 42 + structureRNG.Intn(56),
		eccentricity:   0.66 + structureRNG.Float64()*0.56,
		wobble:         2.6 + structureRNG.Float64()*8.8,
		dotSize:        0.84 + structureRNG.Float64()*0.86,
		colorA:         jitterRGBA(lerpRGB(palette.accentA, palette.bgTop, 0.22), structureRNG, 18),
		colorB:         jitterRGBA(lerpRGB(palette.accentB, palette.bgMid, 0.34), structureRNG, 18),
		alphaBase:      uint8(18 + structureRNG.Intn(28)),
		alphaStep:      uint8(1 + structureRNG.Intn(3)),
	}

	marks := coverMarkProfile{
		enabled:    true,
		count:      28 + structureRNG.Intn(58),
		gridW:      8 + structureRNG.Intn(9),
		gridH:      10 + structureRNG.Intn(10),
		jitter:     8 + structureRNG.Float64()*30,
		sizeMin:    1.7 + structureRNG.Float64()*1.5,
		sizeMax:    3.4 + structureRNG.Float64()*3.0,
		alphaBase:  uint8(40 + structureRNG.Intn(40)),
		alphaRange: uint8(20 + structureRNG.Intn(30)),
		colorA:     jitterRGBA(lerpRGB(palette.accentA, palette.bgTop, 0.18), structureRNG, 22),
		colorB:     jitterRGBA(lerpRGB(palette.accentB, palette.bgMid, 0.30), structureRNG, 22),
		cutout:     palette.markCutout,
	}

	direction := coverArtDirection{
		motif:         int(primary),
		top:           palette.bgTop,
		mid:           palette.bgMid,
		bottom:        palette.bgBottom,
		verticalCurve: 1.18 + structureRNG.Float64()*1.14,
		glows:         glows,
		field:         field,
		flowLayers:    flowLayers,
		lattice:       lattice,
		orbit:         orbit,
		marks:         marks,
		grainCount:    34000 + structureRNG.Intn(32000),
		grainWarm:     palette.grainWarm,
		grainCool:     palette.grainCool,
	}

	applyCoverPrimaryArchetype(&direction, primary, structureRNG)
	for _, modifier := range modifiers {
		applyCoverModifier(&direction, modifier, structureRNG)
	}

	return direction
}

func coverCompositionForSeed(seed uint64) (coverArchetype, []coverModifier) {
	structureRNG := coverStreamRNG(seed, "structure")
	return pickCoverComposition(structureRNG)
}

func pickCoverComposition(rng *rand.Rand) (coverArchetype, []coverModifier) {
	primary := pickWeightedCoverArchetype(rng)
	modifierCount := 0
	roll := rng.Float64()
	switch {
	case roll < 0.58:
		modifierCount = 1
	default:
		modifierCount = 2
	}

	modifierPool := []coverModifier{
		coverModifierDenseFlow,
		coverModifierHighOrbit,
		coverModifierQuietLattice,
		coverModifierMicroMarks,
		coverModifierGrainHeavy,
		coverModifierNegativeSpace,
	}
	rng.Shuffle(len(modifierPool), func(i, j int) {
		modifierPool[i], modifierPool[j] = modifierPool[j], modifierPool[i]
	})

	modifiers := make([]coverModifier, 0, modifierCount)
	for _, modifier := range modifierPool {
		if len(modifiers) >= modifierCount {
			break
		}
		if primary == coverArchetypeSparseGlyph && modifier == coverModifierNegativeSpace {
			continue
		}
		if primary == coverArchetypeBands && modifier == coverModifierHighOrbit {
			continue
		}
		modifiers = append(modifiers, modifier)
	}
	return primary, modifiers
}

func pickWeightedCoverArchetype(rng *rand.Rand) coverArchetype {
	weights := [...]int{18, 16, 14, 18, 17, 17}
	total := 0
	for _, weight := range weights {
		total += weight
	}
	pick := rng.Intn(total)
	for idx, weight := range weights {
		if pick < weight {
			return coverArchetype(idx)
		}
		pick -= weight
	}
	return coverArchetypeSparseGlyph
}

func newCoverPalette(base RGB, rng *rand.Rand) coverPalette {
	baseColor := rgbToColor(base)
	baseHue, baseSat, baseVal := rgbToHSV(baseColor)
	family := coverPaletteFamilies[rng.Intn(len(coverPaletteFamilies))]
	baseHue = wrapHue(baseHue + randCentered(rng, 0.32))

	baseSat = clamp(0.14, 0.50, baseSat*0.62+0.11+family.satBias*0.18+rng.Float64()*0.07)
	baseVal = clamp(0.34, 0.86, baseVal*0.84+0.12+family.valueBias*0.15+randCentered(rng, 0.05))

	top := hsvToRGB(
		wrapHue(baseHue+family.bgShift[0]+randCentered(rng, 0.10)),
		clamp(0.14, 0.34, baseSat*0.50+0.04+family.satBias*0.28+rng.Float64()*0.08),
		clamp(0.74, 0.98, baseVal+0.16+family.valueBias*0.24+rng.Float64()*0.14),
	)
	mid := hsvToRGB(
		wrapHue(baseHue+family.bgShift[1]+randCentered(rng, 0.14)),
		clamp(0.20, 0.50, baseSat*0.74+0.02+family.satBias*0.40+rng.Float64()*0.10),
		clamp(0.46, 0.88, baseVal-0.02+family.valueBias*0.12+randCentered(rng, 0.12)),
	)
	bottom := hsvToRGB(
		wrapHue(baseHue+family.bgShift[2]+randCentered(rng, 0.16)),
		clamp(0.18, 0.54, baseSat*0.82+0.01+family.satBias*0.35+rng.Float64()*0.10),
		clamp(0.08, 0.36, baseVal*0.32+0.02+family.valueBias*0.10+randCentered(rng, 0.10)),
	)
	accentA := hsvToRGB(
		wrapHue(baseHue+family.accentShift[0]+randCentered(rng, 0.18)),
		clamp(0.30, 0.62, baseSat+0.12+family.satBias*0.40+rng.Float64()*0.15),
		clamp(0.62, 0.96, baseVal+0.14+rng.Float64()*0.12),
	)
	accentB := hsvToRGB(
		wrapHue(baseHue+family.accentShift[1]+randCentered(rng, 0.20)),
		clamp(0.28, 0.60, baseSat+0.10+family.satBias*0.38+rng.Float64()*0.16),
		clamp(0.56, 0.92, baseVal+0.08+rng.Float64()*0.12),
	)
	ink := hsvToRGB(
		wrapHue(baseHue+0.50+family.bgShift[2]*0.35),
		clamp(0.22, 0.44, 0.24+baseSat*0.22),
		clamp(0.06, 0.22, 0.12+(1-baseVal)*0.10),
	)

	return coverPalette{
		bgTop:      top,
		bgMid:      mid,
		bgBottom:   bottom,
		accentA:    accentA,
		accentB:    accentB,
		ink:        ink,
		nodeOuter:  blendRGB(top, color.RGBA{R: 255, G: 244, B: 220, A: 255}, 0.26),
		nodeInner:  blendRGB(ink, bottom, 0.34),
		markCutout: blendRGB(bottom, color.RGBA{R: 6, G: 11, B: 20, A: 255}, 0.30),
		grainWarm:  blendRGB(top, color.RGBA{R: 255, G: 248, B: 232, A: 255}, 0.36),
		grainCool:  blendRGB(bottom, ink, 0.44),
	}
}

func applyCoverPrimaryArchetype(
	direction *coverArtDirection,
	primary coverArchetype,
	rng *rand.Rand,
) {
	switch primary {
	case coverArchetypeConstellation:
		direction.lattice.layout = 0
		direction.lattice.neighbors = minInt(5, direction.lattice.neighbors+1)
		direction.lattice.edge.A = minUint8(160, direction.lattice.edge.A+22)
		for i := range direction.flowLayers {
			direction.flowLayers[i].count = maxInt(160, int(float64(direction.flowLayers[i].count)*0.76))
			direction.flowLayers[i].alpha = maxUint8(16, direction.flowLayers[i].alpha-6)
		}
		direction.orbit.enabled = false
		direction.orbit.count = maxInt(4, direction.orbit.count/2)
		direction.orbit.arcCoverage = clamp01(direction.orbit.arcCoverage*0.72 + 0.12)
		direction.marks.count += 12
		direction.field.swirl *= 0.88
	case coverArchetypeVortex:
		direction.lattice.enabled = false
		direction.orbit.count += 14
		direction.orbit.arcCoverage = clamp01(direction.orbit.arcCoverage*1.20 + 0.08)
		direction.orbit.radiusStep *= 0.84
		direction.field.swirl *= 1.42
		for i := range direction.flowLayers {
			direction.flowLayers[i].alpha = minUint8(86, direction.flowLayers[i].alpha+10)
			direction.flowLayers[i].drift *= 1.18
		}
		direction.marks.count = maxInt(8, direction.marks.count-16)
	case coverArchetypeBands:
		direction.lattice.layout = 2
		direction.lattice.enabled = false
		direction.orbit.enabled = false
		direction.orbit.count = maxInt(5, direction.orbit.count/2)
		direction.field.shear *= 0.44
		direction.field.pinch *= 0.74
		for i := range direction.flowLayers {
			direction.flowLayers[i].waveYFreq = 1.8 + rng.Float64()*2.3
			direction.flowLayers[i].waveYAmp += 0.14 + rng.Float64()*0.16
			direction.flowLayers[i].waveXAmp *= 0.66
			direction.flowLayers[i].step *= 0.94
			direction.flowLayers[i].count = maxInt(180, int(float64(direction.flowLayers[i].count)*0.84))
		}
		direction.marks.count = maxInt(8, direction.marks.count-14)
	case coverArchetypeDriftField:
		direction.lattice.enabled = false
		for i := range direction.flowLayers {
			direction.flowLayers[i].count += 130
			direction.flowLayers[i].steps += 18
			direction.flowLayers[i].alpha = minUint8(90, direction.flowLayers[i].alpha+12)
			direction.flowLayers[i].drift *= 1.24
		}
		direction.orbit.count = maxInt(8, direction.orbit.count-4)
		direction.lattice.neighbors = maxInt(2, direction.lattice.neighbors-1)
		direction.marks.count = maxInt(10, direction.marks.count-12)
		direction.grainCount += 5000
	case coverArchetypeRadialMesh:
		direction.lattice.layout = 1
		direction.lattice.rows += 2
		direction.lattice.cols += 1
		direction.lattice.neighbors = minInt(5, direction.lattice.neighbors+1)
		direction.lattice.center = coverVec2{
			x: 0.44 + rng.Float64()*0.12,
			y: 0.38 + rng.Float64()*0.14,
		}
		direction.orbit.enabled = true
		direction.orbit.count += 8
		direction.orbit.eccentricity = clamp(0.56, 1.46, direction.orbit.eccentricity*1.16)
		direction.field.pinch *= 1.24
		for i := range direction.flowLayers {
			direction.flowLayers[i].count = maxInt(110, int(float64(direction.flowLayers[i].count)*0.45))
			direction.flowLayers[i].alpha = maxUint8(14, direction.flowLayers[i].alpha-6)
		}
		direction.marks.count = maxInt(10, direction.marks.count-14)
	case coverArchetypeSparseGlyph:
		for i := range direction.flowLayers {
			direction.flowLayers[i].count = maxInt(120, int(float64(direction.flowLayers[i].count)*0.55))
			direction.flowLayers[i].steps = maxInt(72, int(float64(direction.flowLayers[i].steps)*0.70))
			direction.flowLayers[i].alpha = maxUint8(12, direction.flowLayers[i].alpha-8)
		}
		direction.lattice.enabled = rng.Float64() > 0.68
		if direction.lattice.enabled {
			direction.lattice.neighbors = 2
			direction.lattice.edge.A = maxUint8(14, direction.lattice.edge.A-26)
		}
		direction.orbit.enabled = rng.Float64() > 0.80
		direction.orbit.count = maxInt(5, direction.orbit.count/2)
		direction.orbit.arcCoverage *= 0.66
		direction.marks.count += 40
		direction.marks.sizeMax += 1.5
		direction.marks.jitter *= 0.76
		direction.field.swirl *= 0.72
		direction.grainCount += 9000
	}
}

func applyCoverModifier(
	direction *coverArtDirection,
	modifier coverModifier,
	rng *rand.Rand,
) {
	switch modifier {
	case coverModifierDenseFlow:
		for i := range direction.flowLayers {
			direction.flowLayers[i].count += 96
			direction.flowLayers[i].steps += 16
			direction.flowLayers[i].alpha = minUint8(94, direction.flowLayers[i].alpha+10)
		}
		direction.grainCount += 3000
	case coverModifierHighOrbit:
		direction.orbit.enabled = true
		direction.orbit.count += 10
		direction.orbit.arcCoverage = clamp01(direction.orbit.arcCoverage*1.15 + 0.05)
		direction.orbit.dotSize += 0.12
		direction.orbit.alphaBase = minUint8(120, direction.orbit.alphaBase+6)
	case coverModifierQuietLattice:
		if direction.lattice.enabled {
			direction.lattice.edge.A = maxUint8(12, direction.lattice.edge.A-30)
			direction.lattice.neighbors = maxInt(2, direction.lattice.neighbors-1)
		}
		if rng.Float64() < 0.35 {
			direction.lattice.enabled = false
		}
	case coverModifierMicroMarks:
		direction.marks.count += 26
		direction.marks.sizeMin = clamp(0.8, 4.0, direction.marks.sizeMin*0.78)
		direction.marks.sizeMax = clamp(1.2, 5.2, direction.marks.sizeMax*0.74)
		direction.marks.jitter *= 0.86
		direction.marks.alphaBase = minUint8(108, direction.marks.alphaBase+8)
	case coverModifierGrainHeavy:
		direction.grainCount += 12000
		direction.grainWarm = blendRGB(direction.grainWarm, color.RGBA{R: 255, G: 250, B: 240, A: 255}, 0.12)
	case coverModifierNegativeSpace:
		for i := range direction.flowLayers {
			direction.flowLayers[i].count = maxInt(140, int(float64(direction.flowLayers[i].count)*0.62))
			direction.flowLayers[i].alpha = maxUint8(12, direction.flowLayers[i].alpha-8)
		}
		direction.orbit.count = maxInt(4, direction.orbit.count-6)
		direction.orbit.alphaBase = maxUint8(8, direction.orbit.alphaBase-4)
		direction.marks.count = maxInt(10, direction.marks.count-12)
		direction.verticalCurve = clamp(1.0, 2.8, direction.verticalCurve+0.26)
		for i := range direction.glows {
			direction.glows[i].strength *= 0.72
		}
	}
}

func randCentered(rng *rand.Rand, spread float64) float64 {
	return (rng.Float64()*2 - 1) * spread
}

func paintCoverBackground(img *image.RGBA, w, h int, direction coverArtDirection) {
	for y := range h {
		ty := float64(y) / float64(h-1)
		vertical := blendRGB(
			lerpRGB(direction.top, direction.mid, ty*1.12),
			direction.bottom,
			powClamp(ty, direction.verticalCurve),
		)
		for x := range w {
			tx := float64(x) / float64(w-1)
			col := vertical
			for _, glow := range direction.glows {
				amount := radialFalloff(tx, ty, glow.center.x, glow.center.y, glow.spread)
				col = blendRGB(col, glow.col, amount*glow.strength)
			}
			img.SetRGBA(x, y, col)
		}
	}
}

func drawCoverFlowTrails(
	img *image.RGBA,
	w, h int,
	direction coverArtDirection,
	rng *rand.Rand,
) {
	for _, layer := range direction.flowLayers {
		for i := 0; i < layer.count; i++ {
			x := rng.Float64() * float64(w)
			y := rng.Float64() * float64(h)
			for step := 0; step < layer.steps; step++ {
				nx := x / float64(w)
				ny := y / float64(h)
				angle := coverFieldAngle(nx, ny, direction.field, layer.phase)
				angle += math.Sin((ny+layer.phase)*layer.waveYFreq) * layer.waveYAmp
				angle += math.Cos((nx-layer.phase)*layer.waveXFreq) * layer.waveXAmp

				x += math.Cos(angle) * layer.step
				y += math.Sin(angle) * layer.step
				x += math.Cos((ny-layer.phase)*math.Pi*2) * layer.drift * 0.04
				y += math.Sin((nx+layer.phase)*math.Pi*2) * layer.drift * 0.08
				if x < 1 || x >= float64(w-1) || y < 1 || y >= float64(h-1) {
					break
				}
				t := float64(step) / float64(layer.steps)
				c := lerpRGB(layer.a, layer.b, t)
				c.A = layer.alpha
				drawDisc(img, x, y, layer.radius, c)
			}
		}
	}
}

func drawCoverPuzzleLattice(
	img *image.RGBA,
	w, h int,
	direction coverArtDirection,
	rng *rand.Rand,
) {
	if !direction.lattice.enabled {
		return
	}
	nodes := buildCoverLatticeNodes(w, h, direction.lattice, rng)
	if len(nodes) == 0 {
		return
	}

	for i := range nodes {
		for _, j := range nearestN(nodes, i, direction.lattice.neighbors) {
			if j > i {
				drawLine(img, nodes[i], nodes[j], direction.lattice.edge)
			}
		}
	}
	for _, n := range nodes {
		drawDisc(img, n.x, n.y, direction.lattice.nodeSize, direction.lattice.nodeOuter)
		drawDisc(img, n.x, n.y, direction.lattice.coreSize, direction.lattice.nodeInner)
	}
}

func buildCoverLatticeNodes(
	w, h int,
	profile coverLatticeProfile,
	rng *rand.Rand,
) []coverVec2 {
	nodes := make([]coverVec2, 0, profile.cols*profile.rows)
	switch profile.layout {
	case 1:
		minDim := math.Min(float64(w), float64(h))
		rings := maxInt(4, profile.rows+1)
		spokes := maxInt(8, profile.cols+4)
		center := coverVec2{
			x: profile.center.x * float64(w),
			y: profile.center.y * float64(h),
		}
		for ring := 1; ring <= rings; ring++ {
			ringT := float64(ring) / float64(rings+1)
			radius := ringT * minDim * (0.16 + 0.62*ringT)
			offset := rng.Float64() * math.Pi * 2
			for spoke := range spokes {
				ang := offset + float64(spoke)/float64(spokes)*math.Pi*2
				ang += math.Sin(float64(ring)*0.9+float64(spoke)*0.45) * profile.radialWarp
				jitter := (rng.Float64() - 0.5) * profile.jitterX
				x := center.x + math.Cos(ang)*(radius+jitter)
				y := center.y + math.Sin(ang)*(radius+jitter*0.45)
				nodes = append(nodes, coverVec2{x: x, y: y})
			}
		}
	case 2:
		total := maxInt(38, profile.cols*profile.rows)
		for i := range total {
			tx := (rng.Float64()*0.84 + 0.08)
			ty := (rng.Float64()*0.84 + 0.08)
			warpX := math.Sin(ty*math.Pi*4+float64(i)*0.22) * profile.radialWarp * 0.12
			warpY := math.Cos(tx*math.Pi*3+float64(i)*0.18) * profile.radialWarp * 0.10
			x := (tx+warpX)*float64(w) + (rng.Float64()-0.5)*profile.jitterX
			y := (ty+warpY)*float64(h) + (rng.Float64()-0.5)*profile.jitterY
			nodes = append(nodes, coverVec2{x: x, y: y})
		}
	default:
		for row := 0; row < profile.rows; row++ {
			for col := 0; col < profile.cols; col++ {
				x := (float64(col) + 1) / (float64(profile.cols) + 1) * float64(w)
				y := (float64(row) + 1) / (float64(profile.rows) + 1) * float64(h)
				x += (rng.Float64() - 0.5) * profile.jitterX
				y += (rng.Float64() - 0.5) * profile.jitterY
				nodes = append(nodes, coverVec2{x: x, y: y})
			}
		}
	}
	return nodes
}

func drawCoverOrbitBands(
	img *image.RGBA,
	w, h int,
	direction coverArtDirection,
	rng *rand.Rand,
) {
	if !direction.orbit.enabled {
		return
	}
	center := coverVec2{
		x: float64(w) * direction.orbit.center.x,
		y: float64(h) * direction.orbit.center.y,
	}
	for i := 0; i < direction.orbit.count; i++ {
		radius := direction.orbit.radiusStart + float64(i)*direction.orbit.radiusStep
		start := rng.Float64() * math.Pi * 2
		sweep := direction.orbit.arcCoverage * (0.74 + rng.Float64()*0.52) * math.Pi * 2
		sweep = math.Min(sweep, math.Pi*2)
		segments := direction.orbit.segmentsMin + rng.Intn(direction.orbit.segmentsJitter)
		t := float64(i) / float64(maxInt(1, direction.orbit.count-1))
		col := lerpRGB(direction.orbit.colorA, direction.orbit.colorB, t)
		col.A = minUint8(220, direction.orbit.alphaBase+uint8(i)*direction.orbit.alphaStep)
		for s := range segments {
			a := start + float64(s)/float64(segments)*sweep
			jx := math.Sin(float64(i)*0.43+a*2.9) * direction.orbit.wobble
			jy := math.Cos(float64(i)*0.37+a*2.3) * direction.orbit.wobble * 0.82
			x := center.x + math.Cos(a)*radius*direction.orbit.eccentricity + jx
			y := center.y + math.Sin(a)*radius + jy
			drawDisc(img, x, y, direction.orbit.dotSize, col)
		}
	}
}

func drawCoverSeedMarks(
	img *image.RGBA,
	w, h int,
	direction coverArtDirection,
	rng *rand.Rand,
) {
	if !direction.marks.enabled {
		return
	}
	for i := 0; i < direction.marks.count; i++ {
		gx := 1 + rng.Intn(direction.marks.gridW)
		gy := 1 + rng.Intn(direction.marks.gridH)
		x := float64(gx) / float64(direction.marks.gridW+1) * float64(w)
		y := float64(gy) / float64(direction.marks.gridH+1) * float64(h)
		x += (rng.Float64() - 0.5) * direction.marks.jitter
		y += (rng.Float64() - 0.5) * direction.marks.jitter

		size := direction.marks.sizeMin
		if direction.marks.sizeMax > direction.marks.sizeMin {
			size += rng.Float64() * (direction.marks.sizeMax - direction.marks.sizeMin)
		}
		col := lerpRGB(direction.marks.colorA, direction.marks.colorB, rng.Float64())
		col.A = minUint8(220, direction.marks.alphaBase+uint8(rng.Intn(int(direction.marks.alphaRange)+1)))
		style := (i + rng.Intn(4) + direction.motif) % 4
		switch style {
		case 0:
			drawDisc(img, x, y, size*0.48, col)
		case 1:
			drawDisc(img, x, y, size, col)
			drawDisc(img, x, y, size*0.52, direction.marks.cutout)
		case 2:
			drawLine(img, coverVec2{x: x - size, y: y}, coverVec2{x: x + size, y: y}, col)
			drawLine(img, coverVec2{x: x, y: y - size}, coverVec2{x: x, y: y + size}, col)
		default:
			drawLine(img, coverVec2{x: x - size, y: y - size}, coverVec2{x: x + size, y: y + size}, col)
			drawLine(img, coverVec2{x: x - size, y: y + size}, coverVec2{x: x + size, y: y - size}, col)
			drawDisc(img, x, y, size*0.32, direction.marks.cutout)
		}
	}
}

func drawCoverFilmGrain(
	img *image.RGBA,
	w, h int,
	direction coverArtDirection,
	rng *rand.Rand,
) {
	for i := 0; i < direction.grainCount; i++ {
		x := rng.Intn(w)
		y := rng.Intn(h)
		alpha := uint8(8 + rng.Intn(18))
		if rng.Intn(2) == 0 {
			blendPixel(img, x, y, color.RGBA{
				R: direction.grainWarm.R,
				G: direction.grainWarm.G,
				B: direction.grainWarm.B,
				A: alpha,
			})
			continue
		}
		blendPixel(img, x, y, color.RGBA{
			R: direction.grainCool.R,
			G: direction.grainCool.G,
			B: direction.grainCool.B,
			A: alpha,
		})
	}
}

func coverFieldAngle(x, y float64, field coverFieldProfile, phase float64) float64 {
	ax := math.Sin((x*field.freqX+phase*field.phaseX)*math.Pi*2) * field.ampX
	ay := math.Cos((y*field.freqY-phase*field.phaseY)*math.Pi*2) * field.ampY
	mix := math.Sin((x*field.freqMixX+y*field.freqMixY+phase*field.phaseMix)*math.Pi*2) * field.ampMix
	curl := math.Cos((x*field.freqCurlX-y*field.freqCurlY+phase*field.phaseCurl)*math.Pi*2) * field.ampCurl
	radial := math.Atan2(y-field.pivot.y, x-field.pivot.x)
	distance := math.Hypot(x-field.pivot.x, y-field.pivot.y)
	swirl := radial * field.swirl
	pinch := (0.5 - distance) * field.pinch
	shear := (x - y) * field.shear
	return ax + ay + mix + curl + swirl + pinch + shear
}

func nearestN(nodes []coverVec2, idx, n int) []int {
	n = maxInt(1, n)
	bestDistance := make([]float64, n)
	bestIndex := make([]int, n)
	for i := range bestDistance {
		bestDistance[i] = math.MaxFloat64
		bestIndex[i] = -1
	}
	for j := range nodes {
		if j == idx {
			continue
		}
		dx := nodes[idx].x - nodes[j].x
		dy := nodes[idx].y - nodes[j].y
		distance := dx*dx + dy*dy
		for k := range bestDistance {
			if distance >= bestDistance[k] {
				continue
			}
			for shift := len(bestDistance) - 1; shift > k; shift-- {
				bestDistance[shift] = bestDistance[shift-1]
				bestIndex[shift] = bestIndex[shift-1]
			}
			bestDistance[k] = distance
			bestIndex[k] = j
			break
		}
	}
	out := make([]int, 0, len(bestIndex))
	for _, best := range bestIndex {
		if best >= 0 {
			out = append(out, best)
		}
	}
	return out
}

func drawDisc(img *image.RGBA, cx, cy, radius float64, c color.RGBA) {
	minX := int(math.Floor(cx - radius))
	maxX := int(math.Ceil(cx + radius))
	minY := int(math.Floor(cy - radius))
	maxY := int(math.Ceil(cy + radius))
	r2 := radius * radius
	for y := minY; y <= maxY; y++ {
		for x := minX; x <= maxX; x++ {
			dx := float64(x) + 0.5 - cx
			dy := float64(y) + 0.5 - cy
			if dx*dx+dy*dy <= r2 {
				blendPixel(img, x, y, c)
			}
		}
	}
}

func drawLine(img *image.RGBA, a, b coverVec2, c color.RGBA) {
	dx := b.x - a.x
	dy := b.y - a.y
	steps := int(math.Max(math.Abs(dx), math.Abs(dy)))
	if steps <= 0 {
		blendPixel(img, int(a.x), int(a.y), c)
		return
	}
	for i := 0; i <= steps; i++ {
		t := float64(i) / float64(steps)
		drawDisc(img, a.x+dx*t, a.y+dy*t, 0.82, c)
	}
}

func blendPixel(img *image.RGBA, x, y int, src color.RGBA) {
	if !image.Pt(x, y).In(img.Rect) {
		return
	}
	dst := img.RGBAAt(x, y)
	alpha := float64(src.A) / 255
	inv := 1 - alpha
	img.SetRGBA(x, y, color.RGBA{
		R: uint8(float64(src.R)*alpha + float64(dst.R)*inv),
		G: uint8(float64(src.G)*alpha + float64(dst.G)*inv),
		B: uint8(float64(src.B)*alpha + float64(dst.B)*inv),
		A: 255,
	})
}

func rgbToColor(c RGB) color.RGBA {
	return color.RGBA{R: c.R, G: c.G, B: c.B, A: 255}
}

func rgbToHSV(c color.RGBA) (h, s, v float64) {
	r := float64(c.R) / 255
	g := float64(c.G) / 255
	b := float64(c.B) / 255
	maxC := math.Max(r, math.Max(g, b))
	minC := math.Min(r, math.Min(g, b))
	chroma := maxC - minC

	v = maxC
	if maxC == 0 {
		s = 0
	} else {
		s = chroma / maxC
	}
	if chroma == 0 {
		return 0, s, v
	}

	switch maxC {
	case r:
		h = math.Mod((g-b)/chroma, 6)
	case g:
		h = (b-r)/chroma + 2
	default:
		h = (r-g)/chroma + 4
	}
	h /= 6
	if h < 0 {
		h += 1
	}
	return h, s, v
}

func hsvToRGB(h, s, v float64) color.RGBA {
	h = wrapHue(h)
	s = clamp01(s)
	v = clamp01(v)
	if s == 0 {
		ch := uint8(math.Round(v * 255))
		return color.RGBA{R: ch, G: ch, B: ch, A: 255}
	}

	h6 := h * 6
	segment := math.Floor(h6)
	f := h6 - segment
	p := v * (1 - s)
	q := v * (1 - s*f)
	t := v * (1 - s*(1-f))

	var r, g, b float64
	switch int(segment) % 6 {
	case 0:
		r, g, b = v, t, p
	case 1:
		r, g, b = q, v, p
	case 2:
		r, g, b = p, v, t
	case 3:
		r, g, b = p, q, v
	case 4:
		r, g, b = t, p, v
	default:
		r, g, b = v, p, q
	}

	return color.RGBA{
		R: uint8(math.Round(clamp01(r) * 255)),
		G: uint8(math.Round(clamp01(g) * 255)),
		B: uint8(math.Round(clamp01(b) * 255)),
		A: 255,
	}
}

func wrapHue(h float64) float64 {
	if h >= 0 && h < 1 {
		return h
	}
	h = math.Mod(h, 1)
	if h < 0 {
		h += 1
	}
	return h
}

func lerpRGB(a, b color.RGBA, t float64) color.RGBA {
	t = clamp01(t)
	return color.RGBA{
		R: uint8(float64(a.R) + (float64(b.R)-float64(a.R))*t),
		G: uint8(float64(a.G) + (float64(b.G)-float64(a.G))*t),
		B: uint8(float64(a.B) + (float64(b.B)-float64(a.B))*t),
		A: 255,
	}
}

func blendRGB(base, top color.RGBA, amount float64) color.RGBA {
	return lerpRGB(base, top, clamp01(amount))
}

func radialFalloff(x, y, cx, cy, spread float64) float64 {
	dx := x - cx
	dy := y - cy
	distance := math.Sqrt(dx*dx + dy*dy)
	if distance >= spread {
		return 0
	}
	q := 1 - distance/spread
	return q * q
}

func powClamp(v, p float64) float64 {
	return math.Pow(clamp01(v), p)
}

func clamp01(v float64) float64 {
	if v < 0 {
		return 0
	}
	if v > 1 {
		return 1
	}
	return v
}

func clamp(min, max, v float64) float64 {
	if v < min {
		return min
	}
	if v > max {
		return max
	}
	return v
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func maxUint8(a, b uint8) uint8 {
	if a > b {
		return a
	}
	return b
}

func minUint8(a, b uint8) uint8 {
	if a < b {
		return a
	}
	return b
}

func jitterRGBA(c color.RGBA, rng *rand.Rand, spread int) color.RGBA {
	if spread <= 0 {
		return c
	}
	half := spread / 2
	return color.RGBA{
		R: shiftChannel(c.R, rng.Intn(spread)-half),
		G: shiftChannel(c.G, rng.Intn(spread)-half),
		B: shiftChannel(c.B, rng.Intn(spread)-half),
		A: c.A,
	}
}

func shiftChannel(ch uint8, delta int) uint8 {
	v := int(ch) + delta
	if v < 0 {
		return 0
	}
	if v > 255 {
		return 255
	}
	return uint8(v)
}

func coverStreamRNG(seed uint64, stream string) *rand.Rand {
	h := fnv.New64a()
	var raw [8]byte
	binary.LittleEndian.PutUint64(raw[:], seed)
	_, _ = h.Write(raw[:])
	_, _ = h.Write([]byte(stream))
	return rand.New(rand.NewSource(int64(h.Sum64())))
}

func coverSeedHash(text string) uint64 {
	h := fnv.New64a()
	h.Write([]byte(text))
	return h.Sum64()
}
