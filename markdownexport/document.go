package markdownexport

import (
	"fmt"
	"hash/fnv"
	"math/rand/v2"
	"strings"
	"time"

	"github.com/FelineStateMachine/puzzletea/namegen"
)

type DocumentConfig struct {
	Version       string
	Category      string
	ModeSelection string
	Count         int
	Seed          string
	GeneratedAt   time.Time
}

type PuzzleSection struct {
	Index    int
	GameType string
	Mode     string
	Body     string
}

func BuildDocument(cfg DocumentConfig, puzzles []PuzzleSection) string {
	var b strings.Builder

	seed := cfg.Seed
	if strings.TrimSpace(seed) == "" {
		seed = "none"
	}
	nameRNG := exportNameRNG(cfg)

	fmt.Fprintf(&b, "# PuzzleTea Export\n\n")
	fmt.Fprintf(&b, "- Generated: %s\n", cfg.GeneratedAt.Format(time.RFC3339))
	fmt.Fprintf(&b, "- Version: %s\n", cfg.Version)
	fmt.Fprintf(&b, "- Category: %s\n", cfg.Category)
	fmt.Fprintf(&b, "- Mode Selection: %s\n", cfg.ModeSelection)
	fmt.Fprintf(&b, "- Count: %d\n", cfg.Count)
	fmt.Fprintf(&b, "- Seed: %s\n\n", seed)

	for i, puzzle := range puzzles {
		if i > 0 {
			b.WriteString("\n---\n\n")
		}
		fmt.Fprintf(&b, "## %s - %d\n\n", namegen.GenerateSeeded(nameRNG), puzzle.Index)
		b.WriteString(strings.TrimSpace(puzzle.Body))
		b.WriteString("\n")
	}

	return b.String()
}

func exportNameRNG(cfg DocumentConfig) *rand.Rand {
	seed := cfg.Seed
	if strings.TrimSpace(seed) == "" {
		seed = cfg.GeneratedAt.Format(time.RFC3339Nano)
	}
	return rngFromString("export-names:" + seed)
}

func rngFromString(seed string) *rand.Rand {
	h := fnv.New64a()
	h.Write([]byte(seed))
	s := h.Sum64()
	return rand.New(rand.NewPCG(s, ^s))
}
