# Flux Requirements

## Build Requirements

| Tool | Version | Purpose |
|------|---------|---------|
| Go | 1.22+ | Build Flux CLI and library |
| make | any | Build/install targets on Unix-like systems |

## Runtime Dependencies

Flux can run with one or more external engines plus its built-in data engine.

| Engine | Binary | Required For |
|--------|--------|--------------|
| FFmpeg | `ffmpeg` | Audio/video conversion |
| ImageMagick | `magick` or `convert` | Image conversion |
| Pandoc | `pandoc` | Document conversion |
| PDF to DOCX (python) | `pdf2docx` module | High-fidelity `pdf -> docx` |
| DOCX to PDF (python) | `docx2pdf` module | High-fidelity `docx -> pdf` |
| PDF text extraction | `pdftotext` | Best-effort PDF input conversion |
| Data | built-in | CSV/JSON/YAML/TOML conversion |

Notes:
1. Flux prefers direct conversions, then best-effort multi-step routes.
2. `--engine <name>` is strict and must support the full route.
3. Flux prefers `pdf2docx`/`docx2pdf` for direct PDF<->DOCX conversion.
4. If python modules are missing, PDF input can fall back to text extraction and may lose visual fidelity.

## Platform Setup

### One-command global install (recommended)

If Go is already installed, use these installer scripts to make `flux` available globally from any directory:

macOS/Linux:

```sh
./scripts/install-go.sh
```

Windows PowerShell:

```powershell
powershell -ExecutionPolicy Bypass -File .\scripts\install-go.ps1
```

These scripts run `go install github.com/kelyonnnn17/flux@latest` and ensure Go's bin directory is in your user PATH.

### macOS

Manual install:

```sh
brew install go ffmpeg imagemagick pandoc poppler
```

Project setup:

```sh
./scripts/setup.sh
./scripts/setup.sh -y
```

### Linux

Ubuntu/Debian:

```sh
sudo apt update
sudo apt install -y golang-go ffmpeg imagemagick pandoc poppler-utils
```

Fedora/RHEL:

```sh
sudo dnf install -y golang ffmpeg ImageMagick pandoc poppler-utils
```

Arch:

```sh
sudo pacman -S --noconfirm go ffmpeg imagemagick pandoc poppler
```

Project setup:

```sh
./scripts/setup.sh
./scripts/setup.sh -y
```

### Windows

Chocolatey:

```powershell
choco install golang ffmpeg imagemagick pandoc
```

or winget:

```powershell
winget install --silent --accept-source-agreements --accept-package-agreements GoLang.Go
winget install --silent --accept-source-agreements --accept-package-agreements ffmpeg
winget install --silent --accept-source-agreements --accept-package-agreements ImageMagick.ImageMagick
winget install --silent --accept-source-agreements --accept-package-agreements JohnMacFarlane.Pandoc
winget install --silent --accept-source-agreements --accept-package-agreements oschwartz10612.Poppler
```

Project setup:

```powershell
powershell -ExecutionPolicy Bypass -File .\scripts\setup.ps1
powershell -ExecutionPolicy Bypass -File .\scripts\setup.ps1 -Yes
```

## Install Targets and PATH

Default `make install` target path:

1. Apple Silicon macOS: `/opt/homebrew/bin`
2. Others: `/usr/local/bin`

Override when needed:

```sh
make install BINDIR=$HOME/.local/bin
```

## Use as Go Library

Install package:

```sh
go get github.com/kelyonnnn17/flux/pkg/flux
```

The library exposes:

1. `Convert(src, dst, ConvertOptions)`
2. `CheckDependencies()`
3. `EnsureDependencies(ctx, autoInstall)`

`EnsureDependencies(..., true)` tries to auto-install missing dependencies with:

1. macOS: Homebrew
2. Linux: apt/dnf/pacman
3. Windows: choco/winget

Auto-install requires appropriate privileges (for example sudo/admin shell).

## Verification Checklist

CLI verification:

```sh
go version
make bootstrap
make bootstrap-check
make build
make install
flux --help
flux doctor
flux --no-ui doctor
```

Windows CLI verification:

```powershell
go version
go build -o .\bin\flux.exe main.go
.\bin\flux.exe doctor
```

Library verification (quick):

```sh
go test ./...
go test -tags=integration ./tests/integration/...
```
