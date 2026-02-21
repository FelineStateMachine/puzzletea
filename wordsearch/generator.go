package wordsearch

import (
	"math/rand/v2"
	"sort"
	"strings"
)

type placementStats struct {
	TargetWords      int
	PlacedWords      int
	FailedWords      int
	RandomAttempts   int
	FallbackAttempts int
	FallbackUsed     int
	FallbackPlaced   int
}

func (s placementStats) SuccessRate() float64 {
	if s.TargetWords == 0 {
		return 1
	}
	return float64(s.PlacedWords) / float64(s.TargetWords)
}

func (s placementStats) AttemptsPerWord() float64 {
	if s.TargetWords == 0 {
		return 0
	}
	attempts := s.RandomAttempts + s.FallbackAttempts
	return float64(attempts) / float64(s.TargetWords)
}

type placementAttemptStats struct {
	RandomAttempts   int
	FallbackAttempts int
	UsedFallback     bool
}

type wordBucketSampler struct {
	words     []string
	remaining int
	swaps     map[int]int
}

func (s *wordBucketSampler) resolve(pos int) int {
	if s.swaps == nil {
		return pos
	}
	if mapped, ok := s.swaps[pos]; ok {
		return mapped
	}
	return pos
}

func (s *wordBucketSampler) draw(rng *rand.Rand) string {
	drawPos := rng.IntN(s.remaining)
	wordIndex := s.resolve(drawPos)

	lastPos := s.remaining - 1
	lastIndex := s.resolve(lastPos)

	if drawPos != lastPos {
		if s.swaps == nil {
			s.swaps = map[int]int{}
		}
		s.swaps[drawPos] = lastIndex
	}

	if s.swaps != nil {
		delete(s.swaps, lastPos)
	}

	s.remaining--
	return s.words[wordIndex]
}

func GenerateWordSearch(width, height, wordCount, minLen, maxLen int, allowedDirs []Direction) (grid, []Word) {
	rng := rand.New(rand.NewPCG(rand.Uint64(), rand.Uint64()))
	return GenerateWordSearchSeeded(width, height, wordCount, minLen, maxLen, allowedDirs, rng)
}

func GenerateWordSearchSeeded(width, height, wordCount, minLen, maxLen int, allowedDirs []Direction, rng *rand.Rand) (grid, []Word) {
	g, placedWords, _ := generateWordSearchSeededWithStats(
		width,
		height,
		wordCount,
		minLen,
		maxLen,
		allowedDirs,
		rng,
	)
	return g, placedWords
}

func generateWordSearchSeededWithStats(
	width,
	height,
	wordCount,
	minLen,
	maxLen int,
	allowedDirs []Direction,
	rng *rand.Rand,
) (grid, []Word, placementStats) {
	g := createEmptyGrid(height, width)
	words := selectWordsSeeded(wordCount, minLen, maxLen, rng)
	orderWordsForPlacement(words)

	placedWords := make([]Word, 0, len(words))
	stats := placementStats{TargetWords: len(words)}

	for _, wordText := range words {
		word, attemptStats := tryPlaceWordSeededWithFallback(g, wordText, allowedDirs, 100, rng)

		stats.RandomAttempts += attemptStats.RandomAttempts
		stats.FallbackAttempts += attemptStats.FallbackAttempts
		if attemptStats.UsedFallback {
			stats.FallbackUsed++
		}

		if word == nil {
			stats.FailedWords++
			continue
		}

		stats.PlacedWords++
		if attemptStats.UsedFallback {
			stats.FallbackPlaced++
		}
		placedWords = append(placedWords, *word)
	}

	fillEmptyCellsSeeded(g, rng)

	return g, placedWords, stats
}

func selectWords(count, minLen, maxLen int) []string {
	rng := rand.New(rand.NewPCG(rand.Uint64(), rand.Uint64()))
	return selectWordsSeeded(count, minLen, maxLen, rng)
}

func selectWordsSeeded(count, minLen, maxLen int, rng *rand.Rand) []string {
	if count <= 0 || minLen > maxLen {
		return nil
	}

	buckets := make([]wordBucketSampler, 0, maxLen-minLen+1)
	totalAvailable := 0

	for length := minLen; length <= maxLen; length++ {
		words, ok := wordsByLength[length]
		if !ok || len(words) == 0 {
			continue
		}

		buckets = append(buckets, wordBucketSampler{
			words:     words,
			remaining: len(words),
		})
		totalAvailable += len(words)
	}

	if totalAvailable == 0 {
		return nil
	}

	if count > totalAvailable {
		count = totalAvailable
	}

	selected := make([]string, 0, count)
	for range count {
		bucketIndex := pickBucketByRemaining(buckets, totalAvailable, rng)
		selected = append(selected, buckets[bucketIndex].draw(rng))
		totalAvailable--
	}

	return selected
}

func pickBucketByRemaining(buckets []wordBucketSampler, totalRemaining int, rng *rand.Rand) int {
	randomOffset := rng.IntN(totalRemaining)

	for i := range buckets {
		if randomOffset < buckets[i].remaining {
			return i
		}
		randomOffset -= buckets[i].remaining
	}

	return len(buckets) - 1
}

func orderWordsForPlacement(words []string) {
	sort.SliceStable(words, func(i, j int) bool {
		return len(words[i]) > len(words[j])
	})
}

