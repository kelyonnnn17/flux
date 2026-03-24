package engine

import (
	"errors"
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"
)

var (
	// engineInputFormats maps engine names to formats they can READ (input)
	engineInputFormats = map[string]map[string]bool{
		"pandoc": {
			"pdf":  false, // PDF input NOT supported
			"docx": true, "odt": true, "md": true, "txt": true, "tex": true,
			"epub": true, "html": true, "rst": true,
		},
		"imagemagick": {
			"jpg": true, "jpeg": true, "png": true, "gif": true,
			"tiff": true, "tif": true, "bmp": true, "webp": true, "svg": true,
		},
		"ffmpeg": {
			"mp4": true, "mkv": true, "avi": true, "mov": true,
			"mp3": true, "wav": true, "webm": true, "m4a": true, "flac": true, "ogg": true,
		},
		"data": {
			"csv": true, "tsv": true, "json": true, "yaml": true, "yml": true, "toml": true,
		},
	}

	// engineOutputFormats maps engine names to formats they can WRITE (output)
	engineOutputFormats = map[string]map[string]bool{
		"pandoc": {
			"pdf": true, "docx": true, "odt": true, "md": true, "txt": true, "tex": true,
			"epub": true, "html": true, "rst": true,
		},
		"imagemagick": {
			"jpg": true, "jpeg": true, "png": true, "gif": true,
			"tiff": true, "tif": true, "bmp": true, "webp": true, "svg": true,
		},
		"ffmpeg": {
			"mp4": true, "mkv": true, "avi": true, "mov": true,
			"mp3": true, "wav": true, "webm": true, "m4a": true, "flac": true, "ogg": true,
		},
		"data": {
			"csv": true, "tsv": true, "json": true, "yaml": true, "yml": true, "toml": true,
		},
	}
)

type Engine interface {
	// Convert performs conversion from src to dst using optional args.
	Convert(src, dst string, args []string) error
}

type EngineFactory struct {
	runner CmdRunner
}

func NewFactory(runner CmdRunner) *EngineFactory {
	return &EngineFactory{runner: runner}
}

// GetEngine returns an Engine implementation based on the name.
func (f *EngineFactory) GetEngine(name string) (Engine, error) {
	switch name {
	case "ffmpeg":
		return &FFmpegAdapter{Runner: f.runner}, nil
	case "imagemagick":
		return &ImageMagickAdapter{Runner: f.runner}, nil
	case "pandoc":
		return &PandocAdapter{Runner: f.runner}, nil
	case "data":
		return &DataAdapter{}, nil
	default:
		return nil, errors.New("unknown engine: " + name)
	}
}

// AutoEngine attempts to select the first available engine from the list.
func (f *EngineFactory) AutoEngine(preferred []string) (Engine, error) {
	for _, name := range preferred {
		if binaryExists(name) {
			return f.GetEngine(name)
		}
	}
	return nil, errors.New("no conversion engine found. Install one: brew install ffmpeg imagemagick pandoc | apt install ffmpeg imagemagick pandoc | choco install ffmpeg imagemagick pandoc")
}

// RouteByFormat returns the preferred engine order for converting src to dst,
// based on file extensions. Used when engine is "auto".
// PRD: Documents -> Pandoc, Images -> ImageMagick, Audio/Video -> FFmpeg, Data -> Go.
func RouteByFormat(src, dst string) []string {
	srcExt := strings.ToLower(strings.TrimPrefix(filepath.Ext(src), "."))
	dstExt := strings.ToLower(strings.TrimPrefix(filepath.Ext(dst), "."))
	if isDataExt(srcExt) && isDataExt(dstExt) {
		return []string{"data", "ffmpeg", "imagemagick", "pandoc"}
	}
	ext := srcExt
	switch ext {
	case "pdf", "docx", "odt", "md", "tex", "epub", "html", "rst":
		return []string{"pandoc", "imagemagick", "ffmpeg"}
	case "jpg", "jpeg", "png", "gif", "tiff", "tif", "bmp", "webp", "svg":
		return []string{"imagemagick", "ffmpeg", "pandoc"}
	case "mp4", "mkv", "avi", "mov", "mp3", "wav", "webm", "m4a", "flac", "ogg":
		return []string{"ffmpeg", "imagemagick", "pandoc"}
	case "csv", "tsv", "json", "yaml", "toml":
		return []string{"data", "ffmpeg", "imagemagick", "pandoc"}
	default:
		return []string{"ffmpeg", "imagemagick", "pandoc"}
	}
}

func isDataExt(ext string) bool {
	switch ext {
	case "csv", "tsv", "json", "yaml", "yml", "toml":
		return true
	}
	return false
}

func binaryExists(name string) bool {
	if name == "data" {
		return true // data converter is always available
	}
	switch name {
	case "ffmpeg", "pandoc":
		if _, err := exec.LookPath(name); err == nil {
			return true
		}
	case "pdftotext":
		if _, err := exec.LookPath("pdftotext"); err == nil {
			return true
		}
	case "imagemagick":
		if _, err := exec.LookPath("magick"); err == nil {
			return true
		}
		if _, err := exec.LookPath("convert"); err == nil {
			return true
		}
	}
	return false
}

// CanConvert checks if converting from src to dst format is possible.
// Returns the best engine name, a suggested workaround, and an error.
func CanConvert(src, dst string) (string, string, error) {
	route, err := PlanConversion(src, dst, "auto")
	if err == nil {
		return route.PrimaryEngine(), "", nil
	}
	srcExt := normalizeFormat(src)
	dstExt := normalizeFormat(dst)
	alternative := suggestAlternative(srcExt, dstExt)
	return "", alternative, fmt.Errorf("%s->%s conversion not supported", srcExt, dstExt)
}

// CanEngineConvert checks if a specific engine can convert src->dst, based only on
// format extensions (no runtime binary availability checks).
func CanEngineConvert(src, dst, engineName string) (bool, error) {
	srcExt := normalizeFormat(src)
	dstExt := normalizeFormat(dst)

	if srcExt == dstExt {
		return false, fmt.Errorf("source and target format are the same (%s)", srcExt)
	}

	engineName = strings.ToLower(engineName)
	if _, ok := engineInputFormats[engineName]; !ok {
		return false, fmt.Errorf("unknown engine: %s", engineName)
	}

	_, err := PlanConversion(src, dst, engineName)
	if err != nil {
		return false, nil
	}
	return true, nil
}

func suggestAlternative(srcExt, dstExt string) string {
	// Special case: PDF input
	if srcExt == "pdf" {
		// Suggest converting PDF to intermediate format
		if outputs := engineOutputFormats["imagemagick"]; outputs["png"] {
			return fmt.Sprintf("PDF cannot be read directly. Workaround: (1) flux file.pdf png, then (2) flux file.png %s", dstExt)
		}
	}

	// Check what output format the target engine supports
	for _, engineName := range []string{"pandoc", "imagemagick", "ffmpeg", "data"} {
		outputs := engineOutputFormats[engineName]
		if outputs[dstExt] {
			// Find a common intermediate format
			for intermediate := range engineOutputFormats["imagemagick"] {
				if engineInputFormats[engineName][intermediate] {
					return fmt.Sprintf("Workaround: (1) flux file.%s %s, then (2) flux file.%s %s", srcExt, intermediate, intermediate, dstExt)
				}
			}
		}
	}

	return "No conversion path found. Check supported formats: flux lf"
}
