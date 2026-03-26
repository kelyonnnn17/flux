param(
    [switch]$Check
)

$ErrorActionPreference = "Stop"
Set-Location (Join-Path $PSScriptRoot "..")

$venvDir = if ($env:VENV_DIR) { $env:VENV_DIR } else { ".venv" }
$venvPython = Join-Path $venvDir "Scripts\python.exe"
$reqFile = "requirements/python-runtime.txt"

function Get-PythonExecutable {
    if (Get-Command python -ErrorAction SilentlyContinue) {
        return (Get-Command python).Source
    }
    if (Get-Command py -ErrorAction SilentlyContinue) {
        return (Get-Command py).Source
    }
    return $null
}

if (-not (Test-Path $venvDir)) {
    if ($Check) {
        throw "Missing $venvDir. Run: .\scripts\bootstrap-python.ps1"
    }

    $pythonExe = Get-PythonExecutable
    if (-not $pythonExe) {
        throw "python/py not found. Install Python 3.10+ and retry."
    }

    Write-Host "==> Creating virtual environment at $venvDir"
    if ($pythonExe -like "*\py.exe") {
        & $pythonExe -3 -m venv $venvDir
    } else {
        & $pythonExe -m venv $venvDir
    }
}

if (-not (Test-Path $venvPython)) {
    throw "Python executable not found at $venvPython"
}

if (-not $Check) {
    Write-Host "==> Installing Python runtime dependencies"
    & $venvPython -m pip install --upgrade pip
    & $venvPython -m pip install -r $reqFile
}

& $venvPython -c "import importlib; importlib.import_module('pdf2docx'); importlib.import_module('docx2pdf')"

Write-Host "OK Python runtime ready at $venvDir"
Write-Host "OK Optional: `$env:FLUX_PYTHON = '$((Resolve-Path $venvPython).Path)'"
