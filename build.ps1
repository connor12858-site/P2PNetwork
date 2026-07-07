param(
    [string]$OutDir = "./bin",
    [string]$v = "1.0"
)

$ErrorActionPreference = 'Stop'

$root = Split-Path -Parent $MyInvocation.MyCommand.Path
Set-Location $root

if (-not (Test-Path -Path $OutDir)) {
    New-Item -ItemType Directory -Path $OutDir | Out-Null
}

foreach ($type in Get-ChildItem -Path .\cmd) {
    Write-Host "Building $($type.Name)..."
    $output = "$($type.Name)-$v.exe"
    if ($type.Name -eq "gui") {
        go build -ldflags "-H=windowsgui" -o (Join-Path $OutDir $output) "./cmd/$($type.Name)"
    } else {
        go build -o (Join-Path $OutDir $output) "./cmd/$($type.Name)"
    }

}

# Write-Host "Building peer..."
# $peerOutput = "peer-" + $v + ".exe"
# go build -o (Join-Path $OutDir $peerOutput) ./cmd/peer

# Write-Host "Building bootstrap..."
# $bootstrapOutput = "bootstrap-" + $v + ".exe"
# go build -o (Join-Path $OutDir $bootstrapOutput) ./cmd/bootstrap

Write-Host "Build complete. Binaries placed in:" (Resolve-Path $OutDir)
