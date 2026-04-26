package app

import (
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/FelineStateMachine/puzzletea/config"
	"github.com/FelineStateMachine/puzzletea/difficulty"
	"github.com/FelineStateMachine/puzzletea/game"
	"github.com/FelineStateMachine/puzzletea/puzzle"
	"github.com/FelineStateMachine/puzzletea/registry"

	"charm.land/bubbles/v2/textinput"
)

const defaultCreateElo = 1200

type createState struct {
	tree      []createTreeNode
	checked   map[string]bool
	expanded  map[string]bool
	cursor    int
	eloInput  textinput.Model
	seedInput textinput.Model
	focus     createFocus
}

type createFocus int

const (
	createFocusTree createFocus = iota
	createFocusElo
	createFocusSeed
	createFocusGenerate
)

type createTreeNode struct {
	id       string
	title    string
	leaf     *createLeaf
	children []createTreeNode
}

type createLeaf struct {
	id       string
	title    string
	gameType string
	variant  registry.VariantEntry
}

type createVisibleNode struct {
	node  createTreeNode
	depth int
}

func newCreateState(cfg config.CreateConfig, width int) createState {
	elo := cfg.Elo
	if err := difficulty.ValidateElo(difficulty.Elo(elo)); err != nil {
		elo = defaultCreateElo
	}

	eloInput := textinput.New()
	eloInput.Prompt = ""
	eloInput.CharLimit = 4
	eloInput.SetWidth(8)
	eloInput.SetValue(strconv.Itoa(elo))

	seedInput := textinput.New()
	seedInput.Prompt = ""
	seedInput.Placeholder = "optional seed"
	seedInput.CharLimit = 64
	seedInput.SetWidth(min(width, 48))

	tree := buildCreateTree(registry.Entries())
	validLeafIDs := createLeafIDSet(tree)
	state := createState{
		tree:      tree,
		checked:   make(map[string]bool),
		expanded:  make(map[string]bool),
		eloInput:  eloInput,
		seedInput: seedInput,
		focus:     createFocusTree,
	}
	for _, id := range cfg.SelectedLeafIDs {
		if !validLeafIDs[id] {
			continue
		}
		state.checked[id] = true
		state.expandAncestors(id)
	}
	if len(state.checked) == 0 {
		state.expandTopLevel()
	}
	return state
}

func buildCreateTree(entries []registry.Entry) []createTreeNode {
	builders := make(map[string]*createNodeBuilder)
	order := make([]string, 0, len(entries))
	for _, entry := range entries {
		topTitle, branchTitle := createGrouping(entry.Definition.Name)
		topID := "group:" + puzzle.NormalizeName(topTitle)
		builder, ok := builders[topID]
		if !ok {
			builder = &createNodeBuilder{id: topID, title: topTitle}
			builders[topID] = builder
			order = append(order, topID)
		}
		builder.addBranch(entry, branchTitle)
	}

	nodes := make([]createTreeNode, 0, len(order))
	for _, id := range order {
		nodes = append(nodes, builders[id].node())
	}
	return nodes
}

type createNodeBuilder struct {
	id       string
	title    string
	branches []createTreeNode
}

func (b *createNodeBuilder) addBranch(entry registry.Entry, branchTitle string) {
	leaves := createLeavesForEntry(entry, branchTitle)
	if branchTitle == "" {
		b.branches = append(b.branches, leaves...)
		return
	}

	branchID := b.id + "/branch:" + puzzle.NormalizeName(branchTitle)
	if len(leaves) == 1 {
		b.branches = append(b.branches, leaves[0])
		return
	}
	b.branches = append(b.branches, createTreeNode{
		id:       branchID,
		title:    branchTitle,
		children: leaves,
	})
}

func (b *createNodeBuilder) node() createTreeNode {
	if len(b.branches) == 1 && b.branches[0].leaf != nil {
		return b.branches[0]
	}
	return createTreeNode{id: b.id, title: b.title, children: b.branches}
}

func createGrouping(gameType string) (topTitle, branchTitle string) {
	switch gameType {
	case "Takuzu", "Takuzu+":
		return "Takuzu", gameType
	case "Sudoku", "Sudoku RGB":
		return "Sudoku", gameType
	default:
		return gameType, ""
	}
}

