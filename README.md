# Flux Universal File Converter

A cross-platform CLI for converting files across formats using FFmpeg, ImageMagick, and Pandoc.

## Quick Start

```sh
./scripts/setup.sh      # install deps, build, verify (interactive)
./scripts/setup.sh -y   # same, install engines without prompting
make setup              # alias for ./scripts/setup.sh
```

## Building

```sh
make build   # builds the binary at ./bin/flux
./bin/flux --help
```

## Docker

Run without installing dependencies:

```sh
docker build -t flux .
docker run --rm flux convert -i input.json -o output.yaml
```

See [REQUIREMENTS.md](REQUIREMENTS.md) for full dependency list.

## Usage

Convert a file between formats:

```sh
flux convert -i input.jpg -o output.png
flux convert -i document.md -o document.pdf
flux convert -i video.mp4 -o video.mkv
```

Convert data files (CSV, JSON, YAML, TOML):

```sh
flux convert -i data.json -o data.yaml
flux convert -i sheet.csv -o sheet.json
flux convert -i config.toml -o config.yaml --from toml --to yaml
```

Batch conversion (multiple files or globs):

```sh
flux convert -i a.jpg b.jpg c.jpg -o png
flux convert -i *.json -o yaml --force
```

Pipe / stdin support:

```sh
cat file.json | flux convert -o - --from json --to yaml > out.yaml
```

Override the conversion engine (default is auto, which picks by file type):

```sh
flux convert -i file.pdf -o file.html --engine pandoc
flux convert -i image.png -o image.jpg --engine imagemagick
flux convert -i audio.mp3 -o audio.wav --engine ffmpeg
flux convert -i data.json -o data.csv --engine data
```

## Commands

- `flux convert` – convert files between formats
- `flux doctor` – check installed engines and versions
- `flux list-formats` – show supported conversions per engine
- `flux info <file>` – inspect file metadata

## Configuration

Optional config at `~/.config/flux/config.yaml`:

```yaml
log_level: info
```

## Requirements

At least one of: FFmpeg, ImageMagick, or Pandoc on `$PATH`. Install examples:

```sh
# macOS
brew install ffmpeg imagemagick pandoc

# Ubuntu/Debian
apt install ffmpeg imagemagick pandoc

# Windows
choco install ffmpeg imagemagick pandoc
```
