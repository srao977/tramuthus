<#
.SYNOPSIS
Exports the two authoritative governed h_h_stage_emit_values runs to CSV.
.DESCRIPTION
Reads h_h_stage_emit_values through mongosh using the exact RUN_A and RUN_B
pipeline identifiers. The export is read-only, sorted deterministically, and
preserves stage-specific input, output, and capital_allocation documents as
compact JSON columns.
.PARAMETER OutputPath
Destination CSV path. Defaults to C:\Users\chino\DSE_JEH_exports.
.EXAMPLE
.\scripts\Export-HHStageEmitValuesCsv.ps1
#>
[CmdletBinding()]
param(
    [Parameter()]
    [ValidateNotNullOrEmpty()]
    [string]$OutputPath = "C:\Users\chino\DSE_JEH_exports\h_h_stage_emit_values_RUN_A_RUN_B_2026-09-14.csv",

    [Parameter()]
    [ValidateNotNullOrEmpty()]
    [string]$MongoUri = $(if ($env:BAR_SEQ_LAB_MONGO_URI) { $env:BAR_SEQ_LAB_MONGO_URI } else { "mongodb://127.0.0.1:27017" }),

    [Parameter()]
    [ValidatePattern('^[A-Za-z0-9_-]+$')]
    [string]$Database = $(if ($env:BAR_SEQ_LAB_MONGO_DB) { $env:BAR_SEQ_LAB_MONGO_DB } else { "bar_sequence_db" }),

    [Parameter()]
    [ValidatePattern('^[A-Za-z0-9_-]+$')]
    [string]$Collection = "h_h_stage_emit_values"
)

$ErrorActionPreference = "Stop"
$runAID = "DPE-GOVERNED-RUN-A-20260914-001"
$runBID = "DPE-GOVERNED-RUN-B-20260914-001"
$expectedRunTypes = @{
    $runAID = "RUN_A"
    $runBID = "RUN_B"
}

$OutputPath = [System.IO.Path]::GetFullPath($OutputPath)
$outputDirectory = Split-Path -Parent $OutputPath
New-Item -ItemType Directory -Path $outputDirectory -Force | Out-Null

$mongosh = Get-Command mongosh -ErrorAction SilentlyContinue
if (-not $mongosh) {
    throw "mongosh was not found on PATH. Install MongoDB Shell or add mongosh.exe to PATH."
}

$query = @"
const runIds = ['$runAID', '$runBID'];
const rows = db.getSiblingDB('$Database').getCollection('$Collection').find(
  { pipeline_run_id: { `$in: runIds } },
  {
    _id: 0,
    pipeline_run_id: 1,
    run_type: 1,
    collection_run_id: 1,
    symbol: 1,
    sequence_no: 1,
    stage: 1,
    stage_order: 1,
    event_time: 1,
    emit_type: 1,
    input: 1,
    output: 1,
    capital_allocation: 1
  }
).sort({ pipeline_run_id: 1, symbol: 1, sequence_no: 1, stage_order: 1, stage: 1, emit_type: 1 }).toArray();

function compactExtendedJson(value) {
  return value === undefined ? '' : EJSON.stringify(value, null, 0, { relaxed: true });
}

const exportRows = rows.map((row) => ({
  pipeline_run_id: row.pipeline_run_id,
  run_type: row.run_type,
  collection_run_id: row.collection_run_id,
  symbol: row.symbol,
  sequence_no: Number(row.sequence_no),
  stage: row.stage,
  stage_order: Number(row.stage_order),
  event_time: row.event_time.toISOString(),
  emit_type: row.emit_type === null || row.emit_type === undefined ? '' : row.emit_type,
  input_json: compactExtendedJson(row.input),
  output_json: compactExtendedJson(row.output),
  capital_allocation_json: compactExtendedJson(row.capital_allocation)
}));

print(JSON.stringify({ count: exportRows.length, rows: exportRows }));
"@

Write-Host ("Exporting authoritative RUN_A and RUN_B records from {0}.{1}..." -f $Database, $Collection)
$shellOutput = & $mongosh.Source $MongoUri --quiet --norc --eval $query
if ($LASTEXITCODE -ne 0) {
    throw "mongosh failed with exit code $LASTEXITCODE."
}

$json = ($shellOutput -join [Environment]::NewLine).Trim()
if (-not $json) {
    throw "mongosh returned no export payload."
}
$payload = $json | ConvertFrom-Json
$records = @($payload.rows)
if ($records.Count -ne [int]$payload.count) {
    throw "Mongo payload verification failed: expected $($payload.count) rows, parsed $($records.Count)."
}
if ($records.Count -eq 0) {
    throw "No records were found for the two authoritative pipeline runs."
}

foreach ($record in $records) {
    if (-not $expectedRunTypes.ContainsKey([string]$record.pipeline_run_id)) {
        throw "Unexpected pipeline_run_id in export payload: $($record.pipeline_run_id)"
    }
    $expectedRunType = $expectedRunTypes[[string]$record.pipeline_run_id]
    if ([string]$record.run_type -ne $expectedRunType) {
        throw "Run metadata mismatch for $($record.pipeline_run_id): expected $expectedRunType, found $($record.run_type)."
    }
}

$exportedPipelineIDs = @($records | Select-Object -ExpandProperty pipeline_run_id -Unique)
if ($exportedPipelineIDs.Count -ne 2 -or $exportedPipelineIDs -notcontains $runAID -or $exportedPipelineIDs -notcontains $runBID) {
    throw "Export payload does not contain both authoritative pipeline runs."
}

$columns = @(
    "pipeline_run_id",
    "run_type",
    "collection_run_id",
    "symbol",
    "sequence_no",
    "stage",
    "stage_order",
    "event_time",
    "emit_type",
    "input_json",
    "output_json",
    "capital_allocation_json"
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

$runACount = @($csvRows | Where-Object { $_.pipeline_run_id -eq $runAID }).Count
$runBCount = @($csvRows | Where-Object { $_.pipeline_run_id -eq $runBID }).Count
$file = Get-Item -LiteralPath $OutputPath
$hash = Get-FileHash -LiteralPath $OutputPath -Algorithm SHA256

Write-Host ("Export complete: {0:N0} rows" -f $records.Count)
Write-Host ("RUN_A rows: {0:N0}" -f $runACount)
Write-Host ("RUN_B rows: {0:N0}" -f $runBCount)
Write-Host ("CSV: {0}" -f $file.FullName)
Write-Host ("Size: {0:N0} bytes" -f $file.Length)
Write-Host ("SHA256: {0}" -f $hash.Hash)