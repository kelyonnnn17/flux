package engine

import (
	"os/exec"
	"strings"
)

// EngineInfo holds engine name, path, and version.
type EngineInfo struct {
	Name    string
	Path    string
	Version string
}

// Doctor returns status of all engines.
func Doctor() []EngineInfo {
	var out []EngineInfo
	for _, name := range []string{"ffmpeg", "imagemagick", "pandoc", "data"} {
		info := EngineInfo{Name: name}
		if name == "data" {
			info.Path = "built-in"
			info.Version = "go"
			out = append(out, info)
			continue
		}
		path, err := exec.LookPath(name)
		if name == "imagemagick" {
			if err != nil {
				path, err = exec.LookPath("magick")
			}
			if err != nil {
				path, err = exec.LookPath("convert")
			}
		}
		if err != nil {
			info.Path = "not found"
			out = append(out, info)
			continue
		}
		info.Path = path
		info.Version = getVersion(name, path)
		out = append(out, info)
	}
	return out
}

func getVersion(name, path string) string {
	cmd := exec.Command(path, "-version")
	if name == "imagemagick" {
		cmd = exec.Command(path, "-version")
	}
	out, err := cmd.CombinedOutput()
	if err != nil {
		return "?"
	}
	lines := strings.SplitN(string(out), "\n", 2)
	return strings.TrimSpace(lines[0])
}
