package pdfexport

import (
	"math"

	"codeberg.org/go-pdf/fpdf"
)

type difficultyStarState uint8

const (
	difficultyStarEmpty difficultyStarState = iota
	difficultyStarHalf
	difficultyStarFull
)

type puzzleSubtitleLayout struct {
	labelText  string
	labelWidth float64
	starStates []difficultyStarState
	starsWidth float64
	totalWidth float64
}

func scoreToDifficultyStarStates(score float64) []difficultyStarState {
	score = clampDifficultyScore(score)

	units := int(math.Round(score * 10))
	states := make([]difficultyStarState, 5)
	for i := range states {
		switch {
		case units >= 2:
			states[i] = difficultyStarFull
			units -= 2
		case units == 1:
			states[i] = difficultyStarHalf
			units = 0
		default:
			states[i] = difficultyStarEmpty
		}
	}
	return states
}

func puzzleDifficultySubtitleLayout(pdf *fpdf.Fpdf, puzzle Puzzle) puzzleSubtitleLayout {
	label := "Difficulty:"
	if !isMixedModes(puzzle.ModeSelection) {
		label = "Mode: " + puzzle.ModeSelection + " | " + label
	}

	layout := puzzleSubtitleLayout{
		labelText:  label,
		starStates: scoreToDifficultyStarStates(puzzle.DifficultyScore),
	}
	if pdf == nil {
		return layout
	}

	setPuzzleSubtitleStyle(pdf)
	layout.labelWidth = pdf.GetStringWidth(label)
	layout.starsWidth = difficultyStarsWidth()
	layout.totalWidth = layout.labelWidth + difficultyTextToStarsGapMM + layout.starsWidth
	return layout
}

func renderPuzzleDifficultySubtitle(pdf *fpdf.Fpdf, pageW, y float64, puzzle Puzzle) {
	if pdf == nil {
		return
	}

	layout := puzzleDifficultySubtitleLayout(pdf, puzzle)
	rowHeight := 5.0
	startX := (pageW - layout.totalWidth) / 2
	if startX < 0 {
		startX = 0
	}

	setPuzzleSubtitleStyle(pdf)
	pdf.SetXY(startX, y)
	pdf.CellFormat(layout.labelWidth, rowHeight, layout.labelText, "", 0, "L", false, 0, "")

	starX := startX + layout.labelWidth + difficultyTextToStarsGapMM
	starCenterY := y + rowHeight/2
	renderDifficultyStars(pdf, starX, starCenterY, layout.starStates)
}

func renderDifficultyStars(pdf *fpdf.Fpdf, x, centerY float64, states []difficultyStarState) {
	if pdf == nil {
		return
	}

	textR, textG, textB := pdf.GetTextColor()
	emptyOutline := 132
	pdf.SetLineWidth(difficultyStarOutlineMM)

	for i, state := range states {
		starX := x + float64(i)*(difficultyStarSizeMM+difficultyStarGapMM)
		points := difficultyStarPoints(starX, centerY, difficultyStarSizeMM)
		top := centerY - difficultyStarSizeMM/2

		switch state {
		case difficultyStarFull:
			pdf.SetDrawColor(textR, textG, textB)
			fillDifficultyStar(pdf, points, starX, top, difficultyStarSizeMM, difficultyStarSizeMM, textR, textG, textB)
			pdf.Polygon(points, "D")
		case difficultyStarHalf:
			pdf.SetDrawColor(emptyOutline, emptyOutline, emptyOutline)
			pdf.Polygon(points, "D")
			fillDifficultyStar(pdf, points, starX, top, difficultyStarSizeMM/2, difficultyStarSizeMM, textR, textG, textB)
		default:
			pdf.SetDrawColor(emptyOutline, emptyOutline, emptyOutline)
			pdf.Polygon(points, "D")
		}
	}
}

func fillDifficultyStar(
	pdf *fpdf.Fpdf,
	points []fpdf.PointType,
	x,
	y,
	w,
	h float64,
	r,
	g,
	b int,
) {
	if pdf == nil || len(points) == 0 || w <= 0 || h <= 0 {
		return
	}

	pdf.ClipPolygon(points, false)
	pdf.ClipRect(x, y, w, h, false)
	pdf.SetFillColor(r, g, b)
	pdf.Rect(x, y, w, h, "F")
	pdf.ClipEnd()
	pdf.ClipEnd()
}

func difficultyStarsWidth() float64 {
	return float64(5)*difficultyStarSizeMM + float64(4)*difficultyStarGapMM
}

func difficultyStarPoints(x, centerY, size float64) []fpdf.PointType {
	cx := x + size/2
	cy := centerY
	outer := size / 2
	inner := outer * 0.48
	points := make([]fpdf.PointType, 0, 10)

	for i := range 10 {
		radius := outer
		if i%2 == 1 {
			radius = inner
		}

		angle := -math.Pi/2 + float64(i)*math.Pi/5
		points = append(points, fpdf.PointType{
			X: cx + math.Cos(angle)*radius,
			Y: cy + math.Sin(angle)*radius,
		})
	}

	return points
}
