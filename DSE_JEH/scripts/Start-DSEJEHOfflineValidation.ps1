param(
    [string]$CollectionRunID = "20260911T161623Z-1",
    [string]$MongoURI = "mongodb://127.0.0.1:27017",
    [string]$MongoDatabase = "bar_sequence_db",
    [string]$MongoCollection = "bar_sequence",
    [string]$ServerAddress = "127.0.0.1:50052",
    [string]$StatusInterval = "10s",
    [string]$OutputPath = "exports/offline_20260911T161623Z-1_phase_evidence.jsonl",
    [bool]$RemainRunning = $true
)

$ErrorActionPreference = "Stop"

if ([string]::IsNullOrWhiteSpace($CollectionRunID)) {
    throw "CollectionRunID is required."
}
if ([string]::IsNullOrWhiteSpace($MongoURI)) {
    throw "MongoURI is required."
}
if ([string]::IsNullOrWhiteSpace($MongoDatabase)) {
    throw "MongoDatabase is required."
}
if ([string]::IsNullOrWhiteSpace($MongoCollection)) {
    throw "MongoCollection is required."
}
if ([string]::IsNullOrWhiteSpace($ServerAddress)) {
    throw "ServerAddress is required."
}
if ([string]::IsNullOrWhiteSpace($StatusInterval)) {
    throw "StatusInterval is required."
}

$startScript = Join-Path $PSScriptRoot "Start-DSEJEHTransSat1.ps1"
if (-not (Test-Path $startScript)) {
    throw "Governed start script not found at $startScript."
}

$env:BAR_SEQ_LAB_MONGO_URI = $MongoURI.Trim()
$env:BAR_SEQ_LAB_MONGO_DB = $MongoDatabase.Trim()
$env:BAR_SEQ_LAB_MONGO_COLLECTION = $MongoCollection.Trim()
$env:DSE_JEH_OFFLINE_COLLECTION_RUN_ID = $CollectionRunID.Trim()
$env:DSE_JEH_SERVER_ADDRESS = $ServerAddress.Trim()
$env:DSE_JEH_STATUS_INTERVAL = $StatusInterval.Trim()
$env:DSE_JEH_REMAIN_RUNNING = $RemainRunning.ToString().ToLowerInvariant()

if (-not [string]::IsNullOrWhiteSpace($OutputPath)) {
    $root = Split-Path -Parent $PSScriptRoot
    $resolvedOutput = $OutputPath
    if (-not [System.IO.Path]::IsPathRooted($resolvedOutput)) {
        $resolvedOutput = Join-Path $root $resolvedOutput
    }
    $env:DSE_JEH_OUTPUT = [System.IO.Path]::GetFullPath($resolvedOutput)
}

Write-Host "DSE_JEH_TransSat_1 OFFLINE validation"
Write-Host "Mongo database: $($env:BAR_SEQ_LAB_MONGO_DB)"
Write-Host "Mongo collection: $($env:BAR_SEQ_LAB_MONGO_COLLECTION)"
Write-Host "Collection run: $($env:DSE_JEH_OFFLINE_COLLECTION_RUN_ID)"
Write-Host "Runtime endpoint: $($env:DSE_JEH_SERVER_ADDRESS)"
Write-Host "Remain running after replay: $($env:DSE_JEH_REMAIN_RUNNING)"
if (-not [string]::IsNullOrWhiteSpace($env:DSE_JEH_OUTPUT)) {
    Write-Host "Phase evidence: $($env:DSE_JEH_OUTPUT)"
}
Write-Host "Use scripts/Stop-DSEJEHTransSat1.ps1 from another terminal to request graceful shutdown."

& $startScript -Mode OFFLINE -OfflineCollectionRunID $env:DSE_JEH_OFFLINE_COLLECTION_RUN_ID
exit $LASTEXITCODE