package app

import (
	"strings"
	"testing"

	"github.com/FelineStateMachine/puzzletea/stats"
	"github.com/FelineStateMachine/puzzletea/ui"

	"charm.land/lipgloss/v2"
)

func TestHelpViewportSizeKeepsPanelInsideInsetBounds(t *testing.T) {
	const (
		windowWidth  = 120
		windowHeight = 40
	)

	helpWidth, helpHeight := helpViewportSize(windowWidth, windowHeight)
	line := strings.Repeat("x", helpWidth)
	content := strings.TrimSuffix(strings.Repeat(line+"\n", helpHeight), "\n")
	panel := ui.Panel("Guide", content, "esc back")

	if got := lipgloss.Width(panel); got > windowWidth-(helpPanelInsetX*2) {
		t.Fatalf("panel width = %d, want <= %d", got, windowWidth-(helpPanelInsetX*2))
	}
	if got := lipgloss.Height(panel); got > windowHeight-(helpPanelInsetY*2) {
		t.Fatalf("panel height = %d, want <= %d", got, windowHeight-(helpPanelInsetY*2))
	}
}

func TestHelpViewportSizeNeverFallsBelowOne(t *testing.T) {
	helpWidth, helpHeight := helpViewportSize(1, 1)
	if helpWidth != 1 {
		t.Fatalf("help width = %d, want 1", helpWidth)
	}
	if helpHeight != 1 {
		t.Fatalf("help height = %d, want 1", helpHeight)
	}
}

func TestStatsViewportSizeMatchesHelpWidth(t *testing.T) {
	helpWidth, _ := helpViewportSize(120, 40)
	statsWidth, _ := statsViewportSize(120, 40, nil)
	if statsWidth != helpWidth {
		t.Fatalf("stats width = %d, want %d", statsWidth, helpWidth)
	}
}

func TestStatsViewportSizeReservesBannerHeight(t *testing.T) {
	_, helpHeight := helpViewportSize(120, 40)
	cards := []stats.Card{{GameType: "Nurikabe"}}
	_, statsHeight := statsViewportSize(120, 40, cards)
	if statsHeight != helpHeight-5 {
		t.Fatalf("stats height = %d, want %d", statsHeight, helpHeight-5)
	}
}

func TestGameSelectViewContentSizeIsStableAcrossCategories(t *testing.T) {
	l := ui.InitCategoryList(GameCategories, "Select Category")
	m := model{
		state:  gameSelectView,
		width:  120,
		height: 40,
		nav: navigationState{
			gameSelectList: l,
		},
	}
	m = m.updateCategoryDetailViewport()

	wantWidth := lipgloss.Width(m.gameSelectViewContent())
	wantHeight := lipgloss.Height(m.gameSelectViewContent())

	for i := range m.nav.gameSelectList.Items() {
		m.nav.gameSelectList.Select(i)
		m = m.updateCategoryDetailViewport()

		if got := lipgloss.Width(m.gameSelectViewContent()); got != wantWidth {
			t.Fatalf("selection %d width = %d, want %d", i, got, wantWidth)
		}
		if got := lipgloss.Height(m.gameSelectViewContent()); got != wantHeight {
			t.Fatalf("selection %d height = %d, want %d", i, got, wantHeight)
		}
	}
}
