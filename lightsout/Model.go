// Package lightsout implements the lights out toggle puzzle game.
package lightsout

import (
	"fmt"
	"math/rand/v2"

	"github.com/FelineStateMachine/puzzletea/game"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type Model struct {
	grid         [][]bool
	initialGrid  [][]bool
	width        int
	height       int
	cursor       game.Cursor
	keys         KeyMap
	modeTitle    string
	showFullHelp bool
}

var _ game.Gamer = Model{}

func New(w, h int) (Model, error) {
	grid := Generate(w, h)
	return newFromGrid(w, h, grid), nil
}

func NewSeeded(w, h int, rng *rand.Rand) (Model, error) {
	grid := GenerateSeeded(w, h, rng)
	return newFromGrid(w, h, grid), nil
}

func newFromGrid(w, h int, grid [][]bool) Model {
	initial := make([][]bool, h)
	for y := range h {
		initial[y] = make([]bool, w)
		copy(initial[y], grid[y])
	}
	return Model{
		grid:        grid,
		initialGrid: initial,
		width:       w,
		height:      h,
		cursor:      game.Cursor{X: w / 2, Y: h / 2},
		keys:        DefaultKeyMap,
	}
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (game.Gamer, tea.Cmd) {
	switch msg := msg.(type) {
	case game.HelpToggleMsg:
		m.showFullHelp = msg.Show
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keys.Toggle):
			if !m.IsSolved() {
				Toggle(m.grid, m.cursor.X, m.cursor.Y)
			}
		default:
			m.cursor.Move(m.keys.CursorKeyMap, msg, m.width-1, m.height-1)
		}
	}
	m.updateKeyBindings()
	return m, nil
}

func (m Model) View() string {
	solved := m.IsSolved()

	title := game.TitleBarView("Lights Out", m.modeTitle, solved)
	grid := gridView(m.grid, m.cursor, solved)
	status := statusBarView(m.showFullHelp)

	return lipgloss.JoinVertical(lipgloss.Center, title, grid, status)
}

func (m Model) SetTitle(t string) game.Gamer {
	m.modeTitle = t
	return m
}

func (m Model) IsSolved() bool {
	return IsSolved(m.grid)
}

func (m Model) Reset() game.Gamer {
	for y := range m.height {
		copy(m.grid[y], m.initialGrid[y])
	}
	m.cursor = game.Cursor{X: m.width / 2, Y: m.height / 2}
	return m
}

func (m Model) GetDebugInfo() string {
	status := "In Progress"
	if m.IsSolved() {
		status = "Solved"
	}

	lightsOn := 0
	for _, row := range m.grid {
		for _, cell := range row {
			if cell {
				lightsOn++
			}
		}
	}

	return game.DebugHeader("Lights Out", [][2]string{
		{"Status", status},
		{"Cursor", fmt.Sprintf("(%d, %d)", m.cursor.X, m.cursor.Y)},
		{"Grid Size", fmt.Sprintf("%d x %d", m.width, m.height)},
		{"Lights On", fmt.Sprintf("%d / %d", lightsOn, m.width*m.height)},
	})
}
