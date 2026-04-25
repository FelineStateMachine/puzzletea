package app

import (
	"strings"

	"github.com/FelineStateMachine/puzzletea/puzzle"
	"github.com/FelineStateMachine/puzzletea/registry"
	"github.com/FelineStateMachine/puzzletea/theme"
	"github.com/FelineStateMachine/puzzletea/ui"

	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

const randomSeedModeLabel = "Random"

func buildSeedModeOptions(definitions []puzzle.Definition) []seedModeOption {
	options := []seedModeOption{{label: randomSeedModeLabel}}

	for _, def := range definitions {
		if !definitionHasSeededModes(def) {
			continue
		}

		for _, mode := range def.Modes {
			if !mode.Seeded {
				continue
			}

			options = append(options, seedModeOption{
				key:      seedModeKey(def.Name, ""),
				label:    def.Name,
				gameType: def.Name,
			})
			break
		}
	}

	return options
}

func seedModeKey(gameType, modeTitle string) string {
	if modeTitle == "" {
		return puzzle.NormalizeName(gameType)
	}
	return puzzle.NormalizeName(gameType) + "::" + puzzle.NormalizeName(modeTitle)
}

func definitionHasSeededModes(def puzzle.Definition) bool {
	for _, mode := range def.Modes {
		if mode.Seeded {
			return true
		}
	}
	return false
}

func findSeedModeIndex(options []seedModeOption, key string) int {
	if key == "" {
		return 0
	}
	for i := range options {
		if options[i].key == key {
			return i
		}
	}
	return 0
}

func (m model) enterSeedInputView() (model, tea.Cmd) {
	ti := textinput.New()
	ti.Placeholder = "any word or phrase"
	ti.CharLimit = 64
	ti.SetWidth(min(m.width, 48))

	options := buildSeedModeOptions(registry.Definitions())
	index := findSeedModeIndex(options, m.seed.lastModeKey)

	m.seed.input = ti
	m.seed.modeOptions = options
	m.seed.modeIndex = index
	m.seed.focus = seedFocusText
	m.state = seedInputView
	m = m.initScreen(seedInputView)
	return m, m.seed.input.Focus()
}

func (m model) currentSeedMode() seedModeOption {
	return currentSeedMode(m.seed)
}

func currentSeedMode(seed seedState) seedModeOption {
	if len(seed.modeOptions) == 0 {
		return seedModeOption{label: randomSeedModeLabel}
	}
	if seed.modeIndex < 0 || seed.modeIndex >= len(seed.modeOptions) {
		return seed.modeOptions[0]
	}
	return seed.modeOptions[seed.modeIndex]
}

func (m model) moveSeedMode(step int) model {
	if len(m.seed.modeOptions) == 0 {
		return m
	}

	index := m.seed.modeIndex + step
	for index < 0 {
		index += len(m.seed.modeOptions)
	}
	m.seed.modeIndex = index % len(m.seed.modeOptions)
	m.seed.lastModeKey = m.currentSeedMode().key
	return m
}

func (m model) setSeedFocus(focus seedInputFocus) (model, tea.Cmd) {
	m.seed.focus = focus
	if focus == seedFocusText {
		return m, m.seed.input.Focus()
	}
	m.seed.input.Blur()
	return m, nil
}

func (m model) handleSeedInputUpdate(msg tea.Msg) (model, tea.Cmd) {
	keyMsg, ok := msg.(tea.KeyPressMsg)
	if ok {
		switch keyMsg.String() {
		case "up", "k":
			return m.setSeedFocus(seedFocusText)
		case "down", "j":
			return m.setSeedFocus(seedFocusMode)
		case "left", "h":
			if m.seed.focus == seedFocusMode {
				return m.moveSeedMode(-1), nil
			}
		case "right", "l":
			if m.seed.focus == seedFocusMode {
				return m.moveSeedMode(1), nil
			}
		}
	}

	if m.seed.focus != seedFocusText {
		return m, nil
	}

	var cmd tea.Cmd
	m.seed.input, cmd = m.seed.input.Update(msg)
	return m, cmd
}

func seedInputBody(seed seedState) string {
	selector := renderSeedModeTitle(currentSeedMode(seed).label, seed.modeIndex)
	modePrefix := "  "
	if seed.focus == seedFocusMode {
		modePrefix = ui.CursorStyle().Render("> ")
	}

	inputView := seed.input.View()
	if seed.focus != seedFocusText {
		inputView = strings.Replace(inputView, ">", " ", 1)
	}

	return lipgloss.JoinVertical(
		lipgloss.Left,
		inputView,
		"",
		modePrefix+selector,
	)
}

func renderSeedModeTitle(label string, colorIndex int) string {
	colors := theme.Current().ThemeColors()
	if len(colors) == 0 {
		return ui.PanelTitle().Render("game: " + label)
	}

	bg := colors[((colorIndex%len(colors))+len(colors))%len(colors)]
	chip := lipgloss.NewStyle().
		Bold(true).
		Foreground(theme.TextOnBG(bg)).
		Background(bg).
		Render(label)

	var b strings.Builder
	b.WriteString(ui.PanelTitle().Render("game: "))
	b.WriteString(chip)
	return b.String()
}
