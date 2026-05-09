param(
    [string]$OutDir = "./bin"
)

$ErrorActionPreference = 'Stop'

$root = Split-Path -Parent $MyInvocation.MyCommand.Path
Set-Location $root

if (-not (Test-Path -Path $OutDir)) {
    New-Item -ItemType Directory -Path $OutDir | Out-Null
}

Write-Host "Building peer..."
go build -o (Join-Path $OutDir "peer-node.exe") ./cmd/peer

Write-Host "Building bootstrap..."
go build -o (Join-Path $OutDir "bootstrap-node.exe") ./cmd/bootstrap

Write-Host "Build complete. Binaries placed in:" (Resolve-Path $OutDir)
