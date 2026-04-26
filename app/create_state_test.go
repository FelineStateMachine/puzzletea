package app

import (
	"testing"

	"github.com/FelineStateMachine/puzzletea/config"
	"github.com/FelineStateMachine/puzzletea/difficulty"
)

func TestCreateStateStartsCollapsedAndRestoresCheckedLeaves(t *testing.T) {
	const leafID = "nonogram/nonogram"
	state := newCreateState(config.CreateConfig{
		SelectedLeafIDs: []string{leafID},
		Elo:             1500,
	}, 80)

	if !state.checked[leafID] {
		t.Fatalf("expected %q to be checked", leafID)
	}
	if got := state.eloInput.Value(); got != "1500" {
		t.Fatalf("elo input = %q, want 1500", got)
	}
	for _, item := range state.visibleNodes() {
		if item.node.leaf != nil && item.node.leaf.id == leafID {
			return
		}
	}
	t.Fatalf("visible nodes = %#v, want restored top-level leaf %q", state.visibleNodes(), leafID)
}

func TestCreateStateIgnoresUnknownRestoredLeafIDs(t *testing.T) {
	const validLeafID = "nonogram/nonogram"
	state := newCreateState(config.CreateConfig{
		SelectedLeafIDs: []string{
			"nonogram/5x5",
			validLeafID,
			"stale/game/mode",
		},
	}, 80)

	if got := state.selectedCount(); got != 1 {
		t.Fatalf("selected count = %d, want 1", got)
	}
	if got := state.selectedLeafIDs(); len(got) != 1 || got[0] != validLeafID {
		t.Fatalf("selected leaf IDs = %#v, want [%q]", got, validLeafID)
	}
	if len(state.checked) != 1 || !state.checked[validLeafID] {
		t.Fatalf("checked map = %#v, want only %q", state.checked, validLeafID)
	}
}

func TestCreateStateDefaultStartsCollapsed(t *testing.T) {
	state := newCreateState(config.CreateConfig{}, 80)

	if len(state.visibleNodes()) != len(state.tree) {
		t.Fatalf("visible node count = %d, want %d", len(state.visibleNodes()), len(state.tree))
	}
}

func TestCreateTreeGroupsTakuzuAndSudokuVariants(t *testing.T) {
	state := newCreateState(config.CreateConfig{}, 80)

	takuzu := requireCreateNode(t, state.tree, "Takuzu")
	if !hasCreateChild(takuzu, "Takuzu") || !hasCreateChild(takuzu, "Takuzu+") {
		t.Fatalf("Takuzu group children = %#v, want Takuzu and Takuzu+", childTitles(takuzu))
	}
	for _, child := range takuzu.children {
		if child.leaf == nil {
			t.Fatalf("Takuzu child %q = %#v, want leaf", child.title, child)
		}
	}

	sudoku := requireCreateNode(t, state.tree, "Sudoku")
	if !hasCreateChild(sudoku, "Sudoku") || !hasCreateChild(sudoku, "Sudoku RGB") {
		t.Fatalf("Sudoku group children = %#v, want Sudoku and Sudoku RGB", childTitles(sudoku))
	}
	for _, child := range sudoku.children {
		if child.leaf == nil {
			t.Fatalf("Sudoku child %q = %#v, want leaf", child.title, child)
		}
	}
}

func TestCreateTreePromotesSingleVariantGameToTopLevelLeaf(t *testing.T) {
	state := newCreateState(config.CreateConfig{}, 80)
	nonogram := requireCreateNode(t, state.tree, "Nonogram")

	if nonogram.leaf == nil {
		t.Fatalf("Nonogram node = %#v, want top-level leaf", nonogram)
	}
	if len(nonogram.children) != 0 {
		t.Fatalf("Nonogram children = %#v, want none", childTitles(nonogram))
	}
}

func TestCreateResolveLeafVariantUsesSelectedVariant(t *testing.T) {
	state := newCreateState(config.CreateConfig{}, 80)
	nonogram := requireCreateNode(t, state.tree, "Nonogram")
	leaf := nonogram.leaf
	if leaf == nil {
		t.Fatal("Nonogram node is not a leaf")
	}

	_, modeTitle, err := state.resolveLeafVariant(*leaf, difficulty.Elo(700))
	if err != nil {
		t.Fatal(err)
	}
	if modeTitle != "Nonogram" {
		t.Fatalf("modeTitle = %q, want Nonogram", modeTitle)
	}
}

func TestCreateStateDescendantLeafCountsIncludeCollapsedDescendants(t *testing.T) {
	state := newCreateState(config.CreateConfig{}, 80)
	nonogram := requireCreateNode(t, state.tree, "Nonogram")
	state.checked[nonogram.leaf.id] = true

	selected, total := state.descendantLeafCounts(nonogram)
	if selected != 1 || total != 1 {
		t.Fatalf("descendant counts = %d/%d, want 1/1", selected, total)
	}
}

func TestCreateStateToggleDescendantLeavesSelectsAllWhenAnyUnchecked(t *testing.T) {
	state := newCreateState(config.CreateConfig{}, 80)
	sudoku := requireCreateNode(t, state.tree, "Sudoku")
	state.checked[firstDescendantLeafID(t, sudoku.children[0])] = true

	state.toggleDescendantLeaves(sudoku)

	selected, total := state.descendantLeafCounts(sudoku)
	if selected != total || total == 0 {
		t.Fatalf("descendant counts after select-all = %d/%d, want all selected", selected, total)
	}
}

func TestCreateStateToggleDescendantLeavesClearsFullySelectedParent(t *testing.T) {
	state := newCreateState(config.CreateConfig{}, 80)
	nonogram := requireCreateNode(t, state.tree, "Nonogram")
	state.toggleDescendantLeaves(nonogram)

	state.toggleDescendantLeaves(nonogram)

	selected, total := state.descendantLeafCounts(nonogram)
	if selected != 0 || total == 0 {
		t.Fatalf("descendant counts after clear = %d/%d, want 0 selected", selected, total)
	}
}

func TestCreateStateToggleDescendantLeavesWorksForGroupedLeaf(t *testing.T) {
	state := newCreateState(config.CreateConfig{}, 80)
	takuzu := requireCreateNode(t, state.tree, "Takuzu")
	takuzuPlus := requireCreateNode(t, takuzu.children, "Takuzu+")

	state.toggleDescendantLeaves(takuzuPlus)

	selected, total := state.descendantLeafCounts(takuzuPlus)
	if selected != total || total == 0 {
		t.Fatalf("Takuzu+ descendant counts after toggle = %d/%d, want all selected", selected, total)
	}
}

func requireCreateNode(t *testing.T, nodes []createTreeNode, title string) createTreeNode {
	t.Helper()
	for _, node := range nodes {
		if node.title == title {
			return node
		}
	}
	t.Fatalf("missing create node %q in %#v", title, childTitles(createTreeNode{children: nodes}))
	return createTreeNode{}
}

func hasCreateChild(node createTreeNode, title string) bool {
	for _, child := range node.children {
		if child.title == title {
			return true
		}
	}
	return false
}

func childTitles(node createTreeNode) []string {
	titles := make([]string, 0, len(node.children))
	for _, child := range node.children {
		titles = append(titles, child.title)
	}
	return titles
}

func firstDescendantLeafID(t *testing.T, node createTreeNode) string {
	t.Helper()
	if node.leaf != nil {
		return node.leaf.id
	}
	for _, child := range node.children {
		if id := firstDescendantLeafID(t, child); id != "" {
			return id
		}
	}
	t.Fatalf("node %q has no descendant leaf", node.title)
	return ""
}
