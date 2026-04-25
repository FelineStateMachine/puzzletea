package spellpuzzle

import (
	"fmt"
	"math/rand/v2"
	"regexp"
	"strings"

	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/FelineStateMachine/puzzletea/game"
)

var ansiPattern = regexp.MustCompile(`\x1b\[[0-9;]*m`)

type feedbackKind int

const (
	feedbackNone feedbackKind = iota
	feedbackError
	feedbackSuccess
)

type feedback struct {
	Kind    feedbackKind `json:"kind"`
	Message string       `json:"message"`
}

type Model struct {
	modeTitle string
	board     board
	bank      []rune

	placements []WordPlacement
	boardIndex map[string][]int

	allowedWords map[string]struct{}
	bonusSet     map[string]struct{}
	bonusWords   []string

	trace          []int
	bankCursor     int
	bankColorSlots []int
	keys           KeyMap

	solved       bool
	showFullHelp bool
	feedback     feedback

	mouseDragging bool
	hoverTile     int

	termWidth   int
	termHeight  int
	bankOriginX int
	bankOriginY int
	originValid bool
}

var _ game.Gamer = Model{}

func New(mode SpellPuzzleMode, puzzle GeneratedPuzzle) (game.Gamer, error) {
	model := &Model{
		modeTitle:      mode.Title(),
		bank:           []rune(puzzle.Bank),
		placements:     append([]WordPlacement(nil), puzzle.Placements...),
		bonusWords:     nil,
		bankCursor:     -1,
		bankColorSlots: buildBankColorSlots(len([]rune(puzzle.Bank))),
		keys:           DefaultKeyMap,
		hoverTile:      -1,
	}
	model.rebuildDerivedState()
	return model, nil
}

