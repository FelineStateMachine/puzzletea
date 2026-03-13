package app

import (
	"log"

	"github.com/FelineStateMachine/puzzletea/stats"
	"github.com/FelineStateMachine/puzzletea/theme"
	"github.com/FelineStateMachine/puzzletea/ui"

	"charm.land/bubbles/v2/list"
	"charm.land/bubbles/v2/viewport"
	"github.com/charmbracelet/glamour"
)

const (
	helpPanelInsetX         = 2
	helpPanelInsetY         = 1
	helpPanelHorizontalTrim = 6
	helpPanelVerticalTrim   = 8
	categoryPanelChrome     = 8
	categoryBodyMaxWidth    = 86
	categoryBodyMaxHeight   = 16
	categoryMinListWidth    = 24
	categoryMaxListWidth    = 30
	categoryGapWidth        = 2
	categoryDetailTrimX     = 6
	categoryDetailTrimY     = 4
	categoryStackGapHeight  = 1
	categoryMinSideBySideW  = 72
)

func helpViewportSize(width, height int) (int, int) {
	panelWidth := max(width-(helpPanelInsetX*2), 1)
	panelHeight := max(height-(helpPanelInsetY*2), 1)
	contentWidth := max(panelWidth-helpPanelHorizontalTrim, 1)
	contentHeight := max(panelHeight-helpPanelVerticalTrim, 1)
	return contentWidth, contentHeight
}

func helpSelectListSize(width, height int, l list.Model) (int, int) {
	contentWidth, contentHeight := helpViewportSize(width, height)
	listWidth := min(contentWidth, 64)
	listHeight := min(contentHeight, ui.ListHeight(l))
	return listWidth, listHeight
}

func statsViewportSize(width, height int, cards []stats.Card) (int, int) {
	contentWidth, _ := helpViewportSize(width, height)
	panelHeight := max(height-(helpPanelInsetY*2), 1)
	contentHeight := max(panelHeight-stats.StaticHeight(cards), 1)
	return contentWidth, contentHeight
}

type categoryPickerMetrics struct {
	bodyWidth    int
	bodyHeight   int
	listWidth    int
	listHeight   int
	detailWidth  int
	detailHeight int
	stacked      bool
}

func categoryPickerSize(width, height int) categoryPickerMetrics {
	bodyWidth := min(width, categoryBodyMaxWidth)
	bodyHeight := min(max(height-categoryPanelChrome, 1), categoryBodyMaxHeight)

	if bodyWidth < categoryMinSideBySideW {
		listHeight := min(bodyHeight, categoryPickerListHeight())
		detailHeight := max(bodyHeight-listHeight-categoryStackGapHeight, 1)
		if detailHeight == 1 && bodyHeight > 1 {
			listHeight = max(bodyHeight-categoryStackGapHeight-detailHeight, 1)
		}
		return categoryPickerMetrics{
			bodyWidth:    bodyWidth,
			bodyHeight:   bodyHeight,
			listWidth:    bodyWidth,
			listHeight:   listHeight,
			detailWidth:  bodyWidth,
			detailHeight: detailHeight,
			stacked:      true,
		}
	}

	listWidth := min(categoryMaxListWidth, max(categoryMinListWidth, bodyWidth/3))
	detailWidth := max(bodyWidth-listWidth-categoryGapWidth, 1)
	return categoryPickerMetrics{
		bodyWidth:    bodyWidth,
		bodyHeight:   bodyHeight,
		listWidth:    listWidth,
		listHeight:   bodyHeight,
		detailWidth:  detailWidth,
		detailHeight: bodyHeight,
	}
}

func selectedCategoryName(item list.Item) string {
	entry, ok := selectedCategoryEntry(item)
	if !ok {
		return ""
	}
	return entry.Definition.Name
}

func (m model) updateCategoryDetailViewport() model {
	metrics := categoryPickerSize(m.width, m.height)
	contentWidth := max(metrics.detailWidth-categoryDetailTrimX, 1)
	contentHeight := max(metrics.detailHeight-categoryDetailTrimY, 1)

	if m.nav.categoryDetail.Width() == 0 || m.nav.categoryDetail.Height() == 0 {
		m.nav.categoryDetail = viewport.New(
			viewport.WithWidth(contentWidth),
			viewport.WithHeight(contentHeight),
		)
	}
	m.nav.categoryDetail.SetWidth(contentWidth)
	m.nav.categoryDetail.SetHeight(contentHeight)
	m.nav.categoryDetail.FillHeight = true

	entry, ok := selectedCategoryEntry(m.nav.gameSelectList.SelectedItem())
	if !ok {
		m.nav.categoryDetail.SetContent("")
		return m
	}

	m.nav.categoryDetail.SetContent(renderCategoryDetailContent(entry, contentWidth))
	m.nav.categoryDetail.GotoTop()
	return m
}

func (m model) updateHelpDetailViewport() model {
	helpWidth, helpHeight := helpViewportSize(m.width, m.height)
	palette := theme.Current()
	themeKey := helpMarkdownThemeKey(palette)
	if m.help.renderer == nil || m.help.rendererWidth != helpWidth || m.help.rendererTheme != themeKey {
		renderer, err := glamour.NewTermRenderer(
			glamour.WithStyles(helpMarkdownStyle(palette)),
			glamour.WithWordWrap(helpWidth),
			glamour.WithChromaFormatter("terminal16m"),
		)
		if err != nil {
			log.Printf("failed to create help renderer: %v", err)
			m.help.renderer = nil
			m.help.rendererWidth = 0
			m.help.rendererTheme = ""
		} else {
			m.help.renderer = renderer
			m.help.rendererWidth = helpWidth
			m.help.rendererTheme = themeKey
		}
	}

	rendered := m.help.category.Help
	if m.help.renderer != nil {
		out, err := m.help.renderer.Render(m.help.category.Help)
		if err != nil {
			log.Printf("failed to render help: %v", err)
		} else {
			rendered = out
		}
	}

	m.help.viewport = viewport.New(
		viewport.WithWidth(helpWidth),
		viewport.WithHeight(helpHeight),
	)
	m.help.viewport.SetContent(rendered)
	return m
}

func (m model) updateStatsViewport() model {
	statsWidth, statsHeight := statsViewportSize(m.width, m.height, m.stats.cards)
	m.stats.viewport.SetWidth(statsWidth)
	m.stats.viewport.SetHeight(statsHeight)
	m.stats.viewport.SetContent(ui.RenderStatsCardGrid(m.stats.cards, statsWidth))
	return m
}