func tryPlaceWord(g grid, text string, allowedDirs []Direction, maxAttempts int) *Word {
	rng := rand.New(rand.NewPCG(rand.Uint64(), rand.Uint64()))
	return tryPlaceWordSeeded(g, text, allowedDirs, maxAttempts, rng)
}

func tryPlaceWordSeeded(g grid, text string, allowedDirs []Direction, maxAttempts int, rng *rand.Rand) *Word {
	word, _ := tryPlaceWordSeededWithFallback(g, text, allowedDirs, maxAttempts, rng)
	return word
}

func tryPlaceWordSeededWithFallback(
	g grid,
	text string,
	allowedDirs []Direction,
	maxAttempts int,
	rng *rand.Rand,
) (*Word, placementAttemptStats) {
	word, randomAttempts := tryPlaceWordSeededRandom(g, text, allowedDirs, maxAttempts, rng)
	if word != nil {
		return word, placementAttemptStats{RandomAttempts: randomAttempts}
	}

	fallbackWord, fallbackAttempts := tryPlaceWordSeededFallback(g, text, allowedDirs, rng)
	return fallbackWord, placementAttemptStats{
		RandomAttempts:   randomAttempts,
		FallbackAttempts: fallbackAttempts,
		UsedFallback:     fallbackAttempts > 0,
	}
}

func tryPlaceWordSeededRandom(g grid, text string, allowedDirs []Direction, maxAttempts int, rng *rand.Rand) (*Word, int) {
	text = strings.ToUpper(text)
	if maxAttempts <= 0 || len(text) == 0 || len(allowedDirs) == 0 || len(g) == 0 || len(g[0]) == 0 {
		return nil, 0
	}

	height := len(g)
	width := len(g[0])

	for attempt := 1; attempt <= maxAttempts; attempt++ {
		x := rng.IntN(width)
		y := rng.IntN(height)

		dir := allowedDirs[rng.IntN(len(allowedDirs))]
		dx, dy := dir.Vector()

		if !wordInBounds(x, y, dx, dy, len(text), width, height) {
			continue
		}
		if !canPlaceWordAt(g, text, x, y, dx, dy) {
			continue
		}

		placeWordAt(g, text, x, y, dx, dy)
		return newWord(text, x, y, dx, dy, dir), attempt
	}

	return nil, maxAttempts
}

func tryPlaceWordSeededFallback(g grid, text string, allowedDirs []Direction, rng *rand.Rand) (*Word, int) {
	text = strings.ToUpper(text)
	if len(text) == 0 || len(allowedDirs) == 0 || len(g) == 0 || len(g[0]) == 0 {
		return nil, 0
	}

	height := len(g)
	width := len(g[0])

	dirs := append([]Direction(nil), allowedDirs...)
	rng.Shuffle(len(dirs), func(i, j int) {
		dirs[i], dirs[j] = dirs[j], dirs[i]
	})

	startX := rng.IntN(width)
	startY := rng.IntN(height)

	attempts := 0
	for _, dir := range dirs {
		dx, dy := dir.Vector()
		for yStep := range height {
			y := (startY + yStep) % height
			for xStep := range width {
				x := (startX + xStep) % width
				attempts++

				if !wordInBounds(x, y, dx, dy, len(text), width, height) {
					continue
				}
				if !canPlaceWordAt(g, text, x, y, dx, dy) {
					continue
				}

				placeWordAt(g, text, x, y, dx, dy)
				return newWord(text, x, y, dx, dy, dir), attempts
			}
		}
	}

	return nil, attempts
}

func wordInBounds(x, y, dx, dy, wordLen, width, height int) bool {
	endX := x + dx*(wordLen-1)
	endY := y + dy*(wordLen-1)
	return endX >= 0 && endX < width && endY >= 0 && endY < height
}

func canPlaceWordAt(g grid, text string, x, y, dx, dy int) bool {
	for i := range len(text) {
		cx := x + dx*i
		cy := y + dy*i
		existing := g.Get(cx, cy)
		want := rune(text[i])
		if existing != ' ' && existing != want {
			return false
		}
	}
	return true
}

func placeWordAt(g grid, text string, x, y, dx, dy int) {
	for i := range len(text) {
		cx := x + dx*i
		cy := y + dy*i
		g.Set(cx, cy, rune(text[i]))
	}
}

func newWord(text string, x, y, dx, dy int, dir Direction) *Word {
	endX := x + dx*(len(text)-1)
	endY := y + dy*(len(text)-1)

	return &Word{
		Text:      text,
		Start:     Position{X: x, Y: y},
		End:       Position{X: endX, Y: endY},
		Direction: dir,
		Found:     false,
	}
}

func fillEmptyCells(g grid) {
	rng := rand.New(rand.NewPCG(rand.Uint64(), rand.Uint64()))
	fillEmptyCellsSeeded(g, rng)
}

func fillEmptyCellsSeeded(g grid, rng *rand.Rand) {
	letters := "ABCDEFGHIJKLMNOPQRSTUVWXYZ"

	for y := range g {
		for x := range g[y] {
			if g[y][x] == ' ' {
				g[y][x] = rune(letters[rng.IntN(len(letters))])
			}
		}
	}
}
