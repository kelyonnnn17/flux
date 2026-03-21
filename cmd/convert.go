package cmd

import (
    "fmt"
    "os"
    "path/filepath"
    "github.com/spf13/cobra"
    "github.com/kelyonnnn17/flux/internal/data"
    "github.com/kelyonnnn17/flux/internal/engine"
    "github.com/kelyonnnn17/flux/internal/spinner"
)

var (
    inputPath  string
    outputPath string
    engineFlag string
    fromFlag   string
    toFlag     string
)

var convertCmd = &cobra.Command{
    Use:   "convert",
    Short: "Convert files between formats",
    RunE: func(cmd *cobra.Command, args []string) error {
        fromFormat := fromFlag
        if fromFormat == "" {
            fromFormat = data.FormatFromExt(inputPath)
        }
        toFormat := toFlag
        if toFormat == "" {
            toFormat = data.FormatFromExt(outputPath)
        }
        if engineFlag == "data" || (engineFlag == "auto" && isDataConversion(inputPath, outputPath)) {
            sp := spinner.New("converting ...")
            sp.Start()
            defer sp.Stop()
            if err := data.Convert(inputPath, outputPath, fromFormat, toFormat); err != nil {
                return fmt.Errorf("conversion failed: %w", err)
            }
            fmt.Fprintf(os.Stdout, "Conversion successful: %s -> %s\n", inputPath, outputPath)
            return nil
        }
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

func isDataConversion(src, dst string) bool {
    return data.IsDataFormat(filepath.Ext(src)) && data.IsDataFormat(filepath.Ext(dst))
}

func init() {
    rootCmd.AddCommand(convertCmd)
    convertCmd.Flags().StringVarP(&inputPath, "input", "i", "", "Input file path (required)")
    convertCmd.Flags().StringVarP(&outputPath, "output", "o", "", "Output file path (required)")
    convertCmd.Flags().StringVar(&engineFlag, "engine", "auto", "Conversion engine (ffmpeg|imagemagick|pandoc|data|auto)")
    convertCmd.Flags().StringVar(&fromFlag, "from", "", "Input format (csv|json|yaml|toml); infer from path if empty")
    convertCmd.Flags().StringVar(&toFlag, "to", "", "Output format (csv|json|yaml|toml); infer from path if empty")
    convertCmd.MarkFlagRequired("input")
    convertCmd.MarkFlagRequired("output")
}
