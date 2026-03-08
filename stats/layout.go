package stats

const (
	CardInnerWidth = 30
	CardHeight     = 9
)

const (
	panelChrome   = 8
	bannerHeight  = 5
	CardFullWidth = CardInnerWidth + 4
)

const (
	cardColumnGap = 2
	cardRowGap    = 2
)

func StaticHeight(cards []Card) int {
	if len(cards) == 0 {
		return panelChrome
	}
	return panelChrome + bannerHeight
}

func ContentWidth(termWidth int) int {
	available := max(termWidth-6, CardFullWidth)
	cols := cardColumnCount(available)
	width := cardGridWidth(cols)
	if width > available {
		return available
	}
	return width
}

func ViewportHeight(availableHeight int) int {
	maxRows := CardHeight*2 + CardHeight/2
	minRows := CardHeight
	h := min(availableHeight, maxRows)
	return max(h, minRows)
}

func cardColumnCount(width int) int {
	if width <= CardFullWidth {
		return 1
	}
	cols := (width + cardColumnGap) / (CardFullWidth + cardColumnGap)
	return max(cols, 1)
}

func cardGridWidth(cols int) int {
	if cols < 1 {
		cols = 1
	}
	return cols*CardFullWidth + (cols-1)*cardColumnGap
}
