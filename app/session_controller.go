package app

import (
	"context"
	"math/rand/v2"

	"github.com/FelineStateMachine/puzzletea/game"
	sessionflow "github.com/FelineStateMachine/puzzletea/session"
	"github.com/FelineStateMachine/puzzletea/store"

	tea "charm.land/bubbletea/v2"
)

type sessionController struct {
	model *model
}

func newSessionController(m *model) sessionController {
	return sessionController{model: m}
}

func (c sessionController) beginSpawnContext() (context.Context, int64) {
	c.cancelActiveSpawn()
	c.model.session.spawnJobID++
	jobID := c.model.session.spawnJobID
	ctx, cancel := context.WithTimeout(context.Background(), spawnTimeout)
	c.model.session.spawnCancel = cancel
	c.model.session.generating = true
	return ctx, jobID
}

func (c sessionController) cancelActiveSpawn() {
	if c.model.session.spawnCancel != nil {
		c.model.session.spawnCancel()
		c.model.session.spawnCancel = nil
	}
	if c.model.session.generating {
		c.model.session.spawnJobID++
	}
	c.model.session.generating = false
	c.model.session.spawn = nil
}

func (c sessionController) startSpawn(spawner game.Spawner, request spawnRequest) tea.Cmd {
	ctx, jobID := c.beginSpawnContext()
	c.model.session.spawn = &request
	c.model.state = generatingView
	*c.model = c.model.clearNotice()
	*c.model = c.model.initScreen(generatingView)
	return tea.Batch(c.model.spinner.Tick, spawnCmd(spawner, ctx, jobID))
}

func (c sessionController) startSeededSpawn(spawner game.SeededSpawner, rng *rand.Rand, request spawnRequest) tea.Cmd {
	ctx, jobID := c.beginSpawnContext()
	c.model.session.spawn = &request
	c.model.state = generatingView
	*c.model = c.model.clearNotice()
	*c.model = c.model.initScreen(generatingView)
	return tea.Batch(c.model.spinner.Tick, spawnSeededCmd(spawner, rng, ctx, jobID))
}

func (c sessionController) handleSpawnComplete(jobID int64, msg game.SpawnCompleteMsg) tea.Cmd {
	if jobID != c.model.session.spawnJobID {
		return nil
	}

	c.model.session.generating = false
	if c.model.session.spawnCancel != nil {
		c.model.session.spawnCancel()
		c.model.session.spawnCancel = nil
	}
	request := c.model.session.spawn
	c.model.session.spawn = nil

	if c.model.state != generatingView {
		return nil
	}

	if msg.Err != nil {
		*c.model = c.model.setErrorf("Could not generate puzzle: %v", msg.Err)
		if request != nil {
			c.model.state = request.returnState
		}
		return nil
	}

	if request == nil {
		*c.model = c.model.setErrorf("Internal error: missing spawn request metadata")
		return nil
	}

	c.activateGame(msg.Game.SetTitle(request.name), 0, false, gameOpenOptions{
		returnState: request.exitState,
		weeklyInfo:  request.weeklyInfo,
	})
	*c.model = c.model.clearNotice()

	rec, err := sessionflow.CreateRecord(
		c.model.store,
		c.model.session.game,
		request.name,
		request.gameType,
		request.modeTitle,
		request.run,
	)
	if err != nil {
		*c.model = c.model.setErrorf("Started puzzle, but could not create a save record: %v", err)
	} else {
		c.model.session.activeGameID = rec.ID
	}
	return nil
}

func (c sessionController) activateGame(g game.Gamer, activeGameID int64, completionSaved bool, options gameOpenOptions) {
	g, _ = g.Update(game.HelpToggleMsg{Show: c.model.help.showFull})
	g, _ = g.Update(tea.WindowSizeMsg{Width: c.model.width, Height: c.model.height})

	c.model.session.game = g
	c.model.session.activeGameID = activeGameID
	c.model.session.completionSaved = completionSaved
	c.model.session.returnState = options.returnState
	c.model.session.weeklyAdvance = options.weeklyInfo
	c.model.state = gameView
}

func (c sessionController) importRecord(rec store.GameRecord, options gameOpenOptions) bool {
	g, err := sessionflow.ImportRecord(&rec)
	if err != nil {
		*c.model = c.model.setErrorf("Could not load saved puzzle %q: %v", rec.Name, err)
		return false
	}

	activeGameID := rec.ID
	completionSaved := rec.Status == store.StatusCompleted
	if options.readOnly {
		activeGameID = 0
		completionSaved = true
	}

	c.activateGame(g, activeGameID, completionSaved, options)
	*c.model = c.model.clearNotice()
	return true
}

func (c sessionController) updateActiveGame(msg tea.Msg) tea.Cmd {
	var cmd tea.Cmd
	c.model.session.game, cmd = c.model.session.game.Update(msg)
	if c.model.debug.enabled {
		c.model.debug.info = c.model.renderDebugInfo()
	}
	c.persistCompletionIfSolved()
	return cmd
}

func (c sessionController) persistCompletionIfSolved() {
	if c.model.session.game == nil || c.model.session.activeGameID == 0 ||
		c.model.session.completionSaved || !c.model.session.game.IsSolved() {
		return
	}

	c.model.session.completionSaved = true
	saveData, err := c.model.session.game.GetSave()
	if err == nil {
		if err := c.model.store.UpdateSaveState(c.model.session.activeGameID, string(saveData)); err != nil {
			*c.model = c.model.setErrorf("Puzzle completed, but saving the final state failed: %v", err)
		}
	} else {
		*c.model = c.model.setErrorf("Puzzle completed, but reading the final state failed: %v", err)
	}
	if err := c.model.store.UpdateStatus(c.model.session.activeGameID, store.StatusCompleted); err != nil {
		*c.model = c.model.setErrorf("Puzzle completed, but updating completion status failed: %v", err)
	}
}

func (c sessionController) saveCurrentGame(status store.GameStatus) {
	if c.model.session.game == nil {
		return
	}
	if c.model.session.activeGameID == 0 {
		c.clearActiveGame()
		return
	}

	saveData, err := c.model.session.game.GetSave()
	if err != nil {
		*c.model = c.model.setErrorf("Could not save puzzle progress: %v", err)
		return
	}
	if err := c.model.store.UpdateSaveState(c.model.session.activeGameID, string(saveData)); err != nil {
		*c.model = c.model.setErrorf("Could not save puzzle progress: %v", err)
	}
	if !(c.model.session.completionSaved && status != store.StatusCompleted) {
		if err := c.model.store.UpdateStatus(c.model.session.activeGameID, status); err != nil {
			*c.model = c.model.setErrorf("Could not update puzzle status: %v", err)
		}
	}
	c.clearActiveGame()
}

func (c sessionController) clearActiveGame() {
	c.model.session.activeGameID = 0
	c.model.session.game = nil
	c.model.session.completionSaved = false
	c.model.session.returnState = mainMenuView
	c.model.session.weeklyAdvance = nil
}
