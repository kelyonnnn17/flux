package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/kelyonnnn17/flux/internal/config"
	"github.com/kelyonnnn17/flux/internal/format"
	"github.com/spf13/cobra"
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
	os.Args = rewriteArgsForShortcut(os.Args)
	if _, err := config.Load(); err != nil {
		fmt.Fprintf(os.Stderr, "warning: config load failed: %v\n", err)
	}
	if err := rootCmd.Execute(); err != nil {
		format.Init(noColorFlag)
		format.Error("%s", err)
		os.Exit(1)
	}
}

func rewriteArgsForShortcut(args []string) []string {
	if len(args) < 3 {
		return args
	}

	knownCommands := map[string]bool{
		"convert":      true,
		"c":            true,
		"doctor":       true,
		"d":            true,
		"list-formats": true,
		"lf":           true,
		"info":         true,
		"i":            true,
		"help":         true,
		"completion":   true,
	}

	firstNonFlag := -1
	for i := 1; i < len(args); i++ {
		tok := args[i]
		if tok == "--" {
			return args
		}
		if tok == "--engine" {
			i++
			continue
		}
		if strings.HasPrefix(tok, "-") {
			continue
		}
		firstNonFlag = i
		break
	}

	if firstNonFlag == -1 {
		return args
	}
	if knownCommands[args[firstNonFlag]] {
		return args
	}
	if len(args[firstNonFlag:]) < 2 {
		return args
	}

	out := make([]string, 0, len(args)+1)
	out = append(out, args[:firstNonFlag]...)
	out = append(out, "convert")
	out = append(out, args[firstNonFlag:]...)
	return out
}
