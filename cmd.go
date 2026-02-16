package main

import (
	"fmt"
	"log"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/FelineStateMachine/puzzletea/game"
	"github.com/FelineStateMachine/puzzletea/resolve"
	"github.com/FelineStateMachine/puzzletea/store"

	tea "charm.land/bubbletea/v2"
	"github.com/spf13/cobra"
)

// version is set at build time via ldflags:
//
//	go build -ldflags "-X main.version=v1.0.0"
var version = "dev"

var (
	flagNew      string
	flagContinue string
)

var rootCmd = &cobra.Command{
	Use:     "puzzletea",
	Version: version,
	Short:   "A terminal-based puzzle game framework",
	Long:    "PuzzleTea is a terminal-based puzzle game framework featuring Nonogram, Sudoku, Word Search, Hashiwokakero, and Lights Out.",
	RunE: func(cmd *cobra.Command, args []string) error {
		if flagNew != "" {
			parts := strings.SplitN(flagNew, ":", 2)
			gameArg := parts[0]
			var modeArg string
			if len(parts) > 1 {
				modeArg = parts[1]
			}
			return launchNewGame(gameArg, modeArg)
		}
		if flagContinue != "" {
			return continueGame(flagContinue)
		}
		// Default: launch TUI menu.
		s, err := store.Open(store.DefaultDBPath())
		if err != nil {
			return err
		}
		defer s.Close()
		p := tea.NewProgram(initialModel(s))
		_, err = p.Run()
		return err
	},
}

var newCmd = &cobra.Command{
	Use:   "new <game> [mode]",
	Short: "Start a new puzzle game",
	Long: fmt.Sprintf("Start a new puzzle game, optionally specifying the difficulty mode.\n\nAvailable games:\n  %s",
		strings.Join(resolve.CategoryNames(GameCategories), "\n  ")),
	Args: cobra.RangeArgs(1, 2),
	RunE: func(cmd *cobra.Command, args []string) error {
		gameArg := args[0]
		var modeArg string
		if len(args) > 1 {
			modeArg = args[1]
		}
		return launchNewGame(gameArg, modeArg)
	},
}

var continueCmd = &cobra.Command{
	Use:   "continue <name>",
	Short: "Resume a saved game by name",
	Long:  "Resume a previously saved game using its unique name.\nUse 'puzzletea list' to see available saved games.",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return continueGame(args[0])
	},
}

var listAll bool

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List saved games",
	Long:  "Display a table of saved games. Use --all to include abandoned games.",
	RunE: func(cmd *cobra.Command, args []string) error {
		s, err := store.Open(store.DefaultDBPath())
		if err != nil {
			return err
		}
		defer s.Close()

		var games []store.GameRecord
		if listAll {
			games, err = s.ListAllGames()
		} else {
			games, err = s.ListGames()
		}
		if err != nil {
			return err
		}

		if len(games) == 0 {
			fmt.Println("No saved games.")
			return nil
		}

		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintln(w, "NAME\tGAME\tMODE\tSTATUS\tLAST UPDATED")
		for _, g := range games {
			fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n",
				g.Name,
				g.GameType,
				g.Mode,
				g.Status,
				g.UpdatedAt.Local().Format("Jan 02 15:04"),
			)
		}
		return w.Flush()
	},
}

func init() {
	rootCmd.Flags().StringVar(&flagNew, "new", "", "start a new game (game:mode)")
	rootCmd.Flags().StringVar(&flagContinue, "continue", "", "resume a saved game by name")

	listCmd.Flags().BoolVar(&listAll, "all", false, "include abandoned games")

	rootCmd.AddCommand(newCmd, continueCmd, listCmd)
}

// launchNewGame resolves the game/mode, spawns a new game, and launches the TUI.
func launchNewGame(gameArg, modeArg string) error {
	cat, err := resolve.Category(gameArg, GameCategories)
	if err != nil {
		return err
	}

	spawner, modeTitle, err := resolve.Mode(cat, modeArg)
	if err != nil {
		return err
	}

	g, err := spawner.Spawn()
	if err != nil {
		return fmt.Errorf("failed to spawn game: %w", err)
	}

	s, err := store.Open(store.DefaultDBPath())
	if err != nil {
		return err
	}
	defer s.Close()

	name := generateUniqueName(s)
	g = g.SetTitle(name)

	initialState, err := g.GetSave()
	if err != nil {
		return fmt.Errorf("failed to get initial save: %w", err)
	}

	rec := &store.GameRecord{
		Name:         name,
		GameType:     cat.Name,
		Mode:         modeTitle,
		InitialState: string(initialState),
		SaveState:    string(initialState),
		Status:       store.StatusNew,
	}
	if err := s.CreateGame(rec); err != nil {
		log.Printf("failed to create game record: %v", err)
	}

	m := initialModelWithGame(s, g, rec.ID, false)
	p := tea.NewProgram(m)
	_, err = p.Run()
	return err
}

// continueGame looks up a saved game by name and launches the TUI.
func continueGame(name string) error {
	s, err := store.Open(store.DefaultDBPath())
	if err != nil {
		return err
	}
	defer s.Close()

	rec, err := s.GetGameByName(name)
	if err != nil {
		return err
	}
	if rec == nil {
		return fmt.Errorf("no saved game found with name %q\nRun 'puzzletea list' to see available games.", name)
	}

	importFn, ok := game.Registry[rec.GameType]
	if !ok {
		return fmt.Errorf("unknown game type %q in save data", rec.GameType)
	}

	g, err := importFn([]byte(rec.SaveState))
	if err != nil {
		return fmt.Errorf("failed to import game: %w", err)
	}
	g = g.SetTitle(rec.Name)

	completed := rec.Status == store.StatusCompleted
	m := initialModelWithGame(s, g, rec.ID, completed)
	p := tea.NewProgram(m)
	_, err = p.Run()
	return err
}
