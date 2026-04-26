package difficulty

import "math"

func CandidateCount(elo Elo) int {
	score := Score01(elo)
	if score <= 0 {
		return 1
	}
	return min(8, 1+int(math.Round(score*7)))
}

func BetterCandidate(candidate, best Report, target Elo, hasBest bool) bool {
	if !hasBest {
		return true
	}
	candidateDelta := absInt(int(candidate.ActualElo) - int(target))
	bestDelta := absInt(int(best.ActualElo) - int(target))
	if candidateDelta != bestDelta {
		return candidateDelta < bestDelta
	}
	return confidenceRank(candidate.Confidence) > confidenceRank(best.Confidence)
}

func absInt(v int) int {
	if v < 0 {
		return -v
	}
	return v
}

func confidenceRank(confidence Confidence) int {
	switch confidence {
	case ConfidenceHigh:
		return 3
	case ConfidenceMedium:
		return 2
	case ConfidenceLow:
		return 1
	default:
		return 0
	}
}
