# Flux Universal File Converter

Flux is a cross-platform converter for documents, images, audio/video, and structured data.

- Documents: Pandoc
- Images: ImageMagick
- Audio/Video: FFmpeg
- Data (CSV/JSON/YAML/TOML): built-in Go engine

## Quick Start (CLI)

### Global Install (like pip)

Use this when you want `flux` available from any directory after one install command.

macOS/Linux:

```sh
./scripts/install-go.sh
```

Windows PowerShell:

```powershell
powershell -ExecutionPolicy Bypass -File .\scripts\install-go.ps1
```

This installs via `go install github.com/kelyonnnn17/flux@latest` and updates your user PATH to include Go's binary directory if needed.

### macOS and Linux

```sh
./scripts/setup.sh
./scripts/setup.sh -y
make setup
flux --help
```

### Windows (PowerShell)

```powershell
powershell -ExecutionPolicy Bypass -File .\scripts\setup.ps1
powershell -ExecutionPolicy Bypass -File .\scripts\setup.ps1 -Yes
flux --help
```

What setup does:

1. Installs runtime engines (optional prompt or forced with `-y` / `-Yes`)
2. Builds Flux
3. Installs Flux into a PATH location
4. Runs verification (`flux doctor`)

If you already have Go and only need a global install, prefer the `install-go` scripts above.

## Build and Install

```sh
make build
make install
```

Install path defaults:

1. Apple Silicon macOS: `/opt/homebrew/bin`
2. Other systems: `/usr/local/bin`

Override install directory:

```sh
make install BINDIR=$HOME/.local/bin
```

## Usage

Shortest syntax:

```sh
flux input.jpg png
flux docs/guide.md pdf
flux media/clip.mov mp4
```

Explicit command syntax:

```sh
flux convert input.jpg png
flux convert document.md pdf
flux convert video.mp4 mkv
```

Data conversion:

```sh
flux convert -i data.json -o data.yaml
flux convert -i sheet.csv -o sheet.json
flux convert -i config.toml -o config.yaml --from toml --to yaml
```

Batch conversion:

```sh
flux convert -i a.jpg b.jpg c.jpg -o png
flux convert -i *.json -o yaml --force
```

Pipe/stdin:

```sh
cat file.json | flux convert -o - --from json --to yaml > out.yaml
```

Engine override:

```sh
flux convert -i file.pdf -o file.html --engine pandoc
flux convert -i image.png -o image.jpg --engine imagemagick
flux convert -i audio.mp3 -o audio.wav --engine ffmpeg
flux convert -i data.json -o data.csv --engine data
```

Document formatting presets:

```sh
flux convert -i notes.md -o notes.html --engine pandoc --format-style professional
flux convert -i notes.md -o notes.pdf --engine pandoc --format-style technical
flux convert -i notes.md -o notes.html --engine pandoc --format-style none
```

## Commands

1. `flux convert` - Convert files
2. `flux doctor` - Check engine availability and versions
3. `flux list-formats` - Show format coverage per engine
4. `flux info <file>` - Inspect metadata

Aliases:

```sh
flux c file.md pdf
flux d
flux lf
flux i file.pdf
```

## Use as a Go Library

Flux now exposes a public library package:

```sh
go get github.com/kelyonnnn17/flux/pkg/flux
```

Library example with auto dependency installation:

```go
package main

import (
	"context"
	"log"

	fluxlib "github.com/kelyonnnn17/flux/pkg/flux"
)

func main() {
	ctx := context.Background()

	// Try to auto-install missing runtime engines using brew/apt/dnf/pacman/choco/winget.
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

Note: auto-install still depends on OS package-manager permissions (for example `sudo` on Linux or admin shell on Windows).

## Terminal UX

Animated terminal UI is enabled for interactive commands.

- Disable animations: `--no-ui`
- Disable colors: `--no-color` or `NO_COLOR=1`

Examples:

```sh
flux doctor
flux --no-ui doctor
NO_COLOR=1 flux doctor
```

## Docker

```sh
docker build -t flux .
docker run --rm flux convert -i input.json -o output.yaml
```

## More Details

See [REQUIREMENTS.md](REQUIREMENTS.md) for complete dependency, platform, and verification guidance.
