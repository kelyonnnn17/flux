package flux

import (
	"errors"
	"fmt"
	"path/filepath"

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

	fromFormat := opts.From
	if fromFormat == "" {
		fromFormat = data.FormatFromExt(src)
	}
	toFormat := opts.To
	if toFormat == "" {
		toFormat = data.FormatFromExt(dst)
	}

	if engineFlag != "data" && !isDataConversion(src, dst) {
		_, workaround, err := engine.CanConvert(src, dst)
		if err != nil {
			if workaround != "" {
				return fmt.Errorf("%v. %s", err, workaround)
			}
			return err
		}
	}

	if engineFlag == "data" || (engineFlag == "auto" && isDataConversion(src, dst)) {
		return data.Convert(src, dst, fromFormat, toFormat)
	}

	factory := engine.NewFactory(engine.NewDefaultRunner())
	var eng engine.Engine
	var err error
	selectedEngine := engineFlag
	if engineFlag == "auto" {
		preferred := engine.RouteByFormat(src, dst)
		eng, err = factory.AutoEngine(preferred)
		if err != nil {
			return err
		}
		selectedEngine = engineNameFromAdapter(eng)
	} else {
		eng, err = factory.GetEngine(engineFlag)
		if err != nil {
			return err
		}
	}

	args := []string{}
	outExt := filepath.Ext(dst)
	if selectedEngine == "pandoc" && docfmt.IsDocumentFormat(outExt) {
		formatter := docfmt.NewDocumentFormatter(opts.FormatStyle)
		args = formatter.PandocArgs(dst)
	}

	return eng.Convert(src, dst, args)
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
