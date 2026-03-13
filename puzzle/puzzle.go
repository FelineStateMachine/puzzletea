package puzzle

import "strings"

type GameID string

type ModeID string

type ModeDef struct {
	ID          ModeID
	Title       string
	Description string
	Seeded      bool
}

type ModeSpec struct {
	ID          ModeID
	Title       string
	Description string
	Seeded      bool
}

type Definition struct {
	ID           GameID
	Name         string
	Description  string
	Aliases      []string
	Modes        []ModeDef
	DailyModeIDs []ModeID
}

type DefinitionSpec struct {
	ID           GameID
	Name         string
	Description  string
	Aliases      []string
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
		Modes:        append([]ModeDef(nil), spec.Modes...),
		DailyModeIDs: append([]ModeID(nil), spec.DailyModeIDs...),
	}
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

func (d Definition) HasMode(id ModeID) bool {
	for _, mode := range d.Modes {
		if mode.ID == id {
			return true
		}
	}
	return false
}
