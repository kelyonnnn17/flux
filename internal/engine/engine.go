package engine

import (
    "errors"
    "os/exec"
    "path/filepath"
    "strings"
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
