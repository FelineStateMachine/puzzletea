package sudokurgb

import (
	"crypto/sha256"
	"encoding/binary"
	"math"
	"math/rand/v2"
	"strconv"

	"github.com/FelineStateMachine/puzzletea/difficulty"
	"github.com/FelineStateMachine/puzzletea/game"
)

var _ game.EloSpawner = SudokuRGBMode{}

type eloSolveStats struct {
	Nodes    int
	Branches int
	MaxDepth int
}

func (s SudokuRGBMode) SpawnElo(seed string, elo difficulty.Elo) (game.Gamer, difficulty.Report, error) {
	if err := difficulty.ValidateElo(elo); err != nil {
		return nil, difficulty.Report{}, err
	}

	mode := sudokuRGBModeForElo(s, elo)
	var bestProvided []cell
	var bestReport difficulty.Report
	haveBest := false
	for candidate := range difficulty.CandidateCount(elo) {
		provided := GenerateProvidedCellsSeeded(mode, sudokuRGBEloRNG(sudokuRGBCandidateSeed(seed, candidate), elo))
		report := scoreSudokuRGBElo(elo, provided)
		if difficulty.BetterCandidate(report, bestReport, elo, haveBest) {
			bestProvided = provided
			bestReport = report
			haveBest = true
		}
	}
	g, err := New(mode, bestProvided)
	if err != nil {
		return nil, difficulty.Report{}, err
	}
	return g, bestReport, nil
}

func sudokuRGBModeForElo(base SudokuRGBMode, elo difficulty.Elo) SudokuRGBMode {
	score := difficulty.Score01(elo)
	provided := 60 - int(math.Round(score*36))
	if provided < 24 {
		provided = 24
	}
	mode := base
	mode.ProvidedCount = provided
	return mode
}

func sudokuRGBEloRNG(seed string, elo difficulty.Elo) *rand.Rand {
	sum := sha256.Sum256([]byte("sudokurgb\x00" + seed + "\x00" + strconv.Itoa(int(elo))))
	return rand.New(rand.NewPCG(
		binary.LittleEndian.Uint64(sum[0:8]),
		binary.LittleEndian.Uint64(sum[8:16]),
	))
}

func sudokuRGBCandidateSeed(seed string, candidate int) string {
	if candidate == 0 {
		return seed
	}
	return seed + "\x00candidate:" + strconv.Itoa(candidate)
}

func scoreSudokuRGBElo(target difficulty.Elo, provided []cell) difficulty.Report {
	g := newGrid(provided)
	solutions, stats := countSolutionsWithEloStats(&g, 2)
	metrics := difficulty.Metrics{
		"clue_count":     float64(len(provided)),
		"unknown_count":  float64(gridSize*gridSize - len(provided)),
		"solution_count": float64(solutions),
		"solver_nodes":   float64(stats.Nodes),
		"branches":       float64(stats.Branches),
		"max_depth":      float64(stats.MaxDepth),
	}

	actual := sudokuRGBActualElo(metrics)
	confidence := difficulty.ConfidenceHigh
	if solutions != 1 {
		confidence = difficulty.ConfidenceLow
	}

	return difficulty.Report{
		TargetElo:  target,
		ActualElo:  actual,
		Confidence: confidence,
		Metrics:    metrics,
	}
}

func countSolutionsWithEloStats(g *grid, limit int) (int, eloSolveStats) {
	if limit <= 0 {
		return 0, eloSolveStats{}
	}

	state, ok := buildQuotaState(g)
	if !ok {
		return 0, eloSolveStats{}
	}

	var stats eloSolveStats
	count := countSolutionsRecWithEloStats(g, limit, &state, 0, &stats)
	return count, stats
}

func countSolutionsRecWithEloStats(g *grid, limit int, state *quotaState, depth int, stats *eloSolveStats) int {
	if limit <= 0 {
		return 0
	}

	stats.Nodes++
	if depth > stats.MaxDepth {
		stats.MaxDepth = depth
	}

	x, y, candidates, found := findMostConstrainedEmptyCell(g, state)
	if !found {
		return 1
	}
	if candidates == 0 {
		return 0
	}

	count := 0
	candidateCount := countCandidateBits(candidates)
	if candidateCount > 1 {
		stats.Branches++
	}
	for value := 1; value <= valueCount; value++ {
		if candidates&(1<<value) == 0 {
			continue
		}
		placeValue(g, state, value, x, y)
		count += countSolutionsRecWithEloStats(g, limit-count, state, depth+1, stats)
		clearValue(g, state, value, x, y)
		if count >= limit {
			return count
		}
	}
	return count
}

func sudokuRGBActualElo(metrics difficulty.Metrics) difficulty.Elo {
	score := 0.55*normalizeSudokuRGBMetric(metrics["unknown_count"], 21, 57) +
		0.25*normalizeSudokuRGBMetric(math.Log10(metrics["solver_nodes"]+1), 1.2, 4.0) +
		0.15*normalizeSudokuRGBMetric(metrics["branches"], 0, 120) +
		0.05*normalizeSudokuRGBMetric(metrics["max_depth"], 0, 51)
	return difficulty.ClampElo(difficulty.Elo(math.Round(score * float64(difficulty.SoftCapElo))))
}

func normalizeSudokuRGBMetric(value, low, high float64) float64 {
	if value <= low {
		return 0
	}
	if value >= high {
		return 1
	}
	return (value - low) / (high - low)
}
