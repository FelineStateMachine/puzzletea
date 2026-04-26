package app

import (
	"strconv"
	"strings"

	sessionflow "github.com/FelineStateMachine/puzzletea/session"
	"github.com/FelineStateMachine/puzzletea/theme"
	"github.com/FelineStateMachine/puzzletea/ui"

	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

type createScreen struct {
	width  int
	height int
	create createState
}

const createFooterHint = "↑/↓ navigate • enter expand/select/generate • space select • tab field • click select/generate • esc back"

type createRect struct {
	X int
	Y int
	W int
	H int
}

func (r createRect) contains(x, y int) bool {
	return x >= r.X && x < r.X+r.W && y >= r.Y && y < r.Y+r.H
}

type createSettingsRects struct {
	Elo         createRect
	Seed        createRect
	TotalHeight int
}

type createContentMetrics struct {
	Width         int
	Height        int
	Settings      createSettingsRects
	Tree          createRect
	TreeStart     int
	TreeEnd       int
	Pool          createRect
	Summary       createRect
	SummaryButton createRect
}

func (s createScreen) Resize(width, height int) screenModel {
	s.width = width
	s.height = height
	contentWidth, _ := createContentBounds(width, height)
	_, sideWidth := createLayoutWidths(contentWidth)
	s.create.seedInput.SetWidth(max(10, sideWidth-8))
	return s
}

func (s createScreen) Update(msg tea.Msg) (screenModel, tea.Cmd, screenAction) {
	if clickMsg, ok := msg.(tea.MouseClickMsg); ok {
		contentWidth, contentHeight := createContentBounds(s.width, s.height)
		if next, cmd, action, handled := s.handleMouseClick(clickMsg, contentWidth, contentHeight); handled {
			return next, cmd, action
		}
	}

	keyMsg, ok := msg.(tea.KeyPressMsg)
	if ok {
		switch {
		case key.Matches(keyMsg, rootKeys.Enter):
			return s.handleEnter()
		case key.Matches(keyMsg, rootKeys.Escape):
			return s, nil, createBackAction{create: s.create}
		case keyMsg.String() == "tab":
			s.create.focus = s.nextFocus(1)
			return s.syncFocus()
		case keyMsg.String() == "shift+tab":
			s.create.focus = s.nextFocus(-1)
			return s.syncFocus()
		case keyMsg.String() == "up" || keyMsg.String() == "k":
			if s.create.focus == createFocusTree {
				s.moveCursor(-1)
				return s, nil, nil
			}
		case keyMsg.String() == "down" || keyMsg.String() == "j":
			if s.create.focus == createFocusTree {
				s.moveCursor(1)
				return s, nil, nil
			}
		case keyMsg.String() == " " || keyMsg.String() == "space":
			if s.create.focus == createFocusTree {
				s.toggleCurrentTreeNode()
				return s, nil, nil
			}
		}
	}

	return s.updateFocusedInput(msg)
}

func (s createScreen) handleMouseClick(msg tea.MouseClickMsg, contentWidth, contentHeight int) (screenModel, tea.Cmd, screenAction, bool) {
	mouse := msg.Mouse()
	content := createBody(s.create, s.width, s.height)
	_, _, contentX, contentY := createPanelContentOrigin(s.width, s.height, content)
	localX := mouse.X - contentX
	localY := mouse.Y - contentY
	if localX < 0 || localY < 0 {
		return s, nil, nil, false
	}

	metrics := createMetrics(s.create, contentWidth, contentHeight)
	if metrics.Settings.Elo.contains(localX, localY) {
		s.create.focus = createFocusElo
		s.create.eloInput.SetCursor(createFieldCursor(localX, metrics.Settings.Elo, "Elo"))
		next, cmd, action := s.syncFocus()
		return next, cmd, action, true
	}
	if metrics.Settings.Seed.contains(localX, localY) {
		if !s.create.seedEnabled() {
			return s, nil, nil, true
		}
		s.create.focus = createFocusSeed
		s.create.seedInput.SetCursor(createFieldCursor(localX, metrics.Settings.Seed, "Seed"))
		next, cmd, action := s.syncFocus()
		return next, cmd, action, true
	}
	if metrics.SummaryButton.contains(localX, localY) {
		s.create.focus = createFocusGenerate
		next, cmd, action := s.syncFocus()
		if s.create.selectedCount() == 0 {
			return next, cmd, action, true
		}
		return next, cmd, createGenerateAction{create: s.create}, true
	}
	if index, ok := createTreeIndexAt(localX, localY, metrics, s.create); ok {
		s.create.focus = createFocusTree
		s.create.cursor = index
		next, cmd, action := s.syncFocus()
		nextScreen := next.(createScreen)
		return nextScreen.handleEnterWithHandled(cmd, action)
	}
	return s, nil, nil, false
}

func (s createScreen) handleEnterWithHandled(cmd tea.Cmd, action screenAction) (screenModel, tea.Cmd, screenAction, bool) {
	next, enterCmd, enterAction := s.handleEnter()
	return next, tea.Batch(cmd, enterCmd), firstScreenAction(action, enterAction), true
}

func firstScreenAction(actions ...screenAction) screenAction {
	for _, action := range actions {
		if action != nil {
			return action
		}
	}
	return nil
}

func createFieldCursor(localX int, rect createRect, label string) int {
	labelWidth := lipgloss.Width(label) + 1
	return max(localX-rect.X-2-labelWidth, 0)
}

func createPanelContentOrigin(screenWidth, screenHeight int, content string) (panelX, panelY, contentX, contentY int) {
	panel := ui.Panel("Create", content, createFooterHint)
	panelWidth := lipgloss.Width(panel)
	panelHeight := lipgloss.Height(panel)
	panelX = max((screenWidth-panelWidth)/2, 0)
	panelY = max((screenHeight-panelHeight)/2, 0)

	frame := ui.PanelFrame()
	contentX = panelX + frame.GetHorizontalFrameSize()/2
	contentY = panelY + 1 + 2 + 2
	return panelX, panelY, contentX, contentY
}

func (s createScreen) handleEnter() (screenModel, tea.Cmd, screenAction) {
	if s.create.focus == createFocusGenerate {
		if s.create.selectedCount() == 0 {
			return s, nil, nil
		}
		return s, nil, createGenerateAction{create: s.create}
	}
	if s.create.focus != createFocusTree {
		return s, nil, nil
	}

	visible := s.create.visibleNodes()
	if s.create.cursor < 0 || s.create.cursor >= len(visible) {
		return s, nil, nil
	}
	node := visible[s.create.cursor].node
	if node.leaf == nil {
		s.create.expanded[node.id] = !s.create.expanded[node.id]
		return s, nil, nil
	}
	s.toggleLeaf(node)
	return s, nil, nil
}

func (s createScreen) nextFocus(delta int) createFocus {
	focuses := []createFocus{createFocusTree, createFocusElo, createFocusGenerate}
	if s.create.seedEnabled() {
		focuses = []createFocus{createFocusTree, createFocusElo, createFocusSeed, createFocusGenerate}
	}

	current := 0
	for i, focus := range focuses {
		if focus == s.create.focus {
			current = i
			break
		}
	}
	next := current + delta
	for next < 0 {
		next += len(focuses)
	}
	return focuses[next%len(focuses)]
}

func (s createScreen) syncFocus() (screenModel, tea.Cmd, screenAction) {
	s.create.eloInput.Blur()
	s.create.seedInput.Blur()
	switch s.create.focus {
	case createFocusElo:
		return s, s.create.eloInput.Focus(), nil
	case createFocusSeed:
		if s.create.seedEnabled() {
			return s, s.create.seedInput.Focus(), nil
		}
		s.create.focus = createFocusTree
		return s, nil, nil
	case createFocusGenerate:
		return s, nil, nil
	default:
		return s, nil, nil
	}
}

func (s createScreen) updateFocusedInput(msg tea.Msg) (screenModel, tea.Cmd, screenAction) {
	var cmd tea.Cmd
	switch s.create.focus {
	case createFocusElo:
		s.create.eloInput, cmd = s.create.eloInput.Update(msg)
	case createFocusSeed:
		if s.create.seedEnabled() {
			s.create.seedInput, cmd = s.create.seedInput.Update(msg)
		}
	}
	return s, cmd, nil
}

func (s *createScreen) moveCursor(delta int) {
	visible := s.create.visibleNodes()
	if len(visible) == 0 {
		s.create.cursor = 0
		return
	}
	s.create.cursor += delta
	if s.create.cursor < 0 {
		s.create.cursor = len(visible) - 1
	}
	if s.create.cursor >= len(visible) {
		s.create.cursor = 0
	}
}

func (s *createScreen) toggleCurrentTreeNode() {
	visible := s.create.visibleNodes()
	if s.create.cursor < 0 || s.create.cursor >= len(visible) {
		return
	}
	node := visible[s.create.cursor].node
	if node.leaf == nil {
		s.create.toggleDescendantLeaves(node)
		s.afterCreateSelectionChanged()
		return
	}
	s.toggleLeaf(node)
}

func (s *createScreen) toggleLeaf(node createTreeNode) {
	if node.leaf == nil {
		return
	}
	s.create.checked[node.leaf.id] = !s.create.checked[node.leaf.id]
	s.afterCreateSelectionChanged()
}

func (s *createScreen) afterCreateSelectionChanged() {
	if !s.create.seedEnabled() {
		s.create.seedInput.Blur()
		if s.create.focus == createFocusSeed {
			s.create.focus = createFocusTree
		}
	}
}

func (s createScreen) View(notice noticeState) string {
	return renderPanelView(
		s.width,
		s.height,
		notice,
		"Create",
		createBody(s.create, s.width, s.height),
		createFooterHint,
	)
}

func createBody(state createState, width, height int) string {
	contentWidth, contentHeight := createContentBounds(width, height)
	treeWidth, sideWidth := createLayoutWidths(contentWidth)
	settings := createSettingsPanel(state, contentWidth)
	tree := createTreePanel(state, treeWidth, contentHeight)
	sidebar := createPoolPanel(state, sideWidth, contentHeight)
	summary := createSummaryPanel(state, contentWidth)

	var middle string
	if contentWidth < 88 {
		middle = lipgloss.JoinVertical(lipgloss.Left, tree, sidebar)
	} else {
		middle = lipgloss.JoinHorizontal(lipgloss.Top, tree, "   ", sidebar)
	}
	return lipgloss.JoinVertical(lipgloss.Left, settings, middle, summary)
}

func createContentBounds(width, height int) (contentWidth, contentHeight int) {
	contentWidth = max(width-14, 24)
	contentHeight = max(height-14, 10)
	return contentWidth, contentHeight
}

func createLayoutWidths(contentWidth int) (treeWidth, sideWidth int) {
	sideWidth = min(42, max(30, contentWidth/3))
	treeWidth = max(24, contentWidth-sideWidth-3)
	return treeWidth, sideWidth
}

func createMetrics(state createState, width, height int) createContentMetrics {
	settings := createSettingsRectsForWidth(state, width)
	treeWidth, sideWidth := createLayoutWidths(width)
	tree := createTreePanel(state, treeWidth, height)
	pool := createPoolPanel(state, sideWidth, height)
	summary := createSummaryPanel(state, width)
	treeHeight := lipgloss.Height(tree)
	poolHeight := lipgloss.Height(pool)
	summaryHeight := lipgloss.Height(summary)

	treeY := settings.TotalHeight
	poolX := 0
	poolY := treeY + treeHeight
	if width >= 88 {
		poolX = treeWidth + 3
		poolY = treeY
	}
	summaryY := settings.TotalHeight + treeHeight + poolHeight
	if width >= 88 {
		summaryY = settings.TotalHeight + max(treeHeight, poolHeight)
	}

	button := renderCreateGenerateChip(state)
	buttonWidth := lipgloss.Width(button)
	start, end := createVisibleWindow(state, height)
	return createContentMetrics{
		Width:     width,
		Height:    height,
		Settings:  settings,
		Tree:      createRect{X: 0, Y: treeY, W: treeWidth, H: treeHeight},
		TreeStart: start,
		TreeEnd:   end,
		Pool:      createRect{X: poolX, Y: poolY, W: sideWidth, H: poolHeight},
		Summary:   createRect{X: 0, Y: summaryY, W: width, H: summaryHeight},
		SummaryButton: createRect{
			X: max(width-buttonWidth, 0),
			Y: summaryY,
			W: buttonWidth,
			H: summaryHeight,
		},
	}
}

func createSettingsRectsForWidth(state createState, width int) createSettingsRects {
	inlineHeight := lipgloss.Height(createInlineField("Elo", state.eloInput.View(), state.focus == createFocusElo, max(width, 18)))
	if width < 52 {
		return createSettingsRects{
			Elo:         createRect{X: 0, Y: 0, W: width, H: inlineHeight},
			Seed:        createRect{X: 0, Y: inlineHeight, W: width, H: inlineHeight},
			TotalHeight: inlineHeight * 2,
		}
	}

	eloWidth := 18
	seedWidth := max(width-eloWidth-2, 24)
	return createSettingsRects{
		Elo:         createRect{X: 0, Y: 0, W: eloWidth, H: inlineHeight},
		Seed:        createRect{X: eloWidth + 2, Y: 0, W: seedWidth, H: inlineHeight},
		TotalHeight: inlineHeight,
	}
}

func createVisibleWindow(state createState, height int) (start, end int) {
	visible := state.visibleNodes()
	if len(visible) == 0 {
		return 0, 0
	}
	maxRows := max(8, height-9)
	if len(visible) <= maxRows {
		return 0, len(visible)
	}
	start = min(max(0, state.cursor-maxRows+1), len(visible)-maxRows)
	return start, start + maxRows
}

func createTreeIndexAt(localX, localY int, metrics createContentMetrics, state createState) (int, bool) {
	if !metrics.Tree.contains(localX, localY) {
		return 0, false
	}
	row := localY - metrics.Tree.Y - 2
	if row < 0 || row >= metrics.TreeEnd-metrics.TreeStart {
		return 0, false
	}
	visible := state.visibleNodes()
	if metrics.TreeStart > 0 && row == 0 {
		return 0, false
	}
	if metrics.TreeEnd < len(visible) && row == metrics.TreeEnd-metrics.TreeStart-1 {
		return 0, false
	}
	index := metrics.TreeStart + row
	if index < 0 || index >= len(visible) {
		return 0, false
	}
	return index, true
}

func createTreePanel(state createState, width, height int) string {
	lines := make([]string, 0)
	for i, item := range state.visibleNodes() {
		lines = append(lines, renderCreateTreeLine(state, item, i == state.cursor))
	}
	if len(lines) == 0 {
		lines = append(lines, "No Elo-capable games available.")
	}

	start, end := createVisibleWindow(state, height)
	if len(state.visibleNodes()) > 0 && len(lines) > end-start {
		lines = lines[start:end]
		if start > 0 {
			lines[0] = ui.DimItemStyle().Render("...")
		}
		if end < len(state.visibleNodes()) {
			lines[len(lines)-1] = ui.DimItemStyle().Render("...")
		}
	}

	return createSection("Puzzle Pool", strings.Join(lines, "\n"), state.focus == createFocusTree, width)
}

func createSettingsPanel(state createState, width int) string {
	seed := state.seedInput.View()
	if !state.seedEnabled() {
		seed = "disabled for multi-select"
	}

	if width < 52 {
		return lipgloss.JoinVertical(lipgloss.Left,
			createInlineField("Elo", state.eloInput.View(), state.focus == createFocusElo, width),
			createInlineTextField("Seed", seed, state.focus == createFocusSeed && state.seedEnabled(), !state.seedEnabled(), width),
		)
	}

	eloWidth := 18
	seedWidth := max(width-eloWidth-2, 24)
	return lipgloss.JoinHorizontal(lipgloss.Top,
		createInlineField("Elo", state.eloInput.View(), state.focus == createFocusElo, eloWidth),
		"  ",
		createInlineTextField("Seed", seed, state.focus == createFocusSeed && state.seedEnabled(), !state.seedEnabled(), seedWidth),
	)
}

func createPoolPanel(state createState, width, height int) string {
	body := lipgloss.JoinVertical(
		lipgloss.Left,
		ui.DimItemStyle().Render(strconv.Itoa(state.selectedCount())+" selected"),
		renderSelectedLeafSummary(state, width-4, max(3, height-7)),
	)
	return createSection("Selected", body, false, width)
}

func createSummaryPanel(state createState, width int) string {
	totalText := ui.PanelTitle().Render("Total: " + strconv.Itoa(state.selectedCount()))
	selected := selectedCreateDetail(state)
	button := renderCreateGenerateChip(state)
	detailWidth := max(width-lipgloss.Width(totalText)-lipgloss.Width(button)-4, 12)
	detail := lipgloss.NewStyle().
		Width(detailWidth).
		MaxWidth(detailWidth).
		Render(ui.DimItemStyle().Render(truncatePlainCreateLine(selected, detailWidth)))
	return lipgloss.JoinHorizontal(lipgloss.Top, totalText, "   ", detail, " ", button)
}

func renderSelectedLeafSummary(state createState, width, maxRows int) string {
	leaves := state.selectedLeaves()
	if len(leaves) == 0 {
		return ui.DimItemStyle().Render("No leaves selected")
	}

	limit := min(len(leaves), maxRows)
	lines := make([]string, 0, limit+1)
	for _, leaf := range leaves[:limit] {
		lines = append(lines, ui.DimItemStyle().Render(truncatePlainCreateLine(leaf.gameType+" / "+leaf.title, width)))
	}
	if len(leaves) > limit {
		lines = append(lines, ui.DimItemStyle().Render("+"+strconv.Itoa(len(leaves)-limit)+" more"))
	}
	return strings.Join(lines, "\n")
}

func selectedCreateDetail(state createState) string {
	leaves := state.selectedLeaves()
	if len(leaves) == 0 {
		return "Selected: none"
	}
	if len(leaves) == 1 {
		return "Selected: " + leaves[0].gameType + " / " + leaves[0].title
	}
	return "Selected: random from checked pool"
}

func renderCreateGenerateChip(state createState) string {
	if state.selectedCount() == 0 {
		return ui.DimItemStyle().Render(renderChoiceChip("Generate", false))
	}
	return renderChoiceChip("Generate", state.focus == createFocusGenerate)
}

func renderCreateTreeLine(state createState, item createVisibleNode, cursor bool) string {
	node := item.node
	prefix := strings.Repeat("  ", item.depth)
	cursorText := "  "
	if cursor && state.focus == createFocusTree {
		cursorText = ui.CursorStyle().Render("> ")
	}
	if node.leaf == nil {
		selected, total := state.descendantLeafCounts(node)
		marker := "[ ] "
		if total > 0 && selected == total {
			marker = "[x] "
		} else if selected > 0 {
			marker = "[-] "
		}
		count := ""
		if total > 1 {
			count = " " + ui.DimItemStyle().Render(strconv.Itoa(selected)+"/"+strconv.Itoa(total))
		}
		return cursorText + prefix + marker + node.title + count
	}

	box := "[ ] "
	if state.checked[node.leaf.id] {
		box = "[x] "
	}
	return cursorText + prefix + box + node.title
}

func createInlineField(label, value string, focused bool, width int) string {
	return createInlineTextField(label, value, focused, false, width)
}

func createInlineTextField(label, value string, focused, dimmed bool, width int) string {
	border := theme.Current().Border
	if focused {
		border = theme.Current().Accent
	}

	innerWidth := max(width-4, 1)
	labelWidth := lipgloss.Width(label) + 1
	valueStyle := lipgloss.NewStyle().Foreground(theme.Current().FG)
	if dimmed {
		valueStyle = ui.DimItemStyle()
	}
	if focused {
		valueStyle = valueStyle.Foreground(theme.Current().Accent).Bold(true)
	}
	body := lipgloss.JoinHorizontal(
		lipgloss.Top,
		ui.DimItemStyle().Render(label),
		" ",
		renderSingleLine(valueStyle.Render(value), innerWidth-labelWidth),
	)
	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(border).
		Padding(0, 1).
		Width(width).
		Height(1).
		Render(body)
}

func createSection(title, content string, focused bool, width int) string {
	border := theme.Current().Border
	titleStyle := ui.PanelTitle()
	if focused {
		border = theme.Current().Accent
		titleStyle = ui.SelectedItemStyle()
	}
	body := lipgloss.JoinVertical(lipgloss.Left, titleStyle.Render(title), content)
	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(border).
		Padding(0, 1).
		Width(width).
		Render(body)
}

func truncatePlainCreateLine(s string, width int) string {
	if width <= 1 || len(s) <= width {
		return s
	}
	return s[:width-1] + "…"
}

type createGenerateAction struct {
	create createState
}

type createBackAction struct {
	create createState
}

func (a createBackAction) applyToModel(m model) (model, tea.Cmd) {
	m.create = a.create
	m = m.persistCreatePresetIfValid()
	m.state = playMenuView
	return m.resizeActiveScreen(), nil
}

func (a createGenerateAction) applyToModel(m model) (model, tea.Cmd) {
	m.create = a.create
	return m.handleCreateGenerate()
}

func normalizedCreateSeed(value string) string {
	return sessionflow.NormalizeSeed(value)
}
