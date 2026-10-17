# Coffee Bun (bun) Installer for Windows PowerShell
# Usage:
#   irm https://raw.githubusercontent.com/KennethanCeyer/coffee-bun/main/install.ps1 | iex

$ErrorActionPreference = "Stop"

Write-Host "`nCoffee Bun (bun) - Windows Bootstrap Installer`n" -ForegroundColor Cyan

# 1. Verify Go is available
if (-not (Get-Command "go" -ErrorAction SilentlyContinue)) {
    Write-Host "  [..] Go not found. Installing Go via winget..." -ForegroundColor Yellow
    winget install --id GoLang.Go -e --accept-source-agreements --accept-package-agreements --disable-interactivity
}

$goVer = (go version)
Write-Host "  [OK] Go is ready ($goVer)" -ForegroundColor Green

# 2. Prepare installation directory
$binDir = Join-Path $HOME ".local\bin"
if (-not (Test-Path $binDir)) {
    New-Item -ItemType Directory -Force -Path $binDir | Out-Null
}

# 3. Determine source directory
if ((Test-Path ".git") -and (Test-Path "go.mod")) {
    $srcDir = (Get-Location).Path
    Write-Host "  [..] Building Coffee Bun from current repository..." -ForegroundColor Cyan
} else {
    $srcDir = Join-Path $HOME ".coffee-bun"
    if (Test-Path $srcDir) {
        Write-Host "  [..] Updating Coffee Bun source at $srcDir..." -ForegroundColor Cyan
        git -C $srcDir pull --rebase
    } else {
        Write-Host "  [..] Cloning Coffee Bun repository to $srcDir..." -ForegroundColor Cyan
        git clone https://github.com/KennethanCeyer/coffee-bun.git $srcDir
    }
}

# 4. Build bun.exe and coffee-bun.exe
$bunExe = Join-Path $binDir "bun.exe"
$aliasExe = Join-Path $binDir "coffee-bun.exe"

if (Test-Path $bunExe) { Remove-Item -Force $bunExe }
if (Test-Path $aliasExe) { Remove-Item -Force $aliasExe }

Push-Location $srcDir
try {
    go build -ldflags="-s -w" -o $bunExe ./cmd/bun
    Copy-Item -Force $bunExe $aliasExe
} finally {
    Pop-Location
}

# 5. Add ~/.local/bin to User PATH if not already present
$userPath = [Environment]::GetEnvironmentVariable("Path", "User")
if ($userPath -notlike "*$binDir*") {
    [Environment]::SetEnvironmentVariable("Path", "$binDir;$userPath", "User")
    $env:Path = "$binDir;$env:Path"
}

Write-Host "`n[OK] Coffee Bun installed successfully to $bunExe!`n" -ForegroundColor Green
Write-Host "Run 'bun' or 'bun check' to get started.`n"
