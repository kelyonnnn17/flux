package cmd

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/kelyonnnn17/flux/internal/data"
	"github.com/kelyonnnn17/flux/internal/engine"
	"github.com/kelyonnnn17/flux/internal/format"
	"github.com/kelyonnnn17/flux/internal/spinner"
	"github.com/spf13/cobra"
)

var (
	inputPaths []string
	outputPath string
	fromFlag   string
	toFlag     string
	forceFlag  bool
	quietFlag  bool
)

var convertCmd = &cobra.Command{
	Use:     "convert [input_path] [output_format_or_path]",
	Aliases: []string{"c"},
	Short:   "Convert files between formats",
	RunE:    runConvert,
}

func init() {
	rootCmd.AddCommand(convertCmd)
	convertCmd.Flags().StringSliceVarP(&inputPaths, "input", "i", nil, "Input file path(s); use - for stdin, supports globs")
	convertCmd.Flags().StringVarP(&outputPath, "output", "o", "", "Output path or extension (e.g. png); use - for stdout")
	convertCmd.Flags().StringVar(&fromFlag, "from", "", "Input format (csv|json|yaml|toml); required for pipe")
	convertCmd.Flags().StringVar(&toFlag, "to", "", "Output format (csv|json|yaml|toml); required for pipe")
	convertCmd.Flags().BoolVarP(&forceFlag, "force", "f", false, "Overwrite existing files without prompting")
	convertCmd.Flags().BoolVarP(&quietFlag, "quiet", "q", false, "Suppress output; exit code only")
}

func runConvert(c *cobra.Command, args []string) error {
	inputs, out, err := mergeConvertArgs(args, inputPaths, outputPath)
	if err != nil {
		return err
	}

	engineFlag, _ := c.Root().PersistentFlags().GetString("engine")
	if len(inputs) == 0 {
		if isStdinPipe() {
			return runPipeMode(out)
		}
		return fmt.Errorf("input required: use flux convert <input_path> <output_format_or_path> or -i <path>")
	}
	if inputs[0] == "-" {
		if len(inputs) > 1 {
			return fmt.Errorf("stdin mode accepts only one input")
		}
		return runPipeMode(out)
	}
	if out == "" {
		return fmt.Errorf("output required: use <output_format_or_path> or -o")
	}
	resolvedInputs, err := expandGlobs(inputs)
	if err != nil {
		return err
	}
	if len(resolvedInputs) == 0 {
		return fmt.Errorf("no matching input files")
	}
	return runBatchMode(resolvedInputs, out, engineFlag)
}

func mergeConvertArgs(args []string, existingInputs []string, existingOutput string) ([]string, string, error) {
	if len(args) > 2 {
		return nil, "", fmt.Errorf("too many arguments: expected flux convert <input_path> <output_format_or_path>")
	}

	mergedInputs := append([]string(nil), existingInputs...)
	mergedOutput := existingOutput

	if len(args) >= 1 {
		if len(existingInputs) > 0 {
			return nil, "", fmt.Errorf("cannot mix positional input with -i/--input")
		}
		mergedInputs = []string{args[0]}
	}

	if len(args) == 2 {
		if existingOutput != "" {
			return nil, "", fmt.Errorf("cannot mix positional output with -o/--output")
		}
		mergedOutput = args[1]
	}

	return mergedInputs, mergedOutput, nil
}

func expandGlobs(paths []string) ([]string, error) {
	var out []string
	for _, p := range paths {
		matches, err := filepath.Glob(p)
		if err != nil {
			return nil, fmt.Errorf("glob %s: %w", p, err)
		}
		for _, m := range matches {
			info, err := os.Stat(m)
			if err != nil {
				continue
			}
			if info.Mode().IsRegular() {
				out = append(out, m)
			}
		}
	}
	return out, nil
}

