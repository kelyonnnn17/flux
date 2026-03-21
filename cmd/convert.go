package cmd

import (
    "fmt"
    "os"
    "github.com/spf13/cobra"
    "github.com/kelyonnnn17/flux/internal/engine"
    "github.com/kelyonnnn17/flux/internal/spinner"
)

var (
    inputPath  string
    outputPath string
    engineFlag string
)

var convertCmd = &cobra.Command{
    Use:   "convert",
    Short: "Convert files between formats",
    RunE: func(cmd *cobra.Command, args []string) error {
        // Determine engine selection
        factory := engine.NewFactory(engine.NewDefaultRunner())
        var eng engine.Engine
        var err error
        if engineFlag == "auto" {
            preferred := engine.RouteByFormat(inputPath, outputPath)
            eng, err = factory.AutoEngine(preferred)
        } else {
            eng, err = factory.GetEngine(engineFlag)
        }
        if err != nil {
            return fmt.Errorf("engine selection failed: %w", err)
        }
        sp := spinner.New("converting ...")
        sp.Start()
        defer sp.Stop()
        if err := eng.Convert(inputPath, outputPath, []string{}); err != nil {
            return fmt.Errorf("conversion failed: %w", err)
        }
        fmt.Fprintf(os.Stdout, "Conversion successful: %s -> %s\n", inputPath, outputPath)
        return nil
    },
}

func init() {
    rootCmd.AddCommand(convertCmd)
    convertCmd.Flags().StringVarP(&inputPath, "input", "i", "", "Input file path (required)")
    convertCmd.Flags().StringVarP(&outputPath, "output", "o", "", "Output file path (required)")
    convertCmd.Flags().StringVar(&engineFlag, "engine", "auto", "Conversion engine (ffmpeg|imagemagick|pandoc|auto)")
    convertCmd.MarkFlagRequired("input")
    convertCmd.MarkFlagRequired("output")
}
