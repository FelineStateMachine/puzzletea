package nonogram

import (
	"context"
	"crypto/sha256"
	"encoding/binary"
	"errors"
	"math"
	"math/rand/v2"
	"strconv"
	"time"

	"github.com/FelineStateMachine/puzzletea/difficulty"
	"github.com/FelineStateMachine/puzzletea/game"
)

var _ game.EloSpawner = NonogramMode{}

func (n NonogramMode) SpawnElo(seed string, elo difficulty.Elo) (game.Gamer, difficulty.Report, error) {
	if err := difficulty.ValidateElo(elo); err != nil {
		return nil, difficulty.Report{}, err
	}

	mode := nonogramModeForElo(elo)
	hints := GenerateRandomTomographySeeded(mode, nonogramEloRNG(seed, elo))
	if len(hints.rows) == 0 || len(hints.cols) == 0 {
		return nil, difficulty.Report{}, errors.New("unable to generate Elo nonogram")
	}

	gamer, err := New(mode, hints)
	if err != nil {
		return nil, difficulty.Report{}, err
	}
	return gamer, nonogramDifficultyReport(elo, mode, hints), nil
}

func nonogramModeForElo(elo difficulty.Elo) NonogramMode {
	score := difficulty.Score01(elo)

	size := 5
	switch {
	case score >= 0.75:
		size = 15
	case score >= 0.35:
		size = 10
	}
	density := 0.66 - score*0.26

	return NewMode(
		"Elo "+strconv.Itoa(int(elo)),
		"Elo-targeted Nonogram puzzle.",
		size,
		size,
		density,
	)
}

func nonogramEloRNG(seed string, elo difficulty.Elo) *rand.Rand {
	sum := sha256.Sum256([]byte("nonogram\x00" + seed + "\x00" + strconv.Itoa(int(elo))))
	return rand.New(rand.NewPCG(
		binary.LittleEndian.Uint64(sum[0:8]),
		binary.LittleEndian.Uint64(sum[8:16]),
	))
}

func nonogramDifficultyReport(target difficulty.Elo, mode NonogramMode, hints Hints) difficulty.Report {
	metrics := nonogramDifficultyMetrics(mode, hints)
	return difficulty.Report{
		TargetElo:  target,
		ActualElo:  nonogramActualElo(metrics),
		Confidence: difficulty.ConfidenceMedium,
		Metrics:    metrics,
	}
}

func nonogramDifficultyMetrics(mode NonogramMode, hints Hints) difficulty.Metrics {
	cells := mode.Width * mode.Height
	filled := 0
	rowRuns := 0
	colRuns := 0
	rowPossibilities := 0
	colPossibilities := 0
	maxLinePossibilities := 0

	for _, row := range hints.rows {
		filled += hintSum(row)
		rowRuns += nonzeroHintCount(row)
		possibilities := len(generateLinePossibilities(mode.Width, row))
		rowPossibilities += possibilities
		maxLinePossibilities = max(maxLinePossibilities, possibilities)
	}
	for _, col := range hints.cols {
		colRuns += nonzeroHintCount(col)
		possibilities := len(generateLinePossibilities(mode.Height, col))
		colPossibilities += possibilities
		maxLinePossibilities = max(maxLinePossibilities, possibilities)
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(max(mode.Width, mode.Height))*time.Second)
	solutions := countSolutions(hints, mode.Width, mode.Height, 2, ctx)
	cancel()

	totalLines := len(hints.rows) + len(hints.cols)
	totalPossibilities := rowPossibilities + colPossibilities
	return difficulty.Metrics{
		"width":                  float64(mode.Width),
		"height":                 float64(mode.Height),
		"cells":                  float64(cells),
		"density":                nonogramRatio(filled, cells),
		"target_density":         mode.Density,
		"row_runs":               float64(rowRuns),
		"col_runs":               float64(colRuns),
		"total_runs":             float64(rowRuns + colRuns),
		"avg_runs_per_line":      nonogramRatio(rowRuns+colRuns, totalLines),
		"line_possibilities":     float64(totalPossibilities),
		"avg_line_possibilities": nonogramRatio(totalPossibilities, totalLines),
		"max_line_possibilities": float64(maxLinePossibilities),
		"solutions":              float64(solutions),
	}
}

func nonogramActualElo(metrics difficulty.Metrics) difficulty.Elo {
	score := 0.34*normalizeNonogramMetric(metrics["cells"], 25, 225) +
		0.24*normalizeNonogramMetric(metrics["avg_line_possibilities"], 1, 120) +
		0.18*normalizeNonogramMetric(metrics["max_line_possibilities"], 1, 600) +
		0.14*(1-normalizeNonogramMetric(metrics["density"], 0.35, 0.72)) +
		0.10*normalizeNonogramMetric(metrics["avg_runs_per_line"], 1, 5)

	return difficulty.ClampElo(difficulty.Elo(math.Round(score * float64(difficulty.SoftCapElo))))
}

func hintSum(hint []int) int {
	total := 0
	for _, v := range hint {
		total += v
	}
	return total
}

func nonzeroHintCount(hint []int) int {
	if len(hint) == 1 && hint[0] == 0 {
		return 0
	}
	return len(hint)
}

func nonogramRatio(numerator, denominator int) float64 {
	if denominator == 0 {
		return 0
	}
	return float64(numerator) / float64(denominator)
}

func normalizeNonogramMetric(value, low, high float64) float64 {
	if high <= low || value <= low {
		return 0
	}
	if value >= high {
		return 1
	}
	return (value - low) / (high - low)
}
