package app

import tea "charm.land/bubbletea/v2"

// routeManager owns the route/screen management boundary: current view state,
// cached screen instances, window dimensions, and screen lifecycle operations
// (lookup, initialization, resize, and window-size handling).
//
// It follows the same *model pattern as sessionController: methods read and
// mutate model fields via the pointer. The model keeps thin delegation
// wrappers so call sites remain unchanged.
type routeManager struct {
	model *model
}

func newRouteManager(m *model) routeManager {
	return routeManager{model: m}
}

// activeScreen returns the cached screen for the current route, or nil if
// no screen has been initialized.
func (r routeManager) activeScreen() screenModel {
	return r.model.screens[r.model.state]
}

// initScreen creates a fresh screen for the given state using the registry
// factory, resizes it to current dimensions, and stores it in the cache.
func (r routeManager) initScreen(state viewState) {
	factory, ok := screenRegistry[state]
	if !ok {
		return
	}
	if r.model.screens == nil {
		r.model.screens = make(map[viewState]screenModel)
	}
	r.model.screens[state] = factory(*r.model).Resize(r.model.width, r.model.height)
}

// resizeActiveScreen resizes the screen at the current route to the current
// window dimensions. Safe to call when no screen is cached.
func (r routeManager) resizeActiveScreen() {
	screen := r.model.screens[r.model.state]
	if screen == nil {
		return
	}
	r.model.screens[r.model.state] = screen.Resize(r.model.width, r.model.height)
}

// handleWindowSize updates window dimensions and resizes the active screen.
// In gameView the game handles its own sizing, so the screen cache is left
// untouched.
func (r routeManager) handleWindowSize(msg tea.WindowSizeMsg) {
	r.model.width = msg.Width
	r.model.height = msg.Height
	if r.model.state == gameView {
		return
	}
	r.resizeActiveScreen()
}
