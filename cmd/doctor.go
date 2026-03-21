package cmd

import (
	"github.com/spf13/cobra"
	"github.com/kelyonnnn17/flux/internal/engine"
	"github.com/kelyonnnn17/flux/internal/format"
)

var doctorCmd = &cobra.Command{
	Use:   "doctor",
	Short: "Check installed engines and versions",
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
