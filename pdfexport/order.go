package pdfexport

import (
	"hash/fnv"
	"math/rand/v2"
	"sort"
	"strings"
	"time"
)

func OrderPuzzlesForPrint(puzzles []Puzzle, seed string) []Puzzle {
	ordered := append([]Puzzle(nil), puzzles...)
	if len(ordered) <= 1 {
		return ordered
	}

	sort.SliceStable(ordered, func(i, j int) bool {
		if ordered[i].DifficultyScore != ordered[j].DifficultyScore {
			return ordered[i].DifficultyScore < ordered[j].DifficultyScore
		}
		if c := strings.Compare(normalizeToken(ordered[i].Category), normalizeToken(ordered[j].Category)); c != 0 {
			return c < 0
		}
		if c := strings.Compare(normalizeToken(ordered[i].ModeSelection), normalizeToken(ordered[j].ModeSelection)); c != 0 {
			return c < 0
		}
		if c := strings.Compare(ordered[i].SourceFileName, ordered[j].SourceFileName); c != 0 {
			return c < 0
		}
		return ordered[i].Index < ordered[j].Index
	})

	rng := seededRand(seed)

	const bandSize = 6
	for start := 0; start < len(ordered); start += bandSize {
		end := min(start+bandSize, len(ordered))
		shuffleBand(ordered[start:end], rng)
	}

	// Reduce same-category runs across band edges while preserving
	// the overall difficulty trajectory.
	for i := 1; i < len(ordered); i++ {
		if !sameCategory(ordered[i-1], ordered[i]) {
			continue
		}
		for j := i + 1; j < min(i+4, len(ordered)); j++ {
			if sameCategory(ordered[i-1], ordered[j]) {
				continue
			}
			ordered[i], ordered[j] = ordered[j], ordered[i]
			break
		}
	}

	return ordered
}

func shuffleBand(puzzles []Puzzle, rng *rand.Rand) {
	if len(puzzles) <= 1 {
		return
	}

	perm := rng.Perm(len(puzzles))
	shuffled := make([]Puzzle, len(puzzles))
	for i, idx := range perm {
		shuffled[i] = puzzles[idx]
	}
	copy(puzzles, shuffled)

	for i := 1; i < len(puzzles); i++ {
		if !sameCategory(puzzles[i-1], puzzles[i]) {
			continue
		}
		for j := i + 1; j < len(puzzles); j++ {
			if sameCategory(puzzles[i-1], puzzles[j]) {
				continue
			}
			puzzles[i], puzzles[j] = puzzles[j], puzzles[i]
			break
		}
	}
}

func seededRand(seed string) *rand.Rand {
	if strings.TrimSpace(seed) == "" {
		seed = time.Now().Format(time.RFC3339Nano)
	}
	h := fnv.New64a()
	h.Write([]byte(seed))
	s := h.Sum64()
	return rand.New(rand.NewPCG(s, ^s))
}

func sameCategory(a, b Puzzle) bool {
	return normalizeToken(a.Category) == normalizeToken(b.Category)
}

func normalizeToken(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	s = strings.ReplaceAll(s, "-", " ")
	s = strings.ReplaceAll(s, "_", " ")
	return strings.Join(strings.Fields(s), " ")
}
