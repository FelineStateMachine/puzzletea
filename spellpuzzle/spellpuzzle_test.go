package spellpuzzle

import (
	"encoding/json"
	"math/rand/v2"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
)

func newTestRNG(seed uint64) *rand.Rand {
	return rand.New(rand.NewPCG(seed, seed+1))
}

func testPuzzle(t *testing.T) Model {
	t.Helper()

	model := Model{
		modeTitle: "Test",
		bank:      []rune("MATES"),
		placements: []WordPlacement{
			{Text: "MATE", Start: Position{X: 0, Y: 0}, Orientation: Horizontal},
			{Text: "META", Start: Position{X: 0, Y: 0}, Orientation: Vertical},
		},
		keys:      DefaultKeyMap,
		hoverTile: -1,
	}
	model.rebuildDerivedState()
	return model
}

func TestNewPuzzleStartsWithoutBankCursor(t *testing.T) {
	m := testPuzzle(t)
	if m.bankCursor != -1 {
		t.Fatalf("bank cursor = %d, want -1", m.bankCursor)
	}
	if len(m.bankColorSlots) != len(m.bank) {
		t.Fatalf("bank color slot count = %d, want %d", len(m.bankColorSlots), len(m.bank))
	}
}

func TestGeneratePuzzleSeededDeterministic(t *testing.T) {
	mode := NewMode("Easy", "", 7, 6, 4)
	rngA := newTestRNG(12)
	rngB := newTestRNG(12)

	puzzleA, err := GeneratePuzzleSeeded(mode, rngA)
	if err != nil {
		t.Fatal(err)
	}
	puzzleB, err := GeneratePuzzleSeeded(mode, rngB)
	if err != nil {
		t.Fatal(err)
	}

	if puzzleA.Bank != puzzleB.Bank {
		t.Fatalf("bank mismatch: %q vs %q", puzzleA.Bank, puzzleB.Bank)
	}
	if len(puzzleA.Placements) != len(puzzleB.Placements) {
		t.Fatalf("placement count mismatch: %d vs %d", len(puzzleA.Placements), len(puzzleB.Placements))
	}
	for i := range puzzleA.Placements {
		if puzzleA.Placements[i] != puzzleB.Placements[i] {
			t.Fatalf("placement[%d] mismatch: %+v vs %+v", i, puzzleA.Placements[i], puzzleB.Placements[i])
		}
	}
}

func TestGeneratedWordsRespectBankAndBonusMinimum(t *testing.T) {
	modes := []SpellPuzzleMode{
		NewMode("Beginner", "", 6, 4, 3),
		NewMode("Easy", "", 7, 6, 4),
		NewMode("Medium", "", 8, 8, 6),
		NewMode("Hard", "", 9, 9, 8),
	}

	for _, mode := range modes {
		rng := newTestRNG(uint64(mode.BankSize * 11))
		puzzle, err := GeneratePuzzleSeeded(mode, rng)
		if err != nil {
			t.Fatalf("%s: %v", mode.Title(), err)
		}
		if len(puzzle.Placements) != mode.BoardWordCount {
			t.Fatalf("%s placement count = %d, want %d", mode.Title(), len(puzzle.Placements), mode.BoardWordCount)
		}
		seenPlacements := make(map[WordPlacement]struct{}, len(puzzle.Placements))
		for _, placement := range puzzle.Placements {
			if _, ok := seenPlacements[placement]; ok {
				t.Fatalf("%s contains duplicate placement %+v", mode.Title(), placement)
			}
			seenPlacements[placement] = struct{}{}
		}
		if len(puzzle.BonusWords) < mode.MinBonusWords {
			t.Fatalf("%s bonus count = %d, want at least %d", mode.Title(), len(puzzle.BonusWords), mode.MinBonusWords)
		}
		for _, placement := range puzzle.Placements {
			if !canSpellWord(placement.Text, countLetters(puzzle.Bank)) {
				t.Fatalf("%s placement %q cannot be spelled from bank %q", mode.Title(), placement.Text, puzzle.Bank)
			}
		}
	}
}

func TestGeneratedBankIsScrambledAwayFromPlacedWords(t *testing.T) {
	mode := NewMode("Easy", "", 7, 6, 4)
	puzzle, err := GeneratePuzzleSeeded(mode, newTestRNG(21))
	if err != nil {
		t.Fatal(err)
	}

	for _, placement := range puzzle.Placements {
		if puzzle.Bank == placement.Text {
			t.Fatalf("bank %q matches placed word %q", puzzle.Bank, placement.Text)
		}
		if puzzle.Bank == reverseWord(placement.Text) {
			t.Fatalf("bank %q matches reversed placed word %q", puzzle.Bank, placement.Text)
		}
	}
}

