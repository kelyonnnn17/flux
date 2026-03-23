package cmd

import (
	"fmt"
	"os"

	"github.com/kelyonnnn17/flux/internal/format"
	"github.com/kelyonnnn17/flux/internal/ui"
	"github.com/spf13/cobra"
)

var infoCmd = &cobra.Command{
	Use:     "info [file]",
	Aliases: []string{"i"},
	Short:   "Inspect file metadata",
	Args:    cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		path := args[0]
		if ui.ShouldRender() {
			return ui.Run(ui.Spec{
				Command: "flux info " + path,
				Running: "reading metadata",
				Run: func() (ui.Result, error) {
					info, err := os.Stat(path)
					if err != nil {
						return ui.Result{ErrorHint: "verify path and file permissions"}, fmt.Errorf("stat %s: %w", path, err)
					}

					return ui.Result{
						Meta: []ui.Meta{
							{Key: "file", Value: path},
							{Key: "size", Value: fmt.Sprintf("%d bytes", info.Size())},
							{Key: "mode", Value: info.Mode().String()},
							{Key: "modified", Value: info.ModTime().String()},
						},
						Success: "metadata loaded",
					}, nil
				},
			})
		}

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
