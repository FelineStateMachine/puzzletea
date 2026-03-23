package app

import (
	"path/filepath"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/FelineStateMachine/puzzletea/config"
	"github.com/FelineStateMachine/puzzletea/packexport"
	"github.com/FelineStateMachine/puzzletea/puzzle"
)

func TestHandleExportEnterLoadsPersistedDefaults(t *testing.T) {
	tmp := t.TempDir()
	t.Chdir(tmp)

	m := model{
		cfg: &config.Config{
			Export: config.ExportConfig{
				Title:         "Saved Sampler",
				Volume:        3,
				SheetLayout:   "duplex-booklet",
				PDFOutputPath: filepath.Join(tmp, "saved.pdf"),
				Counts: map[puzzle.GameID]map[puzzle.ModeID]int{
					puzzle.CanonicalGameID("Sudoku"): {
						puzzle.CanonicalModeID("Easy"): 4,
					},
				},
			},
		},
		width:  100,
		height: 30,
	}

	next, _ := m.handleExportEnter()
	got := next.(model)

	if got.state != exportView {
		t.Fatalf("state = %d, want %d", got.state, exportView)
	}
	if got.export.values.Title != "Saved Sampler" {
		t.Fatalf("title = %q, want %q", got.export.values.Title, "Saved Sampler")
	}
	if got.export.values.Volume != "3" {
		t.Fatalf("volume = %q, want %q", got.export.values.Volume, "3")
	}
	if got.export.cards[findExportCardIndex(got.export.cards, puzzle.CanonicalGameID("Sudoku"))].Buckets[0] != 4 {
		t.Fatal("expected persisted count to populate the form")
	}
	if !got.export.initialized {
		t.Fatal("expected export editor to be initialized")
	}
}

func TestHandleExportCompleteSuccessShowsSuccessNotice(t *testing.T) {
	tmp := t.TempDir()
	t.Chdir(tmp)

	m := model{
		state:  exportRunningView,
		width:  100,
		height: 30,
		export: exportState{
			jobID:       7,
			initialized: true,
			values:      exportValuesFromSpec(packexport.DefaultSpec(tmp)),
			cards:       buildExportCardsFromSpec(packexport.DefaultSpec(tmp)),
		},
	}

	next, _ := m.handleExportComplete(exportCompleteMsg{
		jobID: 7,
		result: packexport.Result{
			TotalCount:    10,
			PDFOutputPath: filepath.Join(tmp, "sample.pdf"),
		},
	})
	got := next.(model)

	if got.state != exportView {
		t.Fatalf("state = %d, want %d", got.state, exportView)
	}
	if got.notice.level != noticeLevelSuccess {
		t.Fatalf("notice level = %q, want %q", got.notice.level, noticeLevelSuccess)
	}
	if got.notice.message == "" {
		t.Fatal("expected success notice")
	}
	if strings.Contains(got.notice.message, ".jsonl") {
		t.Fatalf("notice message = %q, did not expect jsonl path when jsonl export is disabled", got.notice.message)
	}
}

func TestHandleExportCompleteSuccessIncludesJSONLWhenPresent(t *testing.T) {
	tmp := t.TempDir()
	t.Chdir(tmp)

	jsonlPath := filepath.Join(tmp, "sample.jsonl")
	m := model{
		state:  exportRunningView,
		width:  100,
		height: 30,
		export: exportState{
			jobID:       8,
			initialized: true,
			values:      exportValuesFromSpec(packexport.DefaultSpec(tmp)),
			cards:       buildExportCardsFromSpec(packexport.DefaultSpec(tmp)),
		},
	}

	next, _ := m.handleExportComplete(exportCompleteMsg{
		jobID: 8,
		result: packexport.Result{
			TotalCount:      10,
			PDFOutputPath:   filepath.Join(tmp, "sample.pdf"),
			JSONLOutputPath: jsonlPath,
		},
	})
	got := next.(model)

	if !strings.Contains(got.notice.message, jsonlPath) {
		t.Fatalf("notice message = %q, want jsonl output path", got.notice.message)
	}
}

func TestHandleExportCompleteFailureShowsErrorNotice(t *testing.T) {
	tmp := t.TempDir()
	t.Chdir(tmp)

	m := model{
		state:  exportRunningView,
		width:  100,
		height: 30,
		export: exportState{
			jobID:       9,
			initialized: true,
			values:      exportValuesFromSpec(packexport.DefaultSpec(tmp)),
			cards:       buildExportCardsFromSpec(packexport.DefaultSpec(tmp)),
		},
	}

	next, _ := m.handleExportComplete(exportCompleteMsg{
		jobID: 9,
		err:   assertiveError("boom"),
	})
	got := next.(model)

	if got.notice.level != noticeLevelError {
		t.Fatalf("notice level = %q, want %q", got.notice.level, noticeLevelError)
	}
	if got.state != exportView {
		t.Fatalf("state = %d, want %d", got.state, exportView)
	}
}

