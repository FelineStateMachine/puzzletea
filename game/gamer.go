// Package game defines the plugin interface for puzzle games.
package game

import (
	"math/rand/v2"

	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/list"
	tea "charm.land/bubbletea/v2"
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

// BaseMode provides the common title/description fields and the three
// Mode-interface methods so that concrete mode structs can embed it
// instead of duplicating boilerplate.
type BaseMode struct {
	title       string
	description string
}

// NewBaseMode creates a BaseMode with the given title and description.
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

// SpawnCompleteMsg is sent when an async Spawn() call finishes.
type SpawnCompleteMsg struct {
	Game Gamer
	Err  error
}

// HelpToggleMsg is sent from the root model to games when the user toggles
// the full help display with Ctrl+H.
type HelpToggleMsg struct{ Show bool }

// Registry maps game type names to their import functions.
var Registry = map[string]func([]byte) (Gamer, error){}

// Register adds an import function for a game type to the registry.
func Register(name string, fn func([]byte) (Gamer, error)) {
	Registry[name] = fn
}
