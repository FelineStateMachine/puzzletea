package cmd

import (
	"fmt"
	"text/tabwriter"

	"github.com/FelineStateMachine/puzzletea/store"
	"github.com/spf13/cobra"
)

var listAll bool

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List saved games",
	Long:  "Display a table of saved games. Use --all to include abandoned games.",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg := loadActiveConfig()
		s, err := openStoreFn(cfg.DBPath)
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
			_, err := fmt.Fprintln(cmd.OutOrStdout(), "No saved games.")
			return err
		}

		w := tabwriter.NewWriter(cmd.OutOrStdout(), 0, 0, 2, ' ', 0)
		if _, err := fmt.Fprintln(w, "NAME\tGAME\tMODE\tSTATUS\tLAST UPDATED"); err != nil {
			return err
		}
		for _, g := range games {
			if _, err := fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n",
				g.Name,
				g.GameType,
				g.Mode,
				g.Status,
				g.UpdatedAt.Local().Format("Jan 02 15:04"),
			); err != nil {
				return err
			}
		}
		return w.Flush()
	},
}

func init() {
	listCmd.Flags().BoolVar(&listAll, "all", false, "include abandoned games")
}