func TestHandleExportSubmitRejectsInvalidVolumeInput(t *testing.T) {
	tmp := t.TempDir()
	spec := packexport.DefaultSpec(tmp)
	state := buildInitialExportState(exportValuesFromSpec(spec), buildExportCardsFromSpec(spec), 160)
	state.volumeInput.SetValue("abc")

	m := model{
		state:  exportView,
		width:  100,
		height: 30,
		export: state,
	}

	next, _ := m.handleExportSubmit()
	got := next.(model)

	if got.state != exportView {
		t.Fatalf("state = %d, want %d", got.state, exportView)
	}
	if got.notice.level != noticeLevelError {
		t.Fatalf("notice level = %q, want %q", got.notice.level, noticeLevelError)
	}
	if !strings.Contains(got.notice.message, "volume") {
		t.Fatalf("notice message = %q, want volume validation error", got.notice.message)
	}
}

func TestExportHeaderInputUsesThreeLineTextarea(t *testing.T) {
	tmp := t.TempDir()
	spec := packexport.DefaultSpec(tmp)
	state := buildInitialExportState(exportValuesFromSpec(spec), buildExportCardsFromSpec(spec), 160)

	if got := state.headerInput.Height(); got != 3 {
		t.Fatalf("headerInput.Height() = %d, want 3", got)
	}
}

func TestExportCardLayoutTargetsDenseDesktop(t *testing.T) {
	tmp := t.TempDir()
	spec := packexport.DefaultSpec(tmp)
	state := buildInitialExportState(exportValuesFromSpec(spec), buildExportCardsFromSpec(spec), 190)

	if got := state.cardColumns(190); got != 5 {
		t.Fatalf("cardColumns(190) = %d, want 5", got)
	}
	if got := state.visibleCardRows(190, 34); got < 3 {
		t.Fatalf("visibleCardRows(190, 34) = %d, want at least 3", got)
	}
}

func TestRenderCardUsesCompactDifficultyPairs(t *testing.T) {
	tmp := t.TempDir()
	spec := packexport.DefaultSpec(tmp)
	state := buildInitialExportState(exportValuesFromSpec(spec), buildExportCardsFromSpec(spec), 190)
	state.focus = exportFocusCards
	state.cardIndex = findExportCardIndex(state.cards, puzzle.CanonicalGameID("Sudoku"))
	state.bucketIndex = 0

	rendered := state.renderCard(state.cardIndex, exportCardWidth(190, state.cardColumns(190)))

	if !strings.Contains(rendered, "▸") || !strings.Contains(rendered, "◂") {
		t.Fatal("expected selected difficulty game cursor markers in card")
	}
	if !strings.Contains(rendered, "E2") {
		t.Fatal("expected compact difficulty pair in card")
	}
}

func TestExportCardArrowNavigationTraversesBucketsAcrossGames(t *testing.T) {
	tmp := t.TempDir()
	spec := packexport.DefaultSpec(tmp)
	state := buildInitialExportState(exportValuesFromSpec(spec), buildExportCardsFromSpec(spec), 190)
	state.focus = exportFocusCards
	state.cardIndex = 0
	state.bucketIndex = 2

	if handled := state.handleNavigationKey(tea.KeyPressMsg{Code: tea.KeyRight}, 190, 30); !handled {
		t.Fatal("expected right arrow to be handled")
	}
	if state.cardIndex != 1 || state.bucketIndex != 0 {
		t.Fatalf("selection = (%d,%d), want (1,0)", state.cardIndex, state.bucketIndex)
	}

	if handled := state.handleNavigationKey(tea.KeyPressMsg{Code: tea.KeyLeft}, 190, 30); !handled {
		t.Fatal("expected left arrow to be handled")
	}
	if state.cardIndex != 0 || state.bucketIndex != 2 {
		t.Fatalf("selection = (%d,%d), want (0,2)", state.cardIndex, state.bucketIndex)
	}
}

