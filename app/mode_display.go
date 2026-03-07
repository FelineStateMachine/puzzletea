package app

import (
	"strings"

	"charm.land/bubbles/v2/list"
	"github.com/FelineStateMachine/puzzletea/game"
)

type modeDisplayItem struct {
	item        list.Item
	title       string
	description string
	filterValue string
}

func (i modeDisplayItem) Title() string       { return i.title }
func (i modeDisplayItem) Description() string { return i.description }
func (i modeDisplayItem) FilterValue() string { return i.filterValue }
func (i modeDisplayItem) Original() list.Item { return i.item }

func buildModeDisplayItems(cat game.Category) []list.Item {
	titles := modeDisplayTitles(cat)
	items := make([]list.Item, 0, len(cat.Modes))

	for idx, item := range cat.Modes {
		mode, ok := item.(game.Mode)
		if !ok {
			items = append(items, item)
			continue
		}

		displayTitle := titles[idx]
		filterValue := displayTitle + " " + mode.Description()
		if displayTitle != mode.Title() {
			filterValue += " " + mode.Title()
		}

		items = append(items, modeDisplayItem{
			item:        item,
			title:       displayTitle,
			description: mode.Description(),
			filterValue: strings.TrimSpace(filterValue),
		})
	}

	return items
}

func modeDisplayTitles(cat game.Category) []string {
	counts := make(map[string]int, len(cat.Modes))
	titles := make([]string, 0, len(cat.Modes))

	for _, item := range cat.Modes {
		mode, ok := item.(game.Mode)
		if !ok {
			continue
		}
		base, _ := trimGridSizeSuffix(mode.Title())
		counts[game.NormalizeName(base)]++
	}

	for _, item := range cat.Modes {
		mode, ok := item.(game.Mode)
		if !ok {
			continue
		}

		title := mode.Title()
		base, hadSize := trimGridSizeSuffix(title)
		if hadSize && counts[game.NormalizeName(base)] == 1 {
			title = base
		}
		titles = append(titles, title)
	}

	return titles
}

func unwrapModeDisplayItem(item list.Item) list.Item {
	displayItem, ok := item.(interface{ Original() list.Item })
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
