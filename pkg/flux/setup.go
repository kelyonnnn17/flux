package flux

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
)

// Dependency describes an external runtime tool used by flux.
type Dependency struct {
	Name      string
	Binary    string
	Installed bool
	Path      string
	Version   string
}

// SetupReport contains dependency state before/after optional auto-install.
type SetupReport struct {
	OS               string
	Dependencies     []Dependency
	InstallAttempted bool
	InstallCommands  []string
}

// CheckDependencies returns currently available runtime engines.
func CheckDependencies() SetupReport {
	deps := []Dependency{
		probeBinary("ffmpeg", "ffmpeg"),
		probeImageMagick(),
		probeBinary("pandoc", "pandoc"),
		probePythonModule("pdf2docx"),
		probePythonModule("docx2pdf"),
	}
	return SetupReport{OS: runtime.GOOS, Dependencies: deps}
}

// EnsureDependencies validates required external binaries.
// If autoInstall is true, it attempts installation using common package managers.
func EnsureDependencies(ctx context.Context, autoInstall bool) (SetupReport, error) {
	report := CheckDependencies()
	missing := missingNames(report.Dependencies)
	if len(missing) == 0 {
		return report, nil
	}

	if !autoInstall {
		return report, fmt.Errorf("missing dependencies: %s", strings.Join(missing, ", "))
	}

	commands, err := installCommandsForOS(missing)
	if err != nil {
		return report, err
	}
	report.InstallAttempted = true
	report.InstallCommands = commands

	for _, cmd := range commands {
		if err := runShellCommand(ctx, cmd); err != nil {
			return report, fmt.Errorf("dependency install failed while running %q: %w", cmd, err)
		}
	}

	report = CheckDependencies()
	remaining := missingNames(report.Dependencies)
	if len(remaining) > 0 {
		return report, fmt.Errorf("still missing after install attempt: %s", strings.Join(remaining, ", "))
	}
	return report, nil
}

func probeBinary(name, bin string) Dependency {
	dep := Dependency{Name: name, Binary: bin}
	if p, err := exec.LookPath(bin); err == nil {
		dep.Installed = true
		dep.Path = p
		dep.Version = detectVersion(bin)
	}
	return dep
}

func probeImageMagick() Dependency {
	dep := Dependency{Name: "imagemagick", Binary: "magick|convert"}
	if p, err := exec.LookPath("magick"); err == nil {
		dep.Installed = true
		dep.Path = p
		dep.Version = detectVersion("magick")
		return dep
	}
	if p, err := exec.LookPath("convert"); err == nil {
		dep.Installed = true
		dep.Path = p
		dep.Version = detectVersion("convert")
	}
	return dep
}

func probePythonModule(module string) Dependency {
	dep := Dependency{Name: module, Binary: "python module"}
	pythonBin, err := findPythonBinary()
	if err != nil {
		return dep
	}

	check := exec.Command(pythonBin, "-c", "import "+module)
	if err := check.Run(); err != nil {
		return dep
	}

	dep.Installed = true
	dep.Path = pythonBin + " (module)"
	dep.Version = detectPythonModuleVersion(pythonBin, module)
	return dep
}

func findPythonBinary() (string, error) {
	if p, err := exec.LookPath("python3"); err == nil {
		return p, nil
	}
	if p, err := exec.LookPath("python"); err == nil {
		return p, nil
	}
	return "", errors.New("python3/python not found")
}

func detectPythonModuleVersion(pythonBin, module string) string {
	cmd := exec.Command(pythonBin, "-c", "import "+module+" as m; print(getattr(m, '__version__', ''))")
	if out, err := cmd.Output(); err == nil {
		return strings.TrimSpace(string(out))
	}
	return ""
}

func detectVersion(bin string) string {
	cmd := exec.Command(bin, "-version")
	if out, err := cmd.Output(); err == nil {
		line := strings.Split(strings.TrimSpace(string(out)), "\n")
		if len(line) > 0 {
			return line[0]
		}
	}
	cmd = exec.Command(bin, "--version")
	if out, err := cmd.Output(); err == nil {
		line := strings.Split(strings.TrimSpace(string(out)), "\n")
		if len(line) > 0 {
			return line[0]
		}
	}
	return ""
}

func missingNames(deps []Dependency) []string {
	missing := []string{}
	for _, d := range deps {
		if !d.Installed {
			missing = append(missing, d.Name)
		}
	}
	return missing
}

func installCommandsForOS(missing []string) ([]string, error) {
	systemPkgNames := map[string]string{
		"ffmpeg":      "ffmpeg",
		"imagemagick": "imagemagick",
		"pandoc":      "pandoc",
	}
	pythonModuleNames := map[string]string{
		"pdf2docx": "pdf2docx",
		"docx2pdf": "docx2pdf",
	}

	systemPackages := []string{}
	pythonModules := []string{}
	for _, name := range missing {
		if p, ok := systemPkgNames[name]; ok {
			systemPackages = append(systemPackages, p)
		}
		if p, ok := pythonModuleNames[name]; ok {
			pythonModules = append(pythonModules, p)
		}
	}
	if len(systemPackages) == 0 && len(pythonModules) == 0 {
		return nil, errors.New("no installable packages found")
	}

	commands := []string{}
	if len(systemPackages) > 0 {
		systemCommands, err := systemInstallCommandsForOS(systemPackages)
		if err != nil {
			return nil, err
		}
		commands = append(commands, systemCommands...)
	}

	if len(pythonModules) > 0 {
		pythonBin, err := findPythonBinary()
		if err != nil {
			return nil, fmt.Errorf("python required to install modules: %w", err)
		}
		commands = append(commands, pythonBin+" -m pip install --upgrade "+strings.Join(pythonModules, " "))
	}

	return commands, nil
}

func systemInstallCommandsForOS(packages []string) ([]string, error) {
	switch runtime.GOOS {
	case "darwin":
		if _, err := exec.LookPath("brew"); err != nil {
			return nil, errors.New("homebrew is required on macOS for auto-install")
		}
		return []string{"brew install " + strings.Join(packages, " ")}, nil
	case "linux":
		if _, err := exec.LookPath("apt-get"); err == nil {
			return []string{"sudo apt-get update", "sudo apt-get install -y " + strings.Join(packages, " ")}, nil
		}
		if _, err := exec.LookPath("dnf"); err == nil {
			return []string{"sudo dnf install -y " + strings.Join(packages, " ")}, nil
		}
		if _, err := exec.LookPath("pacman"); err == nil {
			return []string{"sudo pacman -S --noconfirm " + strings.Join(packages, " ")}, nil
		}
		return nil, errors.New("unsupported linux package manager for auto-install")
	case "windows":
		if _, err := exec.LookPath("choco"); err == nil {
			return []string{"choco install -y " + strings.Join(packages, " ")}, nil
		}
		if _, err := exec.LookPath("winget"); err == nil {
			cmds := []string{}
			for _, p := range packages {
				cmds = append(cmds, fmt.Sprintf("winget install --silent --accept-source-agreements --accept-package-agreements %s", p))
			}
			return cmds, nil
		}
		return nil, errors.New("choco or winget is required on windows for auto-install")
	default:
		return nil, fmt.Errorf("unsupported os for auto-install: %s", runtime.GOOS)
	}
}

func runShellCommand(ctx context.Context, command string) error {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "windows":
		cmd = exec.CommandContext(ctx, "cmd", "/C", command)
	default:
		cmd = exec.CommandContext(ctx, "sh", "-c", command)
	}
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	return cmd.Run()
}
