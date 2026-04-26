package app

import (
	"strconv"
	"testing"

	"charm.land/bubbles/v2/textinput"
	"github.com/FelineStateMachine/puzzletea/config"
	"github.com/FelineStateMachine/puzzletea/difficulty"
	"github.com/FelineStateMachine/puzzletea/game"
	"github.com/FelineStateMachine/puzzletea/puzzle"
	"github.com/FelineStateMachine/puzzletea/registry"
	sessionflow "github.com/FelineStateMachine/puzzletea/session"
	"github.com/FelineStateMachine/puzzletea/store"
)

type createTestEloSpawner struct{}

func (createTestEloSpawner) SpawnElo(string, difficulty.Elo) (game.Gamer, difficulty.Report, error) {
	return &escapeTrackingGame{}, difficulty.Report{
		TargetElo:  1200,
		ActualElo:  1200,
		Confidence: difficulty.ConfidenceHigh,
	}, nil
}

func TestCreateGenerateBlocksWithoutSelection(t *testing.T) {
	m := model{create: createTestState(nil, 1200), store: openAppTestStore(t), cfg: config.Default()}

	got, _ := m.handleCreateGenerate()
	if got.notice.message == "" {
		t.Fatal("expected notice for empty selection")
	}
}

func TestCreateGenerateUsesSingleSelectedLeafAndSeed(t *testing.T) {
	s := openAppTestStore(t)
	state := createTestState([]createLeaf{createTestLeaf("a", "Alpha", "Tiny")}, 1300)
	state.seedInput.SetValue("daily create seed")
	m := model{create: state, store: s, cfg: config.Default()}

	got, cmd := m.handleCreateGenerate()
	if cmd == nil {
		t.Fatal("expected generation command")
	}
	if got.session.spawn == nil {
		t.Fatal("missing spawn metadata")
	}
	if got.session.spawn.gameType != "Alpha" || got.session.spawn.modeTitle != "Tiny" {
		t.Fatalf("spawn = %#v, want Alpha/Tiny", got.session.spawn)
	}
	if got.session.spawn.run.Kind != store.RunKindSeeded {
		t.Fatalf("run kind = %q, want seeded", got.session.spawn.run.Kind)
	}
}

func TestCreateGenerateMultipleSelectionIgnoresSeedAndUsesInjectedRandomChoice(t *testing.T) {
	s := openAppTestStore(t)
	state := createTestState([]createLeaf{
		createTestLeaf("a", "Alpha", "Tiny"),
		createTestLeaf("b", "Beta", "Large"),
	}, 1300)
	state.seedInput.SetValue("ignored")
	m := model{
		create: state,
		store:  s,
		cfg:    config.Default(),
		createRandIndex: func(n int) int {
			if n != 2 {
				t.Fatalf("random choice bound = %d, want 2", n)
			}
			return 1
		},
	}

	got, cmd := m.handleCreateGenerate()
	if cmd == nil {
		t.Fatal("expected generation command")
	}
	if got.session.spawn.gameType != "Beta" {
		t.Fatalf("spawn gameType = %q, want Beta", got.session.spawn.gameType)
	}
	if got.session.spawn.run.Kind != store.RunKindNormal {
		t.Fatalf("run kind = %q, want normal", got.session.spawn.run.Kind)
	}
	if got.session.spawn.run.SeedText != "" {
		t.Fatalf("seed text = %q, want blank", got.session.spawn.run.SeedText)
	}
}

func TestCreateGenerateSeededSingleLeafResumesExistingRecord(t *testing.T) {
	s := openAppTestStore(t)
	leaf := createTestLeaf("lights", "Lights Out", "Easy")
	state := createTestState([]createLeaf{leaf}, 1000)
	state.seedInput.SetValue("resume create")
	name := sessionflow.SeededNameForCreateLeaf("resume create", leaf.gameType, leaf.id, 1000)
	rec := &store.GameRecord{
		Name:         name,
		GameType:     "Lights Out",
		Mode:         "Easy",
		InitialState: `{"grid":[[true]],"mode_title":"Easy"}`,
		SaveState:    `{"grid":[[true]],"mode_title":"Easy"}`,
		Status:       store.StatusAbandoned,
		RunKind:      store.RunKindSeeded,
		SeedText:     "resume create",
	}
	if err := s.CreateGame(rec); err != nil {
		t.Fatal(err)
	}
	m := model{create: state, store: s, cfg: config.Default()}

	got, cmd := m.handleCreateGenerate()
	if cmd != nil {
		t.Fatal("expected resume without generation command")
	}
	if got.state != gameView {
		t.Fatalf("state = %d, want gameView", got.state)
	}
	saved, err := s.GetDailyGame(name)
	if err != nil {
		t.Fatal(err)
	}
	if saved.Status != store.StatusInProgress {
		t.Fatalf("status = %q, want in_progress", saved.Status)
	}
}

func createTestState(leaves []createLeaf, elo int) createState {
	eloInput := textinput.New()
	eloInput.SetValue("1200")
	if elo > 0 {
		eloInput.SetValue(strconv.Itoa(elo))
	}
	seedInput := textinput.New()

	children := make([]createTreeNode, 0, len(leaves))
	checked := make(map[string]bool, len(leaves))
	for _, leaf := range leaves {
		leaf := leaf
		children = append(children, createTreeNode{id: leaf.id, title: leaf.title, leaf: &leaf})
		checked[leaf.id] = true
	}

	return createState{
		tree: []createTreeNode{{
			id:       "group:test",
			title:    "Test",
			children: children,
		}},
		checked:   checked,
		expanded:  map[string]bool{"group:test": true},
		eloInput:  eloInput,
		seedInput: seedInput,
	}
}

func createTestLeaf(id, gameType, modeTitle string) createLeaf {
	spawner := createTestEloSpawner{}
	return createLeaf{
		id:        id,
		title:     modeTitle,
		gameType:  gameType,
		modeTitle: modeTitle,
		modes: []registry.ModeEntry{{
			Definition: puzzle.ModeDef{
				Title:     modeTitle,
				PresetElo: ptrElo(1200),
			},
			Elo: spawner,
		}},
	}
}

func ptrElo(elo difficulty.Elo) *difficulty.Elo {
	return &elo
}
