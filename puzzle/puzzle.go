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

type Definition struct {
	ID           GameID
	Name         string
	Description  string
	Aliases      []string
	Modes        []ModeDef
	DailyModeIDs []ModeID
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