func TestExportCardDigitsEditSelectedBucket(t *testing.T) {
	tmp := t.TempDir()
	spec := packexport.DefaultSpec(tmp)
	state := buildInitialExportState(exportValuesFromSpec(spec), buildExportCardsFromSpec(spec), 190)
	state.focus = exportFocusCards
	state.cardIndex = findExportCardIndex(state.cards, puzzle.CanonicalGameID("Nurikabe"))
	state.bucketIndex = 1

	state.cards[state.cardIndex].Buckets[state.bucketIndex] = 0
	if handled := state.handleNavigationKey(tea.KeyPressMsg{Code: '4', Text: "4"}, 190, 30); !handled {
		t.Fatal("expected digit key to be handled")
	}
	if got := state.cards[state.cardIndex].Buckets[state.bucketIndex]; got != 4 {
		t.Fatalf("bucket count = %d, want 4", got)
	}

	if handled := state.handleNavigationKey(tea.KeyPressMsg{Code: tea.KeyBackspace}, 190, 30); !handled {
		t.Fatal("expected backspace to be handled")
	}
	if got := state.cards[state.cardIndex].Buckets[state.bucketIndex]; got != 0 {
		t.Fatalf("bucket count after backspace = %d, want 0", got)
	}
}

func TestRenderSettingsShowsFullJSONLOffLabel(t *testing.T) {
	tmp := t.TempDir()
	spec := packexport.DefaultSpec(tmp)
	spec.JSONLOutputPath = ""
	state := buildInitialExportState(exportValuesFromSpec(spec), buildExportCardsFromSpec(spec), 190)

	rendered := state.renderSettings(190)
	if !strings.Contains(rendered, "JSONL") || !strings.Contains(rendered, "Off") {
		t.Fatal("expected JSONL toggle to render the full Off label in settings")
	}
}

func TestRenderSummaryIncludesSelectedValue(t *testing.T) {
	tmp := t.TempDir()
	spec := packexport.DefaultSpec(tmp)
	state := buildInitialExportState(exportValuesFromSpec(spec), buildExportCardsFromSpec(spec), 190)
	state.focus = exportFocusCards
	state.cardIndex = findExportCardIndex(state.cards, puzzle.CanonicalGameID("Sudoku"))
	state.bucketIndex = 0

	rendered := state.renderSummary(190)
	if !strings.Contains(rendered, "Sudoku / Easy = 2") {
		t.Fatal("expected summary to include selected game, difficulty, and value")
	}
}

func TestExportMouseClickOnSummaryButtonSubmits(t *testing.T) {
	tmp := t.TempDir()
	spec := packexport.DefaultSpec(tmp)
	state := buildInitialExportState(exportValuesFromSpec(spec), buildExportCardsFromSpec(spec), 190)

	screenWidth, screenHeight := 220, 40
	contentWidth, contentHeight := exportContentBounds(screenWidth, screenHeight)
	metrics := state.contentMetrics(contentWidth, contentHeight)
	buttonWidth := lipgloss.Width(renderChoiceChip("Export", false))
	_, _, contentX, contentY := exportPanelContentOrigin(screenWidth, screenHeight, state.view(contentWidth, contentHeight))

	click := tea.MouseClickMsg{
		X: contentX + metrics.Width - buttonWidth + 1,
		Y: contentY + metrics.SettingsHeight,
	}
	_, action, handled := state.handleMouseClick(click, screenWidth, screenHeight, contentWidth, contentHeight)
	if !handled {
		t.Fatal("expected export button click to be handled")
	}
	if _, ok := action.(exportSubmitAction); !ok {
		t.Fatalf("action = %T, want exportSubmitAction", action)
	}
}

func TestExportMouseClickFocusesTextFields(t *testing.T) {
	tmp := t.TempDir()
	spec := packexport.DefaultSpec(tmp)
	spec.Seed = "sample-seed"
	spec.JSONLOutputPath = filepath.Join(tmp, "out", "sample.jsonl")
	state := buildInitialExportState(exportValuesFromSpec(spec), buildExportCardsFromSpec(spec), 220)

	screenWidth, screenHeight := 220, 40
	contentWidth, contentHeight := exportContentBounds(screenWidth, screenHeight)
	metrics := state.contentMetrics(contentWidth, contentHeight)
	_, _, contentX, contentY := exportPanelContentOrigin(screenWidth, screenHeight, state.view(contentWidth, contentHeight))

	tests := []struct {
		name     string
		rect     exportRect
		focus    exportFocus
		position func() int
	}{
		{"title", metrics.Settings.Title, exportFocusTitle, func() int { return state.titleInput.Position() }},
		{"advert", metrics.Settings.Advert, exportFocusAdvert, func() int { return state.advertInput.Position() }},
		{"seed", metrics.Settings.Seed, exportFocusSeed, func() int { return state.seedInput.Position() }},
		{"pdf", metrics.Settings.PDFPath, exportFocusPDFPath, func() int { return state.pdfPathInput.Position() }},
		{"jsonl path", metrics.Settings.JSONLPath, exportFocusJSONLPath, func() int { return state.jsonlPathInput.Position() }},
	}

	for _, tt := range tests {
		click := tea.MouseClickMsg{
			X: contentX + tt.rect.X + min(tt.rect.W-2, 14),
			Y: contentY + tt.rect.Y + 1,
		}
		_, _, handled := state.handleMouseClick(click, screenWidth, screenHeight, contentWidth, contentHeight)
		if !handled {
			t.Fatalf("%s click was not handled", tt.name)
		}
		if state.focus != tt.focus {
			t.Fatalf("%s focus = %v, want %v", tt.name, state.focus, tt.focus)
		}
		if got := tt.position(); got <= 0 {
			t.Fatalf("%s cursor position = %d, want > 0", tt.name, got)
		}
	}
}

