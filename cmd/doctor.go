package cmd

import (
	"github.com/kelyonnnn17/flux/internal/engine"
	"github.com/kelyonnnn17/flux/internal/format"
	"github.com/spf13/cobra"
)

var doctorCmd = &cobra.Command{
	Use:     "doctor",
	Aliases: []string{"d"},
	Short:   "Check installed engines and versions",
	RunE: func(cmd *cobra.Command, args []string) error {
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