func TestGeneratedCrosswordIsConnectedAndIntersecting(t *testing.T) {
	mode := NewMode("Medium", "", 8, 8, 6)
	puzzle, err := GeneratePuzzleSeeded(mode, newTestRNG(99))
	if err != nil {
		t.Fatal(err)
	}

	seen := make(map[Position]int)
	for _, placement := range puzzle.Placements {
		for _, pos := range placement.Positions() {
			seen[pos]++
		}
	}

	intersectionCount := 0
	for _, count := range seen {
		if count > 1 {
			intersectionCount++
		}
	}
	if intersectionCount == 0 {
		t.Fatal("expected at least one crossword intersection")
	}

	board := buildBoard(puzzle.Placements)
	if board.Width == 0 || board.Height == 0 {
		t.Fatal("expected non-empty board")
	}
}

func TestBoardEdgesOutlineWordZones(t *testing.T) {
	m := testPuzzle(t)

	if !m.board.hasVerticalEdge(0, 0) {
		t.Fatal("expected left outer edge for occupied word cell")
	}
	if !m.board.hasHorizontalEdge(0, 0) {
		t.Fatal("expected top outer edge for occupied word cell")
	}
	if m.board.hasVerticalEdge(1, 0) {
		t.Fatal("expected no edge inside the horizontal word run")
	}
	if m.board.hasHorizontalEdge(0, 1) {
		t.Fatal("expected no edge inside the vertical word run")
	}
}

func TestSubmitBoardWordAndBonusWord(t *testing.T) {
	m := testPuzzle(t)
	m.trace = []int{0, 1, 2, 3}
	m.bankCursor = 3
	m.submitTrace()
	if !m.placements[m.boardIndex["MATE"][0]].Found {
		t.Fatal("expected MATE to be revealed")
	}
	if m.feedback.Message != "word added to board" {
		t.Fatalf("feedback = %q", m.feedback.Message)
	}
	if m.bankCursor != -1 {
		t.Fatalf("bank cursor after submit = %d, want -1", m.bankCursor)
	}

	m.clearFeedback()
	m.trace = []int{0, 3, 4, 1}
	m.submitTrace()
	if len(m.bonusWords) != 1 || m.bonusWords[0] != "MESA" {
		t.Fatalf("bonus words = %v, want [MESA]", m.bonusWords)
	}
	if m.feedback.Message != "bonus word counted" {
		t.Fatalf("feedback = %q", m.feedback.Message)
	}
}

func TestSubmitNotAWordAndAlreadyFound(t *testing.T) {
	m := testPuzzle(t)
	m.trace = []int{0, 0, 0}
	m.submitTrace()
	if m.feedback.Message != "not a word" {
		t.Fatalf("feedback = %q, want not a word", m.feedback.Message)
	}

	m.trace = []int{0, 1, 2, 3}
	m.submitTrace()
	m.clearFeedback()
	m.trace = []int{0, 1, 2, 3}
	m.submitTrace()
	if m.feedback.Message != "word already found" {
		t.Fatalf("feedback = %q, want word already found", m.feedback.Message)
	}
}

func TestKeyboardInputAndSolveState(t *testing.T) {
	m := testPuzzle(t)

	next, _ := m.Update(tea.KeyPressMsg{Code: 'm', Text: "m"})
	got := next.(Model)
	if got.traceWord() != "M" {
		t.Fatalf("trace after typing m = %q, want M", got.traceWord())
	}

	next, _ = got.Update(tea.KeyPressMsg{Code: 'a', Text: "a"})
	got = next.(Model)
	next, _ = got.Update(tea.KeyPressMsg{Code: 't', Text: "t"})
	got = next.(Model)
	next, _ = got.Update(tea.KeyPressMsg{Code: 'e', Text: "e"})
	got = next.(Model)
	next, _ = got.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	got = next.(Model)
	if got.feedback.Message != "word added to board" {
		t.Fatalf("submit feedback = %q", got.feedback.Message)
	}

	next, _ = got.Update(tea.KeyPressMsg{Code: 'm', Text: "m"})
	got = next.(Model)
	next, _ = got.Update(tea.KeyPressMsg{Code: 'e', Text: "e"})
	got = next.(Model)
	next, _ = got.Update(tea.KeyPressMsg{Code: 't', Text: "t"})
	got = next.(Model)
	next, _ = got.Update(tea.KeyPressMsg{Code: 'a', Text: "a"})
	got = next.(Model)
	next, _ = got.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	got = next.(Model)
	if !got.solved {
		t.Fatal("expected puzzle to be solved after both board words")
	}
}