func TestExportMouseClickRepositionsTextCursor(t *testing.T) {
	tmp := t.TempDir()
	spec := packexport.DefaultSpec(tmp)
	state := buildInitialExportState(exportValuesFromSpec(spec), buildExportCardsFromSpec(spec), 220)

	screenWidth, screenHeight := 220, 40
	contentWidth, contentHeight := exportContentBounds(screenWidth, screenHeight)
	metrics := state.contentMetrics(contentWidth, contentHeight)
	_, _, contentX, contentY := exportPanelContentOrigin(screenWidth, screenHeight, state.view(contentWidth, contentHeight))

	first := tea.MouseClickMsg{
		X: contentX + metrics.Settings.Title.X + 8,
		Y: contentY + metrics.Settings.Title.Y + 1,
	}
	second := tea.MouseClickMsg{
		X: contentX + metrics.Settings.Title.X + 20,
		Y: contentY + metrics.Settings.Title.Y + 1,
	}
	_, _, handled := state.handleMouseClick(first, screenWidth, screenHeight, contentWidth, contentHeight)
	if !handled {
		t.Fatal("expected first title click to be handled")
	}
	firstPos := state.titleInput.Position()
	_, _, handled = state.handleMouseClick(second, screenWidth, screenHeight, contentWidth, contentHeight)
	if !handled {
		t.Fatal("expected second title click to be handled")
	}
	if state.titleInput.Position() <= firstPos {
		t.Fatalf("title cursor did not move forward: first=%d second=%d", firstPos, state.titleInput.Position())
	}
}

func TestExportMouseClickFocusesHeaderTextarea(t *testing.T) {
	tmp := t.TempDir()
	spec := packexport.DefaultSpec(tmp)
	spec.Header = "line one\nline two\nline three"
	state := buildInitialExportState(exportValuesFromSpec(spec), buildExportCardsFromSpec(spec), 220)

	screenWidth, screenHeight := 220, 40
	contentWidth, contentHeight := exportContentBounds(screenWidth, screenHeight)
	metrics := state.contentMetrics(contentWidth, contentHeight)
	_, _, contentX, contentY := exportPanelContentOrigin(screenWidth, screenHeight, state.view(contentWidth, contentHeight))

	click := tea.MouseClickMsg{
		X: contentX + metrics.Settings.Header.X + 18,
		Y: contentY + metrics.Settings.Header.Y + 4,
	}
	_, _, handled := state.handleMouseClick(click, screenWidth, screenHeight, contentWidth, contentHeight)
	if !handled {
		t.Fatal("expected header click to be handled")
	}
	if state.focus != exportFocusHeader {
		t.Fatalf("focus = %v, want exportFocusHeader", state.focus)
	}
	if got := state.headerInput.Line(); got != 2 {
		t.Fatalf("header line = %d, want 2", got)
	}
	if got := state.headerInput.LineInfo().CharOffset; got <= 0 {
		t.Fatalf("header char offset = %d, want > 0", got)
	}
}

func TestExportMouseClickTogglesJSONLImmediately(t *testing.T) {
	tmp := t.TempDir()
	spec := packexport.DefaultSpec(tmp)
	state := buildInitialExportState(exportValuesFromSpec(spec), buildExportCardsFromSpec(spec), 220)

	screenWidth, screenHeight := 220, 40
	contentWidth, contentHeight := exportContentBounds(screenWidth, screenHeight)
	metrics := state.contentMetrics(contentWidth, contentHeight)
	_, _, contentX, contentY := exportPanelContentOrigin(screenWidth, screenHeight, state.view(contentWidth, contentHeight))

	click := tea.MouseClickMsg{
		X: contentX + metrics.Settings.JSONL.X + 4,
		Y: contentY + metrics.Settings.JSONL.Y + 1,
	}
	_, _, handled := state.handleMouseClick(click, screenWidth, screenHeight, contentWidth, contentHeight)
	if !handled {
		t.Fatal("expected JSONL click to be handled")
	}
	if !state.values.JSONLEnabled {
		t.Fatal("expected JSONL click to toggle enabled")
	}
	if state.focus != exportFocusJSONLToggle {
		t.Fatalf("focus = %v, want exportFocusJSONLToggle", state.focus)
	}
}

