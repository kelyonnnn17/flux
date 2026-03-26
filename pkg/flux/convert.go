package flux

import (
	"errors"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/kelyonnnn17/flux/internal/data"
	"github.com/kelyonnnn17/flux/internal/engine"
	docfmt "github.com/kelyonnnn17/flux/internal/format"
)

// ConvertOptions controls conversion behavior for library users.
type ConvertOptions struct {
	// Engine can be: auto, pdf2docx, docx2pdf, ffmpeg, imagemagick, pandoc, data.
	// Empty defaults to auto.
	Engine string

	// From and To are optional data-format hints for pipe/data workflows.
	From string
	To   string

	// FormatStyle is used for document conversions via pandoc.
	// Supported: professional, technical, developer, none. Empty defaults to professional.
	FormatStyle string

	// ReferenceDoc points to a DOCX template used when writing DOCX output.
	// If empty and source is DOCX, the source document is used as reference.
	ReferenceDoc string
}

// Convert converts src file to dst file using the same routing logic as the CLI.
func Convert(src, dst string, opts ConvertOptions) error {
	if src == "" || dst == "" {
		return errors.New("src and dst are required")
	}

	engineFlag := opts.Engine
	if engineFlag == "" {
		engineFlag = "auto"
	}
	engineFlag = strings.ToLower(engineFlag)

	fromFormat := opts.From
	if fromFormat == "" {
		fromFormat = data.FormatFromExt(src)
	}
	toFormat := opts.To
	if toFormat == "" {
		toFormat = data.FormatFromExt(dst)
	}

	srcExt := strings.ToLower(strings.TrimPrefix(filepath.Ext(src), "."))
	dstExt := strings.ToLower(strings.TrimPrefix(filepath.Ext(dst), "."))

	switch engineFlag {
	case "auto":
		if err := validateAutoEngine(src, dst, srcExt, dstExt); err != nil {
			return err
		}
	case "data", "ffmpeg", "imagemagick", "pandoc", "pdf2docx", "docx2pdf":
		ok, err := engine.CanEngineConvert(src, dst, engineFlag)
		if err != nil {
			return err
		}
		if !ok {
			bestEngine, workaround, canErr := engine.CanConvert(src, dst)
			if canErr != nil {
				return withWorkaround(canErr, workaround)
			}
			if bestEngine == "" {
				bestEngine = "auto"
			}
			return fmt.Errorf("engine %s cannot convert %s -> %s; try --engine %s or --engine auto", engineFlag, srcExt, dstExt, bestEngine)
		}
	default:
		return fmt.Errorf("unknown engine: %s", engineFlag)
	}

	if engineFlag == "data" || (engineFlag == "auto" && isDataConversion(src, dst)) {
		return data.Convert(src, dst, fromFormat, toFormat)
	}

	route, err := engine.PlanConversion(src, dst, engineFlag)
	if err != nil {
		if engineFlag != "auto" {
			return fmt.Errorf("engine %s cannot perform route %s -> %s; try --engine auto", engineFlag, srcExt, dstExt)
		}
		return err
	}

	factory := engine.NewFactory(engine.NewDefaultRunner())

	executeRoute := func(r engine.ConversionRoute) error {
		return engine.ExecuteRoute(r, src, dst, factory, func(step engine.RouteStep, isFinal bool) []string {
			if step.Engine != "pandoc" || !isFinal {
				return nil
			}
			if !docfmt.IsDocumentFormat("." + step.ToFormat) {
				return nil
			}
			formatter := docfmt.NewDocumentFormatter(opts.FormatStyle)
			return formatter.PandocArgsWithContext(src, dst, opts.ReferenceDoc)
		})
	}

	err = executeRoute(route)
	if err == nil {
		return nil
	}

	if engineFlag == "auto" && srcExt == "docx" && dstExt == "pdf" && len(route.Steps) == 1 && route.Steps[0].Engine == "docx2pdf" {
		fallbackRoute, planErr := engine.PlanConversion(src, dst, "pandoc")
		if planErr == nil {
			if fallbackErr := executeRoute(fallbackRoute); fallbackErr == nil {
				return nil
			}
		}
	}

	return err
}

func isDataConversion(in, out string) bool {
	from := data.FormatFromExt(in)
	to := data.FormatFromExt(out)
	if from == "" || to == "" {
		return false
	}
	return true
}

func validateAutoEngine(src, dst, srcExt, dstExt string) error {
	if srcExt == "pdf" && dstExt == "docx" {
		ok, err := engine.CanEngineConvert(src, dst, "pdf2docx")
		if err != nil {
			return err
		}
		if !ok {
			return errors.New("python pdf2docx adapter is required for high-fidelity PDF->DOCX. Run: make setup")
		}
	}

	if isDataConversion(src, dst) {
		return nil
	}

	_, workaround, err := engine.CanConvert(src, dst)
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
