package app

import (
	"strings"

	"charm.land/bubbles/v2/list"
	"github.com/FelineStateMachine/puzzletea/puzzle"
	"github.com/FelineStateMachine/puzzletea/registry"
)

type categoryItem struct {
	entry registry.Entry
}

func (i categoryItem) Title() string       { return i.entry.Definition.Name }
func (i categoryItem) Description() string { return i.entry.Definition.Description }
func (i categoryItem) FilterValue() string {
	parts := []string{i.entry.Definition.Name, i.entry.Definition.Description}
	parts = append(parts, i.entry.Definition.Aliases...)
	return strings.TrimSpace(strings.Join(parts, " "))
}

func buildCategoryItems() []list.Item {
	entries := registry.Entries()
	items := make([]list.Item, 0, len(entries))
	for _, entry := range entries {
		items = append(items, categoryItem{entry: entry})
	}
	return items
}

func selectedCategoryEntry(item list.Item) (registry.Entry, bool) {
	category, ok := item.(categoryItem)
	if !ok {
		return registry.Entry{}, false
	}
	return category.entry, true
}

func categoryEntryByName(name string) (registry.Entry, bool) {
	return registry.Resolve(name)
}

func normalizeCategoryName(name string) string {
	return puzzle.NormalizeName(name)
}
