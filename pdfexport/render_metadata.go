package pdfexport

import "math"

func difficultyScoreOutOfTen(score float64) int {
	if score < 0 {
		score = 0
	}
	if score > 1 {
		score = 1
	}
	return int(math.Round(score * 10))
}

func isMixedModes(mode string) bool {
	return normalizeToken(mode) == "mixed modes"
}
