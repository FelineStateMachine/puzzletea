package game

import (
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
)

// Gamer is the interface that an active game instance must implement.
type Gamer interface {
	GetDebugInfo() string
	GetFullHelp() [][]key.Binding

	GetSave() ([]byte, error)

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
}

func (c Category) Title() string       { return c.Name }
func (c Category) Description() string { return c.Desc }
func (c Category) FilterValue() string { return c.Name }
