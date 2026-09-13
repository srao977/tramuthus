param(
    [Parameter(Mandatory = $true)]
    [ValidateSet("ONLINE", "OFFLINE")]
    [string]$Mode,

    [string]$OfflineCollectionRunID
)

$ErrorActionPreference = "Stop"
$root = Split-Path -Parent $PSScriptRoot
$executable = Join-Path $root "bin/dse-jeh-transsat-1.exe"
if (-not (Test-Path $executable)) {
    throw "Executable not found. Run scripts/Build-DSEJEHTransSat1.ps1 first."
}

$env:DSE_JEH_MODE = $Mode
if ($Mode -eq "OFFLINE") {
    if (-not [string]::IsNullOrWhiteSpace($OfflineCollectionRunID)) {
        $env:DSE_JEH_OFFLINE_COLLECTION_RUN_ID = $OfflineCollectionRunID.Trim()
    }
    if ([string]::IsNullOrWhiteSpace($env:DSE_JEH_OFFLINE_COLLECTION_RUN_ID) -and
        [string]::IsNullOrWhiteSpace($env:DSE_JEH_COLLECTION_RUN_ID)) {
        throw "OFFLINE mode requires -OfflineCollectionRunID or DSE_JEH_OFFLINE_COLLECTION_RUN_ID."
    }
}
& $executable
exit $LASTEXITCODE