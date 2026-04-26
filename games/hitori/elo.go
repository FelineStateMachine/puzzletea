package hitori

import (
	"crypto/sha256"
	"encoding/binary"
	"math"
	"math/rand/v2"
	"strconv"

	"github.com/FelineStateMachine/puzzletea/difficulty"
	"github.com/FelineStateMachine/puzzletea/game"
)

func (h HitoriMode) SpawnElo(seed string, elo difficulty.Elo) (game.Gamer, difficulty.Report, error) {
	if err := difficulty.ValidateElo(elo); err != nil {
		return nil, difficulty.Report{}, err
	}

	mode := hitoriModeForElo(h, elo)
	puzzle, err := GenerateSeeded(mode.Size, mode.BlackRatio, hitoriEloRNG(seed, elo))
	if err != nil {
		return nil, difficulty.Report{}, err
	}

	gamer, err := New(mode, puzzle)
	if err != nil {
		return nil, difficulty.Report{}, err
	}
	return gamer, hitoriDifficultyReport(elo, mode, puzzle), nil
}

func hitoriModeForElo(base HitoriMode, elo difficulty.Elo) HitoriMode {
	score := difficulty.Score01(elo)
	size := 5 + int(math.Round(score*7))
	blackRatio := 0.32 - score*0.04

	mode := base
	mode.BaseMode = game.NewBaseMode(
		"Elo "+strconv.Itoa(int(elo)),
		"Elo-targeted Hitori puzzle.",
	)
	mode.Size = size
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

func hitoriDifficultyReport(target difficulty.Elo, mode HitoriMode, puzzle grid) difficulty.Report {
	metrics := hitoriDifficultyMetrics(mode, puzzle)
	return difficulty.Report{
		TargetElo:  target,
		ActualElo:  hitoriActualElo(metrics),
		Confidence: difficulty.ConfidenceMedium,
		Metrics:    metrics,
	}
}

func hitoriDifficultyMetrics(mode HitoriMode, puzzle grid) difficulty.Metrics {
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
		"solution_count":   float64(countPuzzleSolutions(puzzle, size, 2)),
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
	score := 0.50*hitoriNormalize(metrics["size"], 5, 12) +
		0.28*hitoriNormalize(metrics["duplicate_pct"], 0.20, 1.20) +
		0.17*hitoriNormalize(metrics["target_black_pct"], 0.28, 0.34) +
		0.05*hitoriNormalize(metrics["duplicate_groups"], 4, 36)

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
