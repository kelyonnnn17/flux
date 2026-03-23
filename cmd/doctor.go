package cmd

import (
	"fmt"

	"github.com/kelyonnnn17/flux/internal/engine"
	"github.com/kelyonnnn17/flux/internal/format"
	"github.com/kelyonnnn17/flux/internal/ui"
	"github.com/spf13/cobra"
)

var doctorCmd = &cobra.Command{
	Use:     "doctor",
	Aliases: []string{"d"},
	Short:   "Check installed engines and versions",
	RunE: func(cmd *cobra.Command, args []string) error {
		if ui.ShouldRender() {
			return ui.Run(ui.Spec{
				Command: "flux doctor",
				Running: "checking engines",
				Run: func() (ui.Result, error) {
					infos := engine.Doctor()
					missing := 0
					lines := make([]ui.Line, 0, len(infos))
					for _, info := range infos {
						if info.Path == "not found" {
							missing++
							lines = append(lines, ui.Line{Kind: "warn", Text: fmt.Sprintf("%s: not found", info.Name)})
							continue
						}
						lines = append(lines, ui.Line{Kind: "ok", Text: fmt.Sprintf("%s: %s (%s)", info.Name, info.Path, info.Version)})
					}

					return ui.Result{
						Meta: []ui.Meta{
							{Key: "command", Value: "doctor"},
							{Key: "engines", Value: fmt.Sprintf("%d", len(infos))},
							{Key: "missing", Value: fmt.Sprintf("%d", missing)},
						},
						Lines:   lines,
						Success: "doctor complete",
					}, nil
				},
			})
		}

		format.Primary("Flux engines:")
		for _, info := range engine.Doctor() {
			if info.Path == "not found" {
				format.Info("  %s: not found", info.Name)
			} else {
				format.Info("  %s: %s (%s)", info.Name, info.Path, info.Version)
			}
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(doctorCmd)
}
