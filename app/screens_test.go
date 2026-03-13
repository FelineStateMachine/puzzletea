package app

import (
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/FelineStateMachine/puzzletea/ui"
)

func TestMainMenuScreenRoutesByStableAction(t *testing.T) {
	screen := mainMenuScreen{
		menu: ui.NewMainMenu([]ui.MenuItem{
			{Action: mainMenuActionPlay, ItemTitle: "Launch", Desc: "custom label"},
		}),
	}

	_, _, action := screen.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	if _, ok := action.(openPlayMenuAction); !ok {
		t.Fatalf("action = %T, want %T", action, openPlayMenuAction{})
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
			name: "seeded",
			item: ui.MenuItem{Action: playMenuActionSeeded, ItemTitle: "Named Seed", Desc: "custom label"},
			assert: func(t *testing.T, action screenAction) {
				t.Helper()
				if _, ok := action.(openSeedInputAction); !ok {
					t.Fatalf("action = %T, want %T", action, openSeedInputAction{})
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
