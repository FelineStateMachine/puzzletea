package app

import (
	"strings"
	"testing"

	"github.com/FelineStateMachine/puzzletea/stats"
	"github.com/FelineStateMachine/puzzletea/ui"

	"charm.land/lipgloss/v2"
)

func TestHelpViewportSizeKeepsPanelWithinWindowBounds(t *testing.T) {
	const (
		windowWidth  = 120
		windowHeight = 40
	)

	helpWidth, helpHeight := helpViewportSize(windowWidth, windowHeight)
	line := strings.Repeat("x", helpWidth)
	content := strings.TrimSuffix(strings.Repeat(line+"\n", helpHeight), "\n")
	panel := ui.Panel("Guide", content, "esc back")

	if got := lipgloss.Width(panel); got > windowWidth {
		t.Fatalf("panel width = %d, want <= %d", got, windowWidth)
	}
	if got := lipgloss.Height(panel); got > windowHeight {
		t.Fatalf("panel height = %d, want <= %d", got, windowHeight)
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

func TestHelpSelectListSizeKeepsPanelWithinWindowBounds(t *testing.T) {
	const (
		windowWidth  = 120
		windowHeight = 20
	)

	helpList := ui.InitList(gameCategoryItems, "How to Play")
	listWidth, listHeight := helpSelectListSize(windowWidth, windowHeight, helpList)
	helpList.SetSize(listWidth, listHeight)

	panel := ui.Panel(
		"How to Play",
		helpList.View(),
		"↑/↓ navigate • enter select • esc back",
	)

	if got := lipgloss.Width(panel); got > windowWidth {
		t.Fatalf("panel width = %d, want <= %d", got, windowWidth)
	}
	if got := lipgloss.Height(panel); got > windowHeight {
		t.Fatalf("panel height = %d, want <= %d", got, windowHeight)
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

	wantWidth := lipgloss.Width(gameSelectViewContent(m.width, m.height, m.nav.gameSelectList, m.nav.categoryDetail, m.notice))
	wantHeight := lipgloss.Height(gameSelectViewContent(m.width, m.height, m.nav.gameSelectList, m.nav.categoryDetail, m.notice))

	for i := range m.nav.gameSelectList.Items() {
		m.nav.gameSelectList.Select(i)
		m = m.updateCategoryDetailViewport()

		if got := lipgloss.Width(gameSelectViewContent(m.width, m.height, m.nav.gameSelectList, m.nav.categoryDetail, m.notice)); got != wantWidth {
			t.Fatalf("selection %d width = %d, want %d", i, got, wantWidth)
		}
		if got := lipgloss.Height(gameSelectViewContent(m.width, m.height, m.nav.gameSelectList, m.nav.categoryDetail, m.notice)); got != wantHeight {
			t.Fatalf("selection %d height = %d, want %d", i, got, wantHeight)
		}
	}
}
