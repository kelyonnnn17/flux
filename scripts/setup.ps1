param(
    [switch]$Yes
)

$ErrorActionPreference = "Stop"
Set-Location (Join-Path $PSScriptRoot "..")

Write-Host "==> Flux setup (Windows)"

function Ensure-Command($Name, $InstallHint) {
    if (-not (Get-Command $Name -ErrorAction SilentlyContinue)) {
        throw "$Name not found. $InstallHint"
    }
}

Ensure-Command "go" "Install Go: https://go.dev/dl/"
Write-Host "OK Go: $(go version)"

function Install-Engines {
    if (Get-Command choco -ErrorAction SilentlyContinue) {
        Write-Host "==> Installing engines with Chocolatey"
        choco install -y ffmpeg imagemagick pandoc poppler
        return
    }

    if (Get-Command winget -ErrorAction SilentlyContinue) {
        Write-Host "==> Installing engines with winget"
        winget install --silent --accept-source-agreements --accept-package-agreements ffmpeg
        winget install --silent --accept-source-agreements --accept-package-agreements ImageMagick.ImageMagick
        winget install --silent --accept-source-agreements --accept-package-agreements JohnMacFarlane.Pandoc
        winget install --silent --accept-source-agreements --accept-package-agreements oschwartz10612.Poppler
        return
    }

    Write-Host "X No supported package manager found (choco or winget). Install ffmpeg, imagemagick, pandoc, and pdftotext manually."
}

if ($Yes) {
    Install-Engines
} else {
    $reply = Read-Host "Install FFmpeg, ImageMagick, Pandoc, Poppler now? [y/N]"
    if ($reply -match '^[Yy]$') {
        Install-Engines
    }
}

Write-Host "==> Preparing Python runtime (.venv + modules)"
powershell -ExecutionPolicy Bypass -File .\scripts\bootstrap-python.ps1

$venvPython = Resolve-Path .\.venv\Scripts\python.exe -ErrorAction SilentlyContinue
if ($venvPython) {
    [Environment]::SetEnvironmentVariable("FLUX_PYTHON", $venvPython.Path, "User")
    $env:FLUX_PYTHON = $venvPython.Path
    Write-Host "OK Set FLUX_PYTHON for current and future terminals"
}

Write-Host "==> Building flux"
go build -o .\bin\flux.exe main.go
Write-Host "OK Built .\\bin\\flux.exe"

$installDir = Join-Path $env:USERPROFILE "bin"
if (-not (Test-Path $installDir)) {
    New-Item -ItemType Directory -Path $installDir | Out-Null
}
Copy-Item .\bin\flux.exe (Join-Path $installDir "flux.exe") -Force
Write-Host "OK Installed $installDir\\flux.exe"

$currentUserPath = [Environment]::GetEnvironmentVariable("Path", "User")
if (-not $currentUserPath) {
    $currentUserPath = ""
}
if ($currentUserPath -notlike "*$installDir*") {
    $newPath = ($currentUserPath.TrimEnd(';') + ";" + $installDir).Trim(';')
    [Environment]::SetEnvironmentVariable("Path", $newPath, "User")
    Write-Host "OK Added $installDir to User PATH"
    Write-Host "OK Open a new terminal to use 'flux' directly"
}

Write-Host "==> Verifying local binary"
.\bin\flux.exe doctor
Write-Host "==> Done"
