package takuzuplus

import (
	"context"
	"crypto/sha256"
	"encoding/binary"
	"math"
	"math/rand/v2"
	"strconv"

	"github.com/FelineStateMachine/puzzletea/difficulty"
	"github.com/FelineStateMachine/puzzletea/game"
	"github.com/FelineStateMachine/puzzletea/games/takuzu"
)

var (
	_ game.EloSpawner            = TakuzuPlusMode{}
	_ game.CancellableEloSpawner = TakuzuPlusMode{}
)

type takuzuPlusEloCounterStats struct {
	Nodes    int
	Branches int
	MaxDepth int
}

func (t TakuzuPlusMode) SpawnElo(seed string, elo difficulty.Elo) (game.Gamer, difficulty.Report, error) {
	return t.SpawnEloContext(context.Background(), seed, elo)
}

func (t TakuzuPlusMode) SpawnEloContext(ctx context.Context, seed string, elo difficulty.Elo) (game.Gamer, difficulty.Report, error) {
	if err := difficulty.ValidateElo(elo); err != nil {
		return nil, difficulty.Report{}, err
	}
	if err := ctx.Err(); err != nil {
		return nil, difficulty.Report{}, err
	}

	mode := takuzuPlusModeForElo(t, elo)
	rng := takuzuPlusEloRNG(seed, elo)
	complete := generateCompleteSeeded(mode.Size, rng)
	puzzle, provided, rels := generatePuzzleSeeded(complete, mode.Size, mode.Prefilled, mode.profile, rng)

	gamer, err := New(mode, puzzle, provided, rels)
	if err != nil {
		return nil, difficulty.Report{}, err
	}
	report := scoreTakuzuPlusElo(ctx, elo, puzzle, provided, rels, mode.Prefilled)
	return gamer, report, nil
}

func takuzuPlusModeForElo(base TakuzuPlusMode, elo difficulty.Elo) TakuzuPlusMode {
	score := difficulty.Score01(elo)
	prefilled := 0.52 - score*0.24
	profileIndex := int(math.Round(score * float64(len(modeProfiles)-1)))
	if profileIndex < 0 {
		profileIndex = 0
	}
	if profileIndex >= len(modeProfiles) {
		profileIndex = len(modeProfiles) - 1
	}

	mode := base
	mode.Prefilled = prefilled
	mode.profile = modeProfiles[profileIndex]
	return mode
}

func takuzuPlusEloRNG(seed string, elo difficulty.Elo) *rand.Rand {
	sum := sha256.Sum256([]byte("takuzuplus\x00" + seed + "\x00" + strconv.Itoa(int(elo))))
	return rand.New(rand.NewPCG(
		binary.LittleEndian.Uint64(sum[0:8]),
		binary.LittleEndian.Uint64(sum[8:16]),
	))
}

func scoreTakuzuPlusElo(
	ctx context.Context,
	target difficulty.Elo,
	puzzle grid,
	provided [][]bool,
	rels relations,
	targetPrefilled float64,
) difficulty.Report {
	solutions, stats := countTakuzuPlusSolutionsWithEloStats(ctx, puzzle.clone(), len(puzzle), 2, rels)
	relationMetrics := analyzeRelations(rels, provided, len(puzzle))
	metrics := takuzuPlusDifficultyMetrics(puzzle, provided, rels, targetPrefilled, solutions, stats, relationMetrics)
	confidence := difficulty.ConfidenceHigh
	if solutions < 0 {
		confidence = difficulty.ConfidenceLow
		metrics["solver_limited"] = 1
	} else if solutions != 1 {
		confidence = difficulty.ConfidenceLow
	}

	return difficulty.Report{
		TargetElo:  target,
		ActualElo:  takuzuPlusActualElo(metrics),
		Confidence: confidence,
		Metrics:    metrics,
	}
}

