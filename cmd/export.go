package cmd

import (
	"context"
	"fmt"
	"strings"

	"github.com/FelineStateMachine/puzzletea/packexport"
	"github.com/spf13/cobra"
)

var exportSpecPath string

var exportCmd = &cobra.Command{
	Use:   "export --spec <path>",
	Short: "Generate a printable puzzle pack from a JSON spec",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, _ []string) error {
		if strings.TrimSpace(exportSpecPath) == "" {
			return fmt.Errorf("--spec is required")
		}

		spec, err := packexport.LoadSpecFile(exportSpecPath)
		if err != nil {
			return err
		}

		result, err := packexport.Run(context.Background(), spec)
		if err != nil {
			return err
		}

		if _, err := fmt.Fprintf(cmd.OutOrStdout(), "wrote %s", result.PDFOutputPath); err != nil {
			return err
		}
		if result.JSONLOutputPath != "" {
			if _, err := fmt.Fprintf(cmd.OutOrStdout(), " and %s", result.JSONLOutputPath); err != nil {
				return err
			}
		}
		_, err = fmt.Fprintln(cmd.OutOrStdout())
		return err
	},
}

func init() {
	exportCmd.Flags().StringVar(&exportSpecPath, "spec", "", "path to a JSON export spec")
}
