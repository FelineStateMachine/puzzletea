package nurikabe

import (
	"context"
	"crypto/sha256"
	"encoding/binary"
	"errors"
	"math"
	"math/rand/v2"
	"strconv"

	"github.com/FelineStateMachine/puzzletea/difficulty"
	"github.com/FelineStateMachine/puzzletea/game"
)

var (
	_ game.EloSpawner            = NurikabeMode{}
	_ game.CancellableEloSpawner = NurikabeMode{}
)

func (n NurikabeMode) SpawnElo(seed string, elo difficulty.Elo) (game.Gamer, difficulty.Report, error) {
	return n.SpawnEloContext(context.Background(), seed, elo)
}

func (n NurikabeMode) SpawnEloContext(ctx context.Context, seed string, elo difficulty.Elo) (game.Gamer, difficulty.Report, error) {
	if err := difficulty.ValidateElo(elo); err != nil {
		return nil, difficulty.Report{}, err
	}

	mode := nurikabeModeForElo(elo)
	puzzle, err := GenerateSeededWithContext(ctx, mode, nurikabeEloRNG(seed, elo))
	if err != nil {
		return nil, difficulty.Report{}, err
	}

	gamer, err := New(mode, puzzle)
	if err != nil {
		return nil, difficulty.Report{}, err
	}
	return gamer, scoreNurikabeElo(ctx, elo, puzzle), nil
}

func nurikabeModeForElo(elo difficulty.Elo) NurikabeMode {
	score := difficulty.Score01(elo)
	size := 5 + int(math.Round(score*7))
	clueDensity := 0.28 - score*0.14
	maxIslandSize := size + int(math.Round(score*4))

	return NewMode(
		"Elo "+strconv.Itoa(int(elo)),
		"Elo-targeted Nurikabe puzzle.",
		size,
		size,
		clueDensity,
		maxIslandSize,
	)
}

func nurikabeEloRNG(seed string, elo difficulty.Elo) *rand.Rand {
	sum := sha256.Sum256([]byte("nurikabe\x00" + seed + "\x00" + strconv.Itoa(int(elo))))
	return rand.New(rand.NewPCG(
		binary.LittleEndian.Uint64(sum[0:8]),
		binary.LittleEndian.Uint64(sum[8:16]),
	))
}

func scoreNurikabeElo(ctx context.Context, target difficulty.Elo, puzzle Puzzle) difficulty.Report {
	metrics := nurikabeDifficultyMetrics(puzzle)

	solutions, stats, err := CountSolutionsContext(ctx, puzzle, 2, 200000)
	metrics["solution_count"] = float64(solutions)
	metrics["solver_nodes"] = float64(stats.Nodes)
	metrics["branches"] = float64(stats.Branches)
	metrics["max_depth"] = float64(stats.MaxDepth)

	confidence := difficulty.ConfidenceHigh
	if err != nil {
		confidence = difficulty.ConfidenceLow
		if errors.Is(err, errNodeLimit) ||
			errors.Is(err, context.Canceled) ||
			errors.Is(err, context.DeadlineExceeded) {
			metrics["solver_limited"] = 1
		}
	} else if solutions != 1 {
		confidence = difficulty.ConfidenceMedium
	}

	return difficulty.Report{
		TargetElo:  target,
		ActualElo:  nurikabeActualElo(metrics),
		Confidence: confidence,
		Metrics:    metrics,
	}
}

func nurikabeDifficultyMetrics(puzzle Puzzle) difficulty.Metrics {
	metrics := difficulty.Metrics{
		"width":         float64(puzzle.Width),
		"height":        float64(puzzle.Height),
		"cells":         float64(puzzle.Width * puzzle.Height),
		"unknown_count": float64(puzzle.Width * puzzle.Height),
	}

	sizes := make([]int, 0, puzzle.Width*puzzle.Height)
	for y := range puzzle.Height {
		for x := range puzzle.Width {
			clue := puzzle.Clues[y][x]
			if clue <= 0 {
				continue
			}
			metrics["clue_count"]++
			metrics["island_cells"] += float64(clue)
			sizes = append(sizes, clue)
			if clue == 1 {
				metrics["singleton_clues"]++
			}
			if float64(clue) > metrics["largest_island"] {
				metrics["largest_island"] = float64(clue)
			}
		}
	}

	if metrics["clue_count"] > 0 {
		metrics["average_island"] = metrics["island_cells"] / metrics["clue_count"]
		metrics["singleton_ratio"] = metrics["singleton_clues"] / metrics["clue_count"]
		metrics["size_entropy"] = componentSizeEntropy(sizes)
		metrics["normalized_spread"] = normalizedSpread(sizes)
	}
	metrics["sea_cells"] = metrics["cells"] - metrics["island_cells"]
	metrics["clue_density"] = metrics["clue_count"] / metrics["cells"]
	metrics["sea_ratio"] = metrics["sea_cells"] / metrics["cells"]
	metrics["unknown_count"] = metrics["cells"] - metrics["clue_count"]

	return metrics
}

func nurikabeActualElo(metrics difficulty.Metrics) difficulty.Elo {
	score := 0.25*normalizeNurikabeMetric(metrics["cells"], 25, 144) +
		0.20*normalizeNurikabeMetric(metrics["unknown_count"], 18, 130) +
		0.20*normalizeNurikabeMetric(math.Log10(metrics["solver_nodes"]+1), 1.2, 5.0) +
		0.15*normalizeNurikabeMetric(metrics["branches"], 0, 180) +
		0.10*normalizeNurikabeMetric(metrics["max_depth"], 0, 80) +
		0.10*normalizeNurikabeMetric(metrics["normalized_spread"], 0.2, 1.8)

	if metrics["solution_count"] != 1 {
		score *= 0.85
	}
	return difficulty.ClampElo(difficulty.Elo(math.Round(score * float64(difficulty.SoftCapElo))))
}

func normalizeNurikabeMetric(value, low, high float64) float64 {
	if value <= low {
		return 0
	}
	if value >= high {
		return 1
	}
	return (value - low) / (high - low)
}