func TestExportMouseClickCyclesLayoutImmediately(t *testing.T) {
	tmp := t.TempDir()
	spec := packexport.DefaultSpec(tmp)
	state := buildInitialExportState(exportValuesFromSpec(spec), buildExportCardsFromSpec(spec), 220)

	screenWidth, screenHeight := 220, 40
	contentWidth, contentHeight := exportContentBounds(screenWidth, screenHeight)
	metrics := state.contentMetrics(contentWidth, contentHeight)
	_, _, contentX, contentY := exportPanelContentOrigin(screenWidth, screenHeight, state.view(contentWidth, contentHeight))

	click := tea.MouseClickMsg{
		X: contentX + metrics.Settings.Layout.X + 5,
		Y: contentY + metrics.Settings.Layout.Y + 1,
	}
	_, _, handled := state.handleMouseClick(click, screenWidth, screenHeight, contentWidth, contentHeight)
	if !handled {
		t.Fatal("expected layout click to be handled")
	}
	if state.values.SheetLayout != "duplex-booklet" {
		t.Fatalf("layout = %q, want %q", state.values.SheetLayout, "duplex-booklet")
	}
}

func TestExportMouseClickSelectsCardBucketWithoutChangingCount(t *testing.T) {
	tmp := t.TempDir()
	spec := packexport.DefaultSpec(tmp)
	state := buildInitialExportState(exportValuesFromSpec(spec), buildExportCardsFromSpec(spec), 220)

	screenWidth, screenHeight := 220, 40
	contentWidth, contentHeight := exportContentBounds(screenWidth, screenHeight)
	metrics := state.contentMetrics(contentWidth, contentHeight)
	_, _, contentX, contentY := exportPanelContentOrigin(screenWidth, screenHeight, state.view(contentWidth, contentHeight))

	target := metrics.CardRects[1]
	before := state.cards[target.Index].Buckets[1]
	click := tea.MouseClickMsg{
		X: contentX + target.Buckets[1].X + 1,
		Y: contentY + target.Buckets[1].Y,
	}
	_, _, handled := state.handleMouseClick(click, screenWidth, screenHeight, contentWidth, contentHeight)
	if !handled {
		t.Fatal("expected card bucket click to be handled")
	}
	if state.cardIndex != target.Index || state.bucketIndex != 1 {
		t.Fatalf("selection = (%d,%d), want (%d,1)", state.cardIndex, state.bucketIndex, target.Index)
	}
	if got := state.cards[target.Index].Buckets[1]; got != before {
		t.Fatalf("bucket count changed on click: got %d want %d", got, before)
	}
}

func TestExportMouseWheelScrollsCardRows(t *testing.T) {
	tmp := t.TempDir()
	spec := packexport.DefaultSpec(tmp)
	state := buildInitialExportState(exportValuesFromSpec(spec), buildExportCardsFromSpec(spec), 220)

	screenWidth, screenHeight := 220, 30
	contentWidth, contentHeight := exportContentBounds(screenWidth, screenHeight)
	metrics := state.contentMetrics(contentWidth, contentHeight)
	_, _, contentX, contentY := exportPanelContentOrigin(screenWidth, screenHeight, state.view(contentWidth, contentHeight))

	if max((len(state.cards)+metrics.CardColumns-1)/metrics.CardColumns, 1) <= metrics.VisibleRows {
		t.Fatal("test setup expected scrollable card grid")
	}

	wheel := tea.MouseWheelMsg{
		X:      contentX + metrics.CardsRect.X + 2,
		Y:      contentY + metrics.CardsRect.Y + 1,
		Button: tea.MouseWheelDown,
	}
	if handled := state.handleMouseWheel(wheel, screenWidth, screenHeight, contentWidth, contentHeight); !handled {
		t.Fatal("expected wheel event to be handled")
	}
	if state.cardRowOffset <= 0 {
		t.Fatalf("cardRowOffset = %d, want > 0", state.cardRowOffset)
	}
}

func findExportCardIndex(cards []exportGameCard, gameID puzzle.GameID) int {
	for i, card := range cards {
		if card.GameID == gameID {
			return i
		}
	}
	return 0
}