func (m *Model) rebuildDerivedState() {
	m.placements = normalizePlacements(m.placements)
	m.board = buildBoard(m.placements)
	m.boardIndex = make(map[string][]int, len(m.placements))
	for i := range m.placements {
		text := m.placements[i].Text
		m.boardIndex[text] = append(m.boardIndex[text], i)
	}

	m.allowedWords = make(map[string]struct{})
	for _, word := range spellableWords(string(m.bank), 3) {
		m.allowedWords[word] = struct{}{}
	}

	m.bonusSet = make(map[string]struct{}, len(m.bonusWords))
	for _, word := range m.bonusWords {
		m.bonusSet[word] = struct{}{}
	}
	m.solved = m.allBoardWordsFound()
	m.normalizeBankState()
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (game.Gamer, tea.Cmd) {
	switch typed := msg.(type) {
	case game.HelpToggleMsg:
		m.showFullHelp = typed.Show
		m.originValid = false
	case tea.WindowSizeMsg:
		m.termWidth = typed.Width
		m.termHeight = typed.Height
		m.originValid = false
	case tea.KeyPressMsg:
		m.clearFeedback()
		m.handleKeyPress(typed)
	case tea.MouseClickMsg:
		m.clearFeedback()
		m.handleMouseClick(typed)
	case tea.MouseMotionMsg:
		m.handleMouseMotion(typed)
	case tea.MouseReleaseMsg:
		m.handleMouseRelease(typed)
	}
	return m, nil
}

func (m *Model) handleKeyPress(msg tea.KeyPressMsg) {
	if !m.solved && m.handleTypedLetter(msg) {
		return
	}

	switch {
	case key.Matches(msg, m.keys.Left):
		m.moveBankCursorLeft()
	case key.Matches(msg, m.keys.Right):
		m.moveBankCursorRight()
	default:
		if m.solved {
			return
		}
		switch {
		case key.Matches(msg, m.keys.Shuffle):
			m.shuffleBank()
		case key.Matches(msg, m.keys.Submit):
			m.submitTrace()
		case key.Matches(msg, m.keys.Back):
			m.backspaceTrace()
		}
	}
}

func (m *Model) moveBankCursorLeft() {
	if len(m.bank) == 0 {
		return
	}
	if m.bankCursor < 0 {
		m.bankCursor = len(m.bank) - 1
		return
	}
	m.bankCursor = (m.bankCursor - 1 + len(m.bank)) % len(m.bank)
}

func (m *Model) moveBankCursorRight() {
	if len(m.bank) == 0 {
		return
	}
	if m.bankCursor < 0 {
		m.bankCursor = 0
		return
	}
	m.bankCursor = (m.bankCursor + 1) % len(m.bank)
}

func (m *Model) handleTypedLetter(msg tea.KeyPressMsg) bool {
	if len(msg.Text) != 1 {
		return false
	}

	typed := msg.Text[0]
	if typed >= 'a' && typed <= 'z' {
		typed = typed - 'a' + 'A'
	}
	if typed < 'A' || typed > 'Z' {
		return false
	}

	index, ok := m.findAvailableTileForLetter(rune(typed))
	if !ok {
		return true
	}
	m.appendTraceTile(index)
	m.bankCursor = index
	return true
}

func (m *Model) handleMouseClick(msg tea.MouseClickMsg) {
	if m.solved {
		return
	}
	if msg.Button != tea.MouseLeft {
		return
	}
	index, ok := m.tileAt(msg.X, msg.Y)
	if !ok {
		return
	}
	m.trace = nil
	m.bankCursor = -1
	m.mouseDragging = true
	m.hoverTile = index
	m.appendTraceTile(index)
}

func (m *Model) handleMouseMotion(msg tea.MouseMotionMsg) {
	if m.solved {
		m.hoverTile = -1
		return
	}
	if !m.mouseDragging {
		m.hoverTile = -1
		return
	}
	index, ok := m.tileAt(msg.X, msg.Y)
	if !ok {
		m.hoverTile = -1
		return
	}
	m.hoverTile = index
	m.appendTraceTile(index)
}

func (m *Model) handleMouseRelease(msg tea.MouseReleaseMsg) {
	if m.solved {
		m.mouseDragging = false
		m.hoverTile = -1
		return
	}
	if msg.Button != tea.MouseLeft || !m.mouseDragging {
		return
	}
	m.mouseDragging = false
	m.hoverTile = -1
	m.submitTrace()
}

func (m *Model) appendTraceTile(index int) {
	if index < 0 || index >= len(m.bank) {
		return
	}
	if m.tileUsage(index) >= m.tileCapacity(index) {
		return
	}
	m.trace = append(m.trace, index)
}

func (m *Model) shuffleBank() {
	if len(m.bank) < 2 {
		return
	}

	perm := make([]int, len(m.bank))
	for i := range perm {
		perm[i] = i
	}

	rng := rand.New(rand.NewPCG(rand.Uint64(), rand.Uint64()))
	changed := false
	for attempt := 0; attempt < 6; attempt++ {
		rng.Shuffle(len(perm), func(i, j int) {
			perm[i], perm[j] = perm[j], perm[i]
		})
		if !isIdentityPermutation(perm) {
			changed = true
			break
		}
	}
	if !changed {
		return
	}

	oldBank := append([]rune(nil), m.bank...)
	oldColorSlots := append([]int(nil), m.bankColorSlots...)
	oldToNew := make([]int, len(m.bank))
	for newIndex, oldIndex := range perm {
		m.bank[newIndex] = oldBank[oldIndex]
		if len(oldColorSlots) == len(m.bank) {
			m.bankColorSlots[newIndex] = oldColorSlots[oldIndex]
		}
		oldToNew[oldIndex] = newIndex
	}

	for i, oldIndex := range m.trace {
		m.trace[i] = oldToNew[oldIndex]
	}
	if m.bankCursor >= 0 && m.bankCursor < len(oldToNew) {
		m.bankCursor = oldToNew[m.bankCursor]
	}
	if m.hoverTile >= 0 && m.hoverTile < len(oldToNew) {
		m.hoverTile = oldToNew[m.hoverTile]
	}
	m.originValid = false
}

func isIdentityPermutation(perm []int) bool {
	for i, value := range perm {
		if i != value {
			return false
		}
	}
	return true
}

func (m Model) findAvailableTileForLetter(letter rune) (int, bool) {
	for index, current := range m.bank {
		if current != letter {
			continue
		}
		if m.tileUsage(index) < m.tileCapacity(index) {
			return index, true
		}
	}
	return 0, false
}

func (m *Model) backspaceTrace() {
	if len(m.trace) == 0 {
		return
	}
	m.trace = m.trace[:len(m.trace)-1]
	if len(m.trace) == 0 {
		m.bankCursor = -1
		return
	}
	m.bankCursor = m.trace[len(m.trace)-1]
}

func (m *Model) submitTrace() {
	word := m.traceWord()
	if word == "" {
		return
	}
	defer func() {
		m.trace = nil
		m.bankCursor = -1
		m.originValid = false
	}()

	if !isExactWord(word) || !m.canTraceSpell(word) {
		m.feedback = feedback{Kind: feedbackError, Message: "not a word"}
		return
	}

	if indexes, ok := m.boardIndex[word]; ok {
		alreadyFound := true
		for _, idx := range indexes {
			if !m.placements[idx].Found {
				alreadyFound = false
				break
			}
		}
		if alreadyFound {
			m.feedback = feedback{Kind: feedbackError, Message: "word already found"}
			return
		}
		for _, idx := range indexes {
			m.placements[idx].Found = true
		}
		m.board = buildBoard(m.placements)
		m.solved = m.allBoardWordsFound()
		m.feedback = feedback{Kind: feedbackSuccess, Message: "word added to board"}
		return
	}

	if _, ok := m.allowedWords[word]; !ok {
		m.feedback = feedback{Kind: feedbackError, Message: "not a word"}
		return
	}
	if _, ok := m.bonusSet[word]; ok {
		m.feedback = feedback{Kind: feedbackError, Message: "word already found"}
		return
	}

	m.bonusSet[word] = struct{}{}
	m.bonusWords = append(m.bonusWords, word)
	m.feedback = feedback{Kind: feedbackSuccess, Message: "bonus word counted"}
}

func (m *Model) clearFeedback() {
	m.feedback = feedback{}
}

func (m Model) canTraceSpell(word string) bool {
	counts := countLetters(word)
	for i, count := range counts {
		if count > countLetters(string(m.bank))[i] {
			return false
		}
	}
	return true
}

func (m Model) traceWord() string {
	var builder strings.Builder
	for _, index := range m.trace {
		if index < 0 || index >= len(m.bank) {
			continue
		}
		builder.WriteRune(m.bank[index])
	}
	return builder.String()
}

func (m Model) tileUsage(index int) int {
	count := 0
	for _, current := range m.trace {
		if current == index {
			count++
		}
	}
	return count
}

func (m Model) tileCapacity(index int) int {
	if index < 0 || index >= len(m.bank) {
		return 0
	}
	return 1
}

func (m Model) allBoardWordsFound() bool {
	if len(m.placements) == 0 {
		return false
	}
	for _, placement := range m.placements {
		if !placement.Found {
			return false
		}
	}
	return true
}

func (m Model) countWordsLeft() int {
	left := 0
	for _, placement := range m.placements {
		if !placement.Found {
			left++
		}
	}
	return left
}

func (m Model) countBonusWordsTotal() int {
	total := len(m.allowedWords) - len(m.boardIndex)
	if total < 0 {
		return 0
	}
	return total
}

func (m Model) recentBonusWords(limit int) []string {
	if limit >= len(m.bonusWords) {
		return append([]string(nil), m.bonusWords...)
	}
	return append([]string(nil), m.bonusWords[len(m.bonusWords)-limit:]...)
}

func (m Model) View() string {
	return renderView(m)
}

func (m Model) GetDebugInfo() string {
	rows := [][2]string{
		{"Bank", string(m.bank)},
		{"Trace", m.traceWord()},
		{"Words Left", fmt.Sprintf("%d", m.countWordsLeft())},
		{"Bonus Count", fmt.Sprintf("%d", len(m.bonusWords))},
		{"Solved", fmt.Sprintf("%v", m.solved)},
	}

	tableRows := make([][]string, 0, len(m.placements))
	for _, placement := range m.placements {
		found := "No"
		if placement.Found {
			found = "Yes"
		}
		tableRows = append(tableRows, []string{
			placement.Text,
			found,
			fmt.Sprintf("(%d,%d)", placement.Start.X, placement.Start.Y),
		})
	}

	out := game.DebugHeader("Spell Puzzle", rows)
	out += game.DebugTable("Board Words", []string{"Word", "Found", "Start"}, tableRows)
	return out
}

func (m Model) SetTitle(t string) game.Gamer {
	m.modeTitle = t
	return m
}

func (m Model) IsSolved() bool {
	return m.solved
}

func (m Model) Reset() game.Gamer {
	for i := range m.placements {
		m.placements[i].Found = false
	}
	m.bonusWords = nil
	m.trace = nil
	m.feedback = feedback{}
	m.hoverTile = -1
	m.bankCursor = -1
	m.rebuildDerivedState()
	return m
}

func (m *Model) normalizeBankState() {
	if len(m.bank) == 0 {
		m.bankCursor = -1
		m.hoverTile = -1
		return
	}
	if len(m.trace) == 0 {
		m.bankCursor = -1
	} else if m.bankCursor < 0 || m.bankCursor >= len(m.bank) {
		m.bankCursor = m.trace[len(m.trace)-1]
	}
	if !m.mouseDragging || m.hoverTile < 0 || m.hoverTile >= len(m.bank) {
		m.hoverTile = -1
	}
	if len(m.bankColorSlots) != len(m.bank) {
		m.bankColorSlots = buildBankColorSlots(len(m.bank))
	}
}

func buildBankColorSlots(length int) []int {
	slots := make([]int, length)
	if length == 0 {
		return slots
	}
	for i := range slots {
		slots[i] = i
	}
	return slots
}

func (m Model) tileAt(screenX, screenY int) (int, bool) {
	startX, startY := m.cachedBankOrigin()
	if screenY != startY {
		return 0, false
	}
	for i := range len(m.bank) {
		tileX := startX + i*4
		if screenX >= tileX && screenX < tileX+3 {
			return i, true
		}
	}
	return 0, false
}

func (m *Model) cachedBankOrigin() (int, int) {
	if m.originValid {
		return m.bankOriginX, m.bankOriginY
	}
	x, y := m.bankOrigin()
	m.bankOriginX, m.bankOriginY = x, y
	m.originValid = true
	return x, y
}

func (m Model) bankOrigin() (int, int) {
	view := renderView(m)
	viewWidth := lipgloss.Width(view)
	viewHeight := lipgloss.Height(view)
	centerX := max((m.termWidth-viewWidth)/2, 0)
	centerY := max((m.termHeight-viewHeight)/2, 0)
	strippedView := stripANSI(view)
	bank := stripANSI(renderBankRow(m))
	for rowIndex, line := range strings.Split(strippedView, "\n") {
		if col := strings.Index(line, bank); col >= 0 {
			return centerX + col, centerY + rowIndex
		}
	}
	return centerX, centerY
}

func stripANSI(text string) string {
	return ansiPattern.ReplaceAllString(text, "")
}
