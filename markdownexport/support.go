package markdownexport

import (
	"errors"
	"strings"
)

var ErrUnsupportedGame = errors.New("game does not support markdown export")

func SupportsGameType(gameType string) bool {
	switch normalizeGameType(gameType) {
	case "hashiwokakero", "hitori", "nonogram", "nurikabe", "shikaku", "sudoku", "takuzu", "word search", "wordsearch":
		return true
	case "lights out", "lights-out", "lightsout", "lights":
		return false
	default:
		return false
	}
}

func normalizeGameType(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	s = strings.ReplaceAll(s, "-", " ")
	s = strings.ReplaceAll(s, "_", " ")
	return strings.Join(strings.Fields(s), " ")
}
