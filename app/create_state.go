package app

import (
	"fmt"
	"regexp"
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

var boardSizePattern = regexp.MustCompile(`(?i)(\d+)\s*[x×]\s*(\d+)`)

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
	id        string
	title     string
	gameType  string
	modeTitle string
	modes     []registry.ModeEntry
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

	state := createState{
		tree:      buildCreateTree(registry.Entries()),
		checked:   make(map[string]bool),
		expanded:  make(map[string]bool),
		eloInput:  eloInput,
		seedInput: seedInput,
		focus:     createFocusTree,
	}
	for _, id := range cfg.SelectedLeafIDs {
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
	b.branches = append(b.branches, createTreeNode{
		id:       branchID,
		title:    branchTitle,
		children: leaves,
	})
}

func (b *createNodeBuilder) node() createTreeNode {
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
	groups := make(map[string][]registry.ModeEntry)
	order := make([]string, 0, len(entry.Modes))
	for _, mode := range entry.Modes {
		if mode.Elo == nil {
			continue
		}
		title := createLeafTitle(entry.Definition.Name, entry.Definition.Description, mode.Definition.Title, mode.Definition.Description)
		if _, ok := groups[title]; !ok {
			order = append(order, title)
		}
		groups[title] = append(groups[title], mode)
	}

	nodes := make([]createTreeNode, 0, len(order))
	for _, title := range order {
		id := createLeafID(entry.Definition.ID, branchTitle, title)
		modes := groups[title]
		nodes = append(nodes, createTreeNode{
			id:    id,
			title: title,
			leaf: &createLeaf{
				id:        id,
				title:     title,
				gameType:  entry.Definition.Name,
				modeTitle: modes[0].Definition.Title,
				modes:     modes,
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

func createLeafTitle(gameType, gameDescription, modeTitle, modeDescription string) string {
	if size := firstBoardSize(modeTitle); size != "" {
		return size
	}
	if size := firstBoardSize(modeDescription); size != "" {
		return size
	}
	if gameType == "Sudoku RGB" {
		return "9x9"
	}
	if size := firstBoardSize(gameDescription); size != "" {
		return size
	}
	return modeTitle
}

func firstBoardSize(s string) string {
	match := boardSizePattern.FindStringSubmatch(s)
	if len(match) != 3 {
		return ""
	}
	return match[1] + "x" + match[2]
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
	ids := make([]string, 0, len(s.checked))
	for id, checked := range s.checked {
		if checked {
			ids = append(ids, id)
		}
	}
	sort.Strings(ids)
	return ids
}

func (s createState) selectedCount() int {
	count := 0
	for _, checked := range s.checked {
		if checked {
			count++
		}
	}
	return count
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

func (s createState) resolveLeafMode(leaf createLeaf, elo difficulty.Elo) (game.EloSpawner, string, error) {
	if len(leaf.modes) == 0 {
		return nil, "", fmt.Errorf("no Elo-capable modes for %s", leaf.title)
	}
	best := leaf.modes[0]
	bestDistance := createPresetDistance(best, elo)
	for _, mode := range leaf.modes[1:] {
		distance := createPresetDistance(mode, elo)
		if distance < bestDistance {
			best = mode
			bestDistance = distance
		}
	}
	if best.Elo == nil {
		return nil, "", fmt.Errorf("%s %s does not support Elo generation", leaf.gameType, leaf.title)
	}
	return best.Elo, best.Definition.Title, nil
}

func createPresetDistance(mode registry.ModeEntry, elo difficulty.Elo) int {
	if mode.Definition.PresetElo == nil {
		return 0
	}
	delta := int(*mode.Definition.PresetElo - elo)
	if delta < 0 {
		return -delta
	}
	return delta
}
