$ErrorActionPreference = "Stop"

$root = Split-Path -Parent $PSScriptRoot
Push-Location $root
try {
    buf lint
    buf generate
    go test ./...
    New-Item -ItemType Directory -Force -Path bin | Out-Null
    go build -o bin/dse-jeh-transsat-1.exe ./cmd/dse-jeh-transsat-1
}
finally {
    Pop-Location
}
