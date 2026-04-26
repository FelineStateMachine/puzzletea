package app

import (
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/FelineStateMachine/puzzletea/ui"
)

func TestMainMenuScreenRoutesByStableAction(t *testing.T) {
	tests := []struct {
		name   string
		item   ui.MenuItem
		assert func(t *testing.T, action screenAction)
	}{
		{
			name: "play",
			item: ui.MenuItem{Action: mainMenuActionPlay, ItemTitle: "Launch", Desc: "custom label"},
			assert: func(t *testing.T, action screenAction) {
				t.Helper()
				if _, ok := action.(openPlayMenuAction); !ok {
					t.Fatalf("action = %T, want %T", action, openPlayMenuAction{})
				}
			},
		},
		{
			name: "export",
			item: ui.MenuItem{Action: mainMenuActionExport, ItemTitle: "Print Pack", Desc: "custom label"},
			assert: func(t *testing.T, action screenAction) {
				t.Helper()
				if _, ok := action.(openExportAction); !ok {
					t.Fatalf("action = %T, want %T", action, openExportAction{})
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			screen := mainMenuScreen{
				menu: ui.NewMainMenu([]ui.MenuItem{tt.item}),
			}

			_, _, action := screen.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
			tt.assert(t, action)
		})
	}
}

func TestMainMenuItemsPlacesExportBetweenPlayAndStats(t *testing.T) {
	if len(mainMenuItems) < 3 {
		t.Fatalf("mainMenuItems length = %d, want at least 3", len(mainMenuItems))
	}
	if got := mainMenuItems[0].Action; got != mainMenuActionPlay {
		t.Fatalf("mainMenuItems[0] = %q, want %q", got, mainMenuActionPlay)
	}
	if got := mainMenuItems[1].Action; got != mainMenuActionExport {
		t.Fatalf("mainMenuItems[1] = %q, want %q", got, mainMenuActionExport)
	}
	if got := mainMenuItems[2].Action; got != mainMenuActionStats {
		t.Fatalf("mainMenuItems[2] = %q, want %q", got, mainMenuActionStats)
	}
}

func TestPlayMenuScreenRoutesByStableAction(t *testing.T) {
	tests := []struct {
		name   string
		item   ui.MenuItem
		assert func(t *testing.T, action screenAction)
	}{
		{
			name: "daily",
			item: ui.MenuItem{Action: playMenuActionDaily, ItemTitle: "Today", Desc: "custom label"},
			assert: func(t *testing.T, action screenAction) {
				t.Helper()
				if _, ok := action.(openDailyAction); !ok {
					t.Fatalf("action = %T, want %T", action, openDailyAction{})
				}
			},
		},
		{
			name: "create",
			item: ui.MenuItem{Action: playMenuActionCreate, ItemTitle: "New", Desc: "custom label"},
			assert: func(t *testing.T, action screenAction) {
				t.Helper()
				if _, ok := action.(openCreateAction); !ok {
					t.Fatalf("action = %T, want %T", action, openCreateAction{})
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			screen := playMenuScreen{
				menu: ui.NewMainMenu([]ui.MenuItem{tt.item}),
			}

			_, _, action := screen.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
			tt.assert(t, action)
		})
	}
}

func TestOptionsMenuScreenRoutesByStableAction(t *testing.T) {
	screen := optionsMenuScreen{
		menu: ui.NewMainMenu([]ui.MenuItem{
			{Action: optionsMenuActionGuides, ItemTitle: "Help Docs", Desc: "custom label"},
		}),
	}

	_, _, action := screen.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	if _, ok := action.(openHelpSelectAction); !ok {
		t.Fatalf("action = %T, want %T", action, openHelpSelectAction{})
	}
}
