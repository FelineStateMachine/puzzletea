package shikaku

import (
	"fmt"

	"github.com/FelineStateMachine/puzzletea/game"

	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

// expansion tracks how far the preview rectangle extends from the clue cell.
type expansion struct {
	clueID                int
	left, right, up, down int
}

// rect converts the expansion to a Rectangle.
func (e expansion) rect(clue *Clue) Rectangle {
	return Rectangle{
		ClueID: e.clueID,
		X:      clue.X - e.left,
		Y:      clue.Y - e.up,
		W:      e.left + 1 + e.right,
		H:      e.up + 1 + e.down,
	}
}

var _ game.Gamer = Model{}

// Model implements game.Gamer for Shikaku.
type Model struct {
	puzzle       Puzzle
	cursor       game.Cursor
	selectedClue *int // nil = navigation mode, non-nil = expansion mode
	expansion    expansion
	keys         KeyMap
	modeTitle    string
	showFullHelp bool

	// Mouse drag support: anchor is the cell where the drag started.
	mouseDragAnchor *[2]int    // nil = not dragging, [col, row] of click origin
	mousePreview    *Rectangle // preview rect computed during drag (for rendering)

	// Screen geometry for mouse hit-testing.
	termWidth, termHeight int

	// Cached grid origin for mouse hit-testing (recomputed on resize/solve).
	originX, originY int
	originValid      bool

	// Debug: last mouse event info.
	lastMouseX, lastMouseY int
	lastMouseBtn           string
	lastMouseGridCol       int
	lastMouseGridRow       int
	lastMouseHit           bool
}

// New creates a new Shikaku game model.
func New(mode ShikakuMode, puzzle Puzzle) game.Gamer {
	puzzle.autoPlaceSingles()
	return Model{
		puzzle:    puzzle,
		cursor:    game.Cursor{X: 0, Y: 0},
		keys:      DefaultKeyMap,
		modeTitle: mode.Title(),
	}
}

func (m Model) Init() tea.Cmd { return nil }

func (m Model) Update(msg tea.Msg) (game.Gamer, tea.Cmd) {
	switch msg := msg.(type) {
	case game.HelpToggleMsg:
		m.showFullHelp = msg.Show

	case tea.WindowSizeMsg:
		m.termWidth = msg.Width
		m.termHeight = msg.Height
		m.originValid = false

	case tea.KeyPressMsg:
		if m.puzzle.IsSolved() {
			m.cursor.Move(m.keys.CursorKeyMap, msg, m.puzzle.Width-1, m.puzzle.Height-1)
		} else if m.selectedClue != nil {
			m = m.handleExpansionMode(msg)
		} else {
			m = m.handleNavMode(msg)
		}

	case tea.MouseClickMsg:
		m = m.handleMouseClick(msg)

	case tea.MouseMotionMsg:
		m = m.handleMouseMotion(msg)

	case tea.MouseReleaseMsg:
		m = m.handleMouseRelease(msg)
	}
	return m, nil
}

func (m Model) handleMouseClick(msg tea.MouseClickMsg) Model {
	m.lastMouseX, m.lastMouseY = msg.X, msg.Y
	m.lastMouseBtn = msg.String()

	col, row, ok := m.screenToGrid(msg.X, msg.Y)
	m.lastMouseGridCol, m.lastMouseGridRow = col, row
	m.lastMouseHit = ok

	if m.puzzle.IsSolved() || !ok {
		return m
	}

	m.cursor.X, m.cursor.Y = col, row

	switch msg.Button {
	case tea.MouseLeft:
		// Start a drag from any cell.
		m.mouseDragAnchor = &[2]int{col, row}
		r := rectFromCorners(col, row, col, row)
		m.mousePreview = &r
		m.selectedClue = nil

		// Sync expansion if the anchor is on a clue (for seamless
		// keyboard takeover if the user releases without moving).
		m.syncExpansionFromPreview()

	case tea.MouseRight:
		// Right-click deletes rectangle at cursor.
		owner := m.puzzle.CellOwner(col, row)
		if owner >= 0 {
			m.puzzle.RemoveRectangle(owner)
			m.originValid = false
		}
	}
	return m
}

func (m Model) handleMouseMotion(msg tea.MouseMotionMsg) Model {
	if m.puzzle.IsSolved() || m.mouseDragAnchor == nil {
		return m
	}

	col, row, ok := m.screenToGrid(msg.X, msg.Y)
	if !ok {
		return m
	}

	// Update the preview rectangle spanning anchor to current cell.
	r := rectFromCorners(m.mouseDragAnchor[0], m.mouseDragAnchor[1], col, row)
	m.mousePreview = &r
	m.syncExpansionFromPreview()
	return m
}

func (m Model) handleMouseRelease(msg tea.MouseReleaseMsg) Model {
	if m.mouseDragAnchor == nil {
		return m
	}

	preview := m.mousePreview
	m.mouseDragAnchor = nil
	m.mousePreview = nil

	if preview == nil {
		m.selectedClue = nil
		return m
	}

	// Find clues inside the drawn rectangle.
	clues := m.puzzle.CluesInRect(*preview)
	if len(clues) != 1 {
		// Not exactly one clue: cancel the preview.
		m.selectedClue = nil
		return m
	}

	clue := clues[0]
	rect := Rectangle{
		ClueID: clue.ID,
		X:      preview.X,
		Y:      preview.Y,
		W:      preview.W,
		H:      preview.H,
	}

	// Place if valid: correct area and no overlap.
	if rect.Area() == clue.Value && !m.puzzle.Overlaps(rect, clue.ID) {
		m.puzzle.SetRectangle(rect)
		m.selectedClue = nil
		m.originValid = false
	} else {
		// Invalid: enter keyboard expansion mode on this clue so the
		// player can fine-tune.
		m.selectedClue = &clue.ID
		m.expansion = expansion{clueID: clue.ID}
		m.expansion.left = clue.X - preview.X
		m.expansion.right = preview.X + preview.W - 1 - clue.X
		m.expansion.up = clue.Y - preview.Y
		m.expansion.down = preview.Y + preview.H - 1 - clue.Y
		m.cursor.X, m.cursor.Y = clue.X, clue.Y
	}
	return m
}

// syncExpansionFromPreview updates selectedClue and expansion to match
// the current mousePreview so that the existing preview rendering works.
// If the preview contains exactly one clue, that clue is selected;
// otherwise selectedClue is cleared (the preview is still rendered via
// mousePreview in gridView).
func (m *Model) syncExpansionFromPreview() {
	if m.mousePreview == nil {
		m.selectedClue = nil
		return
	}

	clues := m.puzzle.CluesInRect(*m.mousePreview)
	if len(clues) == 1 {
		clue := clues[0]
		m.selectedClue = &clue.ID
		m.expansion = expansion{clueID: clue.ID}
		m.expansion.left = clue.X - m.mousePreview.X
		m.expansion.right = m.mousePreview.X + m.mousePreview.W - 1 - clue.X
		m.expansion.up = clue.Y - m.mousePreview.Y
		m.expansion.down = m.mousePreview.Y + m.mousePreview.H - 1 - clue.Y
	} else {
		m.selectedClue = nil
	}
}

// rectFromCorners builds a Rectangle spanning two arbitrary grid cells.
func rectFromCorners(x1, y1, x2, y2 int) Rectangle {
	minX, maxX := x1, x2
	if x2 < x1 {
		minX, maxX = x2, x1
	}
	minY, maxY := y1, y2
	if y2 < y1 {
		minY, maxY = y2, y1
	}
	return Rectangle{
		X: minX,
		Y: minY,
		W: maxX - minX + 1,
		H: maxY - minY + 1,
	}
}

func (m Model) handleNavMode(msg tea.KeyPressMsg) Model {
	switch {
	case key.Matches(msg, m.keys.Select):
		clue := m.puzzle.FindClueAt(m.cursor.X, m.cursor.Y)
		if clue != nil {
			m.selectedClue = &clue.ID
			m.expansion = expansion{clueID: clue.ID}
			// Initialize from existing rectangle if present.
			if r := m.puzzle.FindRectangleForClue(clue.ID); r != nil {
				m.expansion.left = clue.X - r.X
				m.expansion.right = r.X + r.W - 1 - clue.X
				m.expansion.up = clue.Y - r.Y
				m.expansion.down = r.Y + r.H - 1 - clue.Y
			}
		}
	case key.Matches(msg, m.keys.Delete):
		// Delete rectangle at cursor position.
		owner := m.puzzle.CellOwner(m.cursor.X, m.cursor.Y)
		if owner >= 0 {
			m.puzzle.RemoveRectangle(owner)
		}
	default:
		m.cursor.Move(m.keys.CursorKeyMap, msg, m.puzzle.Width-1, m.puzzle.Height-1)
	}
	return m
}

func (m Model) handleExpansionMode(msg tea.KeyPressMsg) Model {
	clue := m.puzzle.FindClueByID(*m.selectedClue)
	if clue == nil {
		m.selectedClue = nil
		return m
	}

	switch {
	// Expand each side independently.
	case key.Matches(msg, m.keys.Right):
		if m.expansion.right < m.puzzle.Width-1-clue.X {
			m.expansion.right++
		}
	case key.Matches(msg, m.keys.Left):
		if m.expansion.left < clue.X {
			m.expansion.left++
		}
	case key.Matches(msg, m.keys.Down):
		if m.expansion.down < m.puzzle.Height-1-clue.Y {
			m.expansion.down++
		}
	case key.Matches(msg, m.keys.Up):
		if m.expansion.up < clue.Y {
			m.expansion.up++
		}
	// Shrink each side independently.
	case key.Matches(msg, m.keys.ShrinkRight):
		if m.expansion.right > 0 {
			m.expansion.right--
		}
	case key.Matches(msg, m.keys.ShrinkLeft):
		if m.expansion.left > 0 {
			m.expansion.left--
		}
	case key.Matches(msg, m.keys.ShrinkDown):
		if m.expansion.down > 0 {
			m.expansion.down--
		}
	case key.Matches(msg, m.keys.ShrinkUp):
		if m.expansion.up > 0 {
			m.expansion.up--
		}
	case key.Matches(msg, m.keys.Select):
		rect := m.expansion.rect(clue)
		if rect.Area() == clue.Value && !m.puzzle.Overlaps(rect, clue.ID) {
			m.puzzle.SetRectangle(rect)
			m.selectedClue = nil
		}
	case key.Matches(msg, m.keys.Cancel):
		m.selectedClue = nil
	case key.Matches(msg, m.keys.Delete):
		m.puzzle.RemoveRectangle(clue.ID)
		m.selectedClue = nil
	}
	return m
}

func (m Model) View() string {
	solved := m.puzzle.IsSolved()
	title := game.TitleBarView("Shikaku", m.modeTitle, solved)
	grid := gridView(m, solved)
	if solved {
		return lipgloss.JoinVertical(lipgloss.Center, title, grid)
	}
	info := infoView(&m.puzzle)
	status := statusBarView(m.selectedClue != nil, m.showFullHelp)
	return lipgloss.JoinVertical(lipgloss.Center, title, grid, info, status)
}

func (m Model) SetTitle(t string) game.Gamer {
	m.modeTitle = t
	return m
}

func (m Model) IsSolved() bool {
	return m.puzzle.IsSolved()
}

func (m Model) Reset() game.Gamer {
	m.puzzle.Rectangles = nil
	m.puzzle.autoPlaceSingles()
	m.cursor = game.Cursor{}
	m.selectedClue = nil
	m.mouseDragAnchor = nil
	m.mousePreview = nil
	m.originValid = false
	return m
}

func (m Model) GetDebugInfo() string {
	status := "In Progress"
	if m.puzzle.IsSolved() {
		status = "Solved"
	}

	selectedStr := "None"
	if m.selectedClue != nil {
		selectedStr = fmt.Sprintf("%d", *m.selectedClue)
	}

	ox, oy := m.gridOrigin()
	hitStr := "miss"
	if m.lastMouseHit {
		hitStr = fmt.Sprintf("(%d, %d)", m.lastMouseGridCol, m.lastMouseGridRow)
	}

	dragStr := "None"
	if m.mouseDragAnchor != nil {
		dragStr = fmt.Sprintf("anchor=(%d, %d)", m.mouseDragAnchor[0], m.mouseDragAnchor[1])
	}

	s := game.DebugHeader("Shikaku", [][2]string{
		{"Status", status},
		{"Grid Size", fmt.Sprintf("%dx%d", m.puzzle.Width, m.puzzle.Height)},
		{"Clues", fmt.Sprintf("%d", len(m.puzzle.Clues))},
		{"Rectangles", fmt.Sprintf("%d", len(m.puzzle.Rectangles))},
		{"Cursor", fmt.Sprintf("(%d, %d)", m.cursor.X, m.cursor.Y)},
		{"Selected Clue", selectedStr},
		{"Term Size", fmt.Sprintf("%d x %d", m.termWidth, m.termHeight)},
		{"Grid Origin", fmt.Sprintf("(%d, %d)", ox, oy)},
		{"Last Mouse", fmt.Sprintf("screen=(%d, %d) btn=%s grid=%s", m.lastMouseX, m.lastMouseY, m.lastMouseBtn, hitStr)},
		{"Mouse Drag", dragStr},
	})

	var rows [][]string
	for _, c := range m.puzzle.Clues {
		r := m.puzzle.FindRectangleForClue(c.ID)
		rectStr := "-"
		rectStatus := "Empty"
		if r != nil {
			rectStr = fmt.Sprintf("(%d,%d) %dx%d", r.X, r.Y, r.W, r.H)
			if r.Area() == c.Value {
				rectStatus = "Correct"
			} else {
				rectStatus = fmt.Sprintf("Area %d", r.Area())
			}
		}
		rows = append(rows, []string{
			fmt.Sprintf("%d", c.ID),
			fmt.Sprintf("(%d,%d)", c.X, c.Y),
			fmt.Sprintf("%d", c.Value),
			rectStr,
			rectStatus,
		})
	}
	s += game.DebugTable("Clues", []string{"ID", "Pos", "Value", "Rectangle", "Status"}, rows)

	return s
}
