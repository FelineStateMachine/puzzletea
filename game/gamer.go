// Package game defines the plugin interface for puzzle games.
package game

import (
	"context"
	"math/rand/v2"

	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"github.com/FelineStateMachine/puzzletea/difficulty"
	"github.com/FelineStateMachine/puzzletea/puzzle"
)

// Gamer is the interface that an active game instance must implement.
type Gamer interface {
	GetDebugInfo() string
	GetFullHelp() [][]key.Binding

	GetSave() ([]byte, error)
	IsSolved() bool
	Reset() Gamer
	SetTitle(string) Gamer

	Init() tea.Cmd
	View() string
	Update(msg tea.Msg) (Gamer, tea.Cmd)
}

// Mode describes a game mode for menu display.
type Mode interface {
	Title() string
	Description() string
	FilterValue() string
}

// Spawner creates a new game instance. Implemented by concrete mode types
// but kept separate from Mode so that Category (which also satisfies Mode)
// does not need a Spawn method.
type Spawner interface {
	Spawn() (Gamer, error)
}

// SeededSpawner creates a new game instance using a deterministic RNG.
// This enables reproducible puzzle generation for daily puzzles.
type SeededSpawner interface {
	Spawner
	SpawnSeeded(rng *rand.Rand) (Gamer, error)
}

// CancellableSpawner optionally supports context-aware generation so callers
// can cancel long-running spawn work.
type CancellableSpawner interface {
	SpawnContext(ctx context.Context) (Gamer, error)
}

// CancellableSeededSpawner optionally supports context-aware deterministic
// generation so callers can cancel long-running seeded spawn work.
type CancellableSeededSpawner interface {
	SeededSpawner
	SpawnSeededContext(ctx context.Context, rng *rand.Rand) (Gamer, error)
}

// EloSpawner creates a deterministic game instance from seed text and Elo.
type EloSpawner interface {
	SpawnElo(seed string, elo difficulty.Elo) (Gamer, difficulty.Report, error)
}

// CancellableEloSpawner optionally supports context-aware Elo generation.
type CancellableEloSpawner interface {
	EloSpawner
	SpawnEloContext(ctx context.Context, seed string, elo difficulty.Elo) (Gamer, difficulty.Report, error)
}

// BaseMode provides the common title/description fields and the three
// Mode-interface methods so that concrete mode structs can embed it
// instead of duplicating boilerplate.
type BaseMode struct {
	title       string
	description string
}

func NewBaseMode(title, description string) BaseMode {
	return BaseMode{title: title, description: description}
}

func (b BaseMode) Title() string       { return b.title }
func (b BaseMode) Description() string { return b.description }
func (b BaseMode) FilterValue() string { return b.title + " " + b.description }

func AdaptImport[T Gamer](fn func([]byte) (T, error)) func([]byte) (Gamer, error) {
	return func(data []byte) (Gamer, error) {
		return fn(data)
	}
}

// NormalizeName lowercases and replaces hyphens/underscores with spaces for
// fuzzy matching of CLI arguments to game names and aliases.
func NormalizeName(s string) string {
	return puzzle.NormalizeName(s)
}

// SpawnCompleteMsg is sent when an async Spawn() call finishes.
type SpawnCompleteMsg struct {
	Game   Gamer
	Report difficulty.Report
	Err    error
}

// HelpToggleMsg is sent from the root model to games when the user toggles
// the full help display with Ctrl+H.
type HelpToggleMsg struct{ Show bool }