func runBatchMode(inputs []string, output string, engineFlag string) error {
	outputIsExt := isOutputExtension(output, len(inputs))
	for _, in := range inputs {
		out := resolveOutputPath(in, output, outputIsExt, len(inputs))
		if out == "" {
			return fmt.Errorf("cannot resolve output for %s", in)
		}
		if !forceFlag {
			if _, err := os.Stat(out); err == nil {
				if !promptOverwrite(out) {
					continue
				}
			}
		}
		if err := convertOne(in, out, engineFlag); err != nil {
			format.Error("%s", err)
			return err
		}
		if !quietFlag {
			format.Success("Converted %s -> %s", in, out)
		}
	}
	return nil
}

func convertOne(in, out string, engineFlag string) error {
	fromFormat := fromFlag
	if fromFormat == "" {
		fromFormat = data.FormatFromExt(in)
	}
	toFormat := toFlag
	if toFormat == "" {
		toFormat = data.FormatFromExt(out)
	}

	// Validate conversion is possible
	if engineFlag != "data" && !isDataConversion(in, out) {
		engineName, workaround, err := engine.CanConvert(in, out)
		if err != nil {
			var msg string
			if workaround != "" {
				msg = fmt.Sprintf("%v. %s", err, workaround)
			} else {
				msg = fmt.Sprintf("%v", err)
			}
			return fmt.Errorf(msg)
		}
		engineName = engineName // Validation passed, use validated engine
	}

	if engineFlag == "data" || (engineFlag == "auto" && isDataConversion(in, out)) {
		sp := spinner.New("data ...")
		sp.Start()
		defer sp.Stop()
		return data.Convert(in, out, fromFormat, toFormat)
	}
	factory := engine.NewFactory(engine.NewDefaultRunner())
	var eng engine.Engine
	var err error
	var engineName string
	if engineFlag == "auto" {
		preferred := engine.RouteByFormat(in, out)
		eng, err = factory.AutoEngine(preferred)
		if err != nil {
			return err
		}
		engineName = preferred[0]
	} else {
		eng, err = factory.GetEngine(engineFlag)
		if err != nil {
			return err
		}
		engineName = engineFlag
	}
	sp := spinner.New(engineName + " ...")
	sp.Start()
	defer sp.Stop()
	return eng.Convert(in, out, []string{})
}

func isOutputExtension(output string, inputCount int) bool {
	if output == "" || output == "-" {
		return false
	}
	if filepath.IsAbs(output) || strings.ContainsRune(output, filepath.Separator) {
		return false
	}
	if inputCount > 1 {
		return true
	}
	return !strings.Contains(output, ".")
}

func resolveOutputPath(in string, output string, outputIsExt bool, inputCount int) string {
	if output == "" || output == "-" {
		return ""
	}
	if outputIsExt {
		ext := strings.TrimPrefix(output, ".")
		return strings.TrimSuffix(in, filepath.Ext(in)) + "." + ext
	}
	if inputCount > 1 {
		dir := filepath.Dir(in)
		base := strings.TrimSuffix(filepath.Base(in), filepath.Ext(in)) + filepath.Ext(output)
		return filepath.Join(dir, base)
	}
	return output
}

func promptOverwrite(out string) bool {
	fmt.Fprintf(os.Stderr, "Overwrite %s? [y/N] ", out)
	scanner := bufio.NewScanner(os.Stdin)
	if scanner.Scan() {
		line := strings.TrimSpace(strings.ToLower(scanner.Text()))
		return line == "y" || line == "yes"
	}
	return false
}

func runPipeMode(output string) error {
	if output != "-" {
		return fmt.Errorf("pipe mode requires -o -")
	}
	if fromFlag == "" || toFlag == "" {
		return fmt.Errorf("pipe mode requires --from and --to")
	}
	if !quietFlag {
		format.Info("Converting stdin (%s) -> stdout (%s)", fromFlag, toFlag)
	}
	return data.ConvertStream(os.Stdin, os.Stdout, fromFlag, toFlag)
}

func isStdinPipe() bool {
	info, err := os.Stdin.Stat()
	if err != nil {
		return false
	}
	return (info.Mode() & os.ModeCharDevice) == 0
}

func isDataConversion(src, dst string) bool {
	return data.IsDataFormat(filepath.Ext(src)) && data.IsDataFormat(filepath.Ext(dst))
}
