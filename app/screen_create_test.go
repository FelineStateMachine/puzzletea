package app

import (
	"strings"
	"testing"

	"github.com/FelineStateMachine/puzzletea/config"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/ansi"
)

func TestCreateScreenSpaceTogglesParentDescendantsAndLeaves(t *testing.T) {
	state := newCreateState(config.CreateConfig{}, 80)
	screen := createScreen{create: state}
	parent := state.visibleNodes()[0].node
	_, parentTotal := state.descendantLeafCounts(parent)

	next, _, _ := screen.Update(tea.KeyPressMsg{Code: tea.KeySpace})
	got := next.(createScreen)
	if got.create.selectedCount() != parentTotal {
		t.Fatalf("selected count after parent space = %d, want %d", got.create.selectedCount(), parentTotal)
	}

	next, _, _ = got.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	got = next.(createScreen)
	got.create.cursor = 1
	next, _, _ = got.Update(tea.KeyPressMsg{Code: tea.KeySpace})
	got = next.(createScreen)
	if got.create.selectedCount() != parentTotal-1 {
		t.Fatalf("selected count after leaf space = %d, want %d", got.create.selectedCount(), parentTotal-1)
	}
}

func TestCreateScreenEnterExpandsAndCollapsesParent(t *testing.T) {
	screen := createScreen{create: newCreateState(config.CreateConfig{}, 80)}
	parentID := screen.create.visibleNodes()[0].node.id

	next, _, _ := screen.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	expanded := next.(createScreen)
	if !expanded.create.expanded[parentID] {
		t.Fatal("expected parent to expand")
	}
	if expanded.create.selectedCount() != 0 {
		t.Fatalf("selected count after parent enter = %d, want 0", expanded.create.selectedCount())
	}

	next, _, _ = expanded.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	collapsed := next.(createScreen)
	if collapsed.create.expanded[parentID] {
		t.Fatal("expected parent to collapse")
	}
}

func TestCreateScreenGenerateButtonProducesAction(t *testing.T) {
	state := newCreateState(config.CreateConfig{SelectedLeafIDs: []string{"nonogram/5x5"}}, 80)
	state.focus = createFocusGenerate
	screen := createScreen{create: state}

	_, _, action := screen.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	if _, ok := action.(createGenerateAction); !ok {
		t.Fatalf("action = %T, want createGenerateAction", action)
	}
}

func TestCreateScreenGenerateButtonDisabledWithoutSelection(t *testing.T) {
	state := newCreateState(config.CreateConfig{Elo: 1200}, 80)
	state.focus = createFocusGenerate
	screen := createScreen{create: state}

	_, _, action := screen.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	if action != nil {
		t.Fatalf("action = %T, want nil", action)
	}
}

func TestCreateScreenMouseClickFocusesFields(t *testing.T) {
	state := newCreateState(config.CreateConfig{SelectedLeafIDs: []string{"nonogram/5x5"}}, 120)
	screen := createScreen{create: state}.Resize(120, 36).(createScreen)
	contentWidth, contentHeight := createContentBounds(screen.width, screen.height)
	metrics := createMetrics(screen.create, contentWidth, contentHeight)
	_, _, contentX, contentY := createPanelContentOrigin(screen.width, screen.height, createBody(screen.create, screen.width, screen.height))

	next, _, _, handled := screen.handleMouseClick(tea.MouseClickMsg{
		X: contentX + metrics.Settings.Seed.X + 8,
		Y: contentY + metrics.Settings.Seed.Y + 1,
	}, contentWidth, contentHeight)
	if !handled {
		t.Fatal("expected seed field click to be handled")
	}
	screen = next.(createScreen)
	if screen.create.focus != createFocusSeed {
		t.Fatalf("focus = %v, want createFocusSeed", screen.create.focus)
	}

	metrics = createMetrics(screen.create, contentWidth, contentHeight)
	next, _, _, handled = screen.handleMouseClick(tea.MouseClickMsg{
		X: contentX + metrics.Settings.Elo.X + 7,
		Y: contentY + metrics.Settings.Elo.Y + 1,
	}, contentWidth, contentHeight)
	if !handled {
		t.Fatal("expected elo field click to be handled")
	}
	screen = next.(createScreen)
	if screen.create.focus != createFocusElo {
		t.Fatalf("focus = %v, want createFocusElo", screen.create.focus)
	}
}

