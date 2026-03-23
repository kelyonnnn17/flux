# Flux Requirements

## Build Requirements

| Tool   | Version   | Purpose                          |
|--------|-----------|----------------------------------|
| Go     | 1.22+     | Build the CLI                    |
| make   | (any)     | Run build targets                |

## Runtime Dependencies (Engines)

At least **one** engine is required. The `data` engine (CSV/JSON/YAML/TOML) is built-in and always available.

| Engine      | Binary         | Purpose                    |
|-------------|----------------|----------------------------|
| FFmpeg      | `ffmpeg`       | Audio, video               |
| ImageMagick | `magick`/`convert` | Images                 |
| Pandoc      | `pandoc`       | Documents                  |
| Data        | built-in       | CSV, JSON, YAML, TOML      |

## Install by Platform

### macOS (Homebrew)

```sh
brew install go ffmpeg imagemagick pandoc
```

### Ubuntu / Debian

```sh
sudo apt update
sudo apt install -y golang-go ffmpeg imagemagick pandoc
```

### Fedora / RHEL

```sh
sudo dnf install -y golang ffmpeg ImageMagick pandoc
```

### Arch Linux

```sh
sudo pacman -S go ffmpeg imagemagick pandoc
```

### Windows (Chocolatey)

```powershell
choco install golang ffmpeg imagemagick pandoc
```

## Verify

After installing, run:

```sh
go version
make build
make install
flux --help
flux doctor
```

`flux doctor` shows which engines are available.
