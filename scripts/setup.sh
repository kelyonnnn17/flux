#!/usr/bin/env bash
# Flux setup: install dependencies, build, verify
# Usage: ./scripts/setup.sh [--yes]   # --yes = install engines without prompting
set -e

cd "$(dirname "$0")/.."

INSTALL_ENGINES=false
[[ "${1:-}" == "--yes" || "${1:-}" == "-y" ]] && INSTALL_ENGINES=true
PATH_UPDATED=false

echo "==> Flux setup"
echo

detect_shell_profile() {
    local shell_name
    shell_name="$(basename "${SHELL:-}")"
    case "$shell_name" in
        zsh) echo "$HOME/.zshrc" ;;
        bash) echo "$HOME/.bashrc" ;;
        *) echo "$HOME/.profile" ;;
    esac
}

is_in_path() {
    local dir="$1"
    case ":$PATH:" in
        *":$dir:"*) return 0 ;;
        *) return 1 ;;
    esac
}

ensure_path_contains() {
    local dir="$1"
    local profile
    profile="$(detect_shell_profile)"

    if is_in_path "$dir"; then
        return 0
    fi

    if [[ ! -f "$profile" ]]; then
        touch "$profile"
    fi

    if ! grep -Fq "# flux-path" "$profile"; then
        {
            echo
            echo "# flux-path"
            echo "export PATH=\"$dir:\$PATH\""
        } >>"$profile"
        echo "OK Added $dir to PATH in $profile"
        PATH_UPDATED=true
    fi

    export PATH="$dir:$PATH"
}

install_flux() {
    local install_dir=""
    local preferred_dirs=()

    case "$(uname -s)" in
        Darwin)
            preferred_dirs=("/opt/homebrew/bin" "/usr/local/bin" "$HOME/.local/bin")
            ;;
        Linux)
            preferred_dirs=("/usr/local/bin" "$HOME/.local/bin")
            ;;
        *)
            preferred_dirs=("$HOME/.local/bin")
            ;;
    esac

    # Prefer a writable directory that is already on PATH so flux works immediately.
    for d in "${preferred_dirs[@]}"; do
        if [[ -d "$d" && -w "$d" ]] && is_in_path "$d"; then
            install_dir="$d"
            break
        fi
    done

    for d in "${preferred_dirs[@]}"; do
        [[ -n "$install_dir" ]] && break
        if [[ -d "$d" && -w "$d" ]]; then
            install_dir="$d"
            break
        fi
    done

    if [[ -z "$install_dir" ]]; then
        if [[ ! -d "$HOME/.local/bin" ]]; then
            mkdir -p "$HOME/.local/bin"
        fi
        install_dir="$HOME/.local/bin"
    fi

    echo "==> Installing flux to $install_dir"
    make install BINDIR="$install_dir"
    ensure_path_contains "$install_dir"
}

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

install_flux

echo "==> Verifying installed command"
flux --help >/dev/null
echo "OK flux command is available"
if [[ "$PATH_UPDATED" == true ]]; then
    echo "OK PATH was updated for future shells. Open a new terminal to pick up that change."
fi
echo

echo "==> Done. Run: flux --help"
