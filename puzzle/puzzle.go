package puzzle

import (
	"strings"

	"github.com/FelineStateMachine/puzzletea/difficulty"
)

type GameID string

type ModeID string

type ModeDef struct {
	ID          ModeID
	Title       string
	Description string
	Seeded      bool
	PresetElo   *difficulty.Elo
}

type VariantID string

type VariantDef struct {
	ID          VariantID
	Title       string
	Description string
	DefaultElo  difficulty.Elo
}

type VariantSpec struct {
	ID          VariantID
	Title       string
	Description string
	DefaultElo  difficulty.Elo
}

type LegacyModeAlias struct {
	ID              ModeID
	Title           string
	Description     string
	TargetVariantID VariantID
	PresetElo       difficulty.Elo
	CLIAliases      []string
}

type LegacyModeAliasSpec struct {
	ID              ModeID
	Title           string
	Description     string
	TargetVariantID VariantID
	PresetElo       difficulty.Elo
	CLIAliases      []string
}

type ModeSpec struct {
	ID          ModeID
	Title       string
	Description string
	Seeded      bool
	PresetElo   *difficulty.Elo
}

type Definition struct {
	ID           GameID
	Name         string
	Description  string
	Aliases      []string
	Variants     []VariantDef
	LegacyModes  []LegacyModeAlias
	Modes        []ModeDef
	DailyModeIDs []ModeID
}

type DefinitionSpec struct {
	ID           GameID
	Name         string
	Description  string
	Aliases      []string
	Variants     []VariantDef
	LegacyModes  []LegacyModeAlias
	Modes        []ModeDef
	DailyModeIDs []ModeID
}

func NewModeDef(spec ModeSpec) ModeDef {
	id := spec.ID
	if id == "" {
		id = CanonicalModeID(spec.Title)
	}
	return ModeDef{
		ID:          id,
		Title:       spec.Title,
		Description: spec.Description,
		Seeded:      spec.Seeded,
		PresetElo:   cloneElo(spec.PresetElo),
	}
}

func NewVariantDef(spec VariantSpec) VariantDef {
	id := spec.ID
	if id == "" {
		id = CanonicalVariantID(spec.Title)
	}
	return VariantDef{
		ID:          id,
		Title:       spec.Title,
		Description: spec.Description,
		DefaultElo:  spec.DefaultElo,
	}
}

func NewLegacyModeAlias(spec LegacyModeAliasSpec) LegacyModeAlias {
	id := spec.ID
	if id == "" {
		id = CanonicalModeID(spec.Title)
	}
	return LegacyModeAlias{
		ID:              id,
		Title:           spec.Title,
		Description:     spec.Description,
		TargetVariantID: spec.TargetVariantID,
		PresetElo:       spec.PresetElo,
		CLIAliases:      append([]string(nil), spec.CLIAliases...),
	}
}

func NewDefinition(spec DefinitionSpec) Definition {
	id := spec.ID
	if id == "" {
		id = CanonicalGameID(spec.Name)
	}
	return Definition{
		ID:           id,
		Name:         spec.Name,
		Description:  spec.Description,
		Aliases:      append([]string(nil), spec.Aliases...),
		Variants:     append([]VariantDef(nil), spec.Variants...),
		LegacyModes:  append([]LegacyModeAlias(nil), spec.LegacyModes...),
		Modes:        append([]ModeDef(nil), spec.Modes...),
		DailyModeIDs: append([]ModeID(nil), spec.DailyModeIDs...),
	}
}

func cloneElo(elo *difficulty.Elo) *difficulty.Elo {
	if elo == nil {
		return nil
	}
	v := *elo
	return &v
}

func SelectModeIDsByIndex(modes []ModeDef, indexes ...int) []ModeID {
	selected := make([]ModeID, 0, len(indexes))
	for _, idx := range indexes {
		if idx < 0 || idx >= len(modes) {
			continue
		}
		selected = append(selected, modes[idx].ID)
	}
	return selected
}

func NormalizeName(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	s = strings.ReplaceAll(s, "-", " ")
	s = strings.ReplaceAll(s, "_", " ")
	return strings.Join(strings.Fields(s), " ")
}

func CanonicalGameID(name string) GameID {
	return GameID(NormalizeName(name))
}

func CanonicalModeID(title string) ModeID {
	return ModeID(NormalizeName(title))
}

func CanonicalVariantID(title string) VariantID {
	return VariantID(NormalizeName(title))
}

func (d Definition) HasMode(id ModeID) bool {
	for _, mode := range d.Modes {
		if mode.ID == id {
			return true
		}
	}
	return false
}

func (d Definition) HasVariant(id VariantID) bool {
	for _, variant := range d.Variants {
		if variant.ID == id {
			return true
		}
	}
	return false
}
