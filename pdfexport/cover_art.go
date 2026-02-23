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
	rng := coverStreamRNG(seed, "direction")
	motif := int(seed % 5)

	warm := blendRGB(
		brightenRGB(base, 0.30+rng.Float64()*0.28),
		color.RGBA{R: 255, G: 230, B: 180, A: 255},
		0.34+rng.Float64()*0.38,
	)
	cool := blendRGB(
		brightenRGB(base, 0.18+rng.Float64()*0.24),
		color.RGBA{R: 88, G: 212, B: 230, A: 255},
		0.30+rng.Float64()*0.42,
	)
	dusk := blendRGB(
		darkenRGB(base, 0.50+rng.Float64()*0.22),
		color.RGBA{R: 9, G: 18, B: 32, A: 255},
		0.44+rng.Float64()*0.30,
	)
	ink := blendRGB(dusk, color.RGBA{R: 22, G: 36, B: 52, A: 255}, 0.40+rng.Float64()*0.25)

	top := blendRGB(warm, color.RGBA{R: 255, G: 255, B: 251, A: 255}, 0.15+rng.Float64()*0.18)
	mid := blendRGB(cool, warm, 0.16+rng.Float64()*0.26)
	bottom := blendRGB(dusk, color.RGBA{R: 4, G: 9, B: 16, A: 255}, 0.30+rng.Float64()*0.22)

	glows := make([]coverGlow, 0, 4)
	glowCount := 3 + rng.Intn(2)
	for i := 0; i < glowCount; i++ {
		col := warm
		if i%2 == 1 {
			col = cool
		}
		col = jitterRGBA(col, rng, 18)
		glows = append(glows, coverGlow{
			center: coverVec2{
				x: 0.08 + rng.Float64()*0.84,
				y: 0.10 + rng.Float64()*0.82,
			},
			spread:   0.42 + rng.Float64()*0.70,
			strength: 0.16 + rng.Float64()*0.50,
			col:      col,
		})
	}

	field := coverFieldProfile{
		freqX:     4.0 + rng.Float64()*5.8,
		freqY:     3.8 + rng.Float64()*5.6,
		freqMixX:  0.5 + rng.Float64()*1.7,
		freqMixY:  0.5 + rng.Float64()*1.7,
		freqCurlX: 0.4 + rng.Float64()*1.9,
		freqCurlY: 0.4 + rng.Float64()*1.9,
		ampX:      0.65 + rng.Float64()*0.95,
		ampY:      0.65 + rng.Float64()*0.95,
		ampMix:    0.28 + rng.Float64()*0.86,
		ampCurl:   0.20 + rng.Float64()*0.86,
		phaseX:    0.25 + rng.Float64()*1.40,
		phaseY:    0.25 + rng.Float64()*1.40,
		phaseMix:  0.20 + rng.Float64()*1.10,
		phaseCurl: 0.20 + rng.Float64()*1.10,
		swirl:     (rng.Float64()*2 - 1) * 0.34,
		shear:     (rng.Float64()*2 - 1) * 0.34,
		pinch:     (rng.Float64()*2 - 1) * 0.52,
		pivot: coverVec2{
			x: 0.22 + rng.Float64()*0.56,
			y: 0.20 + rng.Float64()*0.62,
		},
	}

	flowLayers := make([]coverFlowLayer, 0, 4)
	layerCount := 2 + rng.Intn(3)
	for i := 0; i < layerCount; i++ {
		t := float64(i) / float64(maxInt(1, layerCount-1))
		a := jitterRGBA(lerpRGB(warm, cool, 0.22+t*0.54), rng, 20)
		b := jitterRGBA(lerpRGB(cool, warm, 0.18+t*0.66), rng, 20)
		flowLayers = append(flowLayers, coverFlowLayer{
			count:     250 + rng.Intn(320),
			steps:     90 + rng.Intn(120),
			step:      1.18 + rng.Float64()*1.72,
			alpha:     uint8(14 + rng.Intn(38)),
			radius:    0.76 + rng.Float64()*0.86,
			drift:     0.44 + rng.Float64()*1.68,
			phase:     rng.Float64() * math.Pi * 2,
			waveXFreq: 5.0 + rng.Float64()*7.2,
			waveYFreq: 5.0 + rng.Float64()*7.2,
			waveXAmp:  0.05 + rng.Float64()*0.33,
			waveYAmp:  0.05 + rng.Float64()*0.33,
			a:         a,
			b:         b,
		})
	}

	lattice := coverLatticeProfile{
		enabled:    true,
		layout:     rng.Intn(3),
		cols:       7 + rng.Intn(6),
		rows:       6 + rng.Intn(6),
		neighbors:  2 + rng.Intn(3),
		jitterX:    24 + rng.Float64()*46,
		jitterY:    28 + rng.Float64()*52,
		center:     coverVec2{x: 0.30 + rng.Float64()*0.40, y: 0.26 + rng.Float64()*0.44},
		radialWarp: (rng.Float64()*2 - 1) * 0.34,
		edge:       color.RGBA{R: ink.R, G: ink.G, B: ink.B, A: uint8(40 + rng.Intn(60))},
		nodeOuter:  blendRGB(warm, color.RGBA{R: 255, G: 245, B: 218, A: 255}, 0.35),
		nodeInner:  blendRGB(dusk, color.RGBA{R: 16, G: 30, B: 48, A: 255}, 0.25),
		nodeSize:   3.6 + rng.Float64()*3.6,
		coreSize:   1.5 + rng.Float64()*1.6,
	}

	orbit := coverOrbitProfile{
		enabled:        true,
		center:         coverVec2{x: 0.34 + rng.Float64()*0.32, y: 0.30 + rng.Float64()*0.32},
		count:          8 + rng.Intn(18),
		radiusStart:    40 + rng.Float64()*50,
		radiusStep:     15 + rng.Float64()*18,
		arcCoverage:    0.42 + rng.Float64()*0.58,
		segmentsMin:    38 + rng.Intn(34),
		segmentsJitter: 46 + rng.Intn(52),
		eccentricity:   0.70 + rng.Float64()*0.48,
		wobble:         2.8 + rng.Float64()*8.4,
		dotSize:        0.86 + rng.Float64()*0.76,
		colorA:         jitterRGBA(lerpRGB(warm, cool, 0.32), rng, 16),
		colorB:         jitterRGBA(lerpRGB(cool, warm, 0.70), rng, 16),
		alphaBase:      uint8(12 + rng.Intn(20)),
		alphaStep:      uint8(1 + rng.Intn(2)),
	}

	marks := coverMarkProfile{
		enabled:    true,
		count:      24 + rng.Intn(54),
		gridW:      8 + rng.Intn(8),
		gridH:      10 + rng.Intn(9),
		jitter:     8 + rng.Float64()*28,
		sizeMin:    1.8 + rng.Float64()*1.3,
		sizeMax:    3.8 + rng.Float64()*2.8,
		alphaBase:  uint8(32 + rng.Intn(36)),
		alphaRange: uint8(12 + rng.Intn(32)),
		colorA:     jitterRGBA(lerpRGB(warm, cool, 0.20), rng, 20),
		colorB:     jitterRGBA(lerpRGB(cool, warm, 0.72), rng, 20),
		cutout:     blendRGB(dusk, color.RGBA{R: 5, G: 10, B: 18, A: 255}, 0.35),
	}

	direction := coverArtDirection{
		motif:         motif,
		top:           top,
		mid:           mid,
		bottom:        bottom,
		verticalCurve: 1.24 + rng.Float64()*1.08,
		glows:         glows,
		field:         field,
		flowLayers:    flowLayers,
		lattice:       lattice,
		orbit:         orbit,
		marks:         marks,
		grainCount:    32000 + rng.Intn(34000),
		grainWarm:     blendRGB(warm, color.RGBA{R: 255, G: 246, B: 226, A: 255}, 0.42),
		grainCool:     blendRGB(dusk, color.RGBA{R: 5, G: 12, B: 20, A: 255}, 0.52),
	}

	switch direction.motif {
	case 0:
		direction.lattice.layout = 0
		direction.lattice.neighbors = minInt(5, direction.lattice.neighbors+1)
		direction.orbit.arcCoverage = clamp01(direction.orbit.arcCoverage*0.72 + 0.10)
		direction.marks.count += 10
	case 1:
		for i := range direction.flowLayers {
			direction.flowLayers[i].count += 120
			direction.flowLayers[i].steps += 18
			direction.flowLayers[i].alpha = minUint8(70, direction.flowLayers[i].alpha+12)
			direction.flowLayers[i].drift *= 1.24
		}
		direction.lattice.edge.A = maxUint8(18, direction.lattice.edge.A-22)
		direction.orbit.count = maxInt(6, direction.orbit.count/2)
		direction.grainCount += 6000
	case 2:
		direction.orbit.count += 12
		direction.orbit.arcCoverage = clamp01(direction.orbit.arcCoverage*1.18 + 0.08)
		direction.orbit.radiusStep *= 0.86
		direction.lattice.enabled = rng.Float64() > 0.26
		for i := range direction.flowLayers {
			direction.flowLayers[i].count = maxInt(180, direction.flowLayers[i].count-60)
		}
	case 3:
		direction.lattice.layout = 2
		direction.marks.count += 30
		direction.marks.sizeMax += 1.4
		direction.orbit.arcCoverage *= 0.74
		direction.field.swirl *= 1.25
	case 4:
		direction.lattice.layout = 1
		direction.lattice.rows += 2
		direction.orbit.eccentricity = clamp(0.55, 1.42, direction.orbit.eccentricity*1.16)
		direction.field.pinch *= 1.34
		direction.grainCount += 9000
	}

	return direction
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
			for spoke := 0; spoke < spokes; spoke++ {
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
		for i := 0; i < total; i++ {
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
		for s := 0; s < segments; s++ {
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

func brightenRGB(base RGB, amount float64) color.RGBA {
	return blendRGB(rgbToColor(base), color.RGBA{R: 255, G: 255, B: 255, A: 255}, amount)
}

func darkenRGB(base RGB, amount float64) color.RGBA {
	return blendRGB(rgbToColor(base), color.RGBA{R: 0, G: 0, B: 0, A: 255}, amount)
}

func rgbToColor(c RGB) color.RGBA {
	return color.RGBA{R: c.R, G: c.G, B: c.B, A: 255}
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
