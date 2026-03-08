package app

import (
	"strings"

	"charm.land/bubbles/v2/list"
	"github.com/FelineStateMachine/puzzletea/puzzle"
	"github.com/FelineStateMachine/puzzletea/registry"
)

type modeDisplayItem struct {
	mode        registry.ModeEntry
	title       string
	description string
	filterValue string
}

func (i modeDisplayItem) Title() string                { return i.title }
func (i modeDisplayItem) Description() string          { return i.description }
func (i modeDisplayItem) FilterValue() string          { return i.filterValue }
func (i modeDisplayItem) Original() registry.ModeEntry { return i.mode }

func buildModeDisplayItems(entry registry.Entry) []list.Item {
	titles := modeDisplayTitles(entry)
	items := make([]list.Item, 0, len(entry.Modes))

	for idx, mode := range entry.Modes {
		displayTitle := titles[idx]
		filterValue := displayTitle + " " + mode.Definition.Description
		if displayTitle != mode.Definition.Title {
			filterValue += " " + mode.Definition.Title
		}

		items = append(items, modeDisplayItem{
			mode:        mode,
			title:       displayTitle,
			description: mode.Definition.Description,
			filterValue: strings.TrimSpace(filterValue),
		})
	}

	return items
}

func modeDisplayTitles(entry registry.Entry) []string {
	counts := make(map[string]int, len(entry.Modes))
	titles := make([]string, 0, len(entry.Modes))

	for _, mode := range entry.Modes {
		base, _ := trimGridSizeSuffix(mode.Definition.Title)
		counts[puzzle.NormalizeName(base)]++
	}

	for _, mode := range entry.Modes {
		title := mode.Definition.Title
		base, hadSize := trimGridSizeSuffix(title)
		if hadSize && counts[puzzle.NormalizeName(base)] == 1 {
			title = base
		}
		titles = append(titles, title)
	}

	return titles
}

func unwrapModeDisplayItem(item list.Item) any {
	displayItem, ok := item.(interface{ Original() registry.ModeEntry })
	if !ok {
		return item
	}
	return displayItem.Original()
}

func trimGridSizeSuffix(title string) (string, bool) {
	fields := strings.Fields(title)
	if len(fields) < 2 {
		return title, false
	}

	if !isGridSizeToken(fields[len(fields)-1]) {
		return title, false
	}

	base := strings.TrimSpace(strings.Join(fields[:len(fields)-1], " "))
	if base == "" {
		return title, false
	}

	return base, true
}

func isGridSizeToken(token string) bool {
	width, height, found := strings.Cut(token, "x")
	if !found || width == "" || height == "" {
		return false
	}
	return digitsOnly(width) && digitsOnly(height)
}

func digitsOnly(s string) bool {
	for _, r := range s {
		if r < '0' || r > '9' {
			return false
		}
	}
	return true
}
