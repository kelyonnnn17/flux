#!/usr/bin/env bash
set -euo pipefail

cd "$(dirname "$0")/.."

MODE="apply"
if [[ "${1:-}" == "--check" ]]; then
  MODE="check"
fi

VENV_DIR="${VENV_DIR:-.venv}"
VENV_PYTHON="$VENV_DIR/bin/python"
REQ_FILE="requirements/python-runtime.txt"

find_python() {
  if command -v python3 >/dev/null 2>&1; then
    command -v python3
    return 0
  fi
  if command -v python >/dev/null 2>&1; then
    command -v python
    return 0
  fi
  return 1
}

check_modules() {
  "$VENV_PYTHON" - <<'PY'
import importlib
missing = []
for name in ("pdf2docx", "docx2pdf"):
    try:
        importlib.import_module(name)
    except Exception:
        missing.append(name)
if missing:
    raise SystemExit("missing python modules: " + ", ".join(missing))
print("OK Python modules installed: pdf2docx, docx2pdf")
PY
}

if [[ ! -d "$VENV_DIR" ]]; then
  if [[ "$MODE" == "check" ]]; then
    echo "X Missing $VENV_DIR. Run: make bootstrap"
    exit 1
  fi

  PY_BIN="$(find_python || true)"
  if [[ -z "$PY_BIN" ]]; then
    echo "X python3/python not found. Install Python 3.10+ and retry."
    exit 1
  fi

  echo "==> Creating virtual environment at $VENV_DIR"
  "$PY_BIN" -m venv "$VENV_DIR"
fi

if [[ ! -x "$VENV_PYTHON" ]]; then
  echo "X Python executable not found at $VENV_PYTHON"
  exit 1
fi

if [[ "$MODE" == "apply" ]]; then
  echo "==> Installing Python runtime dependencies"
  "$VENV_PYTHON" -m pip install --upgrade pip
  "$VENV_PYTHON" -m pip install -r "$REQ_FILE"
fi

check_modules

echo "OK Python runtime ready at $VENV_DIR"
echo "OK Optional: export FLUX_PYTHON=$(pwd)/$VENV_PYTHON"
