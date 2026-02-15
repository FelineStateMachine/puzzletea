package ui

// MenuItem is a simple list.Item implementation for menu entries.
type MenuItem struct {
	ItemTitle string
	Desc      string
}

func (i MenuItem) Title() string       { return i.ItemTitle }
func (i MenuItem) Description() string { return i.Desc }
func (i MenuItem) FilterValue() string { return i.ItemTitle }
