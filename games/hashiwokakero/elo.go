package hashiwokakero

import (
	"crypto/sha256"
	"encoding/binary"
	"math"
	"math/rand/v2"
	"strconv"

	"github.com/FelineStateMachine/puzzletea/difficulty"
	"github.com/FelineStateMachine/puzzletea/game"
)

var _ game.EloSpawner = HashiMode{}

func (h HashiMode) SpawnElo(seed string, elo difficulty.Elo) (game.Gamer, difficulty.Report, error) {
	if err := difficulty.ValidateElo(elo); err != nil {
		return nil, difficulty.Report{}, err
	}

	mode := hashiModeForElo(elo)
	puzzle, err := GeneratePuzzleSeeded(mode, hashiEloRNG(seed, elo))
	if err != nil {
		return nil, difficulty.Report{}, err
	}
	return New(mode, puzzle), hashiDifficultyReport(elo, mode, puzzle), nil
}

func hashiModeForElo(elo difficulty.Elo) HashiMode {
	score := difficulty.Score01(elo)
	size := 7 + 2*int(math.Round(score*3))
	density := 0.17 + score*0.23
	target := int(math.Round(float64(size*size) * density))
	minIslands := max(3, target-2)
	maxIslands := min(size*size, target+2)

	return NewMode(
		"Elo "+strconv.Itoa(int(elo)),
		"Elo-targeted Hashiwokakero puzzle.",
		size,
		size,
		minIslands,
		maxIslands,
	)
}

func hashiEloRNG(seed string, elo difficulty.Elo) *rand.Rand {
	gameID := string(Definition.ID)
	if gameID == "" {
		gameID = "hashiwokakero"
	}
	sum := sha256.Sum256([]byte(gameID + "\x00" + seed + "\x00" + strconv.Itoa(int(elo))))
	return rand.New(rand.NewPCG(
		binary.LittleEndian.Uint64(sum[0:8]),
		binary.LittleEndian.Uint64(sum[8:16]),
	))
}

func hashiDifficultyReport(target difficulty.Elo, mode HashiMode, puzzle Puzzle) difficulty.Report {
	metrics := hashiDifficultyMetrics(mode, puzzle)
	return difficulty.Report{
		TargetElo:  target,
		ActualElo:  hashiActualElo(metrics),
		Confidence: difficulty.ConfidenceMedium,
		Metrics:    metrics,
	}
}

func hashiDifficultyMetrics(mode HashiMode, puzzle Puzzle) difficulty.Metrics {
	cells := puzzle.Width * puzzle.Height
	islands := len(puzzle.Islands)
	totalRequired := 0
	maxRequired := 0
	highRequired := 0
	for _, island := range puzzle.Islands {
		totalRequired += island.Required
		if island.Required > maxRequired {
			maxRequired = island.Required
		}
		if island.Required >= 5 {
			highRequired++
		}
	}

	connectablePairs := len(connectableIslandPairs(&puzzle))
	return difficulty.Metrics{
		"width":             float64(puzzle.Width),
		"height":            float64(puzzle.Height),
		"cells":             float64(cells),
		"islands":           float64(islands),
		"island_density":    hashiRatio(islands, cells),
		"target_minlands":   float64(mode.MinIslands),
		"target_maxlands":   float64(mode.MaxIslands),
		"total_required":    float64(totalRequired),
		"avg_required":      hashiRatio(totalRequired, islands),
		"max_required":      float64(maxRequired),
		"high_required":     float64(highRequired),
		"connectable_pairs": float64(connectablePairs),
		"pair_density":      hashiRatio(connectablePairs, max(1, islands)),
	}
}

func hashiActualElo(metrics difficulty.Metrics) difficulty.Elo {
	score := 0.24*normalizeHashi(metrics["width"], 7, 13) +
		0.20*normalizeHashi(metrics["island_density"], 0.14, 0.42) +
		0.17*normalizeHashi(metrics["total_required"]/metrics["cells"], 0.16, 0.58) +
		0.15*normalizeHashi(metrics["avg_required"], 1.6, 5.0) +
		0.13*normalizeHashi(metrics["pair_density"], 1.0, 3.5) +
		0.11*normalizeHashi(metrics["high_required"], 0, metrics["islands"]*0.45)

	return difficulty.ClampElo(difficulty.Elo(math.Round(score * float64(difficulty.SoftCapElo))))
}

func normalizeHashi(value, low, high float64) float64 {
	if high <= low || value <= low {
		return 0
	}
	if value >= high {
		return 1
	}
	return (value - low) / (high - low)
}

func hashiRatio(numerator, denominator int) float64 {
	if denominator == 0 {
		return 0
	}
	return float64(numerator) / float64(denominator)
}
