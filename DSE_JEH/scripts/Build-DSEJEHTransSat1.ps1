$ErrorActionPreference = "Stop"

$root = Split-Path -Parent $PSScriptRoot
Push-Location $root
try {
    buf lint
    if ($LASTEXITCODE -ne 0) { exit $LASTEXITCODE }
    buf generate
    if ($LASTEXITCODE -ne 0) { exit $LASTEXITCODE }
    go test ./...
    if ($LASTEXITCODE -ne 0) { exit $LASTEXITCODE }
    New-Item -ItemType Directory -Force -Path bin | Out-Null
    go build -o bin/dse-jeh-transsat-1.exe ./cmd/dse-jeh-transsat-1
    if ($LASTEXITCODE -ne 0) { exit $LASTEXITCODE }
}
finally {
    Pop-Location
}
