package cmd

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/kelyonnnn17/flux/internal/data"
	"github.com/kelyonnnn17/flux/internal/engine"
	"github.com/kelyonnnn17/flux/internal/format"
	"github.com/kelyonnnn17/flux/internal/spinner"
	"github.com/kelyonnnn17/flux/internal/ui"
	"github.com/spf13/cobra"
)

var (
	inputPaths       []string
	outputPath       string
	fromFlag         string
	toFlag           string
	forceFlag        bool
	quietFlag        bool
	formatStyleFlag  string
	referenceDocFlag string
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
	convertCmd.Flags().StringVar(&formatStyleFlag, "format-style", "professional", "Apply document formatting preset (professional|technical|developer|none); does not auto-generate TOC/section numbering")
	convertCmd.Flags().StringVar(&referenceDocFlag, "reference-doc", "", "Path to DOCX reference template for DOCX output (preserves styles)")
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
	return runBatchMode(resolvedInputs, out, engineFlag, formatStyleFlag, referenceDocFlag)
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

func runBatchMode(inputs []string, output string, engineFlag string, formatStyle string, referenceDoc string) error {
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
		if err := convertOne(in, out, engineFlag, formatStyle, referenceDoc); err != nil {
			return err
		}
		if !quietFlag && !useConvertUI(out) {
			format.Success("Converted %s -> %s", in, out)
		}
	}
	return nil
}

func convertOne(in, out string, engineFlag string, formatStyle string, referenceDoc string) error {
	if useConvertUI(out) {
		return runConvertAnimated(in, out, engineFlag, formatStyle, referenceDoc)
	}
	return convertOneWithSpinner(in, out, engineFlag, formatStyle, referenceDoc)
}

func runConvertAnimated(in, out string, engineFlag string, formatStyle string, referenceDoc string) error {
	started := time.Now()
	return ui.Run(ui.Spec{
		Command: fmt.Sprintf("flux convert %s %s", in, out),
		Running: fmt.Sprintf("converting %s", in),
		Run: func() (ui.Result, error) {
			route, err := convertOneRaw(in, out, engineFlag, formatStyle, referenceDoc)
			if err != nil {
				return ui.Result{ErrorHint: "check supported formats: flux lf"}, err
			}

			meta := []ui.Meta{
				{Key: "from", Value: in},
				{Key: "to", Value: out},
				{Key: "engine", Value: route.PrimaryEngine()},
				{Key: "steps", Value: fmt.Sprintf("%d", len(route.Steps))},
				{Key: "time", Value: time.Since(started).Round(time.Millisecond).String()},
			}
			if st, err := os.Stat(out); err == nil {
				meta = append(meta, ui.Meta{Key: "size", Value: fmt.Sprintf("%d bytes", st.Size())})
			}

			lines := make([]ui.Line, 0, len(route.Warnings))
			for _, w := range route.Warnings {
				lines = append(lines, ui.Line{Kind: "warn", Text: w})
			}

			return ui.Result{
				Meta:    meta,
				Lines:   lines,
				Success: "OK converted",
			}, nil
		},
	})
}

func convertOneWithSpinner(in, out string, engineFlag string, formatStyle string, referenceDoc string) error {
	sp := spinner.New(resolveEngineLabel(in, out, engineFlag) + " ...")
	sp.Start()
	defer sp.Stop()
	_, err := convertOneRaw(in, out, engineFlag, formatStyle, referenceDoc)
	return err
}

