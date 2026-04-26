package netwalk

import (
	"crypto/sha256"
	"encoding/binary"
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

	mode := netwalkModeForElo(elo)
	puzzle, err := GenerateSeededWithDensity(mode.Size, mode.FillRatio, mode.Profile, netwalkEloRNG(seed, elo))
	if err != nil {
		return nil, difficulty.Report{}, err
	}

	gamer, err := New(mode, puzzle)
	if err != nil {
		return nil, difficulty.Report{}, err
	}
	return gamer, netwalkDifficultyReport(elo, mode, puzzle), nil
}

func netwalkModeForElo(elo difficulty.Elo) NetwalkMode {
	score := difficulty.Score01(elo)
	size := 5 + int(math.Round(score*4))*2
	fillRatio := 0.50 + score*0.28
	profile := netwalkProfileForElo(score)

	return NewMode(
		"Elo "+strconv.Itoa(int(elo)),
		"Elo-targeted Netwalk puzzle.",
		size,
		fillRatio,
		profile,
	)
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
	score := 0.24*normalizeFloat(metrics["size"], 5, 13) +
		0.20*normalizeFloat(metrics["density"], 0.20, 0.80) +
		0.18*normalizeFloat(metrics["tee_tiles"]+metrics["cross_tiles"], 0, 18) +
		0.14*normalizeFloat(metrics["elbow_tiles"], 0, 32) +
		0.14*normalizeFloat(metrics["avg_rotation_options"], 1, 4) +
		0.10*normalizeFloat(metrics["initially_rotated_tiles"], 0, metrics["active_tiles"])

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
