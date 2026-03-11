package spellpuzzle

import (
	_ "embed"
	"strings"
)

//go:embed twl06.txt
var rawLexicon string

type letterCounts [26]uint8

type lexiconWord struct {
	Text   string
	Counts letterCounts
}

var (
	allWords          []lexiconWord
	wordsByLength     map[int][]lexiconWord
	exactWordSet      map[string]struct{}
	seedWordsByLength map[int][]string
)

func init() {
	allWords = make([]lexiconWord, 0, 96000)
	wordsByLength = map[int][]lexiconWord{}
	exactWordSet = map[string]struct{}{}
	seedWordsByLength = map[int][]string{}

	lines := strings.Split(rawLexicon, "\n")
	for _, line := range lines {
		word := strings.ToUpper(strings.TrimSpace(line))
		if word == "" {
			continue
		}
		if !isAlphaWord(word) {
			continue
		}
		entry := lexiconWord{
			Text:   word,
			Counts: countLetters(word),
		}
		allWords = append(allWords, entry)
		wordsByLength[len(word)] = append(wordsByLength[len(word)], entry)
		exactWordSet[word] = struct{}{}
		if len(word) >= 3 && len(word) <= 9 {
			seedWordsByLength[len(word)] = append(seedWordsByLength[len(word)], word)
		}
	}
}

func isAlphaWord(word string) bool {
	for _, r := range word {
		if r < 'A' || r > 'Z' {
			return false
		}
	}
	return true
}

func countLetters(word string) letterCounts {
	var counts letterCounts
	for _, r := range word {
		counts[r-'A']++
	}
	return counts
}

func canSpellWord(word string, bankCounts letterCounts) bool {
	counts := countLetters(word)
	for i, want := range counts {
		if want > bankCounts[i] {
			return false
		}
	}
	return true
}

func spellableWords(bank string, minLen int) []string {
	counts := countLetters(bank)
	result := make([]string, 0, 64)
	seen := make(map[string]struct{})
	for _, word := range allWords {
		if len(word.Text) < minLen || len(word.Text) > len(bank) {
			continue
		}
		if !canSpellEntry(word, counts) {
			continue
		}
		if _, ok := seen[word.Text]; ok {
			continue
		}
		seen[word.Text] = struct{}{}
		result = append(result, word.Text)
	}
	return result
}

func canSpellEntry(word lexiconWord, counts letterCounts) bool {
	for i, want := range word.Counts {
		if want > counts[i] {
			return false
		}
	}
	return true
}

func isExactWord(word string) bool {
	_, ok := exactWordSet[word]
	return ok
}
