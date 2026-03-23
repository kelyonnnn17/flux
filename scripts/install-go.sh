#!/usr/bin/env bash
set -euo pipefail

MODULE="github.com/kelyonnnn17/flux"
PROFILE=""

if ! command -v go >/dev/null 2>&1; then
  echo "X Go is required for this installer. Install Go from https://go.dev/dl/"
  exit 1
fi

echo "==> Installing flux via go install"
go install "${MODULE}@latest"

GOBIN="$(go env GOBIN)"
if [[ -z "${GOBIN}" ]]; then
  GOPATH_VAL="$(go env GOPATH)"
  if [[ -z "${GOPATH_VAL}" ]]; then
    GOPATH_VAL="$HOME/go"
  fi
  GOBIN="${GOPATH_VAL}/bin"
fi

if [[ ! -x "${GOBIN}/flux" ]]; then
  echo "X Install completed but ${GOBIN}/flux was not found"
  exit 1
fi

case "$(basename "${SHELL:-}")" in
  zsh) PROFILE="$HOME/.zshrc" ;;
  bash) PROFILE="$HOME/.bashrc" ;;
  fish) PROFILE="$HOME/.config/fish/config.fish" ;;
  *) PROFILE="$HOME/.profile" ;;
esac

if [[ ":$PATH:" != *":${GOBIN}:"* ]]; then
  mkdir -p "$(dirname "$PROFILE")"
  touch "$PROFILE"
  if ! grep -Fq "# flux-gobin" "$PROFILE"; then
    {
      echo
      echo "# flux-gobin"
      echo "export PATH=\"${GOBIN}:\$PATH\""
    } >>"$PROFILE"
  fi
  export PATH="${GOBIN}:$PATH"
  echo "OK Added ${GOBIN} to PATH in ${PROFILE}"
  echo "OK Open a new terminal (or source ${PROFILE})"
fi

echo "OK Installed: ${GOBIN}/flux"
flux --help >/dev/null 2>&1 && echo "OK flux is ready" || true
