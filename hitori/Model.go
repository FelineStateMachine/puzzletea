package hitori

import (
	"fmt"

	"github.com/FelineStateMachine/puzzletea/game"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type Model struct {
	size         int
	grid         grid
	initialGrid  grid
	provided     [][]bool
	cursor       game.Cursor
	solved       bool
	keys         KeyMap
	modeTitle    string
	showFullHelp bool
}

var _ game.Gamer = Model{}

func New(mode HitoriMode, puzzle grid, provided [][]bool) (game.Gamer, error) {
	initial := puzzle.clone()
	m := Model{
		size:        mode.Size,
		grid:        puzzle,
		initialGrid: initial,
		provided:    provided,
		cursor:      game.Cursor{X: mode.Size / 2, Y: mode.Size / 2},
		keys:        DefaultKeyMap,
	}
	m.solved = m.checkSolved()
	return m, nil
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
		case key.Matches(msg, m.keys.Shade):
			if !m.solved && !m.provided[m.cursor.Y][m.cursor.X] {
				if m.grid[m.cursor.Y][m.cursor.X] == shadedCell {
					m.grid[m.cursor.Y][m.cursor.X] = m.initialGrid[m.cursor.Y][m.cursor.X]
				} else {
					m.grid[m.cursor.Y][m.cursor.X] = shadedCell
				}
				m.solved = m.checkSolved()
			}
		case key.Matches(msg, m.keys.Clear):
			if !m.solved && !m.provided[m.cursor.Y][m.cursor.X] {
				m.grid[m.cursor.Y][m.cursor.X] = emptyCell
				m.solved = m.checkSolved()
			}
		default:
			m.cursor.Move(m.keys.CursorKeyMap, msg, m.size-1, m.size-1)
		}
	}
	m.updateKeyBindings()
	return m, nil
}

func (m Model) View() string {
	solved := m.IsSolved()

	title := game.TitleBarView("Hitori", m.modeTitle, solved)
	grid := gridView(m.grid, m.provided, m.cursor, solved)
	status := statusBarView(m.showFullHelp)

	return lipgloss.JoinVertical(lipgloss.Center, title, grid, status)
}

func (m Model) SetTitle(t string) game.Gamer {
	m.modeTitle = t
	return m
}

func (m Model) IsSolved() bool {
	return m.solved
}

func (m Model) Reset() game.Gamer {
	for y := range m.size {
		copy(m.grid[y], m.initialGrid[y])
	}
	m.cursor = game.Cursor{X: m.size / 2, Y: m.size / 2}
	m.solved = m.checkSolved()
	return m
}

func (m Model) GetDebugInfo() string {
	status := "In Progress"
	if m.IsSolved() {
		status = "Solved"
	}

	return game.DebugHeader("Hitori", [][2]string{
		{"Status", status},
		{"Cursor", fmt.Sprintf("(%d, %d)", m.cursor.X, m.cursor.Y)},
		{"Grid Size", fmt.Sprintf("%d x %d", m.size, m.size)},
	})
}

func (m Model) GetFullHelp() [][]key.Binding {
	return [][]key.Binding{
		{m.keys.CursorKeyMap.Up, m.keys.CursorKeyMap.Down, m.keys.CursorKeyMap.Left, m.keys.CursorKeyMap.Right},
		{m.keys.Shade, m.keys.Clear},
	}
}

func (m Model) checkSolved() bool {
	for y := 0; y < m.size; y++ {
		for x := 0; x < m.size; x++ {
			if m.provided[y][x] && m.grid[y][x] == shadedCell {
				return false
			}
		}
	}

	for y := 0; y < m.size; y++ {
		seen := make(map[rune]bool)
		for x := 0; x < m.size; x++ {
			val := m.grid[y][x]
			if val == shadedCell || val == emptyCell {
				continue
			}
			if seen[val] {
				return false
			}
			seen[val] = true
		}
	}

	for x := 0; x < m.size; x++ {
		seen := make(map[rune]bool)
		for y := 0; y < m.size; y++ {
			val := m.grid[y][x]
			if val == shadedCell || val == emptyCell {
				continue
			}
			if seen[val] {
				return false
			}
			seen[val] = true
		}
	}

	for y := 0; y < m.size; y++ {
		for x := 0; x < m.size; x++ {
			if m.grid[y][x] == shadedCell {
				if x+1 < m.size && m.grid[y][x+1] == shadedCell {
					return false
				}
				if y+1 < m.size && m.grid[y+1][x] == shadedCell {
					return false
				}
			}
		}
	}

	return m.isConnected()
}

func (m Model) isConnected() bool {
	var startX, startY int
	found := false
	for y := 0; y < m.size && !found; y++ {
		for x := 0; x < m.size; x++ {
			if m.grid[y][x] != shadedCell {
				startX, startY = x, y
				found = true
				break
			}
		}
	}
	if !found {
		return false
	}

	visited := make([][]bool, m.size)
	for i := range visited {
		visited[i] = make([]bool, m.size)
	}

	stack := [][2]int{{startX, startY}}
	visited[startY][startX] = true
	count := 0

	for len(stack) > 0 {
		curr := stack[len(stack)-1]
		stack = stack[:len(stack)-1]
		count++

		x, y := curr[0], curr[1]
		if x > 0 && !visited[y][x-1] && m.grid[y][x-1] != shadedCell {
			visited[y][x-1] = true
			stack = append(stack, [2]int{x - 1, y})
		}
		if x+1 < m.size && !visited[y][x+1] && m.grid[y][x+1] != shadedCell {
			visited[y][x+1] = true
			stack = append(stack, [2]int{x + 1, y})
		}
		if y > 0 && !visited[y-1][x] && m.grid[y-1][x] != shadedCell {
			visited[y-1][x] = true
			stack = append(stack, [2]int{x, y - 1})
		}
		if y+1 < m.size && !visited[y+1][x] && m.grid[y+1][x] != shadedCell {
			visited[y+1][x] = true
			stack = append(stack, [2]int{x, y + 1})
		}
	}

	total := 0
	for y := 0; y < m.size; y++ {
		for x := 0; x < m.size; x++ {
			if m.grid[y][x] != shadedCell {
				total++
			}
		}
	}

	return count == total
}

func (m *Model) updateKeyBindings() {
	m.keys.Shade.SetEnabled(!m.solved && !m.provided[m.cursor.Y][m.cursor.X])
	m.keys.Clear.SetEnabled(!m.solved && !m.provided[m.cursor.Y][m.cursor.X])
}
