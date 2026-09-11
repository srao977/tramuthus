<#
.SYNOPSIS
Runs one parameterized Bar Sequence Lab phase-angle experiment.
.DESCRIPTION
Builds the native Go batch program and supplies an explicit SERIES_SIZE. It
reads raw Mongo bars and upserts only the separately initialized phase collection.
.PARAMETER SERIES_SIZE
Positive count of ordered observations supplied independently to each symbol solver.
.PARAMETER CollectionRunId
Raw collection run to analyze; defaults to the current Lab reference run.
.EXAMPLE
.\scripts\Start-PhaseAngleSeriesGenerator.ps1 -SERIES_SIZE 6
.EXAMPLE
.\scripts\Start-PhaseAngleSeriesGenerator.ps1 -SERIES_SIZE 48 -CollectionRunId 20260910T191246Z-1
.NOTES
Does not initialize schema, delete data, alter raw records, edit source, commit, or push.
The Go process exit code is preserved as this script's exit code.
#>
[CmdletBinding()]
param(
    [Parameter(Mandatory = $true)]
    [ValidateRange(1, [int]::MaxValue)]
    [int]$SERIES_SIZE,

    [Parameter()]
    [ValidateNotNullOrEmpty()]
    [string]$CollectionRunId = "20260910T191246Z-1"
)

$ErrorActionPreference = "Stop"
$componentRoot = Split-Path -Parent $PSScriptRoot
$binaryDirectory = Join-Path $componentRoot "bin"
$binaryPath = Join-Path $binaryDirectory "phase-angle-series-generator.exe"

$mongoUri = if ($env:BAR_SEQ_LAB_MONGO_URI) { $env:BAR_SEQ_LAB_MONGO_URI } else { "mongodb://127.0.0.1:27017" }
$mongoDb = if ($env:BAR_SEQ_LAB_MONGO_DB) { $env:BAR_SEQ_LAB_MONGO_DB } else { "bar_sequence_db" }
$rawCollection = if ($env:BAR_SEQ_LAB_RAW_COLLECTION) { $env:BAR_SEQ_LAB_RAW_COLLECTION } else { "bar_sequence" }
$phaseCollection = if ($env:BAR_SEQ_LAB_PHASE_COLLECTION) { $env:BAR_SEQ_LAB_PHASE_COLLECTION } else { "bar_sequence_phase_angle_series" }

Write-Host "Bar Sequence Lab - Phase Angle Series Generator"
Write-Host ""
Write-Host ("Collection Run  : {0}" -f $CollectionRunId)
Write-Host ("SERIES_SIZE     : {0}" -f $SERIES_SIZE)
Write-Host ("Mongo DB        : {0}" -f $mongoDb)
Write-Host ("Raw Collection  : {0}" -f $rawCollection)
Write-Host ("Phase Collection: {0}" -f $phaseCollection)
Write-Host "Solver          : EHLERS_DOMINANT_CYCLE_PHASE"
Write-Host "Solver Version  : V0.1"
Write-Host ""

New-Item -ItemType Directory -Path $binaryDirectory -Force | Out-Null
Push-Location $componentRoot
try {
    Write-Host "Building Go batch program..."
    & go build -o $binaryPath ./cmd/phase-angle-series-generator
    if ($LASTEXITCODE -ne 0) {
        Write-Error "Go build failed with exit code $LASTEXITCODE."
        exit $LASTEXITCODE
    }

    Write-Host "Starting experiment..."
    & $binaryPath `
        -series-size $SERIES_SIZE `
        -collection-run-id $CollectionRunId `
        -mongo-uri $mongoUri `
        -mongo-db $mongoDb `
        -raw-collection $rawCollection `
        -phase-collection $phaseCollection
    $processExitCode = $LASTEXITCODE
} finally {
    Pop-Location
}

if ($processExitCode -eq 0) {
    Write-Host "Phase-angle experiment completed successfully."
} else {
    Write-Error "Phase-angle experiment failed with exit code $processExitCode."
}
exit $processExitCode