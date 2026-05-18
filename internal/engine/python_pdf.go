package engine

import (
	"context"
	_ "embed"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

//go:embed scripts/pdf2docx_convert.py
var pdf2docxScript string

//go:embed scripts/docx2pdf_convert.py
var docx2pdfScript string

// PythonPDFAdapter bridges PDF<->DOCX conversions via Python libraries.
type PythonPDFAdapter struct {
	Runner CmdRunner
	Mode   string // pdf2docx or docx2pdf
}

func (a *PythonPDFAdapter) Convert(src, dst string, args []string) error {
	switch a.Mode {
	case "pdf2docx":
		return convertPDFToDOCX(a.Runner, src, dst)
	case "docx2pdf":
		return convertDOCXToPDF(a.Runner, src, dst)
	default:
		return fmt.Errorf("unknown python conversion mode: %s", a.Mode)
	}
}

func convertPDFToDOCX(runner CmdRunner, srcPDF, dstDOCX string) error {
	py, err := resolvePythonExecutableForModule("pdf2docx")
	if err != nil {
		return pythonModuleMissingError("pdf2docx")
	}

	ctx, cancel := engineContext()
	defer cancel()

	cmd := runner.CommandContext(ctx, py, "-c", pdf2docxScript, srcPDF, dstDOCX)
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("pdf2docx error: %w, output: %s", err, string(out))
	}
	return validateOutputFile(dstDOCX, "docx")
}

func convertDOCXToPDF(runner CmdRunner, srcDOCX, dstPDF string) error {
	py, err := resolvePythonExecutableForModule("docx2pdf")
	if err != nil {
		return pythonModuleMissingError("docx2pdf")
	}

	ctx, cancel := engineContext()
	defer cancel()

	// Use a dedicated temp dir so we can find the output without assuming its name.
	tmpDir, err := os.MkdirTemp("", "flux-docx2pdf-*")
	if err != nil {
		return fmt.Errorf("create temp dir for docx2pdf: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	cmd := runner.CommandContext(ctx, py, "-c", docx2pdfScript, srcDOCX, tmpDir)
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("docx2pdf error: %w, output: %s", err, string(out))
	}

	// Find whichever PDF the library produced — don't assume its name.
	produced, err := filepath.Glob(filepath.Join(tmpDir, "*.pdf"))
	if err != nil || len(produced) == 0 {
		return fmt.Errorf("docx2pdf did not produce a PDF in temp dir: %s", tmpDir)
	}
	if err := validateOutputFile(produced[0], "pdf"); err != nil {
		return fmt.Errorf("docx2pdf produced an invalid PDF: %w", err)
	}

	if err := os.MkdirAll(filepath.Dir(dstPDF), 0755); err != nil {
		return fmt.Errorf("create output directory: %w", err)
	}
	_ = os.Remove(dstPDF)
	if err := moveFile(produced[0], dstPDF); err != nil {
		return fmt.Errorf("move docx2pdf output: %w", err)
	}
	return nil
}

// moveFile moves src to dst, falling back to copy+delete if rename fails across filesystems.
func moveFile(src, dst string) error {
	if err := os.Rename(src, dst); err == nil {
		return nil
	}
	// Fallback: copy + fsync + remove (handles cross-filesystem moves).
	srcFile, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("open source: %w", err)
	}
	defer srcFile.Close()
	dstFile, err := os.Create(dst)
	if err != nil {
		return fmt.Errorf("create destination: %w", err)
	}
	if _, err := io.Copy(dstFile, srcFile); err != nil {
		dstFile.Close()
		os.Remove(dst)
		return fmt.Errorf("copy: %w", err)
	}
	if err := dstFile.Sync(); err != nil {
		dstFile.Close()
		os.Remove(dst)
		return fmt.Errorf("fsync: %w", err)
	}
	if err := dstFile.Close(); err != nil {
		os.Remove(dst)
		return fmt.Errorf("close: %w", err)
	}
	os.Remove(src)
	return nil
}

// pythonModuleMissingError returns an actionable error when a Python module is absent,
// pointing users to the project setup script first (preferred over bare pip).
func pythonModuleMissingError(module string) error {
	venvPython := ""
	if cwd, err := os.Getwd(); err == nil {
		candidate := filepath.Join(cwd, ".venv", "bin", "python")
		if _, err := os.Stat(candidate); err == nil {
			venvPython = candidate
		}
	}
	if venvPython != "" {
		return fmt.Errorf(
			"python module %s not found in .venv.\n"+
				"  Run: make bootstrap   (or: %s -m pip install %s)\n"+
				"  Then verify with: flux doctor",
			module, venvPython, module,
		)
	}
	return fmt.Errorf(
		"python module %s not found.\n"+
			"  Recommended: run 'make setup' or 'make bootstrap' to create .venv and install dependencies.\n"+
			"  Alternatively: pip install %s\n"+
			"  Set FLUX_PYTHON to point to the right interpreter if needed.\n"+
			"  Verify with: flux doctor",
		module, module,
	)
}

func resolvePythonExecutable() (string, error) {
	for _, candidate := range candidatePythonExecutables() {
		if candidate != "" {
			return candidate, nil
		}
	}
	return "", fmt.Errorf("python3/python not found in PATH")
}

func pythonModuleAvailable(module string) bool {
	_, err := resolvePythonExecutableForModule(module)
	return err == nil
}

func resolvePythonExecutableForModule(module string) (string, error) {
	for _, py := range candidatePythonExecutables() {
		if moduleImportableWithPython(py, module) {
			return py, nil
		}
	}
	return "", fmt.Errorf("python module %s not available", module)
}

// moduleCheckCache caches per-process results of Python module availability checks.
var moduleCheckCache sync.Map // key: "pythonPath:module" → bool

func moduleImportableWithPython(pythonPath, module string) bool {
	if pythonPath == "" {
		return false
	}
	cacheKey := pythonPath + ":" + module
	if v, ok := moduleCheckCache.Load(cacheKey); ok {
		return v.(bool)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, pythonPath, "-c", "import sys; __import__(sys.argv[1])", module)
	result := cmd.Run() == nil
	moduleCheckCache.Store(cacheKey, result)
	return result
}

func candidatePythonExecutables() []string {
	candidates := make([]string, 0, 5)
	seen := map[string]bool{}
	add := func(path string) {
		if path == "" || seen[path] {
			return
		}
		if _, err := os.Stat(path); err == nil {
			seen[path] = true
			candidates = append(candidates, path)
		}
	}

	if explicit := strings.TrimSpace(os.Getenv("FLUX_PYTHON")); explicit != "" {
		add(explicit)
	}

	if venv := strings.TrimSpace(os.Getenv("VIRTUAL_ENV")); venv != "" {
		add(filepath.Join(venv, "bin", "python"))
	}

	if cwd, err := os.Getwd(); err == nil {
		add(filepath.Join(cwd, ".venv", "bin", "python"))
	}

	if p, err := exec.LookPath("python3"); err == nil {
		add(p)
	}
	if p, err := exec.LookPath("python"); err == nil {
		add(p)
	}

	return candidates
}
