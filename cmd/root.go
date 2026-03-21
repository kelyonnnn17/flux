package cmd

import (
    "fmt"
    "os"
    "github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
    Use:   "flux",
    Short: "Flux Universal File Converter",
    Long:  `Flux is a CLI tool for converting files across formats using FFmpeg, ImageMagick, and Pandoc.`,
    Run: func(cmd *cobra.Command, args []string) {
        fmt.Println("Flux CLI. Use --help for commands.")
    },
}

func init() {
    // Persistent flag for engine selection (default "auto")
    rootCmd.PersistentFlags().String("engine", "auto", "Conversion engine to use: ffmpeg|imagemagick|pandoc|auto")
}

func Execute() {
    if err := rootCmd.Execute(); err != nil {
        fmt.Println(err)
        os.Exit(1)
    }
}
