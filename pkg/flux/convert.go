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
	// Engine can be: auto, ffmpeg, imagemagick, pandoc, data.
	// Empty defaults to auto.
	Engine string

	// From and To are optional data-format hints for pipe/data workflows.
	From string
	To   string

	// FormatStyle is used for document conversions via pandoc.
	// Supported: professional, technical, none. Empty defaults to professional.
	FormatStyle string
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
		if !isDataConversion(src, dst) {
			_, workaround, err := engine.CanConvert(src, dst)
			if err != nil {
				if workaround != "" {
					return fmt.Errorf("%v. %s", err, workaround)
				}
				return err
			}
		}
	case "data", "ffmpeg", "imagemagick", "pandoc":
		ok, err := engine.CanEngineConvert(src, dst, engineFlag)
		if err != nil {
			return err
		}
		if !ok {
			bestEngine, workaround, canErr := engine.CanConvert(src, dst)
			if canErr != nil {
				if workaround != "" {
					return fmt.Errorf("%v. %s", canErr, workaround)
				}
				return canErr
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

	return engine.ExecuteRoute(route, src, dst, factory, func(step engine.RouteStep, isFinal bool) []string {
		if step.Engine != "pandoc" || !isFinal {
			return nil
		}
		if !docfmt.IsDocumentFormat("." + step.ToFormat) {
			return nil
		}
		formatter := docfmt.NewDocumentFormatter(opts.FormatStyle)
		return formatter.PandocArgs(dst)
	})
}

func isDataConversion(in, out string) bool {
	from := data.FormatFromExt(in)
	to := data.FormatFromExt(out)
	if from == "" || to == "" {
		return false
	}
	return true
}

func engineNameFromAdapter(eng engine.Engine) string {
	switch eng.(type) {
	case *engine.PandocAdapter:
		return "pandoc"
	case *engine.FFmpegAdapter:
		return "ffmpeg"
	case *engine.ImageMagickAdapter:
		return "imagemagick"
	case *engine.DataAdapter:
		return "data"
	default:
		return "auto"
	}
}
