package hitori

import (
	"context"
	"crypto/sha256"
	"encoding/binary"
	"math"
	"math/rand/v2"
	"strconv"

	"github.com/FelineStateMachine/puzzletea/difficulty"
	"github.com/FelineStateMachine/puzzletea/game"
)

var _ game.CancellableEloSpawner = HitoriMode{}

func (h HitoriMode) SpawnElo(seed string, elo difficulty.Elo) (game.Gamer, difficulty.Report, error) {
	return h.SpawnEloContext(context.Background(), seed, elo)
}

func (h HitoriMode) SpawnEloContext(ctx context.Context, seed string, elo difficulty.Elo) (game.Gamer, difficulty.Report, error) {
	if err := difficulty.ValidateElo(elo); err != nil {
		return nil, difficulty.Report{}, err
	}
	if err := ctx.Err(); err != nil {
		return nil, difficulty.Report{}, err
	}

	mode := hitoriModeForElo(h, elo)
	var puzzle grid
	var err error
	if elo >= 2400 {
		puzzle = generateFastHitoriEloPuzzle(mode, hitoriEloRNG(seed+"\x00fallback", elo))
	} else {
		puzzle, err = GenerateSeededContext(ctx, mode.Size, mode.BlackRatio, hitoriEloRNG(seed, elo))
		if err != nil {
			if ctx.Err() == nil {
				return nil, difficulty.Report{}, err
			}
			puzzle = generateFastHitoriEloPuzzle(mode, hitoriEloRNG(seed+"\x00fallback", elo))
		}
	}

	gamer, err := New(mode, puzzle)
	if err != nil {
		return nil, difficulty.Report{}, err
	}
	if elo >= 2400 {
		report := hitoriDifficultyReportWithoutCounting(elo, mode, puzzle)
		return gamer, report, nil
	}
	report := hitoriDifficultyReport(ctx, elo, mode, puzzle)
	return gamer, report, nil
}

func hitoriModeForElo(base HitoriMode, elo difficulty.Elo) HitoriMode {
	score := difficulty.Score01(elo)
	blackRatio := 0.32 - score*0.04

	mode := base
	mode.BlackRatio = blackRatio
	return mode
}

func hitoriEloRNG(seed string, elo difficulty.Elo) *rand.Rand {
	sum := sha256.Sum256([]byte("hitori\x00" + seed + "\x00" + strconv.Itoa(int(elo))))
	return rand.New(rand.NewPCG(
		binary.LittleEndian.Uint64(sum[0:8]),
		binary.LittleEndian.Uint64(sum[8:16]),
	))
}

func generateFastHitoriEloPuzzle(mode HitoriMode, rng *rand.Rand) grid {
	baseGrid := generateLatinSquareSeeded(mode.Size, rng)
	mask := generateValidMaskSeeded(mode.Size, mode.BlackRatio, rng)
	return constructPuzzleSeeded(baseGrid, mask, rng)
}

func hitoriDifficultyReport(ctx context.Context, target difficulty.Elo, mode HitoriMode, puzzle grid) difficulty.Report {
	metrics := hitoriDifficultyMetrics(ctx, mode, puzzle)
	confidence := difficulty.ConfidenceMedium
	if metrics["solution_count"] < 0 {
		confidence = difficulty.ConfidenceLow
	}
	return difficulty.Report{
		TargetElo:  target,
		ActualElo:  hitoriActualElo(metrics),
		Confidence: confidence,
		Metrics:    metrics,
	}
}

func hitoriDifficultyReportWithoutCounting(target difficulty.Elo, mode HitoriMode, puzzle grid) difficulty.Report {
	metrics := hitoriDifficultyMetricsWithoutCounting(mode, puzzle)
	metrics["solution_count"] = -1
	metrics["solver_limited"] = 1
	return difficulty.Report{
		TargetElo:  target,
		ActualElo:  hitoriActualElo(metrics),
		Confidence: difficulty.ConfidenceLow,
		Metrics:    metrics,
	}
}

func hitoriDifficultyMetrics(ctx context.Context, mode HitoriMode, puzzle grid) difficulty.Metrics {
	metrics := hitoriDifficultyMetricsWithoutCounting(mode, puzzle)
	metrics["solution_count"] = float64(countPuzzleSolutionsContext(ctx, puzzle, len(puzzle), 2))
	return metrics
}

func hitoriDifficultyMetricsWithoutCounting(mode HitoriMode, puzzle grid) difficulty.Metrics {
	size := len(puzzle)
	cells := size * size
	duplicateCells, duplicateGroups := hitoriDuplicatePressure(puzzle)

	return difficulty.Metrics{
		"size":             float64(size),
		"cells":            float64(cells),
		"target_black_pct": mode.BlackRatio,
		"duplicate_cells":  float64(duplicateCells),
		"duplicate_groups": float64(duplicateGroups),
		"duplicate_pct":    hitoriRatio(duplicateCells, cells),
	}
}

func hitoriDuplicatePressure(puzzle grid) (int, int) {
	duplicateCells := 0
	duplicateGroups := 0

	for _, row := range puzzle {
		cells, groups := hitoriDuplicateLinePressure(row)
		duplicateCells += cells
		duplicateGroups += groups
	}

	size := len(puzzle)
	for x := range size {
		col := make([]rune, size)
		for y := range size {
			col[y] = puzzle[y][x]
		}
		cells, groups := hitoriDuplicateLinePressure(col)
		duplicateCells += cells
		duplicateGroups += groups
	}

	return duplicateCells, duplicateGroups
}

func hitoriDuplicateLinePressure(line []rune) (int, int) {
	counts := make(map[rune]int, len(line))
	for _, value := range line {
		counts[value]++
	}

	cells := 0
	groups := 0
	for _, count := range counts {
		if count < 2 {
			continue
		}
		cells += count
		groups++
	}
	return cells, groups
}

func hitoriActualElo(metrics difficulty.Metrics) difficulty.Elo {
	metricScore := 0.50*hitoriNormalize(metrics["size"], 5, 12) +
		0.28*hitoriNormalize(metrics["duplicate_pct"], 0.20, 1.20) +
		0.17*(1-hitoriNormalize(metrics["target_black_pct"], 0.27, 0.32)) +
		0.05*hitoriNormalize(metrics["duplicate_groups"], 4, 36)
	targetScore := 1 - hitoriNormalize(metrics["target_black_pct"], 0.28, 0.32)
	score := 0.50*targetScore + 0.50*metricScore

	return difficulty.ClampElo(difficulty.Elo(math.Round(score * float64(difficulty.SoftCapElo))))
}

func hitoriNormalize(value, low, high float64) float64 {
	if high <= low || value <= low {
		return 0
	}
	if value >= high {
		return 1
	}
	return (value - low) / (high - low)
}

func hitoriRatio(numerator, denominator int) float64 {
	if denominator == 0 {
		return 0
	}
	return float64(numerator) / float64(denominator)
}
