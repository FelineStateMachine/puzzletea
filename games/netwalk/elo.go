package netwalk

import (
	"crypto/sha256"
	"encoding/binary"
	"errors"
	"math"
	"math/rand/v2"
	"strconv"

	"github.com/FelineStateMachine/puzzletea/difficulty"
	"github.com/FelineStateMachine/puzzletea/game"
)

func (m NetwalkMode) SpawnElo(seed string, elo difficulty.Elo) (game.Gamer, difficulty.Report, error) {
	if err := difficulty.ValidateElo(elo); err != nil {
		return nil, difficulty.Report{}, err
	}

	mode := netwalkModeForElo(m, elo)
	var bestPuzzle Puzzle
	var bestReport difficulty.Report
	haveBest := false
	var lastErr error
	for candidate := range difficulty.CandidateCount(elo) {
		puzzle, err := GenerateSeededWithDensity(mode.Size, mode.FillRatio, mode.Profile, netwalkEloRNG(netwalkCandidateSeed(seed, candidate), elo))
		if err != nil {
			lastErr = err
			continue
		}
		report := netwalkDifficultyReport(elo, mode, puzzle)
		if difficulty.BetterCandidate(report, bestReport, elo, haveBest) {
			bestPuzzle = puzzle
			bestReport = report
			haveBest = true
		}
	}
	if !haveBest {
		if lastErr == nil {
			lastErr = errors.New("unable to generate Elo netwalk")
		}
		return nil, difficulty.Report{}, lastErr
	}

	gamer, err := New(mode, bestPuzzle)
	if err != nil {
		return nil, difficulty.Report{}, err
	}
	return gamer, bestReport, nil
}

func netwalkModeForElo(base NetwalkMode, elo difficulty.Elo) NetwalkMode {
	score := difficulty.Score01(elo)
	fillRatio := 0.34 + score*0.44
	profile := netwalkProfileForElo(score)

	mode := base
	mode.FillRatio = fillRatio
	mode.Profile = profile
	return mode
}

func netwalkProfileForElo(score float64) generateProfile {
	switch {
	case score < 0.20:
		return miniProfile
	case score < 0.42:
		return easyProfile
	case score < 0.65:
		return mediumProfile
	case score < 0.86:
		return hardProfile
	default:
		return expertProfile
	}
}

func netwalkEloRNG(seed string, elo difficulty.Elo) *rand.Rand {
	sum := sha256.Sum256([]byte("netwalk\x00" + seed + "\x00" + strconv.Itoa(int(elo))))
	return rand.New(rand.NewPCG(
		binary.LittleEndian.Uint64(sum[0:8]),
		binary.LittleEndian.Uint64(sum[8:16]),
	))
}

func netwalkCandidateSeed(seed string, candidate int) string {
	if candidate == 0 {
		return seed
	}
	return seed + "\x00candidate:" + strconv.Itoa(candidate)
}

func netwalkDifficultyReport(target difficulty.Elo, mode NetwalkMode, puzzle Puzzle) difficulty.Report {
	metrics := netwalkDifficultyMetrics(mode, puzzle)
	return difficulty.Report{
		TargetElo:  target,
		ActualElo:  netwalkActualElo(metrics),
		Confidence: difficulty.ConfidenceMedium,
		Metrics:    metrics,
	}
}

func netwalkDifficultyMetrics(mode NetwalkMode, puzzle Puzzle) difficulty.Metrics {
	state := analyzePuzzle(puzzle)
	metrics := difficulty.Metrics{
		"size":            float64(puzzle.Size),
		"cells":           float64(puzzle.Size * puzzle.Size),
		"active_tiles":    float64(state.nonEmpty),
		"density":         safeFloatRatio(state.nonEmpty, puzzle.Size*puzzle.Size),
		"target_density":  mode.FillRatio,
		"connected_tiles": float64(state.connected),
		"dangling_edges":  float64(state.dangling),
		"locked_tiles":    float64(state.locked),
	}

	rotationOptions := 0
	for y := range puzzle.Size {
		for x := range puzzle.Size {
			t := puzzle.Tiles[y][x]
			if !isActive(t) {
				continue
			}
			rotationOptions += len(uniqueRotations(t.BaseMask))
			switch degree(t.BaseMask) {
			case 1:
				metrics["leaf_tiles"]++
			case 2:
				if t.BaseMask == north|south || t.BaseMask == east|west {
					metrics["straight_tiles"]++
				} else {
					metrics["elbow_tiles"]++
				}
			case 3:
				metrics["tee_tiles"]++
			case 4:
				metrics["cross_tiles"]++
			}
			if t.Rotation != 0 {
				metrics["initially_rotated_tiles"]++
			}
		}
	}
	metrics["rotation_options"] = float64(rotationOptions)
	metrics["avg_rotation_options"] = safeFloatRatio(rotationOptions, state.nonEmpty)
	return metrics
}

func netwalkActualElo(metrics difficulty.Metrics) difficulty.Elo {
	metricScore := 0.24*normalizeFloat(metrics["size"], 5, 13) +
		0.20*normalizeFloat(metrics["density"], 0.20, 0.80) +
		0.18*normalizeFloat(metrics["tee_tiles"]+metrics["cross_tiles"], 0, 18) +
		0.14*normalizeFloat(metrics["elbow_tiles"], 0, 32) +
		0.14*normalizeFloat(metrics["avg_rotation_options"], 1, 4) +
		0.10*normalizeFloat(metrics["initially_rotated_tiles"], 0, metrics["active_tiles"])
	targetScore := normalizeFloat(metrics["target_density"], 0.34, 0.78)
	score := 0.45*targetScore + 0.55*metricScore

	return difficulty.ClampElo(difficulty.Elo(math.Round(score * float64(difficulty.SoftCapElo))))
}

func normalizeFloat(value, low, high float64) float64 {
	if high <= low || value <= low {
		return 0
	}
	if value >= high {
		return 1
	}
	return (value - low) / (high - low)
}

func safeFloatRatio(numerator, denominator int) float64 {
	if denominator == 0 {
		return 0
	}
	return float64(numerator) / float64(denominator)
}