func TestCreateScreenMouseClickDisabledSeedDoesNotFocusSeed(t *testing.T) {
	state := newCreateState(config.CreateConfig{
		SelectedLeafIDs: []string{"nonogram/5x5", "netwalk/5x5"},
	}, 120)
	screen := createScreen{create: state}.Resize(120, 36).(createScreen)
	contentWidth, contentHeight := createContentBounds(screen.width, screen.height)
	metrics := createMetrics(screen.create, contentWidth, contentHeight)
	_, _, contentX, contentY := createPanelContentOrigin(screen.width, screen.height, createBody(screen.create, screen.width, screen.height))

	next, _, _, handled := screen.handleMouseClick(tea.MouseClickMsg{
		X: contentX + metrics.Settings.Seed.X + 8,
		Y: contentY + metrics.Settings.Seed.Y + 1,
	}, contentWidth, contentHeight)
	if !handled {
		t.Fatal("expected disabled seed field click to be handled")
	}
	got := next.(createScreen)
	if got.create.focus == createFocusSeed {
		t.Fatal("disabled seed click should not focus the seed input")
	}
}

func TestCreateScreenMouseClickGenerateButtonProducesAction(t *testing.T) {
	state := newCreateState(config.CreateConfig{SelectedLeafIDs: []string{"nonogram/5x5"}}, 120)
	screen := createScreen{create: state}.Resize(120, 36).(createScreen)
	contentWidth, contentHeight := createContentBounds(screen.width, screen.height)
	metrics := createMetrics(screen.create, contentWidth, contentHeight)
	_, _, contentX, contentY := createPanelContentOrigin(screen.width, screen.height, createBody(screen.create, screen.width, screen.height))

	_, _, action, handled := screen.handleMouseClick(tea.MouseClickMsg{
		X: contentX + metrics.SummaryButton.X + 1,
		Y: contentY + metrics.SummaryButton.Y,
	}, contentWidth, contentHeight)
	if !handled {
		t.Fatal("expected generate button click to be handled")
	}
	if _, ok := action.(createGenerateAction); !ok {
		t.Fatalf("action = %T, want createGenerateAction", action)
	}
}

func TestCreateScreenMouseClickTreeRowsSelectsAndToggles(t *testing.T) {
	state := newCreateState(config.CreateConfig{}, 120)
	parentID := state.visibleNodes()[0].node.id
	state.expanded[parentID] = true
	screen := createScreen{create: state}.Resize(120, 36).(createScreen)
	contentWidth, contentHeight := createContentBounds(screen.width, screen.height)
	metrics := createMetrics(screen.create, contentWidth, contentHeight)
	_, _, contentX, contentY := createPanelContentOrigin(screen.width, screen.height, createBody(screen.create, screen.width, screen.height))

	leafIndex := -1
	for i, item := range screen.create.visibleNodes() {
		if item.node.leaf != nil {
			leafIndex = i
			break
		}
	}
	if leafIndex < 0 {
		t.Fatal("test setup expected a visible leaf")
	}
	row := leafIndex - metrics.TreeStart
	next, _, _, handled := screen.handleMouseClick(tea.MouseClickMsg{
		X: contentX + metrics.Tree.X + 4,
		Y: contentY + metrics.Tree.Y + 2 + row,
	}, contentWidth, contentHeight)
	if !handled {
		t.Fatal("expected tree leaf click to be handled")
	}
	got := next.(createScreen)
	if got.create.focus != createFocusTree {
		t.Fatalf("focus = %v, want createFocusTree", got.create.focus)
	}
	if got.create.cursor != leafIndex {
		t.Fatalf("cursor = %d, want %d", got.create.cursor, leafIndex)
	}
	if got.create.selectedCount() != 1 {
		t.Fatalf("selected count = %d, want 1", got.create.selectedCount())
	}
}

