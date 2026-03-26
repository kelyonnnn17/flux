#!/usr/bin/env bash
# Flux setup: install dependencies, build, verify
# Usage: ./scripts/setup.sh [--yes]   # --yes = install engines without prompting
set -e

cd "$(dirname "$0")/.."

INSTALL_ENGINES=false
[[ "${1:-}" == "--yes" || "${1:-}" == "-y" ]] && INSTALL_ENGINES=true
PATH_UPDATED=false
FLUX_PYTHON_UPDATED=false

have_imagemagick() {
    command -v magick &>/dev/null || command -v convert &>/dev/null
}

have_all_engines() {
    command -v ffmpeg &>/dev/null && \
        have_imagemagick && \
        command -v pandoc &>/dev/null && \
        command -v pdftotext &>/dev/null
}

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

ensure_flux_python() {
    local profile
    local flux_python
    profile="$(detect_shell_profile)"
    flux_python="$(pwd)/.venv/bin/python"

    if [[ ! -f "$profile" ]]; then
        touch "$profile"
    fi

    if grep -Fq "# flux-python" "$profile"; then
        if ! grep -Fq "export FLUX_PYTHON=\"$flux_python\"" "$profile"; then
            # Replace previous FLUX_PYTHON line if marker exists.
            awk -v new_line="export FLUX_PYTHON=\"$flux_python\"" '
                BEGIN { marker = 0 }
                {
                    if ($0 == "# flux-python") {
                        print $0
                        getline
                        print new_line
                        marker = 1
                        next
                    }
                    print $0
                }
                END {
                    if (marker == 0) {
                        print "# flux-python"
                        print new_line
                    }
                }
            ' "$profile" > "$profile.tmp"
            mv "$profile.tmp" "$profile"
            FLUX_PYTHON_UPDATED=true
        fi
    else
        {
            echo
            echo "# flux-python"
            echo "export FLUX_PYTHON=\"$flux_python\""
        } >>"$profile"
        FLUX_PYTHON_UPDATED=true
    fi

    export FLUX_PYTHON="$flux_python"
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
    if have_all_engines; then
        echo "OK Runtime engines already installed (ffmpeg, imagemagick, pandoc, pdftotext)"
        return
    fi

    case "$(uname -s)" in
        Darwin)
            if command -v brew &>/dev/null; then
                echo "==> Installing engines (macOS Homebrew)..."
                local packages=()
                command -v ffmpeg &>/dev/null || packages+=("ffmpeg")
                have_imagemagick || packages+=("imagemagick")
                command -v pandoc &>/dev/null || packages+=("pandoc")
                command -v pdftotext &>/dev/null || packages+=("poppler")

                if [[ ${#packages[@]} -eq 0 ]]; then
                    echo "OK Runtime engines already installed"
                    return
                fi

                HOMEBREW_NO_AUTO_UPDATE=1 HOMEBREW_NO_INSTALL_CLEANUP=1 brew install "${packages[@]}" || true
            else
                echo "  Install Homebrew and run: brew install ffmpeg imagemagick pandoc poppler"
            fi
            ;;
        Linux)
            if command -v apt-get &>/dev/null; then
                echo "==> Installing engines (apt)..."
                sudo apt-get update -qq
                sudo apt-get install -y ffmpeg imagemagick pandoc poppler-utils 2>/dev/null || true
            elif command -v dnf &>/dev/null; then
                echo "==> Installing engines (dnf)..."
                sudo dnf install -y ffmpeg ImageMagick pandoc poppler-utils 2>/dev/null || true
            elif command -v pacman &>/dev/null; then
                echo "==> Installing engines (pacman)..."
                sudo pacman -S --noconfirm ffmpeg imagemagick pandoc poppler 2>/dev/null || true
            else
                echo "  Install manually: ffmpeg, imagemagick, pandoc, pdftotext (poppler)"
            fi
            ;;
        *)
            echo "  Install manually: ffmpeg, imagemagick, pandoc, pdftotext (poppler)"
            ;;
    esac
}

# Install engines (may need sudo)
if [[ "$INSTALL_ENGINES" == true ]]; then
    install_engines
else
    read -p "Install FFmpeg, ImageMagick, Pandoc, Poppler? [y/N] " -n 1 -r
    echo
    [[ $REPLY =~ ^[Yy]$ ]] && install_engines
fi

# Build
echo "==> Preparing Python runtime (.venv + modules)..."
bash ./scripts/bootstrap-python.sh
ensure_flux_python

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
if command -v flux >/dev/null 2>&1; then
    echo "OK flux command is available"
else
    echo "X flux command is not on PATH yet. Open a new terminal and retry."
    exit 1
fi
if [[ "$PATH_UPDATED" == true ]]; then
    echo "OK PATH was updated for future shells. Open a new terminal to pick up that change."
fi
if [[ "$FLUX_PYTHON_UPDATED" == true ]]; then
    echo "OK FLUX_PYTHON was set in your shell profile for future shells."
fi
echo

echo "==> Done. Run: flux --help"
