#!/usr/bin/env bash
# Flux setup: install dependencies, build, verify
# Usage: ./scripts/setup.sh [--yes]   # --yes = install engines without prompting
set -e

cd "$(dirname "$0")/.."

INSTALL_ENGINES=false
[[ "${1:-}" == "--yes" || "${1:-}" == "-y" ]] && INSTALL_ENGINES=true

echo "==> Flux setup"
echo

# Check Go
if ! command -v go &>/dev/null; then
    echo "X Go not found. Install Go 1.22+: https://go.dev/dl/"
    exit 1
fi
echo "OK Go: $(go version)"

# Detect OS and install engines
install_engines() {
    case "$(uname -s)" in
        Darwin)
            if command -v brew &>/dev/null; then
                echo "==> Installing engines (macOS Homebrew)..."
                brew install ffmpeg imagemagick pandoc 2>/dev/null || true
            else
                echo "  Install Homebrew and run: brew install ffmpeg imagemagick pandoc"
            fi
            ;;
        Linux)
            if command -v apt-get &>/dev/null; then
                echo "==> Installing engines (apt)..."
                sudo apt-get update -qq
                sudo apt-get install -y ffmpeg imagemagick pandoc 2>/dev/null || true
            elif command -v dnf &>/dev/null; then
                echo "==> Installing engines (dnf)..."
                sudo dnf install -y ffmpeg ImageMagick pandoc 2>/dev/null || true
            elif command -v pacman &>/dev/null; then
                echo "==> Installing engines (pacman)..."
                sudo pacman -S --noconfirm ffmpeg imagemagick pandoc 2>/dev/null || true
            else
                echo "  Install manually: ffmpeg, imagemagick, pandoc"
            fi
            ;;
        *)
            echo "  Install manually: ffmpeg, imagemagick, pandoc"
            ;;
    esac
}

# Install engines (may need sudo)
if [[ "$INSTALL_ENGINES" == true ]]; then
    install_engines
else
    read -p "Install FFmpeg, ImageMagick, Pandoc? [y/N] " -n 1 -r
    echo
    [[ $REPLY =~ ^[Yy]$ ]] && install_engines
fi

# Build
echo "==> Building flux..."
make build
echo "OK Built ./bin/flux"
echo

# Verify
echo "==> Checking engines"
./bin/flux doctor
echo

echo "==> Done. Run: ./bin/flux --help"
