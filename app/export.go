package app

import (
	"context"
	"fmt"
	"image/color"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/FelineStateMachine/puzzletea/config"
	"github.com/FelineStateMachine/puzzletea/export/pack"
	puzzlegame "github.com/FelineStateMachine/puzzletea/game"
	"github.com/FelineStateMachine/puzzletea/puzzle"
	"github.com/FelineStateMachine/puzzletea/theme"
	"github.com/FelineStateMachine/puzzletea/ui"

	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/spinner"
	"charm.land/bubbles/v2/textarea"
	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

type exportModel struct {
	focus          exportFocus
	values         exportFormValues
	titleInput     textinput.Model
	headerInput    textarea.Model
	advertInput    textinput.Model
	volumeInput    textinput.Model
	seedInput      textinput.Model
	pdfPathInput   textinput.Model
	jsonlPathInput textinput.Model
	cards          []exportGameCard
	cardIndex      int
	bucketIndex    int
	cardRowOffset  int
	width          int
	height         int
}

type exportFocus int

const (
	exportFocusTitle exportFocus = iota
	exportFocusHeader
	exportFocusAdvert
	exportFocusVolume
	exportFocusLayout
	exportFocusSeed
	exportFocusPDFPath
	exportFocusJSONLToggle
	exportFocusJSONLPath
	exportFocusCards
	exportFocusSubmit
)

var exportSheetLayouts = []string{"half-letter", "duplex-booklet"}

const exportFooterHint = "tab focus • arrows move buckets • digits set counts • +/- tweak • enter next/export • esc back"

const (
	exportPanelWidthPadding  = 10
	exportPanelHeightPadding = 10
	exportCardGap            = 1
	exportCardMinWidth       = 15
	exportCardMaxColumns     = 5
	exportCardTargetRows     = 3
)

type exportFormValues struct {
	Title           string
	Header          string
	Advert          string
	Volume          string
	SheetLayout     string
	Seed            string
	PDFOutputPath   string
	JSONLEnabled    bool
	JSONLOutputPath string
}

type exportGameCard struct {
	GameID      puzzle.GameID
	GameType    string
	Buckets     [3]int
	BucketModes [3][]packexport.ModeCatalog
}

type exportRect struct {
	X int
	Y int
	W int
	H int
}

func (r exportRect) contains(x, y int) bool {
	return x >= r.X && x < r.X+r.W && y >= r.Y && y < r.Y+r.H
}

type exportSettingsRects struct {
	Title       exportRect
	Advert      exportRect
	Header      exportRect
	Volume      exportRect
	Layout      exportRect
	Seed        exportRect
	JSONL       exportRect
	PDFPath     exportRect
	JSONLPath   exportRect
	HasJSONL    bool
	TotalHeight int
}

type exportCardRects struct {
	Index   int
	Rect    exportRect
	Buckets [3]exportRect
}

type exportCompleteMsg struct {
	jobID  int64
	result packexport.Result
	err    error
}

// exportSubmitCmd and exportBackCmd are sentinel commands emitted by
// exportModel.Update to communicate submit/cancel to the root model.
var (
	exportSubmitCmd = func() tea.Msg { return exportSubmitAction{} }
	exportBackCmd   = func() tea.Msg { return backAction{target: mainMenuView} }
)

// Update handles all export form input and returns sentinel commands for
// submit (exportSubmitCmd) and cancel (exportBackCmd).
func (s *exportModel) Update(msg tea.Msg) (exportModel, tea.Cmd) {
	if clickMsg, ok := msg.(tea.MouseClickMsg); ok {
		contentWidth, contentHeight := exportContentBounds(s.width, s.height)
		if cmd, handled := s.handleMouseClick(clickMsg, s.width, s.height, contentWidth, contentHeight); handled {
			return *s, cmd
		}
		return *s, nil
	}
	if wheelMsg, ok := msg.(tea.MouseWheelMsg); ok {
		contentWidth, contentHeight := exportContentBounds(s.width, s.height)
		if handled := s.handleMouseWheel(wheelMsg, s.width, s.height, contentWidth, contentHeight); handled {
			return *s, nil
		}
	}

	keyMsg, ok := msg.(tea.KeyPressMsg)
	if ok {
		switch {
		case key.Matches(keyMsg, rootKeys.Escape):
			return *s, exportBackCmd
		case keyMsg.String() == "tab":
			cmd := s.moveFocus(1)
			return *s, cmd
		case keyMsg.String() == "shift+tab":
			cmd := s.moveFocus(-1)
			return *s, cmd
		case key.Matches(keyMsg, rootKeys.Enter):
			if s.focus == exportFocusSubmit {
				return *s, exportSubmitCmd
			}
			if s.focus == exportFocusHeader {
				break
			}
			if s.focus == exportFocusLayout {
				s.cycleLayout(1)
				return *s, nil
			}
			if s.focus == exportFocusJSONLToggle {
				s.toggleJSONL()
				return *s, nil
			}
			cmd := s.moveFocus(1)
			return *s, cmd
		}

		contentWidth, contentHeight := exportContentBounds(s.width, s.height)
		if handled := s.handleNavigationKey(keyMsg, contentWidth, contentHeight); handled {
			return *s, nil
		}
	}

	cmd := s.updateFocusedInput(msg)
	return *s, cmd
}

type exportScreen struct {
	width  int
	height int
	export exportModel
	// job lifecycle — kept here so exportModel is pure form state
	running bool
	jobID   int64
	cancel  context.CancelFunc
}

func (s exportScreen) Resize(width, height int) screenModel {
	s.width = width
	s.height = height
	s.export.width = width
	s.export.height = height
	contentWidth, contentHeight := exportContentBounds(width, height)
	s.export.resize(contentWidth)
	s.export.ensureCardSelectionVisible(contentWidth, contentHeight)
	return s
}

func (s exportScreen) Update(msg tea.Msg) (screenModel, tea.Cmd, screenAction) {
	next, cmd := s.export.Update(msg)
	s.export = next
	return s, cmd, nil
}

func (s exportScreen) View(notice noticeState) string {
	contentWidth, contentHeight := exportContentBounds(s.width, s.height)
	return renderPanelView(
		s.width,
		s.height,
		notice,
		"Export",
		s.export.view(contentWidth, contentHeight),
		exportFooterHint,
	)
}

type exportRunningScreen struct {
	width   int
	height  int
	spinner spinner.Model
}

func (s exportRunningScreen) Resize(width, height int) screenModel {
	s.width = width
	s.height = height
	return s
}

func (s exportRunningScreen) Update(msg tea.Msg) (screenModel, tea.Cmd, screenAction) {
	var cmd tea.Cmd
	s.spinner, cmd = s.spinner.Update(msg)
	return s, cmd, nil
}

func (s exportRunningScreen) View(notice noticeState) string {
	content := s.spinner.View() + " Exporting puzzle pack..."
	box := ui.GeneratingFrame().Render(appendNoticeContent(s.width, notice, content))
	return ui.CenterView(s.width, s.height, box)
}

func (m model) activeExportScreen() (exportScreen, bool) {
	s, ok := m.screens[exportView].(exportScreen)
	return s, ok
}

func (m model) handleExportEnter() (model, tea.Cmd) {
	if m.screens == nil {
		m.screens = make(map[viewState]screenModel)
	}
	if _, exists := m.screens[exportView]; !exists {
		spec := m.initialExportSpec()
		es := exportScreen{
			width:  m.width,
			height: m.height,
			export: buildInitialExportState(exportValuesFromSpec(spec), buildExportCardsFromSpec(spec), m.width),
		}
		m.screens[exportView] = es.Resize(m.width, m.height)
	} else {
		m.screens[exportView] = m.screens[exportView].Resize(m.width, m.height)
	}
	m.state = exportView
	m = m.clearNotice()
	return m, nil
}

func (m model) handleExportSubmit() (model, tea.Cmd) {
	es, ok := m.activeExportScreen()
	if !ok {
		return m, nil
	}
	spec, cfg := es.export.toSpecAndConfig()
	if err := packexport.ValidateSpec(spec); err != nil {
		return m.setErrorf("Could not start export: %v", err), nil
	}

	m.cfg.Export = cfg
	if err := m.cfg.Save(m.configPath); err != nil {
		m = m.setErrorf("Export settings save failed: %v", err)
	}

	cmd := m.startExport(spec)
	return m, cmd
}

func (m model) handleExportComplete(msg exportCompleteMsg) (model, tea.Cmd) {
	es, ok := m.activeExportScreen()
	if !ok || msg.jobID != es.jobID {
		return m, nil
	}

	if es.cancel != nil {
		es.cancel()
		es.cancel = nil
	}
	es.running = false
	m.screens[exportView] = es
	m.state = exportView

	if msg.err != nil {
		return m.setErrorf("Export failed: %v", msg.err), nil
	}

	message := fmt.Sprintf("Exported %d puzzles to %s", msg.result.TotalCount, msg.result.PDFOutputPath)
	if msg.result.JSONLOutputPath != "" {
		message += fmt.Sprintf(" and %s", msg.result.JSONLOutputPath)
	}
	return m.setSuccessf("%s", message), nil
}

func (m *model) cancelActiveExport() {
	es, ok := m.screens[exportView].(exportScreen)
	if !ok {
		return
	}
	if es.cancel != nil {
		es.cancel()
		es.cancel = nil
	}
	if es.running {
		es.jobID++
	}
	es.running = false
	m.screens[exportView] = es
}

func (m *model) startExport(spec packexport.Spec) tea.Cmd {
	m.cancelActiveExport()
	es, _ := m.screens[exportView].(exportScreen)
	es.jobID++
	jobID := es.jobID
	ctx, cancel := context.WithCancel(context.Background())
	es.cancel = cancel
	es.running = true
	m.screens[exportView] = es
	m.state = exportRunningView
	*m = m.clearNotice()
	*m = m.initScreen(exportRunningView)
	return tea.Batch(m.spinner.Tick, exportCmd(ctx, jobID, spec))
}

func exportCmd(ctx context.Context, jobID int64, spec packexport.Spec) tea.Cmd {
	return func() tea.Msg {
		result, err := packexport.Run(ctx, spec)
		return exportCompleteMsg{jobID: jobID, result: result, err: err}
	}
}

func buildInitialExportState(values exportFormValues, cards []exportGameCard, width int) exportModel {
	contentWidth, _ := exportContentBounds(width, 0)
	state := exportModel{
		focus:  exportFocusCards,
		values: values,
		cards:  cards,
		width:  width,
	}
	state.initInputs(contentWidth)
	return state
}

func exportContentBounds(width, height int) (int, int) {
	return max(width-exportPanelWidthPadding, 24), max(height-exportPanelHeightPadding, 8)
}

func (m model) initialExportSpec() packexport.Spec {
	cwd, err := os.Getwd()
	if err != nil {
		cwd = "."
	}

	base := packexport.DefaultSpec(cwd)
	if !hasSavedExportConfig(m.cfg.Export) {
		return base
	}

	return applySavedExportConfig(base, m.cfg.Export)
}

func hasSavedExportConfig(cfg config.ExportConfig) bool {
	return cfg.Title != "" ||
		cfg.Header != "" ||
		cfg.Advert != "" ||
		cfg.Volume != 0 ||
		cfg.SheetLayout != "" ||
		cfg.Seed != "" ||
		cfg.PDFOutputPath != "" ||
		cfg.JSONLOutputPath != "" ||
		cfg.JSONLEnabled ||
		len(cfg.Counts) > 0
}

func applySavedExportConfig(base packexport.Spec, cfg config.ExportConfig) packexport.Spec {
	if cfg.Title != "" {
		base.Title = cfg.Title
	}
	if cfg.Header != "" {
		base.Header = cfg.Header
	}
	if cfg.Advert != "" {
		base.Advert = cfg.Advert
	}
	if cfg.Volume > 0 {
		base.Volume = cfg.Volume
	}
	if cfg.SheetLayout != "" {
		base.SheetLayout = cfg.SheetLayout
	}
	base.Seed = cfg.Seed
	if cfg.PDFOutputPath != "" {
		base.PDFOutputPath = cfg.PDFOutputPath
	}
	if cfg.JSONLEnabled {
		base.JSONLOutputPath = cfg.JSONLOutputPath
	} else {
		base.JSONLOutputPath = ""
	}
	for gameID, modes := range cfg.Counts {
		if _, ok := base.Counts[gameID]; !ok {
			base.Counts[gameID] = map[puzzle.ModeID]int{}
		}
		for modeID, count := range modes {
			base.Counts[gameID][modeID] = count
		}
	}
	return base
}

func exportValuesFromSpec(spec packexport.Spec) exportFormValues {
	values := exportFormValues{
		Title:           spec.Title,
		Header:          spec.Header,
		Advert:          spec.Advert,
		Volume:          strconv.Itoa(spec.Volume),
		SheetLayout:     spec.SheetLayout,
		Seed:            spec.Seed,
		PDFOutputPath:   spec.PDFOutputPath,
		JSONLEnabled:    strings.TrimSpace(spec.JSONLOutputPath) != "",
		JSONLOutputPath: defaultJSONLPath(spec),
	}
	if strings.TrimSpace(spec.JSONLOutputPath) != "" {
		values.JSONLOutputPath = spec.JSONLOutputPath
	}
	return values
}

func defaultJSONLPath(spec packexport.Spec) string {
	dir := filepath.Dir(spec.PDFOutputPath)
	if dir == "." || dir == "" {
		dir = "out"
	}
	return filepath.Join(dir, "puzzletea-export.jsonl")
}

func buildExportCardsFromSpec(spec packexport.Spec) []exportGameCard {
	catalog := packexport.ExportCatalog()
	cards := make([]exportGameCard, 0, len(catalog))
	for _, game := range catalog {
		card := exportGameCard{
			GameID:   game.GameID,
			GameType: game.GameType,
		}
		for i, mode := range game.Modes {
			bucket := bucketIndexForMode(i, len(game.Modes))
			card.BucketModes[bucket] = append(card.BucketModes[bucket], mode)
			if modes, ok := spec.Counts[mode.GameID]; ok {
				card.Buckets[bucket] += modes[mode.ModeID]
			}
		}
		cards = append(cards, card)
	}
	return cards
}

func bucketIndexForMode(index, total int) int {
	if total <= 0 {
		return 0
	}
	return index * 3 / total
}

func (s *exportModel) initInputs(width int) {
	s.titleInput = newExportInput(s.values.Title, "Sampler title")
	s.headerInput = newExportTextarea(s.values.Header, "Short intro")
	s.advertInput = newExportInput(s.values.Advert, "Footer copy")
	s.volumeInput = newExportInput(s.values.Volume, "1")
	s.seedInput = newExportInput(s.values.Seed, "optional seed")
	s.pdfPathInput = newExportInput(s.values.PDFOutputPath, "output pdf")
	s.jsonlPathInput = newExportInput(s.values.JSONLOutputPath, "optional jsonl")
	s.resize(width)
	s.applyFocus()
}

func newExportInput(value, placeholder string) textinput.Model {
	input := textinput.New()
	input.Prompt = ""
	input.SetValue(value)
	input.Placeholder = placeholder
	input.CharLimit = 512
	styleExportInput(&input)
	return input
}

func newExportTextarea(value, placeholder string) textarea.Model {
	input := textarea.New()
	input.Prompt = ""
	input.SetValue(value)
	input.Placeholder = placeholder
	input.CharLimit = 512
	input.ShowLineNumbers = false
	input.EndOfBufferCharacter = ' '
	input.SetHeight(3)
	styleExportTextarea(&input)
	return input
}

func styleExportInput(input *textinput.Model) {
	palette := theme.Current()
	styles := input.Styles()
	styles.Focused.Text = lipgloss.NewStyle().Foreground(palette.FG)
	styles.Focused.Placeholder = lipgloss.NewStyle().Foreground(palette.TextDim)
	styles.Focused.Prompt = lipgloss.NewStyle().Foreground(palette.Accent)
	styles.Blurred.Text = lipgloss.NewStyle().Foreground(palette.FG)
	styles.Blurred.Placeholder = lipgloss.NewStyle().Foreground(palette.TextDim)
	styles.Blurred.Prompt = lipgloss.NewStyle().Foreground(palette.TextDim)
	styles.Cursor.Color = palette.Accent
	input.SetStyles(styles)
}

func styleExportTextarea(input *textarea.Model) {
	palette := theme.Current()
	styles := input.Styles()
	styles.Focused.Base = lipgloss.NewStyle()
	styles.Focused.Text = lipgloss.NewStyle().Foreground(palette.FG)
	styles.Focused.Placeholder = lipgloss.NewStyle().Foreground(palette.TextDim)
	styles.Focused.Prompt = lipgloss.NewStyle().Foreground(palette.Accent)
	styles.Focused.CursorLine = lipgloss.NewStyle()
	styles.Focused.EndOfBuffer = lipgloss.NewStyle().Foreground(palette.TextDim)
	styles.Blurred.Base = lipgloss.NewStyle()
	styles.Blurred.Text = lipgloss.NewStyle().Foreground(palette.FG)
	styles.Blurred.Placeholder = lipgloss.NewStyle().Foreground(palette.TextDim)
	styles.Blurred.Prompt = lipgloss.NewStyle().Foreground(palette.TextDim)
	styles.Blurred.CursorLine = lipgloss.NewStyle()
	styles.Blurred.EndOfBuffer = lipgloss.NewStyle().Foreground(palette.TextDim)
	styles.Cursor.Color = palette.Accent
	input.SetStyles(styles)
}

func (s *exportModel) resize(width int) {
	styleExportInput(&s.titleInput)
	styleExportTextarea(&s.headerInput)
	styleExportInput(&s.advertInput)
	styleExportInput(&s.volumeInput)
	styleExportInput(&s.seedInput)
	styleExportInput(&s.pdfPathInput)
	styleExportInput(&s.jsonlPathInput)

	fieldW := max(18, min((width-22)/3, 38))
	s.titleInput.SetWidth(fieldW)
	s.headerInput.SetWidth(fieldW)
	s.headerInput.SetHeight(3)
	s.advertInput.SetWidth(fieldW)
	s.volumeInput.SetWidth(6)
	s.seedInput.SetWidth(max(18, min(fieldW, 26)))
	s.pdfPathInput.SetWidth(max(26, min(width-20, 72)))
	s.jsonlPathInput.SetWidth(max(26, min(width-20, 72)))
}

func (s *exportModel) visibleFocuses() []exportFocus {
	focuses := []exportFocus{
		exportFocusTitle,
		exportFocusAdvert,
		exportFocusHeader,
		exportFocusVolume,
		exportFocusLayout,
		exportFocusSeed,
		exportFocusPDFPath,
		exportFocusJSONLToggle,
	}
	if s.values.JSONLEnabled {
		focuses = append(focuses, exportFocusJSONLPath)
	}
	focuses = append(focuses, exportFocusCards, exportFocusSubmit)
	return focuses
}

func (s *exportModel) moveFocus(delta int) tea.Cmd {
	focuses := s.visibleFocuses()
	current := 0
	for i, focus := range focuses {
		if focus == s.focus {
			current = i
			break
		}
	}
	next := (current + delta) % len(focuses)
	if next < 0 {
		next += len(focuses)
	}
	s.focus = focuses[next]
	switch s.focus {
	case exportFocusPDFPath:
		s.pdfPathInput.SetCursor(runeIdxBeforeExt(s.values.PDFOutputPath))
	case exportFocusJSONLPath:
		s.jsonlPathInput.SetCursor(runeIdxBeforeExt(s.values.JSONLOutputPath))
	}
	return s.applyFocus()
}

func (s *exportModel) applyFocus() tea.Cmd {
	s.titleInput.Blur()
	s.headerInput.Blur()
	s.advertInput.Blur()
	s.volumeInput.Blur()
	s.seedInput.Blur()
	s.pdfPathInput.Blur()
	s.jsonlPathInput.Blur()

	switch s.focus {
	case exportFocusTitle:
		return s.titleInput.Focus()
	case exportFocusHeader:
		return s.headerInput.Focus()
	case exportFocusAdvert:
		return s.advertInput.Focus()
	case exportFocusVolume:
		return s.volumeInput.Focus()
	case exportFocusSeed:
		return s.seedInput.Focus()
	case exportFocusPDFPath:
		return s.pdfPathInput.Focus()
	case exportFocusJSONLPath:
		return s.jsonlPathInput.Focus()
	default:
		return nil
	}
}

func (s *exportModel) updateFocusedInput(msg tea.Msg) tea.Cmd {
	var cmd tea.Cmd
	switch s.focus {
	case exportFocusTitle:
		s.titleInput, cmd = s.titleInput.Update(msg)
		s.values.Title = s.titleInput.Value()
	case exportFocusHeader:
		s.headerInput, cmd = s.headerInput.Update(msg)
		s.values.Header = s.headerInput.Value()
	case exportFocusAdvert:
		s.advertInput, cmd = s.advertInput.Update(msg)
		s.values.Advert = s.advertInput.Value()
	case exportFocusVolume:
		s.volumeInput, cmd = s.volumeInput.Update(msg)
		s.values.Volume = s.volumeInput.Value()
	case exportFocusSeed:
		s.seedInput, cmd = s.seedInput.Update(msg)
		s.values.Seed = s.seedInput.Value()
	case exportFocusPDFPath:
		s.pdfPathInput, cmd = s.pdfPathInput.Update(msg)
		s.values.PDFOutputPath = s.pdfPathInput.Value()
	case exportFocusJSONLPath:
		s.jsonlPathInput, cmd = s.jsonlPathInput.Update(msg)
		s.values.JSONLOutputPath = s.jsonlPathInput.Value()
	}
	return cmd
}

func (s *exportModel) handleNavigationKey(msg tea.KeyPressMsg, width, height int) bool {
	switch s.focus {
	case exportFocusLayout:
		switch msg.String() {
		case "left", "h":
			s.cycleLayout(-1)
			return true
		case "right", "l":
			s.cycleLayout(1)
			return true
		}
	case exportFocusJSONLToggle:
		switch msg.String() {
		case "left", "right", "h", "l", " ":
			s.toggleJSONL()
			return true
		}
	case exportFocusCards:
		switch msg.String() {
		case "left":
			s.moveBucketSelection(-1, width, height)
			return true
		case "right":
			s.moveBucketSelection(1, width, height)
			return true
		case "h":
			s.moveBucketSelection(-1, width, height)
			return true
		case "l":
			s.moveBucketSelection(1, width, height)
			return true
		case "up", "k":
			cols := s.cardColumns(width)
			if s.cardIndex-cols >= 0 {
				s.cardIndex -= cols
			}
			s.ensureCardSelectionVisible(width, height)
			return true
		case "down", "j":
			cols := s.cardColumns(width)
			if s.cardIndex+cols < len(s.cards) {
				s.cardIndex += cols
			}
			s.ensureCardSelectionVisible(width, height)
			return true
		case "+", "=":
			s.adjustSelectedBucket(1)
			return true
		case "-":
			s.adjustSelectedBucket(-1)
			return true
		case "backspace", "delete":
			s.backspaceSelectedBucket()
			return true
		case "pgup":
			s.cardRowOffset = max(s.cardRowOffset-1, 0)
			return true
		case "pgdown":
			s.cardRowOffset++
			s.ensureCardSelectionVisible(width, height)
			return true
		case "0", "1", "2", "3", "4", "5", "6", "7", "8", "9":
			s.appendSelectedBucketDigit(msg.String())
			return true
		}
	}
	return false
}

func (s *exportModel) moveBucketSelection(delta, width, height int) {
	if len(s.cards) == 0 || delta == 0 {
		return
	}
	position := s.cardIndex*len(bucketNames) + s.bucketIndex + delta
	maxPosition := len(s.cards)*len(bucketNames) - 1
	position = max(0, min(position, maxPosition))
	s.cardIndex = position / len(bucketNames)
	s.bucketIndex = position % len(bucketNames)
	s.ensureCardSelectionVisible(width, height)
}

func (s *exportModel) appendSelectedBucketDigit(raw string) {
	if len(s.cards) == 0 {
		return
	}
	digit, err := strconv.Atoi(raw)
	if err != nil {
		return
	}
	card := &s.cards[s.cardIndex]
	if strings.TrimSpace(s.values.Seed) != "" && !card.bucketSeeded(s.bucketIndex) {
		return
	}
	card.Buckets[s.bucketIndex] = card.Buckets[s.bucketIndex]*10 + digit
}

func (s *exportModel) backspaceSelectedBucket() {
	if len(s.cards) == 0 {
		return
	}
	card := &s.cards[s.cardIndex]
	card.Buckets[s.bucketIndex] /= 10
}

func (s *exportModel) cycleLayout(delta int) {
	index := 0
	for i, layout := range exportSheetLayouts {
		if layout == s.values.SheetLayout {
			index = i
			break
		}
	}
	index = (index + delta) % len(exportSheetLayouts)
	if index < 0 {
		index += len(exportSheetLayouts)
	}
	s.values.SheetLayout = exportSheetLayouts[index]
}

func (s *exportModel) toggleJSONL() {
	s.values.JSONLEnabled = !s.values.JSONLEnabled
	if !s.values.JSONLEnabled && s.focus == exportFocusJSONLPath {
		s.focus = exportFocusCards
	}
}

func (s *exportModel) adjustSelectedBucket(delta int) {
	if len(s.cards) == 0 {
		return
	}
	card := &s.cards[s.cardIndex]
	if strings.TrimSpace(s.values.Seed) != "" && !card.bucketSeeded(s.bucketIndex) {
		return
	}
	card.Buckets[s.bucketIndex] = max(card.Buckets[s.bucketIndex]+delta, 0)
}

func (c exportGameCard) bucketSeeded(index int) bool {
	if index < 0 || index >= len(c.BucketModes) {
		return true
	}
	if len(c.BucketModes[index]) == 0 {
		return true
	}
	for _, mode := range c.BucketModes[index] {
		if !mode.Seeded {
			return false
		}
	}
	return true
}

func (s *exportModel) cardColumns(width int) int {
	for cols := exportCardMaxColumns; cols >= 1; cols-- {
		if exportCardWidth(width, cols) >= exportCardMinWidth {
			return cols
		}
	}
	return 1
}

func (s *exportModel) ensureCardSelectionVisible(width, height int) {
	cols := s.cardColumns(width)
	if cols <= 0 {
		return
	}
	selectedRow := s.cardIndex / cols
	visibleRows := s.visibleCardRows(width, height)
	if selectedRow < s.cardRowOffset {
		s.cardRowOffset = selectedRow
	}
	if selectedRow >= s.cardRowOffset+visibleRows {
		s.cardRowOffset = selectedRow - visibleRows + 1
	}
	maxRows := max((len(s.cards)+cols-1)/cols, 1)
	s.cardRowOffset = min(max(s.cardRowOffset, 0), maxRows-1)
}

func (s *exportModel) visibleCardRows(width, height int) int {
	if len(s.cards) == 0 {
		return 1
	}

	cardWidth := exportCardWidth(width, s.cardColumns(width))
	cardHeight := lipgloss.Height(s.renderCard(0, cardWidth))
	if cardHeight <= 0 {
		cardHeight = 4
	}

	fixedHeight := lipgloss.Height(s.renderSettings(width)) + lipgloss.Height(s.renderSummary(width))
	available := max(height-fixedHeight-1, cardHeight)

	rows := 1
	used := cardHeight
	for used+1+cardHeight <= available {
		rows++
		used += 1 + cardHeight
	}
	if canFitCardRows(available, cardHeight, exportCardTargetRows) {
		rows = max(rows, exportCardTargetRows)
	}
	return rows
}

func canFitCardRows(available, cardHeight, rows int) bool {
	if available <= 0 || cardHeight <= 0 || rows <= 0 {
		return false
	}
	required := rows * cardHeight
	if rows > 1 {
		required += rows - 1
	}
	return available >= required
}

func (s *exportModel) view(width, height int) string {
	settings := s.renderSettings(width)
	summary := s.renderSummary(width)
	cards := s.renderCards(width, height)
	return lipgloss.JoinVertical(lipgloss.Left, settings, summary, cards)
}

func (s *exportModel) contentMetrics(width, height int) exportContentMetrics {
	settings := s.settingsRects(width)
	settingsHeight := settings.TotalHeight
	summaryHeight := lipgloss.Height(s.renderSummary(width))
	cardWidth := exportCardWidth(width, s.cardColumns(width))
	cardHeight := 0
	if len(s.cards) > 0 {
		cardHeight = lipgloss.Height(s.renderCard(0, cardWidth))
	}
	summaryY := settingsHeight
	summaryButton := exportRect{
		X: max(width-lipgloss.Width(renderChoiceChip("Export", s.focus == exportFocusSubmit)), 0),
		Y: summaryY,
		W: lipgloss.Width(renderChoiceChip("Export", s.focus == exportFocusSubmit)),
		H: summaryHeight,
	}
	cardsY := settingsHeight + summaryHeight
	visibleRows := s.visibleCardRows(width, height)
	cardRects := s.cardRects(width, cardsY, cardWidth, cardHeight, visibleRows)
	return exportContentMetrics{
		Width:          width,
		Height:         height,
		SettingsHeight: settingsHeight,
		SummaryHeight:  summaryHeight,
		Settings:       settings,
		Summary:        exportRect{X: 0, Y: summaryY, W: width, H: summaryHeight},
		SummaryButton:  summaryButton,
		CardsRect: exportRect{
			X: 0,
			Y: cardsY,
			W: width,
			H: max(visibleRows*cardHeight+max(visibleRows-1, 0), 0),
		},
		CardsY:      cardsY,
		CardWidth:   cardWidth,
		CardHeight:  cardHeight,
		CardColumns: s.cardColumns(width),
		VisibleRows: visibleRows,
		CardRects:   cardRects,
	}
}

func (s *exportModel) settingsRects(width int) exportSettingsRects {
	if width < 72 {
		return s.stackedSettingsRects(width)
	}

	y := 0
	inlineHeight := s.inlineFieldHeight(width)
	headerHeight := s.headerFieldHeight(width)

	titleW := max((width-2)/2, 24)
	advertW := max(width-titleW-2, 24)
	row1 := exportSettingsRects{
		Title:  exportRect{X: 0, Y: y, W: titleW, H: inlineHeight},
		Advert: exportRect{X: titleW + 2, Y: y, W: advertW, H: inlineHeight},
	}
	y += inlineHeight
	row1.Header = exportRect{X: 0, Y: y, W: width, H: headerHeight}
	y += headerHeight

	volumeW := 14
	layoutW := 18
	jsonlW := 14
	seedW := max(width-volumeW-layoutW-jsonlW-6, 22)
	row1.Volume = exportRect{X: 0, Y: y, W: volumeW, H: inlineHeight}
	row1.Layout = exportRect{X: volumeW + 2, Y: y, W: layoutW, H: inlineHeight}
	row1.Seed = exportRect{X: volumeW + layoutW + 4, Y: y, W: seedW, H: inlineHeight}
	row1.JSONL = exportRect{X: volumeW + layoutW + seedW + 6, Y: y, W: jsonlW, H: inlineHeight}
	y += inlineHeight

	row1.PDFPath = exportRect{X: 0, Y: y, W: width, H: inlineHeight}
	y += inlineHeight

	if s.values.JSONLEnabled {
		row1.HasJSONL = true
		row1.JSONLPath = exportRect{X: 0, Y: y, W: width, H: inlineHeight}
		y += inlineHeight
	}

	row1.TotalHeight = y
	return row1
}

func (s *exportModel) stackedSettingsRects(width int) exportSettingsRects {
	y := 0
	inlineHeight := s.inlineFieldHeight(width)
	headerHeight := s.headerFieldHeight(width)
	rects := exportSettingsRects{
		Title: exportRect{X: 0, Y: y, W: width, H: inlineHeight},
	}
	y += inlineHeight
	rects.Advert = exportRect{X: 0, Y: y, W: width, H: inlineHeight}
	y += inlineHeight
	rects.Header = exportRect{X: 0, Y: y, W: width, H: headerHeight}
	y += headerHeight

	volumeW, layoutW, jsonlW := 14, 18, 14
	seedW := max(width-volumeW-layoutW-jsonlW-6, 16)
	if seedW < 16 {
		rects.Volume = exportRect{X: 0, Y: y, W: width, H: inlineHeight}
		y += inlineHeight
		rects.Layout = exportRect{X: 0, Y: y, W: width, H: inlineHeight}
		y += inlineHeight
		rects.Seed = exportRect{X: 0, Y: y, W: width, H: inlineHeight}
		y += inlineHeight
		rects.JSONL = exportRect{X: 0, Y: y, W: width, H: inlineHeight}
		y += inlineHeight
	} else {
		rects.Volume = exportRect{X: 0, Y: y, W: volumeW, H: inlineHeight}
		rects.Layout = exportRect{X: volumeW + 2, Y: y, W: layoutW, H: inlineHeight}
		rects.Seed = exportRect{X: volumeW + layoutW + 4, Y: y, W: seedW, H: inlineHeight}
		rects.JSONL = exportRect{X: volumeW + layoutW + seedW + 6, Y: y, W: jsonlW, H: inlineHeight}
		y += inlineHeight
	}

	rects.PDFPath = exportRect{X: 0, Y: y, W: width, H: inlineHeight}
	y += inlineHeight
	if s.values.JSONLEnabled {
		rects.HasJSONL = true
		rects.JSONLPath = exportRect{X: 0, Y: y, W: width, H: inlineHeight}
		y += inlineHeight
	}
	rects.TotalHeight = y
	return rects
}

func (s *exportModel) inlineFieldHeight(width int) int {
	input := s.titleInput
	return lipgloss.Height(s.renderInlineInputField("Title", &input, false, max(width, 14)))
}

func (s *exportModel) headerFieldHeight(width int) int {
	input := s.headerInput
	return lipgloss.Height(s.renderTextareaField("Header", &input, false, max(width, 24)))
}

func (s *exportModel) cardRects(width, cardsY, cardWidth, cardHeight, visibleRows int) []exportCardRects {
	if len(s.cards) == 0 || cardHeight <= 0 {
		return nil
	}
	cols := s.cardColumns(width)
	start := s.cardRowOffset * cols
	end := min(len(s.cards), start+visibleRows*cols)
	layouts := make([]exportCardRects, 0, end-start)
	for i := start; i < end; i++ {
		rel := i - start
		row := rel / cols
		col := rel % cols
		rect := exportRect{
			X: col * (cardWidth + exportCardGap),
			Y: cardsY + row*(cardHeight+1),
			W: cardWidth,
			H: cardHeight,
		}
		layouts = append(layouts, exportCardRects{
			Index:   i,
			Rect:    rect,
			Buckets: s.bucketRectsForCard(i, rect),
		})
	}
	return layouts
}

func (s *exportModel) bucketRectsForCard(index int, rect exportRect) [3]exportRect {
	var rects [3]exportRect
	card := s.cards[index]
	selected := index == s.cardIndex
	x := rect.X + 2
	y := rect.Y + 2
	for bucketIndex, name := range bucketNames {
		part := s.renderDifficultyPair(card, selected, bucketIndex, name)
		w := lipgloss.Width(part)
		rects[bucketIndex] = exportRect{X: x, Y: y, W: w, H: 1}
		x += w + 1
	}
	return rects
}

func (s *exportModel) renderSettings(width int) string {
	if width < 72 {
		return s.renderStackedSettings(width)
	}

	titleW := max((width-2)/2, 24)
	advertW := max(width-titleW-2, 24)
	row1 := lipgloss.JoinHorizontal(lipgloss.Top,
		s.renderInlineInputField("Title", &s.titleInput, s.focus == exportFocusTitle, titleW),
		"  ",
		s.renderInlineInputField("Advert", &s.advertInput, s.focus == exportFocusAdvert, advertW),
	)
	rowHeader := s.renderTextareaField("Header", &s.headerInput, s.focus == exportFocusHeader, width)

	volumeW := 14
	layoutW := 18
	jsonlW := 14
	seedW := max(width-volumeW-layoutW-jsonlW-6, 22)
	row2 := lipgloss.JoinHorizontal(lipgloss.Top,
		s.renderInlineInputField("Volume", &s.volumeInput, s.focus == exportFocusVolume, volumeW),
		"  ",
		s.renderInlineChoiceField("Layout", layoutLabel(s.values.SheetLayout), s.focus == exportFocusLayout, layoutW),
		"  ",
		s.renderInlineInputField("Seed", &s.seedInput, s.focus == exportFocusSeed, seedW),
		"  ",
		s.renderInlineChoiceField("JSONL", onOffLabel(s.values.JSONLEnabled), s.focus == exportFocusJSONLToggle, jsonlW),
	)

	rows := []string{
		row1,
		rowHeader,
		row2,
		s.renderInlineInputField("PDF Path", &s.pdfPathInput, s.focus == exportFocusPDFPath, width),
	}
	if s.values.JSONLEnabled {
		rows = append(rows, s.renderInlineInputField("JSONL Path", &s.jsonlPathInput, s.focus == exportFocusJSONLPath, width))
	}

	return lipgloss.JoinVertical(lipgloss.Left, rows...)
}

func (s *exportModel) renderStackedSettings(width int) string {
	rows := []string{
		s.renderInlineInputField("Title", &s.titleInput, s.focus == exportFocusTitle, width),
		s.renderInlineInputField("Advert", &s.advertInput, s.focus == exportFocusAdvert, width),
		s.renderTextareaField("Header", &s.headerInput, s.focus == exportFocusHeader, width),
	}

	volumeW, layoutW, jsonlW := 14, 18, 14
	seedW := max(width-volumeW-layoutW-jsonlW-6, 16)
	if seedW < 16 {
		rows = append(rows,
			s.renderInlineInputField("Volume", &s.volumeInput, s.focus == exportFocusVolume, width),
			s.renderInlineChoiceField("Layout", layoutLabel(s.values.SheetLayout), s.focus == exportFocusLayout, width),
			s.renderInlineInputField("Seed", &s.seedInput, s.focus == exportFocusSeed, width),
			s.renderInlineChoiceField("JSONL", onOffLabel(s.values.JSONLEnabled), s.focus == exportFocusJSONLToggle, width),
		)
	} else {
		rows = append(rows, lipgloss.JoinHorizontal(lipgloss.Top,
			s.renderInlineInputField("Volume", &s.volumeInput, s.focus == exportFocusVolume, volumeW),
			"  ",
			s.renderInlineChoiceField("Layout", layoutLabel(s.values.SheetLayout), s.focus == exportFocusLayout, layoutW),
			"  ",
			s.renderInlineInputField("Seed", &s.seedInput, s.focus == exportFocusSeed, seedW),
			"  ",
			s.renderInlineChoiceField("JSONL", onOffLabel(s.values.JSONLEnabled), s.focus == exportFocusJSONLToggle, jsonlW),
		))
	}

	rows = append(rows, s.renderInlineInputField("PDF Path", &s.pdfPathInput, s.focus == exportFocusPDFPath, width))
	if s.values.JSONLEnabled {
		rows = append(rows, s.renderInlineInputField("JSONL Path", &s.jsonlPathInput, s.focus == exportFocusJSONLPath, width))
	}
	return lipgloss.JoinVertical(lipgloss.Left, rows...)
}

func (s *exportModel) renderTextareaField(label string, input *textarea.Model, focused bool, width int) string {
	border := theme.Current().Border
	if focused {
		border = theme.Current().Accent
	}

	innerWidth := max(width-4, 1)
	input.SetWidth(innerWidth)
	input.SetHeight(3)

	body := lipgloss.JoinVertical(
		lipgloss.Left,
		ui.DimItemStyle().Render(label),
		input.View(),
	)
	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(border).
		Padding(0, 1).
		Width(width).
		Render(body)
}

func (s *exportModel) renderInlineInputField(label string, input *textinput.Model, focused bool, width int) string {
	border := theme.Current().Border
	if focused {
		border = theme.Current().Accent
	}

	innerWidth := max(width-4, 1)
	labelWidth := lipgloss.Width(label) + 1
	input.SetWidth(max(innerWidth-labelWidth, 1))

	body := lipgloss.JoinHorizontal(
		lipgloss.Top,
		ui.DimItemStyle().Render(label),
		" ",
		renderSingleLine(input.View(), innerWidth-labelWidth),
	)
	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(border).
		Padding(0, 1).
		Width(width).
		Height(1).
		Render(body)
}

func (s *exportModel) renderInlineChoiceField(label, value string, focused bool, width int) string {
	border := theme.Current().Border
	if focused {
		border = theme.Current().Accent
	}

	innerWidth := max(width-4, 1)
	labelWidth := lipgloss.Width(label) + 1
	valueStyle := lipgloss.NewStyle().Foreground(theme.Current().FG)
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

func renderSingleLine(value string, width int) string {
	if width <= 0 {
		return ""
	}
	return lipgloss.NewStyle().
		Width(width).
		MaxWidth(width).
		Height(1).
		MaxHeight(1).
		Render(value)
}

func renderChoiceChip(label string, focused bool) string {
	style := lipgloss.NewStyle().
		Bold(true).
		Padding(0, 1).
		Foreground(theme.Current().FG).
		Border(lipgloss.NormalBorder()).
		BorderForeground(theme.Current().Border)
	if focused {
		style = style.Foreground(theme.TextOnBG(theme.Current().Accent)).Background(theme.Current().Accent).BorderForeground(theme.Current().Accent)
	}
	return style.Render(label)
}

func layoutLabel(layout string) string {
	switch layout {
	case "duplex-booklet":
		return "Booklet"
	default:
		return "Half Letter"
	}
}

func onOffLabel(enabled bool) string {
	if enabled {
		return "On"
	}
	return "Off"
}

func (s *exportModel) renderSummary(width int) string {
	total := 0
	for _, card := range s.cards {
		for _, count := range card.Buckets {
			total += count
		}
	}

	selected := "none"
	seedHint := ""
	if len(s.cards) > 0 {
		card := s.cards[s.cardIndex]
		selected = fmt.Sprintf("%s / %s = %d", card.GameType, bucketNames[s.bucketIndex], card.Buckets[s.bucketIndex])
		if strings.TrimSpace(s.values.Seed) != "" && !card.bucketSeeded(s.bucketIndex) {
			seedHint = "  selected bucket cannot be seeded"
		}
	}

	button := renderChoiceChip("Export", s.focus == exportFocusSubmit)
	totalText := ui.PanelTitle().Render(fmt.Sprintf("Total: %d", total))
	detailWidth := max(width-lipgloss.Width(totalText)-lipgloss.Width(button)-4, 12)
	detailText := ui.DimItemStyle().Render(truncateLabel("Selected: "+selected+seedHint, detailWidth))
	detail := lipgloss.NewStyle().Width(detailWidth).MaxWidth(detailWidth).Render(detailText)
	return lipgloss.JoinHorizontal(lipgloss.Top, totalText, "   ", detail, " ", button)
}

var bucketNames = [3]string{"Easy", "Medium", "Hard"}

func (s *exportModel) renderCards(width, height int) string {
	if len(s.cards) == 0 {
		return ui.DimItemStyle().Render("No exportable games found.")
	}

	cols := s.cardColumns(width)
	cardWidth := exportCardWidth(width, cols)
	start := s.cardRowOffset * cols
	end := min(len(s.cards), start+s.visibleCardRows(width, height)*cols)

	rows := []string{}
	for i := start; i < end; i += cols {
		rowEnd := min(i+cols, end)
		parts := make([]string, 0, rowEnd-i)
		for idx := i; idx < rowEnd; idx++ {
			parts = append(parts, s.renderCard(idx, cardWidth))
		}
		rows = append(rows, lipgloss.JoinHorizontal(lipgloss.Top, joinCardParts(parts)...))
	}

	footer := ""
	totalRows := max((len(s.cards)+cols-1)/cols, 1)
	if totalRows > s.visibleCardRows(width, height) {
		footer = ui.DimItemStyle().Render(fmt.Sprintf("rows %d-%d of %d", s.cardRowOffset+1, min(s.cardRowOffset+s.visibleCardRows(width, height), totalRows), totalRows))
	}

	content := strings.Join(rows, "\n")
	if footer != "" {
		content = lipgloss.JoinVertical(lipgloss.Left, content, footer)
	}
	return content
}

func (s *exportModel) renderCard(index, width int) string {
	card := s.cards[index]
	selected := index == s.cardIndex

	border := theme.Current().Border
	if selected {
		border = theme.Current().Accent
	}

	titleStyle := ui.PanelTitle()
	if selected {
		titleStyle = ui.SelectedItemStyle()
	}

	bucketParts := make([]string, 0, 3)
	for bucketIndex, name := range bucketNames {
		bucketParts = append(bucketParts, s.renderDifficultyPair(card, selected, bucketIndex, name))
	}

	total := card.Buckets[0] + card.Buckets[1] + card.Buckets[2]
	titleWidth := max(width-10, 6)
	title := truncateLabel(card.GameType, titleWidth)
	titleLine := lipgloss.JoinHorizontal(
		lipgloss.Top,
		titleStyle.Width(titleWidth).Render(title),
		" ",
		ui.DimItemStyle().Render(fmt.Sprintf("%d", total)),
	)

	body := lipgloss.JoinVertical(
		lipgloss.Left,
		titleLine,
		lipgloss.JoinHorizontal(lipgloss.Top, bucketParts[0], " ", bucketParts[1], " ", bucketParts[2]),
	)

	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(border).
		Padding(0, 1).
		Width(width).
		Render(body)
}

func (s *exportModel) renderDifficultyPair(card exportGameCard, cardSelected bool, bucketIndex int, bucketName string) string {
	palette := theme.Current()
	bg, fg := exportDifficultyColors(bucketIndex, palette)
	pairStyle := lipgloss.NewStyle().
		Foreground(fg).
		Background(bg).
		Bold(true).
		Padding(0, 0)
	if strings.TrimSpace(s.values.Seed) != "" && !card.bucketSeeded(bucketIndex) {
		pairStyle = pairStyle.Foreground(palette.TextDim).Background(theme.Blend(palette.BG, bg, 0.35))
	}

	pair := pairStyle.Render(fmt.Sprintf("%s%d", string(bucketName[0]), card.Buckets[bucketIndex]))
	if cardSelected && s.focus == exportFocusCards && bucketIndex == s.bucketIndex {
		return puzzlegame.CursorStyle().Render(puzzlegame.CursorLeft) + pair + puzzlegame.CursorStyle().Render(puzzlegame.CursorRight)
	}
	return pair
}

func (s *exportModel) handleMouseClick(msg tea.MouseClickMsg, screenWidth, screenHeight, contentWidth, contentHeight int) (tea.Cmd, bool) {
	mouse := msg.Mouse()
	content := s.view(contentWidth, contentHeight)
	_, _, contentX, contentY := exportPanelContentOrigin(screenWidth, screenHeight, content)
	localX := mouse.X - contentX
	localY := mouse.Y - contentY
	if localX < 0 || localY < 0 {
		return nil, false
	}

	metrics := s.contentMetrics(contentWidth, contentHeight)
	if cmd, handled := s.handleSettingsClick(localX, localY, metrics); handled {
		return cmd, true
	}
	if metrics.SummaryButton.contains(localX, localY) {
		s.focus = exportFocusSubmit
		return tea.Batch(s.applyFocus(), exportSubmitCmd), true
	}
	if cmd, handled := s.handleCardClick(localX, localY, metrics); handled {
		return cmd, true
	}
	return nil, false
}

func exportPanelContentOrigin(screenWidth, screenHeight int, content string) (panelX, panelY, contentX, contentY int) {
	panel := ui.Panel("Export", content, exportFooterHint)
	panelWidth := lipgloss.Width(panel)
	panelHeight := lipgloss.Height(panel)
	panelX = max((screenWidth-panelWidth)/2, 0)
	panelY = max((screenHeight-panelHeight)/2, 0)

	frame := ui.PanelFrame()
	contentX = panelX + frame.GetHorizontalFrameSize()/2
	contentY = panelY + 1 + 2 + 2
	return panelX, panelY, contentX, contentY
}

func (s *exportModel) handleMouseWheel(msg tea.MouseWheelMsg, screenWidth, screenHeight, contentWidth, contentHeight int) bool {
	mouse := msg.Mouse()
	content := s.view(contentWidth, contentHeight)
	_, _, contentX, contentY := exportPanelContentOrigin(screenWidth, screenHeight, content)
	localX := mouse.X - contentX
	localY := mouse.Y - contentY
	if localX < 0 || localY < 0 {
		return false
	}
	metrics := s.contentMetrics(contentWidth, contentHeight)
	if !metrics.CardsRect.contains(localX, localY) {
		return false
	}
	totalRows := max((len(s.cards)+metrics.CardColumns-1)/metrics.CardColumns, 1)
	if totalRows <= metrics.VisibleRows {
		return false
	}
	switch mouse.Button {
	case tea.MouseWheelDown:
		s.cardRowOffset = min(s.cardRowOffset+1, totalRows-metrics.VisibleRows)
	case tea.MouseWheelUp:
		s.cardRowOffset = max(s.cardRowOffset-1, 0)
	default:
		return false
	}
	s.clampSelectionToVisibleRows(metrics.CardColumns, metrics.VisibleRows)
	return true
}

func (s *exportModel) clampSelectionToVisibleRows(cols, visibleRows int) {
	if len(s.cards) == 0 || cols <= 0 || visibleRows <= 0 {
		return
	}
	row := s.cardIndex / cols
	col := s.cardIndex % cols
	if row < s.cardRowOffset {
		row = s.cardRowOffset
	}
	maxVisibleRow := s.cardRowOffset + visibleRows - 1
	if row > maxVisibleRow {
		row = maxVisibleRow
	}
	index := row*cols + col
	if index >= len(s.cards) {
		index = len(s.cards) - 1
	}
	s.cardIndex = max(index, 0)
}

func (s *exportModel) handleSettingsClick(localX, localY int, metrics exportContentMetrics) (tea.Cmd, bool) {
	settings := metrics.Settings
	switch {
	case settings.Title.contains(localX, localY):
		return s.focusTextInput(exportFocusTitle, &s.titleInput, settings.Title, "Title", localX), true
	case settings.Advert.contains(localX, localY):
		return s.focusTextInput(exportFocusAdvert, &s.advertInput, settings.Advert, "Advert", localX), true
	case settings.Header.contains(localX, localY):
		return s.focusTextarea(settings.Header, localX, localY), true
	case settings.Volume.contains(localX, localY):
		return s.focusTextInput(exportFocusVolume, &s.volumeInput, settings.Volume, "Volume", localX), true
	case settings.Layout.contains(localX, localY):
		s.focus = exportFocusLayout
		s.cycleLayout(1)
		return s.applyFocus(), true
	case settings.Seed.contains(localX, localY):
		return s.focusTextInput(exportFocusSeed, &s.seedInput, settings.Seed, "Seed", localX), true
	case settings.JSONL.contains(localX, localY):
		s.focus = exportFocusJSONLToggle
		s.toggleJSONL()
		return s.applyFocus(), true
	case settings.PDFPath.contains(localX, localY):
		return s.focusTextInput(exportFocusPDFPath, &s.pdfPathInput, settings.PDFPath, "PDF Path", localX), true
	case settings.HasJSONL && settings.JSONLPath.contains(localX, localY):
		return s.focusTextInput(exportFocusJSONLPath, &s.jsonlPathInput, settings.JSONLPath, "JSONL Path", localX), true
	default:
		return nil, false
	}
}

func (s *exportModel) focusTextInput(focus exportFocus, input *textinput.Model, rect exportRect, label string, localX int) tea.Cmd {
	s.focus = focus
	labelWidth := lipgloss.Width(label) + 1
	cursorPos := max(localX-rect.X-2-labelWidth, 0)
	input.SetCursor(cursorPos)
	return s.applyFocus()
}

func (s *exportModel) focusTextarea(rect exportRect, localX, localY int) tea.Cmd {
	s.focus = exportFocusHeader
	line := max(localY-rect.Y-2, 0)
	col := max(localX-rect.X-2, 0)
	maxLine := max(s.headerInput.LineCount()-1, 0)
	line = min(line, maxLine)
	s.headerInput.MoveToBegin()
	for i := 0; i < line; i++ {
		s.headerInput.CursorDown()
	}
	s.headerInput.SetCursorColumn(col)
	return s.applyFocus()
}

func (s *exportModel) handleCardClick(localX, localY int, metrics exportContentMetrics) (tea.Cmd, bool) {
	if len(s.cards) == 0 {
		return nil, false
	}
	for _, layout := range metrics.CardRects {
		if !layout.Rect.contains(localX, localY) {
			continue
		}
		for bucketIndex, bucket := range layout.Buckets {
			if bucket.contains(localX, localY) {
				return s.selectCardBucket(layout.Index, bucketIndex, metrics), true
			}
		}
		return s.selectCardBucket(layout.Index, s.bucketIndex, metrics), true
	}
	return nil, false
}

func (s *exportModel) selectCardBucket(cardIndex, bucketIndex int, metrics exportContentMetrics) tea.Cmd {
	s.focus = exportFocusCards
	s.cardIndex = cardIndex
	s.bucketIndex = min(max(bucketIndex, 0), len(bucketNames)-1)
	s.ensureCardSelectionVisible(metrics.Width, metrics.Height)
	return s.applyFocus()
}

type exportContentMetrics struct {
	Width          int
	Height         int
	SettingsHeight int
	SummaryHeight  int
	Settings       exportSettingsRects
	Summary        exportRect
	SummaryButton  exportRect
	CardsRect      exportRect
	CardsY         int
	CardWidth      int
	CardHeight     int
	CardColumns    int
	VisibleRows    int
	CardRects      []exportCardRects
}

func exportDifficultyColors(bucketIndex int, palette theme.Palette) (bg, fg color.Color) {
	switch bucketIndex {
	case 0:
		bg = palette.ErrorBG
	case 1:
		bg = palette.SuccessBG
	default:
		bg = theme.Blend(palette.BG, palette.Secondary, 0.35)
	}
	fg = theme.TextOnBG(bg)
	return bg, fg
}

func exportCardWidth(width, cols int) int {
	if cols <= 0 {
		return width
	}
	return max((width-(cols-1)*exportCardGap)/cols, exportCardMinWidth)
}

func joinCardParts(parts []string) []string {
	if len(parts) == 0 {
		return nil
	}
	joined := make([]string, 0, len(parts)*2-1)
	for i, part := range parts {
		if i > 0 {
			joined = append(joined, strings.Repeat(" ", exportCardGap))
		}
		joined = append(joined, part)
	}
	return joined
}

func truncateLabel(s string, width int) string {
	if width <= 0 {
		return ""
	}
	runes := []rune(s)
	if len(runes) <= width {
		return s
	}
	if width == 1 {
		return string(runes[:1])
	}
	return string(runes[:width-1]) + "…"
}

func (s exportModel) toSpecAndConfig() (packexport.Spec, config.ExportConfig) {
	sync := s
	sync.syncValuesFromInputs()

	spec := sync.values.toExactSpec()
	spec.Counts = expandCardCounts(sync.cards)
	if !sync.values.JSONLEnabled {
		spec.JSONLOutputPath = ""
	}

	cfg := config.ExportConfig{
		Title:           spec.Title,
		Header:          spec.Header,
		Advert:          spec.Advert,
		Volume:          spec.Volume,
		SheetLayout:     spec.SheetLayout,
		Seed:            spec.Seed,
		PDFOutputPath:   spec.PDFOutputPath,
		JSONLEnabled:    sync.values.JSONLEnabled,
		JSONLOutputPath: sync.values.JSONLOutputPath,
		Counts:          spec.Counts,
	}
	return spec, cfg
}

func (s *exportModel) syncValuesFromInputs() {
	s.values.Title = s.titleInput.Value()
	s.values.Header = s.headerInput.Value()
	s.values.Advert = s.advertInput.Value()
	s.values.Volume = s.volumeInput.Value()
	s.values.Seed = s.seedInput.Value()
	s.values.PDFOutputPath = s.pdfPathInput.Value()
	s.values.JSONLOutputPath = s.jsonlPathInput.Value()
}

func (v exportFormValues) toExactSpec() packexport.Spec {
	volume, err := strconv.Atoi(strings.TrimSpace(v.Volume))
	if err != nil {
		volume = -1
	}

	spec := packexport.Spec{
		Title:         strings.TrimSpace(v.Title),
		Header:        strings.TrimSpace(v.Header),
		Advert:        strings.TrimSpace(v.Advert),
		Volume:        volume,
		SheetLayout:   strings.TrimSpace(v.SheetLayout),
		Seed:          strings.TrimSpace(v.Seed),
		PDFOutputPath: strings.TrimSpace(v.PDFOutputPath),
		Counts:        map[puzzle.GameID]map[puzzle.ModeID]int{},
	}
	if v.JSONLEnabled {
		spec.JSONLOutputPath = strings.TrimSpace(v.JSONLOutputPath)
	}
	return spec
}

func expandCardCounts(cards []exportGameCard) map[puzzle.GameID]map[puzzle.ModeID]int {
	counts := make(map[puzzle.GameID]map[puzzle.ModeID]int, len(cards))
	for _, card := range cards {
		if _, ok := counts[card.GameID]; !ok {
			counts[card.GameID] = make(map[puzzle.ModeID]int)
		}
		for bucketIndex, total := range card.Buckets {
			modes := card.BucketModes[bucketIndex]
			if len(modes) == 0 {
				continue
			}
			base := total / len(modes)
			remainder := total % len(modes)
			for i, mode := range modes {
				count := base
				if i < remainder {
					count++
				}
				counts[mode.GameID][mode.ModeID] = count
			}
		}
	}
	return counts
}

// runeIdxBeforeExt returns the rune index just before the file extension so
// that focusing a path field via tab lands the cursor there. This lets users
// type a new base-name suffix (e.g. "2") without overwriting ".pdf"/".jsonl".
func runeIdxBeforeExt(path string) int {
	ext := filepath.Ext(path)
	runes := []rune(path)
	extRunes := len([]rune(ext))
	if extRunes == 0 || extRunes >= len(runes) {
		return len(runes)
	}
	return len(runes) - extRunes
}