func convertOneRaw(in, out string, engineFlag string, formatStyle string, referenceDoc string) (engine.ConversionRoute, error) {
	fromFormat := fromFlag
	if fromFormat == "" {
		fromFormat = data.FormatFromExt(in)
	}
	toFormat := toFlag
	if toFormat == "" {
		toFormat = data.FormatFromExt(out)
	}

	srcExt := strings.ToLower(strings.TrimPrefix(filepath.Ext(in), "."))
	dstExt := strings.ToLower(strings.TrimPrefix(filepath.Ext(out), "."))

	// Validate conversion is possible
	engineFlag = strings.ToLower(engineFlag)
	switch engineFlag {
	case "auto":
		if err := validateAutoEngine(in, out, srcExt, dstExt); err != nil {
			return engine.ConversionRoute{}, err
		}
	case "data", "ffmpeg", "imagemagick", "pandoc", "pdf2docx", "docx2pdf":
		ok, err := engine.CanEngineConvert(in, out, engineFlag)
		if err != nil {
			return engine.ConversionRoute{}, err
		}
		if !ok {
			bestEngine, workaround, canErr := engine.CanConvert(in, out)
			if canErr != nil {
				return engine.ConversionRoute{}, withWorkaround(canErr, workaround)
			}
			if bestEngine == "" {
				bestEngine = "auto"
			}
			return engine.ConversionRoute{}, fmt.Errorf("engine %s cannot convert %s -> %s end-to-end; try --engine %s or --engine auto", engineFlag, srcExt, dstExt, bestEngine)
		}
	default:
		return engine.ConversionRoute{}, fmt.Errorf("unknown engine: %s", engineFlag)
	}

	if engineFlag == "data" || (engineFlag == "auto" && isDataConversion(in, out)) {
		return engine.ConversionRoute{Steps: []engine.RouteStep{{
			Engine:     "data",
			FromFormat: fromFormat,
			ToFormat:   toFormat,
			Cost:       1,
		}}}, data.Convert(in, out, fromFormat, toFormat)
	}

	route, err := engine.PlanConversion(in, out, engineFlag)
	if err != nil {
		if engineFlag != "auto" {
			return engine.ConversionRoute{}, fmt.Errorf("engine %s cannot perform route %s -> %s; try --engine auto", engineFlag, srcExt, dstExt)
		}
		return engine.ConversionRoute{}, err
	}

	factory := engine.NewFactory(engine.NewDefaultRunner())
	if len(route.Warnings) > 0 && !quietFlag && !useConvertUI(out) {
		for _, w := range route.Warnings {
			format.Info("Warning: %s", w)
		}
	}

	executeRoute := func(r engine.ConversionRoute) error {
		return engine.ExecuteRoute(r, in, out, factory, func(step engine.RouteStep, isFinal bool) []string {
			if step.Engine != "pandoc" {
				return nil
			}
			if !isFinal {
				return nil
			}
			if !format.IsDocumentFormat("." + step.ToFormat) {
				return nil
			}
			formatter := format.NewDocumentFormatter(formatStyle)
			return formatter.PandocArgsWithContext(in, out, referenceDoc)
		})
	}

	err = executeRoute(route)
	if err != nil && engineFlag == "auto" && srcExt == "docx" && dstExt == "pdf" && len(route.Steps) == 1 && route.Steps[0].Engine == "docx2pdf" {
		fallbackRoute, planErr := engine.PlanConversion(in, out, "pandoc")
		if planErr == nil {
			if fallbackErr := executeRoute(fallbackRoute); fallbackErr == nil {
				msg := "docx2pdf failed at runtime; fell back to pandoc for DOCX->PDF"
				fallbackRoute.Warnings = append(fallbackRoute.Warnings, msg)
				if !quietFlag && !useConvertUI(out) {
					format.Info("Warning: %s", msg)
				}
				return fallbackRoute, nil
			}
		}
	}
	if err != nil {
		return engine.ConversionRoute{}, err
	}

	return route, nil
}

func validateAutoEngine(in, out, srcExt, dstExt string) error {
	if srcExt == "pdf" && dstExt == "docx" {
		ok, err := engine.CanEngineConvert(in, out, "pdf2docx")
		if err != nil {
			return err
		}
		if !ok {
			return errors.New("python pdf2docx adapter is required for high-fidelity PDF->DOCX. Run: make setup")
		}
	}

	if isDataConversion(in, out) {
		return nil
	}

	_, workaround, err := engine.CanConvert(in, out)
	if err != nil {
		return withWorkaround(err, workaround)
	}
	return nil
}

func withWorkaround(err error, workaround string) error {
	if workaround == "" {
		return err
	}
	return fmt.Errorf("%v. %s", err, workaround)
}

func resolveEngineLabel(in, out, engineFlag string) string {
	if engineFlag != "auto" {
		return engineFlag
	}
	if isDataConversion(in, out) {
		return "data"
	}
	preferred := engine.RouteByFormat(in, out)
	if len(preferred) == 0 {
		return "auto"
	}
	return preferred[0]
}

func useConvertUI(out string) bool {
	if quietFlag {
		return false
	}
	if out == "-" {
		return false
	}
	return ui.ShouldRender()
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
