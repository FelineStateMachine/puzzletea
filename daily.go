package main

import (
	"fmt"
	"math/rand/v2"
	"time"

	"github.com/FelineStateMachine/puzzletea/game"
	"github.com/FelineStateMachine/puzzletea/namegen"
)

// dailyRNG returns a deterministic RNG seeded from the given date.
// The same calendar day always produces the same seed, ensuring
// reproducible daily puzzles regardless of the time of day.
func dailyRNG(today time.Time) *rand.Rand {
	y, m, d := today.Date()
	seed := uint64(y)*10000 + uint64(m)*100 + uint64(d)
	return rand.New(rand.NewPCG(seed, seed))
}

// dailyName builds the unique identifier for today's daily puzzle.
// Format: "Daily <Mon DD YY> <adjective-noun>"
func dailyName(today time.Time, rng *rand.Rand) string {
	datePart := today.Format("Jan _2 06")
	namePart := namegen.GenerateSeeded(rng)
	return fmt.Sprintf("Daily %s %s", datePart, namePart)
}

// dailyMode picks a random game category and mode using the provided RNG.
// Returns the SeededSpawner, the canonical game type name, and the mode title.
func dailyMode(rng *rand.Rand) (game.SeededSpawner, string, string) {
	cat := GameCategories[rng.IntN(len(GameCategories))].(game.Category)
	mode := cat.Modes[rng.IntN(len(cat.Modes))].(game.Mode)
	return mode.(game.SeededSpawner), cat.Name, mode.Title()
}
