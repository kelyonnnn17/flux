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
	for _, name := range []string{"ffmpeg", "imagemagick", "pandoc", "pdf2docx", "docx2pdf", "pdftotext", "data"} {
		info := EngineInfo{Name: name}
		if name == "data" {
			info.Path = "built-in"
			info.Version = "go"
			out = append(out, info)
			continue
		}
		if name == "pdf2docx" || name == "docx2pdf" {
			py, err := resolvePythonExecutable()
			if err != nil || !pythonModuleAvailable(name) {
				info.Path = "not found"
				out = append(out, info)
				continue
			}
			info.Path = py + " (python module)"
			info.Version = getPythonModuleVersion(py, name)
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

func getPythonModuleVersion(pythonPath, module string) string {
	cmd := exec.Command(pythonPath, "-c", "import "+module+" as m; print(getattr(m, '__version__', '?'))")
	out, err := cmd.CombinedOutput()
	if err != nil {
		return "?"
	}
	return strings.TrimSpace(string(out))
}

func getVersion(name, path string) string {
	args := []string{"-version"}
	switch name {
	case "pandoc":
		args = []string{"--version"}
	case "pdftotext":
		args = []string{"-v"}
	}
	cmd := exec.Command(path, args...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return "?"
	}
	lines := strings.SplitN(string(out), "\n", 2)
	return strings.TrimSpace(lines[0])
}
