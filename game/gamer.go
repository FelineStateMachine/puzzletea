package game

import (
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
)

type Gamer interface {
	GetDebugInfo() string
	GetFullHelp() [][]key.Binding

	GetSave() ([]byte, error)

	Init() tea.Cmd
	View() string
	Update(msg tea.Msg) (Gamer, tea.Cmd)
}

type Mode interface {
	Title() string
	Description() string
	FilterValue() string
}
