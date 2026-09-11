<#
.SYNOPSIS
Exports phase-angle records created on one UTC date to CSV.
.DESCRIPTION
Reads bar_sequence_phase_angle_series through mongosh with an explicit UTC
analysis_created_at range. The export is read-only, sorted deterministically,
and includes a stable header even when phase_angle_degrees is null.
.PARAMETER ExportDate
UTC calendar date to export. Defaults to the current UTC date.
.PARAMETER OutputPath
Destination CSV path. Defaults to the generator exports directory.
.EXAMPLE
.\scripts\Export-PhaseAngleSeriesCsv.ps1 -ExportDate 2026-09-11
#>
[CmdletBinding()]
param(
    [Parameter()]
    [datetime]$ExportDate = [datetime]::UtcNow.Date,

    [Parameter()]
    [ValidateNotNullOrEmpty()]
    [string]$OutputPath,

    [Parameter()]
    [ValidateNotNullOrEmpty()]
    [string]$MongoUri = $(if ($env:BAR_SEQ_LAB_MONGO_URI) { $env:BAR_SEQ_LAB_MONGO_URI } else { "mongodb://127.0.0.1:27017" }),

    [Parameter()]
    [ValidatePattern('^[A-Za-z0-9_-]+$')]
    [string]$Database = $(if ($env:BAR_SEQ_LAB_MONGO_DB) { $env:BAR_SEQ_LAB_MONGO_DB } else { "bar_sequence_db" }),

    [Parameter()]
    [ValidatePattern('^[A-Za-z0-9_-]+$')]
    [string]$Collection = $(if ($env:BAR_SEQ_LAB_PHASE_COLLECTION) { $env:BAR_SEQ_LAB_PHASE_COLLECTION } else { "bar_sequence_phase_angle_series" })
)

$ErrorActionPreference = "Stop"
$componentRoot = Split-Path -Parent $PSScriptRoot
$utcStart = $ExportDate.ToUniversalTime().Date
$utcEnd = $utcStart.AddDays(1)

if (-not $OutputPath) {
    $exportDirectory = Join-Path $componentRoot "exports"
    $OutputPath = Join-Path $exportDirectory ("bar_sequence_phase_angle_series_{0}.csv" -f $utcStart.ToString("yyyy-MM-dd"))
}
$OutputPath = [System.IO.Path]::GetFullPath($OutputPath)
$outputDirectory = Split-Path -Parent $OutputPath
New-Item -ItemType Directory -Path $outputDirectory -Force | Out-Null

$mongosh = Get-Command mongosh -ErrorAction SilentlyContinue
if (-not $mongosh) {
    throw "mongosh was not found on PATH. Install MongoDB Shell or add mongosh.exe to PATH."
}

$startIso = $utcStart.ToString("o")
$endIso = $utcEnd.ToString("o")
$query = @"
const start = new Date('$startIso');
const end = new Date('$endIso');
const rows = db.getSiblingDB('$Database').getCollection('$Collection').find(
  { analysis_created_at: { `$gte: start, `$lt: end } },
  {
        _id: 0,
        collection_run_id: 1,
        partition_id: 1,
        symbol: 1,
        generator_sequence_no: 1,
        series_size: 1,
        solver_name: 1,
        solver_version: 1,
        input_series_type: 1,
        phase_angle_degrees: 1,
        phase_observable: 1,
        validity_state: 1,
        analysis_created_at: 1
  }
).sort({ collection_run_id: 1, series_size: 1, partition_id: 1, symbol: 1, generator_sequence_no: 1 }).toArray();
const exportRows = rows.map((row) => ({
    collection_run_id: row.collection_run_id,
    partition_id: row.partition_id,
    symbol: row.symbol,
    generator_sequence_no: Number(row.generator_sequence_no),
    series_size: Number(row.series_size),
    solver_name: row.solver_name,
    solver_version: row.solver_version,
    input_series_type: row.input_series_type,
    phase_angle_degrees: row.phase_angle_degrees === null ? null : Number(row.phase_angle_degrees),
    phase_observable: row.phase_observable,
    validity_state: row.validity_state,
    analysis_created_at: row.analysis_created_at.toISOString()
}));
print(JSON.stringify({ count: exportRows.length, rows: exportRows }));
"@

Write-Host ("Exporting {0}.{1} records created during UTC {2:yyyy-MM-dd}..." -f $Database, $Collection, $utcStart)
$shellOutput = & $mongosh.Source $MongoUri --quiet --norc --eval $query
if ($LASTEXITCODE -ne 0) {
    throw "mongosh failed with exit code $LASTEXITCODE."
}

$json = ($shellOutput -join [Environment]::NewLine).Trim()
if (-not $json) {
    throw "mongosh returned no export payload."
}
$payload = $json | ConvertFrom-Json
$records = @()
foreach ($record in $payload.rows) {
    $records += $record
}
if ($records.Count -ne [int]$payload.count) {
    throw "Mongo payload verification failed: expected $($payload.count) rows, parsed $($records.Count)."
}
$columns = @(
    "collection_run_id",
    "partition_id",
    "symbol",
    "generator_sequence_no",
    "series_size",
    "solver_name",
    "solver_version",
    "input_series_type",
    "phase_angle_degrees",
    "phase_observable",
    "validity_state",
    "analysis_created_at"
)

$records |
    Select-Object -Property $columns |
    Export-Csv -LiteralPath $OutputPath -NoTypeInformation -Encoding UTF8

if (-not (Test-Path -LiteralPath $OutputPath)) {
    throw "CSV export was not created: $OutputPath"
}

$csvRows = @(Import-Csv -LiteralPath $OutputPath)
if ($csvRows.Count -ne $records.Count) {
    throw "CSV verification failed: expected $($records.Count) rows, found $($csvRows.Count)."
}

Write-Host ("Export complete: {0:N0} rows" -f $records.Count)
Write-Host ("CSV: {0}" -f $OutputPath)