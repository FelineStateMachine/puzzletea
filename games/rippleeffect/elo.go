package rippleeffect

import (
	"crypto/sha256"
	"encoding/binary"
	"math"
	"math/rand/v2"
	"strconv"

	"github.com/FelineStateMachine/puzzletea/difficulty"
	"github.com/FelineStateMachine/puzzletea/game"
)

func (m Mode) SpawnElo(seed string, elo difficulty.Elo) (game.Gamer, difficulty.Report, error) {
	if err := difficulty.ValidateElo(elo); err != nil {
		return nil, difficulty.Report{}, err
	}

	mode := rippleEffectModeForElo(m, elo)
	puzzle, err := mode.generatePuzzleSeeded(rippleEffectEloRNG(seed, elo))
	if err != nil {
		return nil, difficulty.Report{}, err
	}

	gamer, err := New(mode, puzzle)
	if err != nil {
		return nil, difficulty.Report{}, err
	}
	return gamer, rippleEffectDifficultyReport(elo, mode, puzzle), nil
}

func rippleEffectModeForElo(base Mode, elo difficulty.Elo) Mode {
	score := difficulty.Score01(elo)
	maxCage := min(base.Size, 3+int(math.Round(score*2)))
	givenRatio := 0.69 - score*0.27
	profile := rippleEffectProfileForElo(score, maxCage)

	mode := base
	mode.MaxCage = maxCage
	mode.GivenRatio = givenRatio
	mode.profile = profile
	return mode
}

func rippleEffectProfileForElo(score float64, maxCage int) generationProfile {
	switch {
	case score < 0.20:
		return generationProfile{
			cageWeights:       []int{0, 5, 6, 3},
			frontierSamples:   3,
			shapeBias:         shapeBiasCompact,
			minGivensByCage:   []int{0, 1, 1, 2},
			maxSingletonCages: -1,
		}
	case score < 0.42:
		return generationProfile{
			cageWeights:       []int{0, 3, 5, 4},
			frontierSamples:   3,
			shapeBias:         shapeBiasCompact,
			minGivensByCage:   []int{0, 1, 1, 2},
			maxSingletonCages: -1,
		}
	case score < 0.65:
		return generationProfile{
			cageWeights:       growIntSlice([]int{0, 0, 3, 5, 4}, maxCage+1),
			frontierSamples:   2,
			shapeBias:         shapeBiasNeutral,
			minGivensByCage:   make([]int, maxCage+1),
			maxSingletonCages: 1,
		}
	case score < 0.86:
		return generationProfile{
			cageWeights:       growIntSlice([]int{0, 0, 1, 4, 5}, maxCage+1),
			frontierSamples:   3,
			shapeBias:         shapeBiasWinding,
			minGivensByCage:   make([]int, maxCage+1),
			maxSingletonCages: 1,
		}
	default:
		return generationProfile{
			cageWeights:       growIntSlice([]int{0, 0, 1, 2, 4, 5}, maxCage+1),
			frontierSamples:   4,
			shapeBias:         shapeBiasWinding,
			minGivensByCage:   make([]int, maxCage+1),
			maxSingletonCages: 1,
		}
	}
}

func rippleEffectEloRNG(seed string, elo difficulty.Elo) *rand.Rand {
	sum := sha256.Sum256([]byte("rippleeffect\x00" + seed + "\x00" + strconv.Itoa(int(elo))))
	return rand.New(rand.NewPCG(
		binary.LittleEndian.Uint64(sum[0:8]),
		binary.LittleEndian.Uint64(sum[8:16]),
	))
}

func rippleEffectDifficultyReport(target difficulty.Elo, mode Mode, puzzle Puzzle) difficulty.Report {
	geo, err := buildGeometry(puzzle.Width, puzzle.Height, puzzle.Cages)
	if err != nil {
		return difficulty.Report{
			TargetElo:  target,
			ActualElo:  target,
			Confidence: difficulty.ConfidenceLow,
			Metrics: difficulty.Metrics{
				"width":  float64(puzzle.Width),
				"height": float64(puzzle.Height),
			},
		}
	}

	solutions := countSolutions(geo, puzzle.Givens, 2)
	metrics := rippleEffectDifficultyMetrics(mode, puzzle, geo, solutions)
	confidence := difficulty.ConfidenceHigh
	if solutions != 1 {
		confidence = difficulty.ConfidenceLow
	}

	return difficulty.Report{
		TargetElo:  target,
		ActualElo:  rippleEffectActualElo(metrics),
		Confidence: confidence,
		Metrics:    metrics,
	}
}

func rippleEffectDifficultyMetrics(mode Mode, puzzle Puzzle, geo *geometry, solutions int) difficulty.Metrics {
	cells := puzzle.Width * puzzle.Height
	givens := countFilledCells(puzzle.Givens)
	empty := cells - givens
	cageCount := len(puzzle.Cages)
	singletons := 0
	maxCageSize := 0
	totalCageSize := 0
	for _, cage := range puzzle.Cages {
		size := len(cage.Cells)
		if size == 1 {
			singletons++
		}
		if size > maxCageSize {
			maxCageSize = size
		}
		totalCageSize += size
	}

	return difficulty.Metrics{
		"width":           float64(puzzle.Width),
		"height":          float64(puzzle.Height),
		"cells":           float64(cells),
		"max_cage":        float64(mode.MaxCage),
		"actual_max_cage": float64(maxCageSize),
		"cage_count":      float64(cageCount),
		"avg_cage_size":   rippleEffectRatio(totalCageSize, cageCount),
		"singleton_cages": float64(singletons),
		"given_count":     float64(givens),
		"empty_count":     float64(empty),
		"given_density":   rippleEffectRatio(givens, cells),
		"target_givens":   mode.GivenRatio,
		"solution_count":  float64(solutions),
		"geometry_cages":  float64(len(geo.cages)),
	}
}

func countFilledCells(g grid) int {
	count := 0
	for y := range g {
		for x := range g[y] {
			if g[y][x] != 0 {
				count++
			}
		}
	}
	return count
}

func rippleEffectActualElo(metrics difficulty.Metrics) difficulty.Elo {
	score := 0.26*rippleEffectNormalize(metrics["cells"], 25, 81) +
		0.22*rippleEffectNormalize(1-metrics["given_density"], 0.20, 0.65) +
		0.18*rippleEffectNormalize(metrics["actual_max_cage"], 3, 5) +
		0.14*rippleEffectNormalize(metrics["avg_cage_size"], 1.5, 4.2) +
		0.12*rippleEffectNormalize(metrics["empty_count"], 8, 52) +
		0.08*(1-rippleEffectNormalize(metrics["singleton_cages"], 0, 6))

	return difficulty.ClampElo(difficulty.Elo(math.Round(score * float64(difficulty.SoftCapElo))))
}

func rippleEffectRatio(numerator, denominator int) float64 {
	if denominator == 0 {
		return 0
	}
	return float64(numerator) / float64(denominator)
}

func rippleEffectNormalize(value, low, high float64) float64 {
	if high <= low || value <= low {
		return 0
	}
	if value >= high {
		return 1
	}
	return (value - low) / (high - low)
}
