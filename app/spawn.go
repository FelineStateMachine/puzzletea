package app

import (
	"context"
	"math/rand/v2"
	"time"

	"github.com/FelineStateMachine/puzzletea/difficulty"
	"github.com/FelineStateMachine/puzzletea/game"

	tea "charm.land/bubbletea/v2"
)

const spawnTimeout = 10 * time.Second

type spawnCompleteMsg struct {
	jobID  int64
	result game.SpawnCompleteMsg
}

func spawnEloCmd(spawner game.EloSpawner, seed string, elo difficulty.Elo, ctx context.Context, jobID int64) tea.Cmd {
	return func() tea.Msg {
		var (
			g      game.Gamer
			report difficulty.Report
			err    error
		)
		if cancellable, ok := spawner.(game.CancellableEloSpawner); ok {
			g, report, err = cancellable.SpawnEloContext(ctx, seed, elo)
		} else {
			g, report, err = spawner.SpawnElo(seed, elo)
		}
		return spawnCompleteMsg{
			jobID:  jobID,
			result: game.SpawnCompleteMsg{Game: g, Report: report, Err: err},
		}
	}
}

// spawnCmd returns a tea.Cmd that runs Spawn() off the main goroutine.
func spawnCmd(spawner game.Spawner, ctx context.Context, jobID int64) tea.Cmd {
	return func() tea.Msg {
		var (
			g   game.Gamer
			err error
		)
		if cancellable, ok := spawner.(game.CancellableSpawner); ok {
			g, err = cancellable.SpawnContext(ctx)
		} else {
			g, err = spawner.Spawn()
		}
		return spawnCompleteMsg{
			jobID:  jobID,
			result: game.SpawnCompleteMsg{Game: g, Err: err},
		}
	}
}

// spawnSeededCmd returns a tea.Cmd that runs SpawnSeeded() off the main goroutine.
// The caller must not use rng after this call — *rand.Rand is not goroutine-safe
// and ownership is transferred to the spawned goroutine.
func spawnSeededCmd(spawner game.SeededSpawner, rng *rand.Rand, ctx context.Context, jobID int64) tea.Cmd {
	return func() tea.Msg {
		var (
			g   game.Gamer
			err error
		)
		if cancellable, ok := spawner.(game.CancellableSeededSpawner); ok {
			g, err = cancellable.SpawnSeededContext(ctx, rng)
		} else {
			g, err = spawner.SpawnSeeded(rng)
		}
		return spawnCompleteMsg{
			jobID:  jobID,
			result: game.SpawnCompleteMsg{Game: g, Err: err},
		}
	}
}
