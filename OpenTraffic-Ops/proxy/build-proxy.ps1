# build-proxy.ps1 - RTM Proxy Windows build script (cross-compile to Linux)
#
# Usage:
#   .\build-proxy.ps1
#   .\build-proxy.ps1 -Version "1.1.0"
#
# Prerequisites:
#   - Go 1.22+ installed and in PATH
#   - Run on Windows (for cross-compiling to Linux)

param(
    [string]$Version = "1.0.0",
    [string]$OutputDir = ".\dist"
)

$ErrorActionPreference = "Stop"

function Info($msg) { Write-Host "[INFO] $msg" -ForegroundColor Cyan }
function OK($msg)   { Write-Host "[OK]   $msg" -ForegroundColor Green }
function Warn($msg) { Write-Host "[WARN] $msg" -ForegroundColor Yellow }
function Err($msg)  { Write-Host "[ERR]  $msg" -ForegroundColor Red }

Info "Checking Go environment..."
$goVersion = go version 2>$null
if (-not $goVersion) {
    Err "Go not found. Please install Go and add it to PATH."
    exit 1
}
OK "Go version: $goVersion"

$BuildTime = Get-Date -Format "yyyy-MM-dd_HH:mm:ss"
$GoVer = (go version).Split(" ")[2]

$Ldflags = "-s -w -X main.proxyVersion=$Version -X main.buildTime=$BuildTime -X main.goVersion=$GoVer"

Info "RTM Proxy version: $Version"
Info "Build time: $BuildTime"

if (-not (Test-Path $OutputDir)) {
    New-Item -ItemType Directory -Path $OutputDir | Out-Null
}

$DistDir = Resolve-Path $OutputDir

$Targets = @(
    @{ GOOS = "linux"; GOARCH = "amd64"; Suffix = "linux-amd64" },
    @{ GOOS = "linux"; GOARCH = "arm64"; Suffix = "linux-arm64" }
)

$SuccessCount = 0
$FailCount = 0

foreach ($target in $Targets) {
    $os = $target.GOOS
    $arch = $target.GOARCH
    $suffix = $target.Suffix
    $outputFile = "rtm-proxy-$suffix"
    $outputPath = Join-Path $DistDir $outputFile

    Write-Host ""
    Info "[$($SuccessCount + $FailCount + 1)/$($Targets.Count)] Building $os/$arch ..."

    $env:GOOS = $os
    $env:GOARCH = $arch
    $env:CGO_ENABLED = "0"

    go build -ldflags "$Ldflags" -o "$outputPath" .

    if ($LASTEXITCODE -eq 0) {
        $fileInfo = Get-Item $outputPath
        $sizeMB = [math]::Round($fileInfo.Length / 1MB, 2)
        OK "Build success: $outputFile ($sizeMB MB)"
        $SuccessCount++
    } else {
        Err "Build failed: $os/$arch"
        $FailCount++
    }
}

$env:GOOS = $null
$env:GOARCH = $null
$env:CGO_ENABLED = $null

Write-Host ""
Write-Host "========================================" -ForegroundColor Cyan
Write-Host "        Build Summary" -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan

if ($SuccessCount -gt 0) {
    OK "Success: $SuccessCount targets"
    $items = Get-ChildItem $DistDir -Filter "rtm-proxy-*"
    foreach ($item in $items) {
        $size = [math]::Round($item.Length / 1MB, 2)
        Write-Host "  - $($item.Name) ($size MB)" -ForegroundColor Green
    }
}

if ($FailCount -gt 0) {
    Warn "Failed: $FailCount targets"
}

Write-Host ""
Info "Output directory: $DistDir"
Write-Host ""
Write-Host "Next steps:" -ForegroundColor Cyan
Write-Host "  1. Upload the binary to your Linux server" -ForegroundColor White
Write-Host "  2. scp $DistDir\rtm-proxy-linux-amd64 user@host:/opt/rtm-proxy/" -ForegroundColor Gray
Write-Host "  3. On server: chmod +x /opt/rtm-proxy/rtm-proxy-linux-amd64" -ForegroundColor Gray
Write-Host "  4. Start: ./rtm-proxy-linux-amd64" -ForegroundColor Gray
