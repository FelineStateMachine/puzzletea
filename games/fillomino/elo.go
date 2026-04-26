package fillomino

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

var _ game.EloSpawner = Mode{}

func (m Mode) SpawnElo(seed string, elo difficulty.Elo) (game.Gamer, difficulty.Report, error) {
	if err := difficulty.ValidateElo(elo); err != nil {
		return nil, difficulty.Report{}, err
	}

	mode := fillominoModeForElo(m, elo)
	var bestPuzzle Puzzle
	var bestReport difficulty.Report
	haveBest := false
	var lastErr error
	for candidate := range fillominoEloCandidateCount(elo) {
		puzzle, err := GeneratePuzzleSeeded(mode.Size, mode.Size, mode.MaxRegion, mode.GivenRatio, fillominoEloRNG(fillominoCandidateSeed(seed, candidate), elo))
		if err != nil {
			lastErr = err
			continue
		}
		report := fillominoDifficultyReport(elo, mode, puzzle)
		if difficulty.BetterCandidate(report, bestReport, elo, haveBest) {
			bestPuzzle = puzzle
			bestReport = report
			haveBest = true
		}
	}
	if !haveBest {
		if lastErr == nil {
			lastErr = errors.New("unable to generate Elo fillomino")
		}
		return nil, difficulty.Report{}, lastErr
	}
	g, err := New(mode, bestPuzzle)
	if err != nil {
		return nil, difficulty.Report{}, err
	}
	return g, bestReport, nil
}

func fillominoModeForElo(base Mode, elo difficulty.Elo) Mode {
	score := difficulty.Score01(elo)
	maxRegion := 5 + int(math.Round(score*3))
	givenRatio := 0.70 - score*0.14
	mode := base
	mode.MaxRegion = min(maxRegion, max(1, mode.Size*mode.Size))
	mode.GivenRatio = givenRatio
	return mode
}

func fillominoEloCandidateCount(elo difficulty.Elo) int {
	if elo >= 2400 {
		return 1
	}
	return difficulty.CandidateCount(elo)
}

func fillominoEloRNG(seed string, elo difficulty.Elo) *rand.Rand {
	sum := sha256.Sum256([]byte("fillomino\x00" + seed + "\x00" + strconv.Itoa(int(elo))))
	return rand.New(rand.NewPCG(
		binary.LittleEndian.Uint64(sum[0:8]),
		binary.LittleEndian.Uint64(sum[8:16]),
	))
}

func fillominoCandidateSeed(seed string, candidate int) string {
	if candidate == 0 {
		return seed
	}
	return seed + "\x00candidate:" + strconv.Itoa(candidate)
}

func fillominoDifficultyReport(target difficulty.Elo, mode Mode, puzzle Puzzle) difficulty.Report {
	solutions := countSolutions(puzzle.Givens, mode.MaxRegion, 2)
	metrics := fillominoDifficultyMetrics(mode, puzzle, solutions)
	confidence := difficulty.ConfidenceHigh
	if solutions != 1 {
		confidence = difficulty.ConfidenceLow
	}
	return difficulty.Report{
		TargetElo:  target,
		ActualElo:  fillominoActualElo(metrics),
		Confidence: confidence,
		Metrics:    metrics,
	}
}

func fillominoDifficultyMetrics(mode Mode, puzzle Puzzle, solutions int) difficulty.Metrics {
	cells := puzzle.Width * puzzle.Height
	givens := 0
	maxRegionValue := 0
	for _, row := range puzzle.Givens {
		for _, value := range row {
			if value == 0 {
				continue
			}
			givens++
			if value > maxRegionValue {
				maxRegionValue = value
			}
		}
	}
	return difficulty.Metrics{
		"width":            float64(puzzle.Width),
		"height":           float64(puzzle.Height),
		"cells":            float64(cells),
		"given_count":      float64(givens),
		"unknown_count":    float64(cells - givens),
		"given_ratio":      fillominoRatio(givens, cells),
		"target_ratio":     mode.GivenRatio,
		"max_region":       float64(mode.MaxRegion),
		"max_given_region": float64(maxRegionValue),
		"solution_count":   float64(solutions),
	}
}

func fillominoActualElo(metrics difficulty.Metrics) difficulty.Elo {
	metricScore := 0.40*fillominoNormalize(metrics["cells"], 25, 144) +
		0.25*(1-fillominoNormalize(metrics["given_ratio"], 0.52, 0.70)) +
		0.20*fillominoNormalize(metrics["max_region"], 5, 9) +
		0.15*fillominoNormalize(metrics["unknown_count"], 8, 70)
	targetScore := 1 - fillominoNormalize(metrics["target_ratio"], 0.56, 0.70)
	score := 0.75*targetScore + 0.25*metricScore
	return difficulty.ClampElo(difficulty.Elo(math.Round(score * float64(difficulty.SoftCapElo))))
}

func fillominoNormalize(value, low, high float64) float64 {
	if high <= low || value <= low {
		return 0
	}
	if value >= high {
		return 1
	}
	return (value - low) / (high - low)
}

func fillominoRatio(numerator, denominator int) float64 {
	if denominator == 0 {
		return 0
	}
	return float64(numerator) / float64(denominator)
}
