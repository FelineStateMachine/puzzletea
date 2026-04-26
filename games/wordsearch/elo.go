package wordsearch

import (
	"fmt"
	"hash/fnv"
	"math"
	"math/rand/v2"

	"github.com/FelineStateMachine/puzzletea/difficulty"
	"github.com/FelineStateMachine/puzzletea/game"
)

var _ game.EloSpawner = WordSearchMode{}

type wordSearchEloSpec struct {
	width       int
	height      int
	wordCount   int
	minWordLen  int
	maxWordLen  int
	allowedDirs []Direction
}

func (w WordSearchMode) SpawnElo(seed string, elo difficulty.Elo) (game.Gamer, difficulty.Report, error) {
	if err := difficulty.ValidateElo(elo); err != nil {
		return nil, difficulty.Report{}, err
	}

	spec := wordSearchSpecForElo(w, elo)
	var bestGrid grid
	var bestWords []Word
	var bestReport difficulty.Report
	haveBest := false
	for candidate := range difficulty.CandidateCount(elo) {
		rng := wordSearchEloRNG(wordSearchCandidateSeed(seed, candidate), elo)
		g, words, stats := generateWordSearchSeededWithStats(
			spec.width,
			spec.height,
			spec.wordCount,
			spec.minWordLen,
			spec.maxWordLen,
			spec.allowedDirs,
			rng,
		)
		report := wordSearchDifficultyReport(elo, spec, words, stats)
		if difficulty.BetterCandidate(report, bestReport, elo, haveBest) {
			bestGrid = g
			bestWords = words
			bestReport = report
			haveBest = true
		}
	}

	mode := w
	mode.Width = spec.width
	mode.Height = spec.height
	mode.WordCount = spec.wordCount
	mode.MinWordLen = spec.minWordLen
	mode.MaxWordLen = spec.maxWordLen
	mode.AllowedDirs = append([]Direction(nil), spec.allowedDirs...)

	gamer, err := New(mode, bestGrid, bestWords)
	if err != nil {
		return nil, difficulty.Report{}, err
	}

	return gamer, bestReport, nil
}

func wordSearchSpecForElo(base WordSearchMode, elo difficulty.Elo) wordSearchEloSpec {
	score := difficulty.Score01(elo)
	dirCount := 3 + int(math.Round(score*5))
	if dirCount > len(selectionDirections) {
		dirCount = len(selectionDirections)
	}

	minLen := 3 + int(math.Floor(score*3))
	maxLen := 5 + int(math.Round(score*5))
	maxAllowedLen := max(base.Width, base.Height)
	minLen = min(minLen, maxAllowedLen)
	maxLen = min(maxLen, maxAllowedLen)
	if maxLen < minLen {
		maxLen = minLen
	}
	wordCount := 5 + int(math.Round(score*13))
	wordCount = min(wordCount, max(1, base.Width*base.Height/minLen))

	return wordSearchEloSpec{
		width:       base.Width,
		height:      base.Height,
		wordCount:   wordCount,
		minWordLen:  minLen,
		maxWordLen:  maxLen,
		allowedDirs: append([]Direction(nil), selectionDirections[:dirCount]...),
	}
}

func wordSearchEloRNG(seed string, elo difficulty.Elo) *rand.Rand {
	h := fnv.New64a()
	_, _ = h.Write([]byte("wordsearch"))
	_, _ = h.Write([]byte{0})
	_, _ = h.Write([]byte(seed))
	_, _ = h.Write([]byte(fmt.Sprintf("\x00%d", elo)))
	s := h.Sum64()
	return rand.New(rand.NewPCG(s, ^s))
}

func wordSearchCandidateSeed(seed string, candidate int) string {
	if candidate == 0 {
		return seed
	}
	return seed + fmt.Sprintf("\x00candidate:%d", candidate)
}

func wordSearchDifficultyReport(
	target difficulty.Elo,
	spec wordSearchEloSpec,
	words []Word,
	stats placementStats,
) difficulty.Report {
	metrics := wordSearchDifficultyMetrics(spec, words, stats)
	actual := wordSearchActualElo(metrics, stats)

	confidence := difficulty.ConfidenceHigh
	if stats.PlacedWords != stats.TargetWords {
		confidence = difficulty.ConfidenceMedium
	}

	return difficulty.Report{
		TargetElo:  target,
		ActualElo:  actual,
		Confidence: confidence,
		Metrics:    metrics,
	}
}

func wordSearchDifficultyMetrics(spec wordSearchEloSpec, words []Word, stats placementStats) difficulty.Metrics {
	occupied := occupiedWordCells(words)
	totalCells := spec.width * spec.height
	avgWordLen, maxWordLen := wordLengthMetrics(words)
	uniqueDirs := uniqueDirectionCount(words)
	fallbackRate := 0.0
	if stats.TargetWords > 0 {
		fallbackRate = float64(stats.FallbackUsed) / float64(stats.TargetWords)
	}

	return difficulty.Metrics{
		"width":             float64(spec.width),
		"height":            float64(spec.height),
		"cells":             float64(totalCells),
		"target_words":      float64(stats.TargetWords),
		"placed_words":      float64(stats.PlacedWords),
		"failed_words":      float64(stats.FailedWords),
		"success_rate":      stats.SuccessRate(),
		"avg_word_length":   avgWordLen,
		"max_word_length":   float64(maxWordLen),
		"unique_directions": float64(uniqueDirs),
		"occupied_cells":    float64(occupied),
		"density":           safeRatio(occupied, totalCells),
		"attempts_per_word": stats.AttemptsPerWord(),
		"fallback_rate":     fallbackRate,
	}
}

func wordSearchActualElo(metrics difficulty.Metrics, stats placementStats) difficulty.Elo {
	score := 0.16*normalize(metrics["cells"], 100, 400) +
		0.24*normalize(metrics["placed_words"], 5, 18) +
		0.22*normalize(metrics["avg_word_length"], 3, 10) +
		0.15*normalize(metrics["unique_directions"], 3, 8) +
		0.16*normalize(metrics["density"], 0.08, 0.36) +
		0.07*normalize(metrics["attempts_per_word"], 1, 20)

	score *= stats.SuccessRate()
	return difficulty.ClampElo(difficulty.Elo(math.Round(score * float64(difficulty.SoftCapElo))))
}

func occupiedWordCells(words []Word) int {
	seen := make(map[Position]struct{})
	for i := range words {
		for _, pos := range words[i].Positions() {
			seen[pos] = struct{}{}
		}
	}
	return len(seen)
}

func wordLengthMetrics(words []Word) (float64, int) {
	if len(words) == 0 {
		return 0, 0
	}

	total := 0
	maxLen := 0
	for _, word := range words {
		length := len(word.Text)
		total += length
		if length > maxLen {
			maxLen = length
		}
	}
	return float64(total) / float64(len(words)), maxLen
}

func uniqueDirectionCount(words []Word) int {
	seen := make(map[Direction]struct{})
	for _, word := range words {
		seen[word.Direction] = struct{}{}
	}
	return len(seen)
}

func safeRatio(numerator, denominator int) float64 {
	if denominator == 0 {
		return 0
	}
	return float64(numerator) / float64(denominator)
}

func normalize(value, min, max float64) float64 {
	if max <= min {
		return 0
	}
	if value <= min {
		return 0
	}
	if value >= max {
		return 1
	}
	return (value - min) / (max - min)
}
