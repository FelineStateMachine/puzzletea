package spellpuzzle

import (
	"crypto/sha256"
	"encoding/binary"
	"math"
	"math/rand/v2"
	"strconv"

	"github.com/FelineStateMachine/puzzletea/difficulty"
	"github.com/FelineStateMachine/puzzletea/game"
)

func (m SpellPuzzleMode) SpawnElo(seed string, elo difficulty.Elo) (game.Gamer, difficulty.Report, error) {
	if err := difficulty.ValidateElo(elo); err != nil {
		return nil, difficulty.Report{}, err
	}

	mode := spellPuzzleModeForElo(m, elo)
	puzzle, err := GeneratePuzzleSeeded(mode, randForElo(seed, elo))
	if err != nil {
		return nil, difficulty.Report{}, err
	}

	report := scoreEloPuzzle(elo, puzzle)
	g, err := New(mode, puzzle)
	if err != nil {
		return nil, difficulty.Report{}, err
	}
	return g, report, nil
}

func spellPuzzleModeForElo(base SpellPuzzleMode, elo difficulty.Elo) SpellPuzzleMode {
	score := difficulty.Score01(elo)
	bankSize := 6 + int(math.Round(score*3))
	boardWords := 4 + int(math.Round(score*5))
	minBonusWords := 3 + int(math.Round(score*5))

	mode := base
	mode.BankSize = bankSize
	mode.BoardWordCount = boardWords
	mode.MinBonusWords = minBonusWords
	return mode
}

func randForElo(seed string, elo difficulty.Elo) *rand.Rand {
	sum := sha256.Sum256([]byte("spellpuzzle\x00" + seed + "\x00" + strconv.Itoa(int(elo))))
	return rand.New(rand.NewPCG(
		binary.LittleEndian.Uint64(sum[0:8]),
		binary.LittleEndian.Uint64(sum[8:16]),
	))
}

func scoreEloPuzzle(target difficulty.Elo, puzzle GeneratedPuzzle) difficulty.Report {
	metrics := spellPuzzleMetrics(puzzle)
	score := normalizedMetric(metrics["bank_size"], 6, 9)*0.20 +
		normalizedMetric(metrics["board_word_count"], 4, 9)*0.20 +
		normalizedMetric(metrics["bonus_word_count"], 3, 16)*0.15 +
		normalizedMetric(metrics["average_word_length"], 3, 9)*0.15 +
		normalizedMetric(metrics["crossing_count"], 1, 8)*0.15 +
		normalizedMetric(metrics["allowed_word_count"], 10, 80)*0.15

	return difficulty.Report{
		TargetElo:  target,
		ActualElo:  difficulty.Elo(math.Round(score * float64(difficulty.SoftCapElo))),
		Confidence: difficulty.ConfidenceMedium,
		Metrics:    metrics,
	}
}

func spellPuzzleMetrics(puzzle GeneratedPuzzle) difficulty.Metrics {
	minLength, maxLength, averageLength := wordLengthStats(puzzle.Placements)
	repeatedLetters, entropy := bankLetterStats(puzzle.Bank)
	board := buildBoard(puzzle.Placements)

	return difficulty.Metrics{
		"bank_size":           float64(len([]rune(puzzle.Bank))),
		"board_word_count":    float64(len(puzzle.Placements)),
		"bonus_word_count":    float64(len(puzzle.BonusWords)),
		"allowed_word_count":  float64(len(puzzle.AllowedWord)),
		"min_word_length":     float64(minLength),
		"max_word_length":     float64(maxLength),
		"average_word_length": averageLength,
		"crossing_count":      float64(crossingCount(puzzle.Placements)),
		"repeated_letters":    float64(repeatedLetters),
		"letter_entropy":      entropy,
		"board_width":         float64(board.Width),
		"board_height":        float64(board.Height),
		"occupied_cell_count": float64(occupiedCellCount(board)),
	}
}

func wordLengthStats(placements []WordPlacement) (minLength, maxLength int, average float64) {
	if len(placements) == 0 {
		return 0, 0, 0
	}

	minLength = len(placements[0].Text)
	total := 0
	for _, placement := range placements {
		length := len(placement.Text)
		minLength = min(minLength, length)
		maxLength = max(maxLength, length)
		total += length
	}
	return minLength, maxLength, float64(total) / float64(len(placements))
}

func bankLetterStats(bank string) (repeatedLetters int, entropy float64) {
	counts := make(map[rune]int)
	for _, letter := range bank {
		counts[letter]++
	}
	for _, count := range counts {
		if count > 1 {
			repeatedLetters += count - 1
		}
		probability := float64(count) / float64(len([]rune(bank)))
		entropy -= probability * math.Log2(probability)
	}
	return repeatedLetters, entropy
}

func crossingCount(placements []WordPlacement) int {
	counts := make(map[Position]int)
	for _, placement := range placements {
		for _, pos := range placement.Positions() {
			counts[pos]++
		}
	}

	crossings := 0
	for _, count := range counts {
		if count > 1 {
			crossings++
		}
	}
	return crossings
}

func occupiedCellCount(board board) int {
	count := 0
	for y := range board.Cells {
		for x := range board.Cells[y] {
			if board.Cells[y][x].Occupied {
				count++
			}
		}
	}
	return count
}

func normalizedMetric(value, low, high float64) float64 {
	if value <= low {
		return 0
	}
	if value >= high {
		return 1
	}
	return (value - low) / (high - low)
}
