package shikaku

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

var _ game.EloSpawner = ShikakuMode{}

type shikakuEloSolveStats struct {
	Nodes    int
	Branches int
	MaxDepth int
}

func (s ShikakuMode) SpawnElo(seed string, elo difficulty.Elo) (game.Gamer, difficulty.Report, error) {
	if err := difficulty.ValidateElo(elo); err != nil {
		return nil, difficulty.Report{}, err
	}

	mode := shikakuModeForElo(s, elo)
	var bestPuzzle Puzzle
	var bestReport difficulty.Report
	haveBest := false
	var lastErr error
	for candidate := range difficulty.CandidateCount(elo) {
		puzzle, err := GeneratePuzzleSeeded(mode.Width, mode.Height, mode.MaxRectSize, shikakuEloRNG(shikakuCandidateSeed(seed, candidate), elo))
		if err != nil {
			lastErr = err
			continue
		}
		report := scoreShikakuElo(elo, mode, puzzle)
		if difficulty.BetterCandidate(report, bestReport, elo, haveBest) {
			bestPuzzle = puzzle
			bestReport = report
			haveBest = true
		}
	}
	if !haveBest {
		if lastErr == nil {
			lastErr = errors.New("unable to generate Elo shikaku")
		}
		return nil, difficulty.Report{}, lastErr
	}
	return New(mode, bestPuzzle), bestReport, nil
}

func shikakuModeForElo(base ShikakuMode, elo difficulty.Elo) ShikakuMode {
	score := difficulty.Score01(elo)
	maxRectSize := 5 + int(math.Round(score*15))
	mode := base
	mode.MaxRectSize = min(maxRectSize, max(1, mode.Width*mode.Height))
	return mode
}

func shikakuEloRNG(seed string, elo difficulty.Elo) *rand.Rand {
	sum := sha256.Sum256([]byte("shikaku\x00" + seed + "\x00" + strconv.Itoa(int(elo))))
	return rand.New(rand.NewPCG(
		binary.LittleEndian.Uint64(sum[0:8]),
		binary.LittleEndian.Uint64(sum[8:16]),
	))
}

func shikakuCandidateSeed(seed string, candidate int) string {
	if candidate == 0 {
		return seed
	}
	return seed + "\x00candidate:" + strconv.Itoa(candidate)
}

func scoreShikakuElo(target difficulty.Elo, mode ShikakuMode, puzzle Puzzle) difficulty.Report {
	solutions, stats := countShikakuSolutionsWithEloStats(puzzle, 2)
	metrics := shikakuDifficultyMetrics(mode, puzzle, solutions, stats)
	confidence := difficulty.ConfidenceHigh
	if solutions != 1 {
		confidence = difficulty.ConfidenceMedium
	}
	return difficulty.Report{
		TargetElo:  target,
		ActualElo:  shikakuActualElo(metrics),
		Confidence: confidence,
		Metrics:    metrics,
	}
}

func shikakuDifficultyMetrics(
	mode ShikakuMode,
	puzzle Puzzle,
	solutions int,
	stats shikakuEloSolveStats,
) difficulty.Metrics {
	cells := puzzle.Width * puzzle.Height
	clueCount := len(puzzle.Clues)
	metrics := difficulty.Metrics{
		"width":          float64(puzzle.Width),
		"height":         float64(puzzle.Height),
		"cells":          float64(cells),
		"clue_count":     float64(clueCount),
		"max_rect_size":  float64(mode.MaxRectSize),
		"solution_count": float64(solutions),
		"solver_nodes":   float64(stats.Nodes),
		"branches":       float64(stats.Branches),
		"max_depth":      float64(stats.MaxDepth),
	}
	if cells == 0 || clueCount == 0 {
		return metrics
	}

	totalCandidates := 0
	maxCandidates := 0
	for _, clue := range puzzle.Clues {
		if clue.Value == 1 {
			metrics["singleton_clues"]++
		}
		if float64(clue.Value) > metrics["largest_clue"] {
			metrics["largest_clue"] = float64(clue.Value)
		}
		candidateCount := len((&puzzle).CandidateRectanglesForClue(clue.ID))
		totalCandidates += candidateCount
		if candidateCount > maxCandidates {
			maxCandidates = candidateCount
		}
	}
	metrics["average_rect_area"] = float64(cells) / float64(clueCount)
	metrics["clue_density"] = float64(clueCount) / float64(cells)
	metrics["singleton_ratio"] = metrics["singleton_clues"] / float64(clueCount)
	metrics["candidate_count"] = float64(totalCandidates)
	metrics["average_candidates"] = float64(totalCandidates) / float64(clueCount)
	metrics["max_candidates"] = float64(maxCandidates)

	return metrics
}