func TestCreateScreenMouseClickTreeParentExpands(t *testing.T) {
	state := newCreateState(config.CreateConfig{}, 120)
	parentID := state.visibleNodes()[0].node.id
	screen := createScreen{create: state}.Resize(120, 36).(createScreen)
	contentWidth, contentHeight := createContentBounds(screen.width, screen.height)
	metrics := createMetrics(screen.create, contentWidth, contentHeight)
	_, _, contentX, contentY := createPanelContentOrigin(screen.width, screen.height, createBody(screen.create, screen.width, screen.height))

	next, _, _, handled := screen.handleMouseClick(tea.MouseClickMsg{
		X: contentX + metrics.Tree.X + 4,
		Y: contentY + metrics.Tree.Y + 2,
	}, contentWidth, contentHeight)
	if !handled {
		t.Fatal("expected tree parent click to be handled")
	}
	got := next.(createScreen)
	if !got.create.expanded[parentID] {
		t.Fatal("expected parent click to expand the tree row")
	}
}

func TestCreateScreenRenderedParentRowsIncludeRollupCounts(t *testing.T) {
	state := newCreateState(config.CreateConfig{}, 80)
	sudoku := requireCreateNode(t, state.tree, "Sudoku")
	state.checked[firstDescendantLeafID(t, sudoku.children[0])] = true

	line := renderCreateTreeLine(state, createVisibleNode{node: sudoku}, false)
	if !strings.Contains(line, "Sudoku") || !strings.Contains(line, "1/2") {
		t.Fatalf("rendered line = %q, want Sudoku with 1/2 rollup", line)
	}
}

func TestCreateScreenRenderedParentRowsHideSingletonRollupCounts(t *testing.T) {
	state := newCreateState(config.CreateConfig{}, 80)
	nonogram := requireCreateNode(t, state.tree, "Nonogram")
	state.checked[firstDescendantLeafID(t, nonogram)] = true

	line := renderCreateTreeLine(state, createVisibleNode{node: nonogram}, false)
	if strings.Contains(line, "1/1") {
		t.Fatalf("rendered line = %q, want singleton parent without rollup count", line)
	}
}

func TestCreateScreenRenderedParentRowsSignalSelectionState(t *testing.T) {
	state := newCreateState(config.CreateConfig{}, 80)
	sudoku := requireCreateNode(t, state.tree, "Sudoku")

	line := renderCreateTreeLine(state, createVisibleNode{node: sudoku}, false)
	if !strings.Contains(line, "[ ] Sudoku") {
		t.Fatalf("unchecked parent line = %q, want unchecked marker", line)
	}

	state.checked[firstDescendantLeafID(t, sudoku.children[0])] = true
	line = renderCreateTreeLine(state, createVisibleNode{node: sudoku}, false)
	if !strings.Contains(line, "[-] Sudoku") {
		t.Fatalf("partial parent line = %q, want partial marker", line)
	}

	state.toggleDescendantLeaves(sudoku)
	line = renderCreateTreeLine(state, createVisibleNode{node: sudoku}, false)
	if !strings.Contains(line, "[x] Sudoku") {
		t.Fatalf("selected parent line = %q, want checked marker", line)
	}
}

func TestCreateScreenEloFieldDoesNotRenderPromptMarker(t *testing.T) {
	state := newCreateState(config.CreateConfig{Elo: 1200}, 80)
	state.focus = createFocusElo
	screen, _, _ := (createScreen{create: state}).syncFocus()
	state = screen.(createScreen).create

	settings := ansi.Strip(createSettingsPanel(state, 80))
	if strings.Contains(settings, "> 1200") || strings.Contains(settings, ">1200") {
		t.Fatalf("settings rendered Elo prompt marker: %q", settings)
	}
	if got := state.eloInput.Value(); got != "1200" {
		t.Fatalf("elo input value = %q, want 1200", got)
	}
}
