package app

import tea "charm.land/bubbletea/v2"

type screenAction interface{ isScreenAction() }

type openPlayMenuAction struct{}

func (openPlayMenuAction) isScreenAction() {}

type openStatsAction struct{}

func (openStatsAction) isScreenAction() {}

type openOptionsMenuAction struct{}

func (openOptionsMenuAction) isScreenAction() {}

type quitAction struct{}

func (quitAction) isScreenAction() {}

type openGameSelectAction struct{}

func (openGameSelectAction) isScreenAction() {}

type openContinueAction struct{}

func (openContinueAction) isScreenAction() {}

type openDailyAction struct{}

func (openDailyAction) isScreenAction() {}

type openWeeklyAction struct{}

func (openWeeklyAction) isScreenAction() {}

type openSeedInputAction struct{}

func (openSeedInputAction) isScreenAction() {}

type openExportAction struct{}

func (openExportAction) isScreenAction() {}

type backAction struct {
	target viewState
}

func (backAction) isScreenAction() {}

type gameSelectEnterAction struct{}

func (gameSelectEnterAction) isScreenAction() {}

type modeSelectEnterAction struct{}

func (modeSelectEnterAction) isScreenAction() {}

type continueEnterAction struct{}

func (continueEnterAction) isScreenAction() {}

type weeklyShiftAction struct {
	delta int
}

func (weeklyShiftAction) isScreenAction() {}

type weeklyEnterAction struct{}

func (weeklyEnterAction) isScreenAction() {}

type helpSelectEnterAction struct{}

func (helpSelectEnterAction) isScreenAction() {}

type openThemeSelectAction struct{}

func (openThemeSelectAction) isScreenAction() {}

type openHelpSelectAction struct{}

func (openHelpSelectAction) isScreenAction() {}

type previewThemeAction struct {
	name string
}

func (previewThemeAction) isScreenAction() {}

type confirmThemeAction struct{}

func (confirmThemeAction) isScreenAction() {}

type seedConfirmAction struct{}

func (seedConfirmAction) isScreenAction() {}

type exportSubmitAction struct{}

func (exportSubmitAction) isScreenAction() {}

type screenModel interface {
	State() viewState
	Resize(width, height int) screenModel
	Update(msg tea.Msg) (screenModel, tea.Cmd, screenAction)
	View(notice noticeState) string
	Apply(m model) model
}

func (m model) activeScreen() screenModel {
	switch m.state {
	case mainMenuView:
		return mainMenuScreen{
			width:  m.width,
			height: m.height,
			menu:   m.nav.mainMenu,
		}
	case playMenuView:
		return playMenuScreen{
			width:  m.width,
			height: m.height,
			menu:   m.nav.playMenu,
		}
	case optionsMenuView:
		return optionsMenuScreen{
			width:  m.width,
			height: m.height,
			menu:   m.nav.optionsMenu,
		}
	case seedInputView:
		return seedInputScreen{
			width:  m.width,
			height: m.height,
			seed:   m.seed,
		}
	case gameSelectView:
		return gameSelectScreen{
			width:  m.width,
			height: m.height,
			list:   m.nav.gameSelectList,
			detail: m.nav.categoryDetail,
		}
	case modeSelectView:
		return modeSelectScreen{
			width:  m.width,
			height: m.height,
			entry:  m.nav.selectedCategory,
			list:   m.nav.modeSelectList,
		}
	case exportView:
		return exportScreen{
			width:  m.width,
			height: m.height,
			export: m.export,
		}
	case continueView:
		return continueScreen{
			width:  m.width,
			height: m.height,
			cont:   m.cont,
		}
	case weeklyView:
		return weeklyScreen{
			width:  m.width,
			height: m.height,
			weekly: m.weekly,
		}
	case helpSelectView:
		return helpSelectScreen{
			width:  m.width,
			height: m.height,
			help:   m.help,
		}
	case helpDetailView:
		return helpDetailScreen{
			width:  m.width,
			height: m.height,
			help:   m.help,
		}
	case statsView:
		return statsScreen{
			width:  m.width,
			height: m.height,
			stats:  m.stats,
		}
	case themeSelectView:
		return themeSelectScreen{
			width:  m.width,
			height: m.height,
			theme:  m.theme,
		}
	case generatingView:
		return generatingScreen{
			width:   m.width,
			height:  m.height,
			spinner: m.spinner,
		}
	case exportRunningView:
		return exportRunningScreen{
			width:   m.width,
			height:  m.height,
			spinner: m.spinner,
		}
	default:
		return nil
	}
}
