package engine

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/kelyonnnn17/flux/internal/data"
)

// ExecuteRoute runs a planned conversion route from src to dst.
// It manages intermediate files and cleanup automatically.
func ExecuteRoute(route ConversionRoute, src, dst string, factory *EngineFactory, argBuilder func(step RouteStep, isFinal bool) []string) error {
	if len(route.Steps) == 0 {
		return fmt.Errorf("conversion route has no steps")
	}

	current := src
	temps := make([]string, 0)
	defer func() {
		for _, p := range temps {
			_ = os.Remove(p)
		}
	}()

	for i, step := range route.Steps {
		isFinal := i == len(route.Steps)-1
		next := dst
		if !isFinal {
			tmp, err := os.CreateTemp("", "flux-step-*."+step.ToFormat)
			if err != nil {
				return fmt.Errorf("create temp intermediate for %s: %w", step.ToFormat, err)
			}
			next = tmp.Name()
			_ = tmp.Close()
			temps = append(temps, next)
		}

		if err := executeStep(factory, step, current, next, argBuilder(step, isFinal)); err != nil {
			return fmt.Errorf("step %d/%d (%s %s->%s) failed: %w", i+1, len(route.Steps), step.Engine, step.FromFormat, step.ToFormat, err)
		}
		current = next
	}

	return nil
}

func executeStep(factory *EngineFactory, step RouteStep, src, dst string, args []string) error {
	switch step.Engine {
	case "data":
		return data.Convert(src, dst, step.FromFormat, step.ToFormat)
	case "pdftotext":
		return extractPDFText(factory.runner, src, dst)
	default:
		eng, err := factory.GetEngine(step.Engine)
		if err != nil {
			return err
		}
		return eng.Convert(src, dst, args)
	}
}

func extractPDFText(runner CmdRunner, srcPDF, dstText string) error {
	if !binaryExists("pdftotext") {
		return fmt.Errorf("pdftotext not found. Install poppler utils and retry PDF conversion")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()
	cmd := runner.CommandContext(ctx, "pdftotext", "-layout", srcPDF, dstText)
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("pdftotext error: %w, output: %s", err, string(out))
	}
	return nil
}