func countShikakuSolutionsWithEloStats(puzzle Puzzle, limit int) (int, shikakuEloSolveStats) {
	if limit <= 0 || puzzle.Width <= 0 || puzzle.Height <= 0 {
		return 0, shikakuEloSolveStats{}
	}

	candidates := make(map[int][]Rectangle, len(puzzle.Clues))
	for _, clue := range puzzle.Clues {
		candidates[clue.ID] = (&puzzle).CandidateRectanglesForClue(clue.ID)
		if len(candidates[clue.ID]) == 0 {
			return 0, shikakuEloSolveStats{}
		}
	}

	covered := make([][]bool, puzzle.Height)
	for y := range puzzle.Height {
		covered[y] = make([]bool, puzzle.Width)
	}
	placed := make(map[int]bool, len(puzzle.Clues))
	stats := shikakuEloSolveStats{}
	count := countShikakuSolutionsRec(puzzle, candidates, covered, placed, 0, limit, 0, &stats)
	return count, stats
}

func countShikakuSolutionsRec(
	puzzle Puzzle,
	candidates map[int][]Rectangle,
	covered [][]bool,
	placed map[int]bool,
	coveredCells int,
	limit int,
	depth int,
	stats *shikakuEloSolveStats,
) int {
	stats.Nodes++
	if depth > stats.MaxDepth {
		stats.MaxDepth = depth
	}
	if len(placed) == len(puzzle.Clues) {
		if coveredCells == puzzle.Width*puzzle.Height {
			return 1
		}
		return 0
	}

	clueID, choices := shikakuNextEloChoices(puzzle, candidates, covered, placed)
	if len(choices) == 0 {
		return 0
	}

	total := 0
	for _, rect := range choices {
		stats.Branches++
		placed[clueID] = true
		markShikakuCovered(covered, rect, true)
		total += countShikakuSolutionsRec(
			puzzle,
			candidates,
			covered,
			placed,
			coveredCells+rect.Area(),
			limit-total,
			depth+1,
			stats,
		)
		markShikakuCovered(covered, rect, false)
		delete(placed, clueID)
		if total >= limit {
			return total
		}
	}
	return total
}

func shikakuNextEloChoices(
	puzzle Puzzle,
	candidates map[int][]Rectangle,
	covered [][]bool,
	placed map[int]bool,
) (int, []Rectangle) {
	bestID := -1
	var best []Rectangle
	for _, clue := range puzzle.Clues {
		if placed[clue.ID] {
			continue
		}
		choices := make([]Rectangle, 0, len(candidates[clue.ID]))
		for _, rect := range candidates[clue.ID] {
			if shikakuRectAvailable(covered, rect) {
				choices = append(choices, rect)
			}
		}
		if bestID == -1 || len(choices) < len(best) {
			bestID = clue.ID
			best = choices
			if len(best) == 0 {
				break
			}
		}
	}
	return bestID, best
}

func shikakuRectAvailable(covered [][]bool, rect Rectangle) bool {
	for y := rect.Y; y < rect.Y+rect.H; y++ {
		for x := rect.X; x < rect.X+rect.W; x++ {
			if covered[y][x] {
				return false
			}
		}
	}
	return true
}

func markShikakuCovered(covered [][]bool, rect Rectangle, value bool) {
	for y := rect.Y; y < rect.Y+rect.H; y++ {
		for x := rect.X; x < rect.X+rect.W; x++ {
			covered[y][x] = value
		}
	}
}

func shikakuActualElo(metrics difficulty.Metrics) difficulty.Elo {
	metricScore := 0.30*normalizeShikakuMetric(metrics["cells"], 25, 144) +
		0.20*normalizeShikakuMetric(metrics["average_candidates"], 1.5, 8.0) +
		0.20*normalizeShikakuMetric(math.Log10(metrics["solver_nodes"]+1), 1.0, 3.2) +
		0.15*normalizeShikakuMetric(metrics["branches"], 8, 80) +
		0.10*normalizeShikakuMetric(metrics["max_rect_size"], 5, 20) +
		0.05*(1-normalizeShikakuMetric(metrics["clue_density"], 0.18, 0.55))
	targetScore := normalizeShikakuMetric(metrics["max_rect_size"], 5, 20)
	score := 0.55*targetScore + 0.45*metricScore
	if metrics["solution_count"] != 1 {
		score *= 0.85
	}
	return difficulty.ClampElo(difficulty.Elo(math.Round(score * float64(difficulty.SoftCapElo))))
}

func normalizeShikakuMetric(value, low, high float64) float64 {
	if high <= low || value <= low {
		return 0
	}
	if value >= high {
		return 1
	}
	return (value - low) / (high - low)
}
