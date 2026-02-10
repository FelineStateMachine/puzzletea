// Package hashiwokakero implements the bridge-connecting puzzle game.
package hashiwokakero

import (
	"fmt"

	"github.com/FelineStateMachine/puzzletea/game"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var _ game.Gamer = Model{}

type Model struct {
	puzzle         Puzzle
	cursorIsland   int  // ID of island the cursor is on
	selectedIsland *int // ID of island selected for bridge building, nil if none
	keys           KeyMap
	modeTitle      string
	showFullHelp   bool
}

func New(mode HashiMode, puzzle Puzzle) game.Gamer {
	cursorID := 0
	if len(puzzle.Islands) > 0 {
		cursorID = puzzle.Islands[0].ID
	}
	return Model{
		puzzle:       puzzle,
		cursorIsland: cursorID,
		keys:         DefaultKeyMap,
		modeTitle:    mode.Title(),
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
		if m.selectedIsland != nil {
			m = m.handleBridgeMode(msg)
		} else {
			m = m.handleNavMode(msg)
		}
	}
	return m, nil
}

func (m Model) handleNavMode(msg tea.KeyMsg) Model {
	switch {
	case key.Matches(msg, m.keys.Up):
		m = m.moveCursor(0, -1)
	case key.Matches(msg, m.keys.Down):
		m = m.moveCursor(0, 1)
	case key.Matches(msg, m.keys.Left):
		m = m.moveCursor(-1, 0)
	case key.Matches(msg, m.keys.Right):
		m = m.moveCursor(1, 0)
	case key.Matches(msg, m.keys.Select):
		id := m.cursorIsland
		m.selectedIsland = &id
	}
	return m
}

func (m Model) handleBridgeMode(msg tea.KeyMsg) Model {
	switch {
	case key.Matches(msg, m.keys.Up):
		m = m.cycleBridge(0, -1)
	case key.Matches(msg, m.keys.Down):
		m = m.cycleBridge(0, 1)
	case key.Matches(msg, m.keys.Left):
		m = m.cycleBridge(-1, 0)
	case key.Matches(msg, m.keys.Right):
		m = m.cycleBridge(1, 0)
	case key.Matches(msg, m.keys.Select), key.Matches(msg, m.keys.Cancel):
		m.selectedIsland = nil
	}
	return m
}

// moveCursor moves the cursor to the nearest island in direction (dx, dy).
func (m Model) moveCursor(dx, dy int) Model {
	adj := m.puzzle.FindAdjacentIsland(m.cursorIsland, dx, dy)
	if adj != nil {
		m.cursorIsland = adj.ID
	}
	return m
}

// cycleBridge cycles the bridge from the selected island toward (dx, dy): 0→1→2→0.
func (m Model) cycleBridge(dx, dy int) Model {
	if m.selectedIsland == nil {
		return m
	}

	adj := m.puzzle.FindAdjacentIsland(*m.selectedIsland, dx, dy)
	if adj == nil {
		return m
	}

	existing := m.puzzle.GetBridge(*m.selectedIsland, adj.ID)
	currentCount := 0
	if existing != nil {
		currentCount = existing.Count
	}

	newCount := (currentCount + 1) % 3

	if newCount > 0 && currentCount == 0 {
		// Adding a new bridge — check for crossings
		if m.puzzle.WouldCross(*m.selectedIsland, adj.ID) {
			return m
		}
	}

	m.puzzle.SetBridge(*m.selectedIsland, adj.ID, newCount)
	return m
}

func (m Model) View() string {
	solved := m.puzzle.IsSolved()

	title := game.TitleBarView("Hashiwokakero", m.modeTitle, solved)
	grid := gridView(m, solved)
	info := infoView(&m.puzzle)
	status := statusBarView(m.selectedIsland != nil, m.showFullHelp)

	return lipgloss.JoinVertical(lipgloss.Left, title, grid, info, status)
}

func (m Model) SetTitle(t string) game.Gamer {
	m.modeTitle = t
	return m
}

func (m Model) IsSolved() bool {
	return m.puzzle.IsSolved()
}

func (m Model) GetDebugInfo() string {
	solved := m.puzzle.IsSolved()
	connected := m.puzzle.IsConnected()

	status := "In Progress"
	if solved {
		status = "Solved"
	}

	connStr := "No"
	if connected {
		connStr = "Yes"
	}

	selectedStr := "None"
	if m.selectedIsland != nil {
		selectedStr = fmt.Sprintf("%d", *m.selectedIsland)
	}

	s := game.DebugHeader("Hashiwokakero", [][2]string{
		{"Status", status},
		{"Grid Size", fmt.Sprintf("%dx%d", m.puzzle.Width, m.puzzle.Height)},
		{"Islands", fmt.Sprintf("%d", len(m.puzzle.Islands))},
		{"Bridges", fmt.Sprintf("%d", len(m.puzzle.Bridges))},
		{"Connected", connStr},
		{"Cursor Island", fmt.Sprintf("%d", m.cursorIsland)},
		{"Selected Island", selectedStr},
	})

	var rows [][]string
	for _, isl := range m.puzzle.Islands {
		current := m.puzzle.BridgeCount(isl.ID)
		islStatus := "Unsatisfied"
		if current == isl.Required {
			islStatus = "Satisfied"
		} else if current > isl.Required {
			islStatus = "Over"
		}
		rows = append(rows, []string{
			fmt.Sprintf("%d", isl.ID),
			fmt.Sprintf("(%d,%d)", isl.X, isl.Y),
			fmt.Sprintf("%d", isl.Required),
			fmt.Sprintf("%d", current),
			islStatus,
		})
	}
	s += game.DebugTable("Islands", []string{"ID", "Pos", "Required", "Current", "Status"}, rows)

	return s
}