func TestKeyboardDisallowsLettersOutsideBank(t *testing.T) {
	m := testPuzzle(t)

	next, _ := m.Update(tea.KeyPressMsg{Code: 'z', Text: "z"})
	got := next.(Model)
	if got.traceWord() != "" {
		t.Fatalf("trace after typing z = %q, want empty", got.traceWord())
	}

	next, _ = got.Update(tea.KeyPressMsg{Code: 'm', Text: "m"})
	got = next.(Model)
	next, _ = got.Update(tea.KeyPressMsg{Code: 'm', Text: "m"})
	got = next.(Model)
	if got.traceWord() != "M" {
		t.Fatalf("trace after reusing unavailable m = %q, want M", got.traceWord())
	}
}

func TestBackspaceClearsBankCursorWhenTraceBecomesEmpty(t *testing.T) {
	m := testPuzzle(t)

	next, _ := m.Update(tea.KeyPressMsg{Code: 'm', Text: "m"})
	got := next.(Model)
	if got.bankCursor != 0 {
		t.Fatalf("bank cursor after typing = %d, want 0", got.bankCursor)
	}

	next, _ = got.Update(tea.KeyPressMsg{Code: tea.KeyBackspace})
	got = next.(Model)
	if got.bankCursor != -1 {
		t.Fatalf("bank cursor after clearing trace = %d, want -1", got.bankCursor)
	}
}

func TestSolvedStateOnlyAllowsCursorMovement(t *testing.T) {
	m := testPuzzle(t)
	m.solved = true

	next, _ := m.Update(tea.KeyPressMsg{Code: 'm', Text: "m"})
	got := next.(Model)
	if got.traceWord() != "" {
		t.Fatalf("trace in solved state = %q, want empty", got.traceWord())
	}

	next, _ = got.Update(tea.KeyPressMsg{Code: '1', Text: "1"})
	got = next.(Model)
	if string(got.bank) != string(m.bank) {
		t.Fatalf("bank changed in solved state: %q vs %q", string(got.bank), string(m.bank))
	}

	next, _ = got.Update(tea.KeyPressMsg{Code: tea.KeyRight})
	got = next.(Model)
	if got.bankCursor != 0 {
		t.Fatalf("bank cursor after right in solved state = %d, want 0", got.bankCursor)
	}
}

func TestShuffleReordersVisibleBankAndPreservesTrace(t *testing.T) {
	m := testPuzzle(t)
	m.trace = []int{0, 1}
	beforeTrace := m.traceWord()
	beforeBank := string(m.bank)
	beforeSlots := append([]int(nil), m.bankColorSlots...)
	beforePairs := make(map[string]int, len(m.bank))
	for i, letter := range m.bank {
		beforePairs[string(letter)+":"+string(rune('0'+beforeSlots[i]))]++
	}

	next, _ := m.Update(tea.KeyPressMsg{Code: '1', Text: "1"})
	got := next.(Model)

	if string(got.bank) == beforeBank {
		t.Fatalf("bank did not change: %q", string(got.bank))
	}
	if got.traceWord() != beforeTrace {
		t.Fatalf("trace after shuffle = %q, want %q", got.traceWord(), beforeTrace)
	}
	afterPairs := make(map[string]int, len(got.bank))
	for i, letter := range got.bank {
		afterPairs[string(letter)+":"+string(rune('0'+got.bankColorSlots[i]))]++
	}
	if len(afterPairs) != len(beforePairs) {
		t.Fatalf("bank color pair cardinality changed: %v vs %v", afterPairs, beforePairs)
	}
	for key, count := range beforePairs {
		if afterPairs[key] != count {
			t.Fatalf("bank color pairing for %q = %d, want %d", key, afterPairs[key], count)
		}
	}
}

func TestMouseTraceSubmission(t *testing.T) {
	m := testPuzzle(t)
	m.termWidth = 100
	m.termHeight = 40
	x0, y := m.bankOrigin()

	next, _ := m.Update(tea.MouseClickMsg{X: x0 + 1, Y: y, Button: tea.MouseLeft})
	got := next.(Model)
	next, _ = got.Update(tea.MouseMotionMsg{X: x0 + 5, Y: y, Button: tea.MouseLeft})
	got = next.(Model)
	next, _ = got.Update(tea.MouseMotionMsg{X: x0 + 9, Y: y, Button: tea.MouseLeft})
	got = next.(Model)
	next, _ = got.Update(tea.MouseMotionMsg{X: x0 + 13, Y: y, Button: tea.MouseLeft})
	got = next.(Model)
	next, _ = got.Update(tea.MouseReleaseMsg{X: x0 + 13, Y: y, Button: tea.MouseLeft})
	got = next.(Model)
	if got.feedback.Message != "word added to board" {
		t.Fatalf("feedback = %q", got.feedback.Message)
	}
}

