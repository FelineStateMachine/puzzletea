package wordsearch

import (
	"math/rand/v2"
	"strings"
)

// GenerateWordSearch creates a new word search grid with the specified parameters
func GenerateWordSearch(width, height, wordCount, minLen, maxLen int, allowedDirs []Direction) (grid, []Word) {
	g := createEmptyGrid(height, width)
	words := selectWords(wordCount, minLen, maxLen)
	placedWords := make([]Word, 0, len(words))

	// Try to place each word
	for _, wordText := range words {
		word := tryPlaceWord(g, wordText, allowedDirs, 100)
		if word != nil {
			placedWords = append(placedWords, *word)
		}
	}

	// Fill remaining empty cells with random letters
	fillEmptyCells(g)

	return g, placedWords
}

// selectWords randomly selects words from the word list
func selectWords(count, minLen, maxLen int) []string {
	// Collect all valid words
	var candidates []string
	for length := minLen; length <= maxLen; length++ {
		if words, ok := wordsByLength[length]; ok {
			candidates = append(candidates, words...)
		}
	}

	// Shuffle and take first 'count' words
	rand.Shuffle(len(candidates), func(i, j int) {
		candidates[i], candidates[j] = candidates[j], candidates[i]
	})

	if count > len(candidates) {
		count = len(candidates)
	}

	return candidates[:count]
}

// tryPlaceWord attempts to place a word in the grid
func tryPlaceWord(g grid, text string, allowedDirs []Direction, maxAttempts int) *Word {
	text = strings.ToUpper(text)
	height := len(g)
	width := len(g[0])

	for range maxAttempts {
		// Random starting position
		x := rand.IntN(width)
		y := rand.IntN(height)

		// Random direction
		dir := allowedDirs[rand.IntN(len(allowedDirs))]
		dx, dy := dir.Vector()

		// Check if word fits in this direction
		endX := x + dx*(len(text)-1)
		endY := y + dy*(len(text)-1)

		if endX < 0 || endX >= width || endY < 0 || endY >= height {
			continue
		}

		// Check if placement is valid (no conflicts, allow letter overlap)
		valid := true
		for i := range len(text) {
			cx := x + dx*i
			cy := y + dy*i
			existing := g.Get(cx, cy)

			if existing != ' ' && existing != rune(text[i]) {
				valid = false
				break
			}
		}

		if !valid {
			continue
		}

		// Place the word
		for i := range len(text) {
			cx := x + dx*i
			cy := y + dy*i
			g.Set(cx, cy, rune(text[i]))
		}

		return &Word{
			Text:      text,
			Start:     Position{X: x, Y: y},
			End:       Position{X: endX, Y: endY},
			Direction: dir,
			Found:     false,
		}
	}

	return nil
}

// fillEmptyCells fills all empty cells with random letters
func fillEmptyCells(g grid) {
	letters := "ABCDEFGHIJKLMNOPQRSTUVWXYZ"

	for y := range g {
		for x := range g[y] {
			if g[y][x] == ' ' {
				g[y][x] = rune(letters[rand.IntN(len(letters))])
			}
		}
	}
}
