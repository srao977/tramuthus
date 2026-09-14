<#
.SYNOPSIS
Exports analysis-friendly flat CSV rows for the two authoritative governed runs.
.DESCRIPTION
Reads h_h_stage_emit_values through mongosh using the exact RUN_A and RUN_B
pipeline identifiers. Nested trace values are flattened into ordinary columns;
audit-only evidence hashes are omitted. The source collection is read-only.
.PARAMETER OutputPath
Destination CSV path. Defaults to C:\Users\chino\DSE_JEH_exports.
.EXAMPLE
.\scripts\Export-HHStageEmitValuesFlatCsv.ps1
#>
[CmdletBinding()]
param(
    [Parameter()]
    [ValidateNotNullOrEmpty()]
    [string]$OutputPath = "C:\Users\chino\DSE_JEH_exports\h_h_stage_emit_values_RUN_A_RUN_B_flat_2026-09-14.csv",

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
  { _id: 0 }
).sort({ pipeline_run_id: 1, symbol: 1, sequence_no: 1, stage_order: 1, stage: 1, emit_type: 1 }).toArray();

function value(document, key) {
  return document !== undefined && document !== null && document[key] !== undefined && document[key] !== null
    ? document[key]
    : '';
}

function numberValue(document, key) {
  const item = value(document, key);
  return item === '' ? '' : Number(item);
}

const exportRows = rows.map((row) => {
  const input = row.input || {};
  const output = row.output || {};
  const capital = row.capital_allocation || {};
  const fact = Array.isArray(output.facts) && output.facts.length > 0 ? output.facts[0] : {};
  return {
    pipeline_run_id: row.pipeline_run_id,
    run_type: row.run_type,
    collection_run_id: row.collection_run_id,
    symbol: row.symbol,
    sequence_no: Number(row.sequence_no),
    stage: row.stage,
    stage_order: Number(row.stage_order),
    event_time_utc: row.event_time.toISOString(),
    emit_type: value(row, 'emit_type'),
    phase_degrees: numberValue(output, 'phase_degrees'),
    close_price: numberValue(output, 'close'),
    signed_circular_displacement_degrees: numberValue(output, 'signed_circular_displacement_degrees'),
    phase_velocity_degrees_per_bar: numberValue(output, 'phase_velocity_degrees_per_bar'),
    phase_direction: value(output, 'direction'),
    boundary_degrees: numberValue(fact, 'degrees'),
    boundary_direction: value(fact, 'direction'),
    boundary_encounter_order: numberValue(fact, 'encounter_order'),
    prior_state: value(input, 'prior_state'),
    current_state: value(output, 'current_state'),
    boundary_policy_outcome: value(output, 'boundary_policy_outcome'),
    action: value(output, 'action'),
    requested_action: value(output, 'requested_action'),
    execution_status: value(output, 'status'),
    reference_price: numberValue(output, 'reference_price'),
    hold_price: numberValue(output, 'price'),
    fill_price: numberValue(output, 'fill_price'),
    quantity: numberValue(output, 'quantity'),
    requested_quantity: numberValue(output, 'requested_quantity'),
    filled_quantity: numberValue(output, 'filled_quantity'),
    starting_capital: numberValue(capital, 'starting_capital'),
    available_capital: numberValue(capital, 'available_capital'),
    cash_remaining: numberValue(capital, 'cash_remaining'),
    active_quantity: numberValue(capital, 'active_quantity'),
    entry_price: numberValue(capital, 'entry_price'),
    peak_price: numberValue(capital, 'peak_price') !== '' ? numberValue(capital, 'peak_price') : numberValue(output, 'peak_price'),
    stop_price: numberValue(output, 'stop_price'),
    current_capital: numberValue(capital, 'current_capital'),
    realized_pnl: numberValue(capital, 'realized_pnl'),
    return_pct: numberValue(capital, 'return_pct'),
    allocation_pct: numberValue(capital, 'allocation_pct'),
    risk_r: numberValue(capital, 'risk_r')
  };
});

print(JSON.stringify({ count: exportRows.length, rows: exportRows }));
"@

Write-Host ("Exporting flat RUN_A and RUN_B analysis from {0}.{1}..." -f $Database, $Collection)
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

$columns = @(
    "pipeline_run_id", "run_type", "collection_run_id", "symbol", "sequence_no",
    "stage", "stage_order", "event_time_utc", "emit_type",
    "phase_degrees", "close_price", "signed_circular_displacement_degrees",
    "phase_velocity_degrees_per_bar", "phase_direction", "boundary_degrees",
    "boundary_direction", "boundary_encounter_order", "prior_state", "current_state",
    "boundary_policy_outcome", "action", "requested_action", "execution_status",
    "reference_price", "hold_price", "fill_price", "quantity", "requested_quantity",
    "filled_quantity", "starting_capital", "available_capital", "cash_remaining",
    "active_quantity", "entry_price", "peak_price", "stop_price", "current_capital",
    "realized_pnl", "return_pct", "allocation_pct", "risk_r"
)

$records |
    Select-Object -Property $columns |
    Export-Csv -LiteralPath $OutputPath -NoTypeInformation -Encoding UTF8

$csvRows = @(Import-Csv -LiteralPath $OutputPath)
if ($csvRows.Count -ne $records.Count) {
    throw "CSV verification failed: expected $($records.Count) rows, found $($csvRows.Count)."
}

$runACount = @($csvRows | Where-Object { $_.pipeline_run_id -eq $runAID }).Count
$runBCount = @($csvRows | Where-Object { $_.pipeline_run_id -eq $runBID }).Count
if ($runACount -eq 0 -or $runBCount -eq 0 -or ($runACount + $runBCount) -ne $csvRows.Count) {
    throw "CSV run verification failed: RUN_A=$runACount, RUN_B=$runBCount, total=$($csvRows.Count)."
}

$jsonColumns = @($csvRows[0].PSObject.Properties.Name | Where-Object { $_ -match 'json' })
if ($jsonColumns.Count -ne 0) {
    throw "Flat CSV unexpectedly contains JSON columns: $($jsonColumns -join ', ')."
}

$file = Get-Item -LiteralPath $OutputPath
$hash = Get-FileHash -LiteralPath $OutputPath -Algorithm SHA256
Write-Host ("Flat export complete: {0:N0} rows, {1} columns" -f $records.Count, $columns.Count)
Write-Host ("RUN_A rows: {0:N0}" -f $runACount)
Write-Host ("RUN_B rows: {0:N0}" -f $runBCount)
Write-Host ("CSV: {0}" -f $file.FullName)
Write-Host ("Size: {0:N0} bytes" -f $file.Length)
Write-Host ("SHA256: {0}" -f $hash.Hash)