func TestMouseHoverDoesNotHighlightWithoutDrag(t *testing.T) {
	m := testPuzzle(t)
	m.termWidth = 100
	m.termHeight = 40
	x, y := m.bankOrigin()

	next, _ := m.Update(tea.MouseMotionMsg{X: x + 1, Y: y})
	got := next.(Model)
	if got.hoverTile != -1 {
		t.Fatalf("hover tile = %d, want -1 without dragging", got.hoverTile)
	}
}

func TestSolvedViewHidesTraceAndStatusRows(t *testing.T) {
	m := testPuzzle(t)
	m.solved = true

	view := m.View()
	if strings.Contains(view, "trace:") {
		t.Fatalf("solved view unexpectedly contains trace row: %q", view)
	}
	if strings.Contains(view, "type: add letters") {
		t.Fatalf("solved view unexpectedly contains status row: %q", view)
	}
}

func TestSaveLoadRoundTrip(t *testing.T) {
	m := testPuzzle(t)
	m.trace = []int{0, 1}
	m.bonusWords = []string{"MESA"}
	m.bankCursor = 1
	m.bankColorSlots = []int{3, 1, 4, 0, 2}
	m.rebuildDerivedState()

	data, err := m.GetSave()
	if err != nil {
		t.Fatal(err)
	}

	got, err := ImportModel(data)
	if err != nil {
		t.Fatal(err)
	}
	if string(got.bank) != string(m.bank) {
		t.Fatalf("bank = %q, want %q", string(got.bank), string(m.bank))
	}
	if len(got.bonusWords) != 1 || got.bonusWords[0] != "MESA" {
		t.Fatalf("bonusWords = %v", got.bonusWords)
	}
	if got.traceWord() != "MA" {
		t.Fatalf("trace = %q, want MA", got.traceWord())
	}
	if len(got.bankColorSlots) != len(m.bankColorSlots) {
		t.Fatalf("bankColorSlots len = %d, want %d", len(got.bankColorSlots), len(m.bankColorSlots))
	}
	for i := range got.bankColorSlots {
		if got.bankColorSlots[i] != m.bankColorSlots[i] {
			t.Fatalf("bankColorSlots[%d] = %d, want %d", i, got.bankColorSlots[i], m.bankColorSlots[i])
		}
	}
}

func TestImportWithoutTraceClearsIdleBankCursor(t *testing.T) {
	data, err := json.Marshal(Save{
		ModeTitle:  "Test",
		Bank:       "MATES",
		Placements: []WordPlacement{{Text: "MATE", Start: Position{X: 0, Y: 0}, Orientation: Horizontal}},
		BankCursor: 0,
	})
	if err != nil {
		t.Fatal(err)
	}

	got, err := ImportModel(data)
	if err != nil {
		t.Fatal(err)
	}
	if got.bankCursor != -1 {
		t.Fatalf("bank cursor after import = %d, want -1", got.bankCursor)
	}
}

func TestImportDeduplicatesDuplicatePlacements(t *testing.T) {
	data, err := json.Marshal(Save{
		ModeTitle: "opal-crane",
		Bank:      "PATACN",
		Placements: []WordPlacement{
			{Text: "CATNAP", Start: Position{X: 3, Y: 0}, Orientation: Vertical, Found: true},
			{Text: "CAPTAN", Start: Position{X: 0, Y: 2}, Orientation: Vertical, Found: true},
			{Text: "CANT", Start: Position{X: 0, Y: 2}, Orientation: Horizontal, Found: false},
			{Text: "CANT", Start: Position{X: 0, Y: 2}, Orientation: Horizontal, Found: true},
		},
	})
	if err != nil {
		t.Fatal(err)
	}

	got, err := ImportModel(data)
	if err != nil {
		t.Fatal(err)
	}
	if len(got.placements) != 3 {
		t.Fatalf("placement count after import = %d, want 3", len(got.placements))
	}
	if got.countWordsLeft() != 0 {
		t.Fatalf("words left = %d, want 0", got.countWordsLeft())
	}
	if !got.solved {
		t.Fatal("expected imported model to be solved after deduping duplicate placement")
	}
}

func TestCountBonusWordsTotal(t *testing.T) {
	m := testPuzzle(t)

	want := len(m.allowedWords) - len(m.boardIndex)
	if got := m.countBonusWordsTotal(); got != want {
		t.Fatalf("bonus total = %d, want %d", got, want)
	}
}

func TestImportModelEmptyData(t *testing.T) {
	data, _ := json.Marshal(Save{})
	got, err := ImportModel(data)
	if err != nil {
		t.Fatal(err)
	}
	if len(got.bank) != 0 {
		t.Fatalf("len(bank) = %d, want 0", len(got.bank))
	}
}
