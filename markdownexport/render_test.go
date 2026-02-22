package markdownexport

import (
	"errors"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/FelineStateMachine/puzzletea/game"
	"github.com/FelineStateMachine/puzzletea/sudoku"
	"github.com/FelineStateMachine/puzzletea/wordsearch"

	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
)

func TestSupportsGameType(t *testing.T) {
	if !SupportsGameType("Sudoku") {
		t.Fatal("expected sudoku to be supported")
	}
	if SupportsGameType("Lights Out") {
		t.Fatal("expected lights out to be unsupported")
	}
}

func TestRenderPuzzleSnippetUnsupported(t *testing.T) {
	_, err := RenderPuzzleSnippet("Lights Out", "", []byte(`{}`))
	if !errors.Is(err, ErrUnsupportedGame) {
		t.Fatalf("expected ErrUnsupportedGame, got %v", err)
	}
}

func TestRenderPuzzleSnippetSudokuUsesProvidedOnly(t *testing.T) {
	data := []byte(`{
		"grid":"500000000\n600000000\n000000000\n000000000\n000000000\n000000000\n000000000\n000000000\n000000000",
		"provided":[{"x":0,"y":0,"v":5}]
	}`)

	snippet, err := RenderPuzzleSnippet("Sudoku", "", data)
	if err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(snippet, "| 1 | 5 | . | . | . | . | . | . | . | . |") {
		t.Fatalf("expected given row in snippet, got:\n%s", snippet)
	}
	if !strings.Contains(snippet, "| 2 | . | . | . | . | . | . | . | . | . |") {
		t.Fatalf("expected user-entered row to be blank in snippet, got:\n%s", snippet)
	}
}

func TestRenderPuzzleSnippetTakuzuUsesProvidedOnly(t *testing.T) {
	data := []byte(`{
		"size":2,
		"state":"01\n10",
		"provided":"#.\n.#",
		"mode_title":"Test"
	}`)

	snippet, err := RenderPuzzleSnippet("Takuzu", "", data)
	if err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(snippet, "| 1 | 0 | . |") {
		t.Fatalf("expected first row to keep only provided value, got:\n%s", snippet)
	}
	if !strings.Contains(snippet, "| 2 | . | 0 |") {
		t.Fatalf("expected second row to keep only provided value, got:\n%s", snippet)
	}
}

func TestRenderPuzzleSnippetNonogramIntegratedHintsLayout(t *testing.T) {
	data := []byte(`{
		"width":3,
		"height":2,
		"row-hints":[[3],[1,1]],
		"col-hints":[[1],[2],[1]]
	}`)

	snippet, err := RenderPuzzleSnippet("Nonogram", "", data)
	if err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(snippet, "### Puzzle Grid with Integrated Hints") {
		t.Fatalf("expected integrated nonogram heading, got:\n%s", snippet)
	}
	if strings.Contains(snippet, "### Row Hints") || strings.Contains(snippet, "### Column Hints") {
		t.Fatalf("expected legacy split hint sections to be removed, got:\n%s", snippet)
	}
	if !strings.Contains(snippet, "| . | . | 1 | 2 | 1 |") {
		t.Fatalf("expected column hints row aligned above puzzle grid, got:\n%s", snippet)
	}
	if !strings.Contains(snippet, "| . | 3 | . | . | . |") {
		t.Fatalf("expected single row hint to be right-aligned near grid, got:\n%s", snippet)
	}
	if !strings.Contains(snippet, "| 1 | 1 | . | . | . |") {
		t.Fatalf("expected multi-value row hint to render beside puzzle row, got:\n%s", snippet)
	}
}

func TestRenderPuzzleSnippetNonogramColumnHintsBottomAligned(t *testing.T) {
	data := []byte(`{
		"width":2,
		"height":1,
		"row-hints":[[1]],
		"col-hints":[[2,1],[3]]
	}`)

	snippet, err := RenderPuzzleSnippet("Nonogram", "", data)
	if err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(snippet, "| . | 2 | . |") {
		t.Fatalf("expected top column hint row to pad shorter hints, got:\n%s", snippet)
	}
	if !strings.Contains(snippet, "| . | 1 | 3 |") {
		t.Fatalf("expected lower column hint row to sit closest to grid, got:\n%s", snippet)
	}
}

func TestDetectGameType(t *testing.T) {
	gameType, err := DetectGameType(sudoku.Model{})
	if err != nil {
		t.Fatal(err)
	}
	if gameType != "Sudoku" {
		t.Fatalf("gameType = %q, want %q", gameType, "Sudoku")
	}

	gameType, err = DetectGameType(&wordsearch.Model{})
	if err != nil {
		t.Fatal(err)
	}
	if gameType != "Word Search" {
		t.Fatalf("gameType = %q, want %q", gameType, "Word Search")
	}

	_, err = DetectGameType(testUnknownGamer{})
	if err == nil {
		t.Fatal("expected unknown gamer detection error")
	}
}

func TestBuildDocument(t *testing.T) {
	doc := BuildDocument(DocumentConfig{
		Version:       "v-test",
		Category:      "Sudoku",
		ModeSelection: "mixed modes",
		Count:         2,
		Seed:          "seed-1",
		GeneratedAt:   time.Date(2026, 2, 22, 10, 0, 0, 0, time.UTC),
	}, []PuzzleSection{
		{Index: 1, GameType: "Sudoku", Mode: "Easy", Body: "body-one"},
		{Index: 2, GameType: "Sudoku", Mode: "Hard", Body: "body-two"},
	})

	if !strings.Contains(doc, "# PuzzleTea Export") {
		t.Fatal("expected export title in markdown document")
	}
	if !strings.Contains(doc, "Version: v-test") {
		t.Fatal("expected version metadata in markdown document")
	}
	if matched := regexp.MustCompile(`## [a-z]+-[a-z]+ - 1`).MatchString(doc); !matched {
		t.Fatal("expected first puzzle heading in adjective-noun pattern")
	}
	if matched := regexp.MustCompile(`\n---\n\n## [a-z]+-[a-z]+ - 2`).MatchString(doc); !matched {
		t.Fatal("expected puzzle separator and second heading in adjective-noun pattern")
	}
	if strings.Contains(doc, "## Puzzle 1 - Sudoku (Easy)") {
		t.Fatal("expected legacy puzzle heading format to be removed")
	}
}

type testUnknownGamer struct{}

func (testUnknownGamer) GetDebugInfo() string                 { return "" }
func (testUnknownGamer) GetFullHelp() [][]key.Binding         { return nil }
func (testUnknownGamer) GetSave() ([]byte, error)             { return nil, nil }
func (testUnknownGamer) IsSolved() bool                       { return false }
func (testUnknownGamer) Reset() game.Gamer                    { return testUnknownGamer{} }
func (testUnknownGamer) SetTitle(string) game.Gamer           { return testUnknownGamer{} }
func (testUnknownGamer) Init() tea.Cmd                        { return nil }
func (testUnknownGamer) View() string                         { return "" }
func (testUnknownGamer) Update(tea.Msg) (game.Gamer, tea.Cmd) { return testUnknownGamer{}, nil }
