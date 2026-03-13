// Package game defines the plugin interface for puzzle games.
package game

import (
	"context"
	"math/rand/v2"

	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/list"
	tea "charm.land/bubbletea/v2"
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

// Category groups related game modes under a heading in the menu.
type Category struct {
	Name  string
	Desc  string
	Modes []list.Item
	Help  string // embedded help.md content rendered in "How to Play"
}

func (c Category) Title() string       { return c.Name }
func (c Category) Description() string { return c.Desc }
func (c Category) FilterValue() string { return c.Name }

// Definition is the package-level metadata for a puzzle game.
type Definition struct {
	Name        string
	Description string
	Aliases     []string
	Modes       []Mode
	DailyModes  []Mode
	Help        string
	Import      func([]byte) (Gamer, error)
}

type DefinitionSpec struct {
	Name             string
	Description      string
	Aliases          []string
	Modes            []Mode
	DailyModeIndexes []int
	Help             string
	Import           func([]byte) (Gamer, error)
}

func NewDefinition(spec DefinitionSpec) Definition {
	return Definition{
		Name:        spec.Name,
		Description: spec.Description,
		Aliases:     append([]string(nil), spec.Aliases...),
		Modes:       append([]Mode(nil), spec.Modes...),
		DailyModes:  SelectModesByIndex(spec.Modes, spec.DailyModeIndexes...),
		Help:        spec.Help,
		Import:      spec.Import,
	}
}

func SelectModesByIndex(modes []Mode, indexes ...int) []Mode {
	selected := make([]Mode, 0, len(indexes))
	for _, idx := range indexes {
		if idx < 0 || idx >= len(modes) {
			continue
		}
		selected = append(selected, modes[idx])
	}
	return selected
}

func AdaptImport[T Gamer](fn func([]byte) (T, error)) func([]byte) (Gamer, error) {
	return func(data []byte) (Gamer, error) {
		return fn(data)
	}
}

func (d Definition) Category() Category {
	modes := make([]list.Item, len(d.Modes))
	for i, m := range d.Modes {
		modes[i] = m
	}
	return Category{
		Name:  d.Name,
		Desc:  d.Description,
		Modes: modes,
		Help:  d.Help,
	}
}

// NormalizeName lowercases and replaces hyphens/underscores with spaces for
// fuzzy matching of CLI arguments to game names and aliases.
func NormalizeName(s string) string {
	return puzzle.NormalizeName(s)
}

// SpawnCompleteMsg is sent when an async Spawn() call finishes.
type SpawnCompleteMsg struct {
	Game Gamer
	Err  error
}

// HelpToggleMsg is sent from the root model to games when the user toggles
// the full help display with Ctrl+H.
type HelpToggleMsg struct{ Show bool }
