package cmd

import (
	"bytes"
	"fmt"
	"path/filepath"
	"testing"

	"github.com/FelineStateMachine/puzzletea/config"
	"github.com/FelineStateMachine/puzzletea/store"
	"github.com/FelineStateMachine/puzzletea/theme"
	"github.com/spf13/cobra"
)

func TestNewContinueAndRootUseActiveConfig(t *testing.T) {
	configPath, dbPath := writeCommandConfig(t)
	restore := snapshotCommandGlobals(t)
	defer restore()

	flagConfigPath = configPath

	var seen []string
	launchNewGameFn = func(gameArg, modeArg, seed string, cfg *config.Config) error {
		seen = append(seen, cfg.DBPath)
		if gameArg == "" {
			t.Fatal("expected game argument to be forwarded")
		}
		return nil
	}
	continueGameFn = func(name string, cfg *config.Config) error {
		seen = append(seen, cfg.DBPath)
		if name == "" {
			t.Fatal("expected continue name to be forwarded")
		}
		return nil
	}

	if err := newCmd.RunE(newCmd, []string{"nonogram", "mini"}); err != nil {
		t.Fatalf("newCmd.RunE returned error: %v", err)
	}
	if err := continueCmd.RunE(continueCmd, []string{"resume-me"}); err != nil {
		t.Fatalf("continueCmd.RunE returned error: %v", err)
	}

	flagNew = "sudoku:easy"
	if err := RootCmd.RunE(RootCmd, nil); err != nil {
		t.Fatalf("RootCmd.RunE returned error: %v", err)
	}
	flagNew = ""

	flagContinue = "resume-me"
	if err := RootCmd.RunE(RootCmd, nil); err != nil {
		t.Fatalf("RootCmd.RunE returned error: %v", err)
	}

	if got, want := len(seen), 4; got != want {
		t.Fatalf("saw %d config loads, want %d", got, want)
	}
	for i, got := range seen {
		if got != dbPath {
			t.Fatalf("config load %d used DB path %q, want %q", i, got, dbPath)
		}
	}
}

func TestListCmdUsesActiveConfig(t *testing.T) {
	configPath, dbPath := writeCommandConfig(t)
	restore := snapshotCommandGlobals(t)
	defer restore()

	flagConfigPath = configPath

	s, err := store.Open(dbPath)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = s.Close() })
	if err := s.CreateGame(&store.GameRecord{
		Name:         "from-custom-config",
		GameType:     "Lights Out",
		Mode:         "Easy",
		InitialState: "{}",
		SaveState:    "{}",
		Status:       store.StatusNew,
		RunKind:      store.RunKindNormal,
	}); err != nil {
		t.Fatal(err)
	}

	var out bytes.Buffer
	cmd := &cobra.Command{}
	cmd.SetOut(&out)
	if err := listCmd.RunE(cmd, nil); err != nil {
		t.Fatalf("listCmd.RunE returned error: %v", err)
	}
	if !bytes.Contains(out.Bytes(), []byte("from-custom-config")) {
		t.Fatalf("list output %q did not include custom DB record", out.String())
	}
}

func TestLoadConfigThemeFlagOverridesPersistedTheme(t *testing.T) {
	configPath, _ := writeCommandConfig(t)
	restore := snapshotCommandGlobals(t)
	defer restore()
	t.Cleanup(func() { _ = theme.Apply("") })

	flagTheme = "Dracula"
	if themeDoc := theme.LookupTheme(flagTheme); themeDoc == nil {
		t.Fatalf("expected theme %q to exist", flagTheme)
	}

	_ = loadConfig(configPath)

	want := theme.LookupTheme(flagTheme).Palette()
	got := theme.Current()
	if fmt.Sprint(got.FG) != fmt.Sprint(want.FG) {
		t.Fatalf("theme override FG = %v, want %v", got.FG, want.FG)
	}
}

func snapshotCommandGlobals(t *testing.T) func() {
	t.Helper()

	prevNew := flagNew
	prevContinue := flagContinue
	prevTheme := flagTheme
	prevConfigPath := flagConfigPath
	prevSetSeed := flagSetSeed
	prevWithSeed := flagWithSeed
	prevExport := flagExport
	prevOutput := flagOutput
	prevListAll := listAll
	prevLaunchNew := launchNewGameFn
	prevLaunchSeeded := launchSeededGameFn
	prevContinueFn := continueGameFn
	prevOpenStore := openStoreFn

	return func() {
		flagNew = prevNew
		flagContinue = prevContinue
		flagTheme = prevTheme
		flagConfigPath = prevConfigPath
		flagSetSeed = prevSetSeed
		flagWithSeed = prevWithSeed
		flagExport = prevExport
		flagOutput = prevOutput
		listAll = prevListAll
		launchNewGameFn = prevLaunchNew
		launchSeededGameFn = prevLaunchSeeded
		continueGameFn = prevContinueFn
		openStoreFn = prevOpenStore
	}
}

func writeCommandConfig(t *testing.T) (string, string) {
	t.Helper()

	configPath := filepath.Join(t.TempDir(), "config.json")
	dbPath := filepath.Join(t.TempDir(), "history.db")
	cfg := &config.Config{
		DBPath: dbPath,
		Theme:  "Nord",
	}
	if err := cfg.Save(configPath); err != nil {
		t.Fatal(err)
	}
	return configPath, dbPath
}
