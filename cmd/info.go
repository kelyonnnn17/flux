package cmd

import (
	"fmt"
	"os"

	"github.com/kelyonnnn17/flux/internal/format"
	"github.com/spf13/cobra"
)

var infoCmd = &cobra.Command{
	Use:     "info [file]",
	Aliases: []string{"i"},
	Short:   "Inspect file metadata",
	Args:    cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		path := args[0]
		info, err := os.Stat(path)
		if err != nil {
			return fmt.Errorf("stat %s: %w", path, err)
		}
		format.Primary("File: %s", path)
		format.Info("  size: %d bytes", info.Size())
		format.Info("  mode: %s", info.Mode())
		format.Info("  modified: %s", info.ModTime())
		return nil
	},
}

func init() {
	rootCmd.AddCommand(infoCmd)
}
