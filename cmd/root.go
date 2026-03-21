package cmd

import (
    "fmt"
    "os"
    "github.com/spf13/cobra"
    "github.com/kelyonnnn17/flux/internal/config"
    "github.com/kelyonnnn17/flux/internal/format"
)

var noColorFlag bool

var rootCmd = &cobra.Command{
    Use:   "flux",
    Short: "Flux Universal File Converter",
    Long:  `Flux is a CLI tool for converting files across formats using FFmpeg, ImageMagick, and Pandoc.`,
    Run: func(cmd *cobra.Command, args []string) {
        format.Primary("Flux CLI. Use --help for commands.")
    },
    PersistentPreRun: func(cmd *cobra.Command, args []string) {
        format.Init(noColorFlag)
    },
}

func init() {
    rootCmd.PersistentFlags().String("engine", "auto", "Conversion engine to use: ffmpeg|imagemagick|pandoc|data|auto")
    rootCmd.PersistentFlags().BoolVar(&noColorFlag, "no-color", false, "Disable ANSI color output")
}

func Execute() {
    if _, err := config.Load(); err != nil {
        fmt.Fprintf(os.Stderr, "warning: config load failed: %v\n", err)
    }
    if err := rootCmd.Execute(); err != nil {
        format.Init(noColorFlag)
        format.Error("%s", err)
        os.Exit(1)
    }
}
