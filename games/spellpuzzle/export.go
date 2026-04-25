package spellpuzzle

import "encoding/json"

type Save struct {
	ModeTitle    string          `json:"mode_title"`
	Bank         string          `json:"bank"`
	BankColors   []int           `json:"bank_colors,omitempty"`
	Placements   []WordPlacement `json:"placements"`
	BonusWords   []string        `json:"bonus_words"`
	Trace        []int           `json:"trace"`
	BankCursor   int             `json:"bank_cursor"`
	Solved       bool            `json:"solved"`
	Feedback     feedback        `json:"feedback"`
	ShowFullHelp bool            `json:"show_full_help"`
}

func ImportModel(data []byte) (*Model, error) {
	var exported Save
	if err := json.Unmarshal(data, &exported); err != nil {
		return nil, err
	}

	model := &Model{
		modeTitle:      exported.ModeTitle,
		bank:           []rune(exported.Bank),
		bankColorSlots: append([]int(nil), exported.BankColors...),
		placements:     append([]WordPlacement(nil), exported.Placements...),
		bonusWords:     append([]string(nil), exported.BonusWords...),
		trace:          append([]int(nil), exported.Trace...),
		bankCursor:     exported.BankCursor,
		solved:         exported.Solved,
		feedback:       exported.Feedback,
		showFullHelp:   exported.ShowFullHelp,
		keys:           DefaultKeyMap,
	}
	model.rebuildDerivedState()
	return model, nil
}

func (m Model) GetSave() ([]byte, error) {
	data := Save{
		ModeTitle:    m.modeTitle,
		Bank:         string(m.bank),
		BankColors:   append([]int(nil), m.bankColorSlots...),
		Placements:   append([]WordPlacement(nil), m.placements...),
		BonusWords:   append([]string(nil), m.bonusWords...),
		Trace:        append([]int(nil), m.trace...),
		BankCursor:   m.bankCursor,
		Solved:       m.solved,
		Feedback:     m.feedback,
		ShowFullHelp: m.showFullHelp,
	}
	return json.Marshal(data)
}
