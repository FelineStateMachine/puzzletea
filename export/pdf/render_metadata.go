package pdfexport

func clampDifficultyScore(score float64) float64 {
	if score < 0 {
		return 0
	}
	if score > 1 {
		return 1
	}
	return score
}

func isMixedModes(mode string) bool {
	return normalizeToken(mode) == "mixed modes"
}