func createLeavesForEntry(entry registry.Entry, branchTitle string) []createTreeNode {
	nodes := make([]createTreeNode, 0, len(entry.Variants))
	for _, variant := range entry.Variants {
		if variant.Elo == nil {
			continue
		}
		title := variant.Definition.Title
		id := createLeafID(entry.Definition.ID, branchTitle, title)
		nodes = append(nodes, createTreeNode{
			id:    id,
			title: title,
			leaf: &createLeaf{
				id:       id,
				title:    title,
				gameType: entry.Definition.Name,
				variant:  variant,
			},
		})
	}
	return nodes
}

func createLeafID(gameID puzzle.GameID, branchTitle, title string) string {
	parts := []string{string(gameID)}
	if branchTitle != "" {
		parts = append(parts, puzzle.NormalizeName(branchTitle))
	}
	parts = append(parts, puzzle.NormalizeName(title))
	return strings.Join(parts, "/")
}

func (s createState) selectedLeaves() []createLeaf {
	leaves := make([]createLeaf, 0, len(s.checked))
	var walk func([]createTreeNode)
	walk = func(nodes []createTreeNode) {
		for _, node := range nodes {
			if node.leaf != nil && s.checked[node.leaf.id] {
				leaves = append(leaves, *node.leaf)
			}
			walk(node.children)
		}
	}
	walk(s.tree)
	return leaves
}

func (s createState) selectedLeafIDs() []string {
	leaves := s.selectedLeaves()
	ids := make([]string, 0, len(leaves))
	for _, leaf := range leaves {
		ids = append(ids, leaf.id)
	}
	sort.Strings(ids)
	return ids
}

func (s createState) selectedCount() int {
	return len(s.selectedLeaves())
}

func (s createState) descendantLeafCounts(node createTreeNode) (selected, total int) {
	var walk func(createTreeNode)
	walk = func(node createTreeNode) {
		if node.leaf != nil {
			total++
			if s.checked[node.leaf.id] {
				selected++
			}
			return
		}
		for _, child := range node.children {
			walk(child)
		}
	}
	walk(node)
	return selected, total
}

func (s *createState) toggleDescendantLeaves(node createTreeNode) {
	selected, total := s.descendantLeafCounts(node)
	if total == 0 {
		return
	}
	checked := selected != total
	var walk func(createTreeNode)
	walk = func(node createTreeNode) {
		if node.leaf != nil {
			s.checked[node.leaf.id] = checked
			return
		}
		for _, child := range node.children {
			walk(child)
		}
	}
	walk(node)
}

func (s createState) seedEnabled() bool {
	return s.selectedCount() == 1
}

func (s createState) elo() (difficulty.Elo, error) {
	value, err := strconv.Atoi(strings.TrimSpace(s.eloInput.Value()))
	if err != nil {
		return 0, fmt.Errorf("Elo must be a number")
	}
	elo := difficulty.Elo(value)
	if err := difficulty.ValidateElo(elo); err != nil {
		return 0, err
	}
	return elo, nil
}

func (s createState) visibleNodes() []createVisibleNode {
	visible := make([]createVisibleNode, 0)
	var walk func([]createTreeNode, int)
	walk = func(nodes []createTreeNode, depth int) {
		for _, node := range nodes {
			visible = append(visible, createVisibleNode{node: node, depth: depth})
			if node.leaf == nil && s.expanded[node.id] {
				walk(node.children, depth+1)
			}
		}
	}
	walk(s.tree, 0)
	return visible
}

func (s *createState) expandTopLevel() {
	for _, node := range s.tree {
		s.expanded[node.id] = false
	}
}

func (s *createState) expandAncestors(leafID string) {
	var walk func([]createTreeNode, []string) bool
	walk = func(nodes []createTreeNode, ancestors []string) bool {
		for _, node := range nodes {
			if node.id == leafID {
				for _, id := range ancestors {
					s.expanded[id] = true
				}
				return true
			}
			if walk(node.children, append(ancestors, node.id)) {
				return true
			}
		}
		return false
	}
	walk(s.tree, nil)
}

func (s createState) resolveLeafVariant(leaf createLeaf, elo difficulty.Elo) (game.EloSpawner, string, error) {
	if leaf.variant.Elo == nil {
		return nil, "", fmt.Errorf("%s %s does not support Elo generation", leaf.gameType, leaf.title)
	}
	return leaf.variant.Elo, leaf.variant.Definition.Title, nil
}

func createLeafIDSet(nodes []createTreeNode) map[string]bool {
	ids := make(map[string]bool)
	var walk func([]createTreeNode)
	walk = func(nodes []createTreeNode) {
		for _, node := range nodes {
			if node.leaf != nil {
				ids[node.leaf.id] = true
				continue
			}
			walk(node.children)
		}
	}
	walk(nodes)
	return ids
}
