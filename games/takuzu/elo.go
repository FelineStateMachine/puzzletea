package takuzu

import (
	"crypto/sha256"
	"encoding/binary"
	"math"
	"math/rand/v2"
	"strconv"

	"github.com/FelineStateMachine/puzzletea/difficulty"
	"github.com/FelineStateMachine/puzzletea/game"
)

var _ game.EloSpawner = TakuzuMode{}

type takuzuEloCounterStats struct {
	Nodes    int
	Branches int
	MaxDepth int
}

func (t TakuzuMode) SpawnElo(seed string, elo difficulty.Elo) (game.Gamer, difficulty.Report, error) {
	if err := difficulty.ValidateElo(elo); err != nil {
		return nil, difficulty.Report{}, err
	}

	mode := takuzuModeForElo(elo)
	rng := takuzuEloRNG(seed, elo)
	complete := GenerateCompleteGridSeeded(mode.Size, rng)
	puzzle, provided := GeneratePuzzleFromCompleteSeeded(complete, mode.Size, mode.Prefilled, rng)

	gamer, err := New(mode, puzzle, provided)
	if err != nil {
		return nil, difficulty.Report{}, err
	}
	return gamer, scoreTakuzuElo(elo, puzzle, provided, mode.Prefilled), nil
}

func takuzuModeForElo(elo difficulty.Elo) TakuzuMode {
	score := difficulty.Score01(elo)
	size := 6 + 2*int(math.Round(score*4))
	prefilled := 0.52 - score*0.24
	return NewMode(
		"Elo "+strconv.Itoa(int(elo)),
		"Elo-targeted Takuzu puzzle.",
		size,
		prefilled,
	)
}

func takuzuEloRNG(seed string, elo difficulty.Elo) *rand.Rand {
	sum := sha256.Sum256([]byte("takuzu\x00" + seed + "\x00" + strconv.Itoa(int(elo))))
	return rand.New(rand.NewPCG(
		binary.LittleEndian.Uint64(sum[0:8]),
		binary.LittleEndian.Uint64(sum[8:16]),
	))
}

func scoreTakuzuElo(target difficulty.Elo, puzzle [][]rune, provided [][]bool, targetPrefilled float64) difficulty.Report {
	solutions, stats := countTakuzuSolutionsWithEloStats(grid(puzzle).clone(), len(puzzle), 2)
	metrics := takuzuDifficultyMetrics(puzzle, provided, targetPrefilled, solutions, stats)
	confidence := difficulty.ConfidenceHigh
	if solutions != 1 {
		confidence = difficulty.ConfidenceLow
	}

	return difficulty.Report{
		TargetElo:  target,
		ActualElo:  takuzuActualElo(metrics),
		Confidence: confidence,
		Metrics:    metrics,
	}
}

func takuzuDifficultyMetrics(
	puzzle [][]rune,
	provided [][]bool,
	targetPrefilled float64,
	solutions int,
	stats takuzuEloCounterStats,
) difficulty.Metrics {
	size := len(puzzle)
	cells := size * size
	clues := 0
	for y := range provided {
		for x := range provided[y] {
			if provided[y][x] {
				clues++
			}
		}
	}
	empties := cells - clues

	return difficulty.Metrics{
		"size":             float64(size),
		"cells":            float64(cells),
		"clue_count":       float64(clues),
		"empty_count":      float64(empties),
		"clue_density":     takuzuRatio(clues, cells),
		"target_prefilled": targetPrefilled,
		"solution_count":   float64(solutions),
		"solver_nodes":     float64(stats.Nodes),
		"branches":         float64(stats.Branches),
		"max_depth":        float64(stats.MaxDepth),
	}
}

func countTakuzuSolutionsWithEloStats(g grid, size, limit int) (int, takuzuEloCounterStats) {
	if limit <= 0 {
		return 0, takuzuEloCounterStats{}
	}
	stats := newLineStats(g, size)
	var counter takuzuEloCounterStats
	count := countTakuzuSolutionsRecWithEloStats(g, size, limit, stats, 0, &counter)
	return count, counter
}

func countTakuzuSolutionsRecWithEloStats(
	g grid,
	size int,
	limit int,
	lineStats *lineStats,
	depth int,
	counter *takuzuEloCounterStats,
) int {
	counter.Nodes++
	if depth > counter.MaxDepth {
		counter.MaxDepth = depth
	}

	choice := selectMRVCell(g, size, lineStats, nil)
	if choice.x < 0 {
		if hasUniqueLines(g, size) {
			return 1
		}
		return 0
	}
	if choice.count == 0 {
		return 0
	}
	if choice.count > 1 {
		counter.Branches++
	}

	total := 0
	for i := range choice.count {
		v := choice.vals[i]
		g[choice.y][choice.x] = v
		lineStats.apply(choice.x, choice.y, v, 1)
		total += countTakuzuSolutionsRecWithEloStats(g, size, limit-total, lineStats, depth+1, counter)
		lineStats.apply(choice.x, choice.y, v, -1)
		g[choice.y][choice.x] = emptyCell
		if total >= limit {
			return total
		}
	}

	return total
}

func takuzuActualElo(metrics difficulty.Metrics) difficulty.Elo {
	score := 0.34*takuzuNormalize(metrics["size"], 6, 14) +
		0.28*(1-takuzuNormalize(metrics["clue_density"], 0.28, 0.52)) +
		0.18*takuzuNormalize(metrics["max_depth"], 0, metrics["empty_count"]) +
		0.12*takuzuNormalize(metrics["branches"], 0, metrics["empty_count"]) +
		0.08*takuzuNormalize(metrics["solver_nodes"], 1, math.Max(2, metrics["empty_count"]*3))

	return difficulty.ClampElo(difficulty.Elo(math.Round(score * float64(difficulty.SoftCapElo))))
}

func takuzuNormalize(value, low, high float64) float64 {
	if high <= low || value <= low {
		return 0
	}
	if value >= high {
		return 1
	}
	return (value - low) / (high - low)
}

func takuzuRatio(numerator, denominator int) float64 {
	if denominator == 0 {
		return 0
	}
	return float64(numerator) / float64(denominator)
}
