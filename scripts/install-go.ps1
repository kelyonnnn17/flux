param(
    [string]$Module = "github.com/kelyonnnn17/flux"
)

$ErrorActionPreference = "Stop"

if (-not (Get-Command go -ErrorAction SilentlyContinue)) {
    throw "Go is required for this installer. Install Go from https://go.dev/dl/"
}

Write-Host "==> Installing flux via go install"
go install "$Module@latest"

$gobin = (go env GOBIN).Trim()
if ([string]::IsNullOrWhiteSpace($gobin)) {
    $gopath = (go env GOPATH).Trim()
    if ([string]::IsNullOrWhiteSpace($gopath)) {
        $gopath = Join-Path $env:USERPROFILE "go"
    }
    $gobin = Join-Path $gopath "bin"
}

$fluxExe = Join-Path $gobin "flux.exe"
if (-not (Test-Path $fluxExe)) {
    throw "Install completed but $fluxExe was not found"
}

$userPath = [Environment]::GetEnvironmentVariable("Path", "User")
if (-not $userPath) { $userPath = "" }
if ($userPath -notlike "*$gobin*") {
    $newPath = ($userPath.TrimEnd(';') + ";" + $gobin).Trim(';')
    [Environment]::SetEnvironmentVariable("Path", $newPath, "User")
    Write-Host "OK Added $gobin to User PATH"
    Write-Host "OK Open a new terminal to use flux"
}

Write-Host "OK Installed: $fluxExe"
& $fluxExe --help | Out-Null
Write-Host "OK flux is ready"
