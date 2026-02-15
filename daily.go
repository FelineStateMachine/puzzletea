package main

import (
	"hash/fnv"
	"math/rand/v2"
	"time"

	"github.com/FelineStateMachine/puzzletea/game"
	"github.com/FelineStateMachine/puzzletea/hashiwokakero"
	"github.com/FelineStateMachine/puzzletea/hitori"
	"github.com/FelineStateMachine/puzzletea/lightsout"
	"github.com/FelineStateMachine/puzzletea/namegen"
	"github.com/FelineStateMachine/puzzletea/nonogram"
	"github.com/FelineStateMachine/puzzletea/sudoku"
	"github.com/FelineStateMachine/puzzletea/takuzu"
	"github.com/FelineStateMachine/puzzletea/wordsearch"
)

// dailyEntry pairs a SeededSpawner with metadata for the eligible daily pool.
type dailyEntry struct {
	spawner  game.SeededSpawner
	gameType string
	mode     string
}

// eligibleDailyModes is the curated pool of quick (<10 min) modes for daily rotation.
var eligibleDailyModes = []dailyEntry{
	{nonogram.NewMode("Standard", "10x10 grid, ~67% filled. Classic size, dense hints.", 10, 10, 0.67), "Nonogram", "Standard"},
	{nonogram.NewMode("Classic", "10x10 grid, ~52% filled. The typical nonogram experience.", 10, 10, 0.52), "Nonogram", "Classic"},
	{sudoku.NewMode("Easy", "38\u201344 clues. Naked Singles.", 38), "Sudoku", "Easy"},
	{sudoku.NewMode("Medium", "32\u201337 clues. Hidden Pairs / Pointing.", 32), "Sudoku", "Medium"},
	{takuzu.NewMode("Medium", "8\u00d78 grid, ~40% clues. Larger grid, moderate deduction.", 8, 0.40), "Takuzu", "Medium"},
	{takuzu.NewMode("Tricky", "10\u00d710 grid, ~38% clues. Uniqueness rule needed.", 10, 0.38), "Takuzu", "Tricky"},
	{hashiwokakero.NewMode("Easy 9x9", "9x9 grid with 12-16 islands.", 9, 9, 12, 16), "Hashiwokakero", "Easy 9x9"},
	{hashiwokakero.NewMode("Medium 7x7", "7x7 grid with 12-15 islands.", 7, 7, 12, 15), "Hashiwokakero", "Medium 7x7"},
	{hitori.NewMode("Easy", "6\u00d76 grid, straightforward logic.", 6, 0.32), "Hitori", "Easy"},
	{hitori.NewMode("Medium", "8\u00d78 grid, moderate challenge.", 8, 0.30), "Hitori", "Medium"},
	{lightsout.NewMode("Medium", "5x5 grid", 5, 5), "Lights Out", "Medium"},
	{lightsout.NewMode("Hard", "7x7 grid", 7, 7), "Lights Out", "Hard"},
	{wordsearch.NewMode("Easy 10x10", "Find 6 words in a 10x10 grid.", 10, 10, 6, 3, 5, []wordsearch.Direction{wordsearch.Right, wordsearch.Down, wordsearch.DownRight}), "Word Search", "Easy 10x10"},
}

// dailySeed returns a deterministic int64 seed derived from the date.
func dailySeed(date time.Time) uint64 {
	dateStr := date.Format("2006-01-02")
	h := fnv.New64a()
	h.Write([]byte(dateStr))
	return h.Sum64()
}

// dailyRNG creates a deterministic RNG seeded from the given date.
func dailyRNG(date time.Time) *rand.Rand {
	seed := dailySeed(date)
	return rand.New(rand.NewPCG(seed, seed))
}

// dailyName generates the daily puzzle name in the format:
// "Daily Feb 14 26 - amber-falcon"
func dailyName(date time.Time, rng *rand.Rand) string {
	return "Daily " + date.Format("Jan _2 06") + " - " + namegen.GenerateSeeded(rng)
}

// dailyMode selects the daily mode from the eligible pool using the seeded RNG.
// Returns the spawner, game type name, and mode title.
func dailyMode(rng *rand.Rand) (game.SeededSpawner, string, string) {
	entry := eligibleDailyModes[rng.IntN(len(eligibleDailyModes))]
	return entry.spawner, entry.gameType, entry.mode
}
