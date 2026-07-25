param(
    [string]$OutDir = "./bin",
    [string]$v = "1.0",
    [string]$cmd = "a"
)

$ErrorActionPreference = 'Stop'

$root = Split-Path -Parent $MyInvocation.MyCommand.Path
Set-Location $root

if (-not (Test-Path -Path $OutDir)) {
    New-Item -ItemType Directory -Path $OutDir | Out-Null
}

foreach ($type in Get-ChildItem -Path .\cmd) {
    if ($cmd -eq "a" -or $cmd -like ("*" + $type.name[0] + "*")) {
        Write-Host "Building $($type.Name)..."
        $output = "$($type.Name)-$v.exe"
        if ($type.Name -eq "gui") {
            go build -ldflags "-H=windowsgui" -o (Join-Path $OutDir $output) "./cmd/$($type.Name)"
        } else {
            go build -o (Join-Path $OutDir $output) "./cmd/$($type.Name)"
        }
    }
}

Write-Host "Build complete. Binaries placed in:" (Resolve-Path $OutDir)