func takuzuPlusDifficultyMetrics(
	puzzle grid,
	provided [][]bool,
	rels relations,
	targetPrefilled float64,
	solutions int,
	stats takuzuPlusEloCounterStats,
	relationMetrics relationMetrics,
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
	relationCount := countRelations(rels)

	return difficulty.Metrics{
		"size":               float64(size),
		"cells":              float64(cells),
		"clue_count":         float64(clues),
		"empty_count":        float64(empties),
		"clue_density":       takuzuPlusRatio(clues, cells),
		"target_prefilled":   targetPrefilled,
		"solution_count":     float64(solutions),
		"solver_nodes":       float64(stats.Nodes),
		"branches":           float64(stats.Branches),
		"max_depth":          float64(stats.MaxDepth),
		"relation_count":     float64(relationCount),
		"same_relations":     float64(relationMetrics.SameCount),
		"diff_relations":     float64(relationMetrics.DifferentCount),
		"horizontal_clues":   float64(relationMetrics.HorizontalCount),
		"vertical_clues":     float64(relationMetrics.VerticalCount),
		"occupied_regions":   float64(relationMetrics.OccupiedRegions),
		"min_relation_gap":   float64(relationMetrics.MinExistingGap),
		"one_endpoint_clues": float64(relationMetrics.EndpointCounts[endpointOneProvided]),
	}
}

func countTakuzuPlusSolutionsWithEloStats(ctx context.Context, g grid, size, limit int, rels relations) (int, takuzuPlusEloCounterStats) {
	if limit <= 0 {
		return 0, takuzuPlusEloCounterStats{}
	}
	var counter takuzuPlusEloCounterStats
	count := countTakuzuPlusSolutionsRecWithEloStats(ctx, g, size, limit, rels, 0, &counter)
	return count, counter
}

func countTakuzuPlusSolutionsRecWithEloStats(
	ctx context.Context,
	g grid,
	size int,
	limit int,
	rels relations,
	depth int,
	counter *takuzuPlusEloCounterStats,
) int {
	if ctx.Err() != nil {
		return -1
	}
	counter.Nodes++
	if depth > counter.MaxDepth {
		counter.MaxDepth = depth
	}

	choice := selectMRVCell(g, size, rels)
	if choice.x < 0 {
		if checkSolvedGrid(g, size, rels) {
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
		val := choice.vals[i]
		g[choice.y][choice.x] = val
		n := countTakuzuPlusSolutionsRecWithEloStats(ctx, g, size, limit-total, rels, depth+1, counter)
		g[choice.y][choice.x] = emptyCell
		if n < 0 {
			return n
		}
		total += n
		if total >= limit {
			return total
		}
	}
	return total
}

func checkSolvedGrid(g grid, size int, rels relations) bool {
	for y := range size {
		for x := range size {
			if g[y][x] == emptyCell {
				return false
			}
		}
	}
	return takuzu.CheckConstraintsGrid(g, size) &&
		takuzu.HasUniqueLinesGrid(g, size) &&
		checkRelations(g, rels)
}

func takuzuPlusActualElo(metrics difficulty.Metrics) difficulty.Elo {
	score := 0.30*takuzuPlusNormalize(metrics["size"], 6, 14) +
		0.24*(1-takuzuPlusNormalize(metrics["clue_density"], 0.28, 0.52)) +
		0.16*takuzuPlusNormalize(metrics["max_depth"], 0, metrics["empty_count"]) +
		0.10*takuzuPlusNormalize(metrics["branches"], 0, metrics["empty_count"]) +
		0.08*takuzuPlusNormalize(metrics["solver_nodes"], 1, math.Max(2, metrics["empty_count"]*3)) +
		0.08*takuzuPlusNormalize(metrics["relation_count"], 2, 18) +
		0.04*takuzuPlusNormalize(metrics["occupied_regions"], 1, 4)

	return difficulty.ClampElo(difficulty.Elo(math.Round(score * float64(difficulty.SoftCapElo))))
}

func takuzuPlusNormalize(value, low, high float64) float64 {
	if high <= low || value <= low {
		return 0
	}
	if value >= high {
		return 1
	}
	return (value - low) / (high - low)
}

func takuzuPlusRatio(numerator, denominator int) float64 {
	if denominator == 0 {
		return 0
	}
	return float64(numerator) / float64(denominator)
}
