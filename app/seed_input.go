package app

import (
	"strings"

	"github.com/FelineStateMachine/puzzletea/catalog"
	"github.com/FelineStateMachine/puzzletea/game"
	"github.com/FelineStateMachine/puzzletea/theme"
	"github.com/FelineStateMachine/puzzletea/ui"

	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

const randomSeedModeLabel = "Random"

func buildSeedModeOptions(definitions []game.Definition) []seedModeOption {
	options := []seedModeOption{{label: randomSeedModeLabel}}

	for _, def := range definitions {
		if !definitionHasSeededModes(def) {
			continue
		}

		for _, item := range def.Modes {
			if _, ok := item.(game.SeededSpawner); !ok {
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
		return game.NormalizeName(gameType)
	}
	return game.NormalizeName(gameType) + "::" + game.NormalizeName(modeTitle)
}

func definitionHasSeededModes(def game.Definition) bool {
	for _, item := range def.Modes {
		if _, ok := item.(game.SeededSpawner); ok {
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

	options := buildSeedModeOptions(catalog.All)
	index := findSeedModeIndex(options, m.nav.lastSeedModeKey)

	m.nav.seedInput = ti
	m.nav.seedModeOptions = options
	m.nav.seedModeIndex = index
	m.nav.seedFocus = seedFocusText
	m.state = seedInputView
	return m, m.nav.seedInput.Focus()
}

func (m model) currentSeedMode() seedModeOption {
	if len(m.nav.seedModeOptions) == 0 {
		return seedModeOption{label: randomSeedModeLabel}
	}
	if m.nav.seedModeIndex < 0 || m.nav.seedModeIndex >= len(m.nav.seedModeOptions) {
		return m.nav.seedModeOptions[0]
	}
	return m.nav.seedModeOptions[m.nav.seedModeIndex]
}

func (m model) moveSeedMode(step int) model {
	if len(m.nav.seedModeOptions) == 0 {
		return m
	}

	index := m.nav.seedModeIndex + step
	for index < 0 {
		index += len(m.nav.seedModeOptions)
	}
	m.nav.seedModeIndex = index % len(m.nav.seedModeOptions)
	m.nav.lastSeedModeKey = m.currentSeedMode().key
	return m
}

func (m model) setSeedFocus(focus seedInputFocus) (model, tea.Cmd) {
	m.nav.seedFocus = focus
	if focus == seedFocusText {
		return m, m.nav.seedInput.Focus()
	}
	m.nav.seedInput.Blur()
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
			if m.nav.seedFocus == seedFocusMode {
				return m.moveSeedMode(-1), nil
			}
		case "right", "l":
			if m.nav.seedFocus == seedFocusMode {
				return m.moveSeedMode(1), nil
			}
		}
	}

	if m.nav.seedFocus != seedFocusText {
		return m, nil
	}

	var cmd tea.Cmd
	m.nav.seedInput, cmd = m.nav.seedInput.Update(msg)
	return m, cmd
}

func (m model) seedInputBody() string {
	selector := renderSeedModeTitle(m.currentSeedMode().label, m.nav.seedModeIndex)
	modePrefix := "  "
	if m.nav.seedFocus == seedFocusMode {
		modePrefix = ui.CursorStyle().Render("> ")
	}

	seedInputView := m.nav.seedInput.View()
	if m.nav.seedFocus != seedFocusText {
		seedInputView = strings.Replace(seedInputView, ">", " ", 1)
	}

	return lipgloss.JoinVertical(
		lipgloss.Left,
		seedInputView,
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
