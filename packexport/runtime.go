package packexport

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math/rand/v2"
	"os"
	"path/filepath"
	"strings"

	"github.com/FelineStateMachine/puzzletea/game"
	"github.com/FelineStateMachine/puzzletea/pdfexport"
	"github.com/FelineStateMachine/puzzletea/puzzle"
	"github.com/FelineStateMachine/puzzletea/resolve"
)

var writePDFFn = pdfexport.WritePDF

type randState struct {
	rng *rand.Rand
}

func newRandState(seed string) *randState {
	return &randState{rng: resolve.RNGFromString(seed)}
}

func (r *randState) RNG() *rand.Rand {
	if r == nil {
		return nil
	}
	return r.rng
}

func spawnSaveData(ctx context.Context, selection modeSelection, state *randState) (json.RawMessage, error) {
	var (
		puzzleGame game.Gamer
		err        error
	)

	if state != nil {
		if selection.seededSpawner == nil {
			return nil, fmt.Errorf("mode does not support deterministic spawning")
		}
		rng := state.RNG()
		if cancellable, ok := selection.seededSpawner.(game.CancellableSeededSpawner); ok {
			puzzleGame, err = cancellable.SpawnSeededContext(ctx, rng)
		} else {
			puzzleGame, err = selection.seededSpawner.SpawnSeeded(rng)
		}
	} else if cancellable, ok := selection.spawner.(game.CancellableSpawner); ok {
		puzzleGame, err = cancellable.SpawnContext(ctx)
	} else {
		puzzleGame, err = selection.spawner.Spawn()
	}
	if err != nil {
		return nil, err
	}

	save, err := puzzleGame.GetSave()
	if err != nil {
		return nil, fmt.Errorf("serialize puzzle: %w", err)
	}
	if !json.Valid(save) {
		return nil, fmt.Errorf("save payload is not valid json")
	}

	return append(json.RawMessage(nil), save...), nil
}

func WriteJSONL(path string, records []pdfexport.JSONLRecord) error {
	content, err := EncodeJSONL(records)
	if err != nil {
		return err
	}

	dir := filepath.Dir(path)
	if dir != "" && dir != "." {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return fmt.Errorf("create output directory: %w", err)
		}
	}

	if err := os.WriteFile(path, content, 0o644); err != nil {
		return fmt.Errorf("write output jsonl: %w", err)
	}
	return nil
}

type stagedOutput struct {
	targetPath string
	tempPath   string
}

func stageJSONL(path string, records []pdfexport.JSONLRecord) (stagedOutput, error) {
	content, err := EncodeJSONL(records)
	if err != nil {
		return stagedOutput{}, err
	}

	staged, err := newStagedOutput(path)
	if err != nil {
		return stagedOutput{}, err
	}
	if err := os.WriteFile(staged.tempPath, content, 0o644); err != nil {
		staged.Cleanup()
		return stagedOutput{}, fmt.Errorf("write output jsonl: %w", err)
	}
	return staged, nil
}

func stagePDF(path string, docs []pdfexport.PackDocument, puzzles []pdfexport.Puzzle, cfg pdfexport.RenderConfig) (stagedOutput, error) {
	staged, err := newStagedOutput(path)
	if err != nil {
		return stagedOutput{}, err
	}
	if err := writePDFFn(staged.tempPath, docs, puzzles, cfg); err != nil {
		staged.Cleanup()
		return stagedOutput{}, err
	}
	return staged, nil
}

func newStagedOutput(targetPath string) (stagedOutput, error) {
	dir := filepath.Dir(targetPath)
	if dir != "" && dir != "." {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return stagedOutput{}, fmt.Errorf("create output directory: %w", err)
		}
	}

	tempFile, err := os.CreateTemp(dir, ".puzzletea-export-*"+filepath.Ext(targetPath))
	if err != nil {
		return stagedOutput{}, fmt.Errorf("create temporary output file: %w", err)
	}
	tempPath := tempFile.Name()
	if err := tempFile.Close(); err != nil {
		return stagedOutput{}, fmt.Errorf("close temporary output file: %w", err)
	}
	if err := os.Remove(tempPath); err != nil && !errors.Is(err, os.ErrNotExist) {
		return stagedOutput{}, fmt.Errorf("prepare temporary output file: %w", err)
	}

	return stagedOutput{
		targetPath: targetPath,
		tempPath:   tempPath,
	}, nil
}

func (s stagedOutput) Cleanup() {
	if s.tempPath == "" {
		return
	}
	_ = os.Remove(s.tempPath)
}

func (s stagedOutput) Publish() error {
	if err := os.Rename(s.tempPath, s.targetPath); err == nil {
		return nil
	}

	if err := os.Remove(s.targetPath); err != nil && !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("replace output file: %w", err)
	}
	if err := os.Rename(s.tempPath, s.targetPath); err != nil {
		return fmt.Errorf("publish output file: %w", err)
	}
	return nil
}

func EncodeJSONL(records []pdfexport.JSONLRecord) ([]byte, error) {
	var b strings.Builder
	for _, record := range records {
		line, err := json.Marshal(record)
		if err != nil {
			return nil, fmt.Errorf("encode jsonl record: %w", err)
		}
		b.Write(line)
		b.WriteByte('\n')
	}
	return []byte(b.String()), nil
}

func flattenPuzzles(docs []pdfexport.PackDocument) []pdfexport.Puzzle {
	puzzles := []pdfexport.Puzzle{}
	for _, doc := range docs {
		puzzles = append(puzzles, doc.Puzzles...)
	}
	return puzzles
}

func normalizeDifficultyToken(s string) string {
	return puzzle.NormalizeName(s)
}
