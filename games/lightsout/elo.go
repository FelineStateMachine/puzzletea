package lightsout

import (
	"crypto/sha256"
	"encoding/binary"
	"math"
	"math/rand/v2"
	"strconv"

	"github.com/FelineStateMachine/puzzletea/difficulty"
	"github.com/FelineStateMachine/puzzletea/game"
)

var _ game.EloSpawner = Mode{}

type eloSpec struct {
	width  int
	height int
}

func (m Mode) SpawnElo(seed string, elo difficulty.Elo) (game.Gamer, difficulty.Report, error) {
	if err := difficulty.ValidateElo(elo); err != nil {
		return nil, difficulty.Report{}, err
	}

	spec := lightsOutSpecForMode(m)
	rng := lightsOutRandForElo(seed, elo)
	grid, report := generateLightsOutEloGrid(spec, rng, elo)

	mode := m

	g := newFromGrid(mode.Width, mode.Height, grid)
	return g, report, nil
}

func lightsOutSpecForMode(mode Mode) eloSpec {
	return eloSpec{width: mode.Width, height: mode.Height}
}

func lightsOutRandForElo(seed string, elo difficulty.Elo) *rand.Rand {
	sum := sha256.Sum256([]byte("lightsout\x00" + seed + "\x00" + strconv.Itoa(int(elo))))
	return rand.New(rand.NewPCG(
		binary.LittleEndian.Uint64(sum[0:8]),
		binary.LittleEndian.Uint64(sum[8:16]),
	))
}

func generateLightsOutEloGrid(spec eloSpec, rng *rand.Rand, target difficulty.Elo) ([][]bool, difficulty.Report) {
	best := GenerateSeeded(spec.width, spec.height, rng)
	bestReport := scoreLightsOutGrid(target, best)

	for range 23 {
		candidate := GenerateSeeded(spec.width, spec.height, rng)
		report := scoreLightsOutGrid(target, candidate)
		if eloDistance(report.ActualElo, target) < eloDistance(bestReport.ActualElo, target) {
			best = candidate
			bestReport = report
		}
	}

	return best, bestReport
}

func scoreLightsOutGrid(target difficulty.Elo, grid [][]bool) difficulty.Report {
	metrics := lightsOutMetrics(grid)
	score := 0.45*normalizeLightsOut(metrics["cells"], 9, 81) +
		0.35*normalizeLightsOut(metrics["min_solution_moves"], 1, 41) +
		0.20*normalizeLightsOut(metrics["lights_on"], 1, 81)

	return difficulty.Report{
		TargetElo:  target,
		ActualElo:  difficulty.ClampElo(difficulty.Elo(math.Round(score * float64(difficulty.SoftCapElo)))),
		Confidence: difficulty.ConfidenceHigh,
		Metrics:    metrics,
	}
}

func lightsOutMetrics(grid [][]bool) difficulty.Metrics {
	height := len(grid)
	width := 0
	if height > 0 {
		width = len(grid[0])
	}
	cells := width * height
	lightsOn := 0
	for y := range grid {
		for x := range grid[y] {
			if grid[y][x] {
				lightsOn++
			}
		}
	}

	minMoves, solved := minSolutionMoves(grid)
	solvable := 0.0
	if solved {
		solvable = 1
	}

	return difficulty.Metrics{
		"width":              float64(width),
		"height":             float64(height),
		"cells":              float64(cells),
		"lights_on":          float64(lightsOn),
		"light_density":      ratio(lightsOn, cells),
		"min_solution_moves": float64(minMoves),
		"move_density":       ratio(minMoves, cells),
		"solvable":           solvable,
	}
}

func minSolutionMoves(grid [][]bool) (int, bool) {
	height := len(grid)
	if height == 0 {
		return 0, true
	}
	width := len(grid[0])
	if width == 0 {
		return 0, true
	}

	best := width*height + 1
	for firstRowMask := range 1 << width {
		work := copyBoolGrid(grid)
		moves := 0

		for x := range width {
			if firstRowMask&(1<<x) != 0 {
				Toggle(work, x, 0)
				moves++
			}
		}

		for y := 1; y < height; y++ {
			for x := range width {
				if work[y-1][x] {
					Toggle(work, x, y)
					moves++
				}
			}
		}

		if IsSolved(work) && moves < best {
			best = moves
		}
	}

	if best == width*height+1 {
		return 0, false
	}
	return best, true
}

func copyBoolGrid(grid [][]bool) [][]bool {
	copied := make([][]bool, len(grid))
	for y := range grid {
		copied[y] = make([]bool, len(grid[y]))
		copy(copied[y], grid[y])
	}
	return copied
}

func normalizeLightsOut(value, minValue, maxValue float64) float64 {
	if maxValue <= minValue {
		return 0
	}
	return math.Max(0, math.Min(1, (value-minValue)/(maxValue-minValue)))
}

func ratio(numerator, denominator int) float64 {
	if denominator == 0 {
		return 0
	}
	return float64(numerator) / float64(denominator)
}

func eloDistance(a, b difficulty.Elo) difficulty.Elo {
	if a > b {
		return a - b
	}
	return b - a
}
