# Flux

Flux is a cross-platform CLI for converting files across formats with a practical engine-first pipeline:

- Documents: Pandoc
- Images: ImageMagick
- Audio/Video: FFmpeg
- Structured data (CSV/JSON/YAML/TOML): native Go engine
- High-fidelity PDF/DOCX: Python adapters (`pdf2docx`, `docx2pdf`)

Flux is designed for developer workflows: simple commands, predictable output, and explicit fallback behavior.

## Features

- One-command setup on macOS/Linux and Windows
- Global install option (`go install`)
- Smart route planning (`--engine auto`) with direct conversions preferred
- Explicit engine forcing (`--engine pandoc`, `--engine ffmpeg`, etc.)
- Batch conversion and glob support
- Stdin/stdout pipe support for data workflows
- DOCX style preservation via `--reference-doc`
- Developer-friendly markdown formatting profile (`--format-style developer`)
- Built-in diagnostics via `flux doctor`
- Public Go library package at `pkg/flux`

## Quick Start

### Option A: One-command local setup (recommended)

macOS/Linux:

```sh
make setup
flux --help
```

Windows PowerShell:

```powershell
powershell -ExecutionPolicy Bypass -File .\scripts\setup.ps1 -Yes
flux --help
```

Setup performs:

1. Runtime dependency install (non-interactive)
2. Python `.venv` bootstrap
3. Python module install (`pdf2docx`, `docx2pdf`)
4. `FLUX_PYTHON` persistence
5. Build + install
6. Health verification with `flux doctor`

### Option B: Global install (Go toolchain path)

macOS/Linux:

```sh
./scripts/install-go.sh
```

Windows PowerShell:

```powershell
powershell -ExecutionPolicy Bypass -File .\scripts\install-go.ps1
```

This installs `github.com/kelyonnnn17/flux` via `go install` and updates user PATH if required.

## Core Commands

- `flux convert` - convert files
- `flux doctor` - check engine availability and versions
- `flux list-formats` - show format support by engine
- `flux info <file>` - inspect file metadata

Aliases:

- `flux c` -> `flux convert`
- `flux d` -> `flux doctor`
- `flux lf` -> `flux list-formats`
- `flux i` -> `flux info`

## Usage

### Short syntax

```sh
flux input.jpg png
flux docs/guide.md pdf
flux media/clip.mov mp4
```

### Explicit syntax

```sh
flux convert input.jpg png
flux convert notes.md pdf
flux convert input.mp4 webm
```

### Data conversion

```sh
flux convert -i data.json -o data.yaml
flux convert -i sheet.csv -o sheet.json
flux convert -i config.toml -o config.yaml --from toml --to yaml
```

### Batch conversion

```sh
flux convert -i a.jpg b.jpg c.jpg -o png
flux convert -i *.json -o yaml --force
```

### Pipe support (stdin/stdout)

```sh
cat file.json | flux convert -o - --from json --to yaml > out.yaml
```

## Engine Selection

Default behavior uses route planning:

```sh
flux convert -i input.docx -o output.pdf --engine auto
```

Force a specific engine:

```sh
flux convert -i file.pdf -o file.docx --engine pdf2docx
flux convert -i file.docx -o file.pdf --engine docx2pdf
flux convert -i file.md -o file.pdf --engine pandoc
flux convert -i file.png -o file.jpg --engine imagemagick
flux convert -i file.mp4 -o file.webm --engine ffmpeg
flux convert -i file.json -o file.csv --engine data
```

Important behavior:

- `--engine auto` prefers direct, high-fidelity routes.
- Forced engines are strict. If forced engine cannot complete the route, Flux returns an actionable error.
- `pdf -> docx` in auto mode expects Python adapter availability for best fidelity.
- `docx -> pdf` in auto mode prefers `docx2pdf` and can fallback when runtime issues occur.

## PDF <-> DOCX Pipeline

Flux prioritizes Python adapters for document fidelity:

- `pdf2docx` for `pdf -> docx`
- `docx2pdf` for `docx -> pdf`

If modules are missing, fallback paths can be used in some cases, but may lose layout fidelity for complex PDFs.

Install modules manually (if needed):

```sh
pip install pdf2docx docx2pdf
```

## Document Formatting

Flux preserves source structure by default and avoids synthetic section generation.

### Style presets

- `professional` (default)
- `technical`
- `developer`
- `none`

Examples:

```sh
flux convert -i notes.md -o notes.pdf --engine pandoc --format-style professional
flux convert -i notes.md -o notes.pdf --engine pandoc --format-style technical
flux convert -i notes.md -o notes.pdf --engine pandoc --format-style developer
flux convert -i notes.md -o notes.html --engine pandoc --format-style none
```

### DOCX style preservation and templates

Use DOCX templates for consistent Word styling:

```sh
flux convert -i notes.md -o notes.docx --reference-doc template.docx --format-style professional
```

If input and output are both DOCX, Flux can reuse input styling as reference automatically.

## Build and Install

```sh
make build
make install
```

Default install targets:

1. Apple Silicon macOS: `/opt/homebrew/bin`
2. Other systems: `/usr/local/bin`

Override destination:

```sh
make install BINDIR=$HOME/.local/bin
```

## Developer Setup

Python runtime bootstrap helpers:

```sh
make bootstrap
make bootstrap-check
make dev-ready
```

`make install` enforces bootstrap checks so local installs remain conversion-ready.

## Terminal UX

- Disable animated UI: `--no-ui`
- Disable ANSI colors: `--no-color` or `NO_COLOR=1`

Examples:

```sh
flux doctor
flux --no-ui doctor
NO_COLOR=1 flux doctor
```

## Go Library

Install:

```sh
go get github.com/kelyonnnn17/flux/pkg/flux
```

Example:

```go
package main

import (
	"context"
	"log"

	fluxlib "github.com/kelyonnnn17/flux/pkg/flux"
)

func main() {
	ctx := context.Background()

	if _, err := fluxlib.EnsureDependencies(ctx, true); err != nil {
		log.Fatalf("dependency setup failed: %v", err)
	}

	err := fluxlib.Convert("README.md", "README.html", fluxlib.ConvertOptions{
		Engine:      "auto",
		FormatStyle: "professional",
	})
	if err != nil {
		log.Fatalf("convert failed: %v", err)
	}
}
```

## Docker

```sh
docker build -t flux .
docker run --rm flux convert -i input.json -o output.yaml
```

## Requirements and Compatibility

See [REQUIREMENTS.md](REQUIREMENTS.md) for full dependency matrix and platform-specific verification details.

## License

MIT License. See [LICENSE](LICENSE).
