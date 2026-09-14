param(
    [Parameter(Mandatory = $true)]
    [ValidateSet("ONLINE", "OFFLINE")]
    [string]$Mode,

    [string]$OfflineCollectionRunID,

    [string]$PipelineRunID,

    [ValidatePattern('^[A-Z0-9][A-Z0-9_-]*$')]
    [string]$RunType,

    [double]$StartingCapital = 100000.00,

    [ValidateRange(0.0, 1.0)]
    [double]$AllocationPct = 1.0,

    [ValidateRange(0.0, 1.0)]
    [double]$RiskR = 0.0
)

$ErrorActionPreference = "Stop"
$root = Split-Path -Parent $PSScriptRoot
$executable = Join-Path $root "bin/dse-jeh-transsat-1.exe"
$runtimeDirectory = Join-Path $root "bin/runtime"
$pidFile = Join-Path $runtimeDirectory "dse-jeh-transsat-1.pid"
$stopFile = Join-Path $runtimeDirectory "dse-jeh-transsat-1.stop"
if (-not (Test-Path $executable)) {
    throw "Executable not found. Run scripts/Build-DSEJEHTransSat1.ps1 first."
}
New-Item -ItemType Directory -Force -Path $runtimeDirectory | Out-Null
if (Test-Path $pidFile) {
    $existingPID = (Get-Content $pidFile -Raw).Trim()
    if ($existingPID -match '^\d+$' -and (Get-Process -Id ([int]$existingPID) -ErrorAction SilentlyContinue)) {
        throw "DSE_JEH_TransSat_1 is already running with PID $existingPID."
    }
    Remove-Item $pidFile -Force
}
Remove-Item $stopFile -Force -ErrorAction SilentlyContinue

$env:DSE_JEH_MODE = $Mode
if ($Mode -eq "OFFLINE") {
    if (-not [string]::IsNullOrWhiteSpace($OfflineCollectionRunID)) {
        $env:DSE_JEH_OFFLINE_COLLECTION_RUN_ID = $OfflineCollectionRunID.Trim()
    }
    if ([string]::IsNullOrWhiteSpace($env:DSE_JEH_OFFLINE_COLLECTION_RUN_ID) -and
        [string]::IsNullOrWhiteSpace($env:DSE_JEH_COLLECTION_RUN_ID)) {
        throw "OFFLINE mode requires -OfflineCollectionRunID or DSE_JEH_OFFLINE_COLLECTION_RUN_ID."
    }
    if ([string]::IsNullOrWhiteSpace($PipelineRunID)) {
        throw "OFFLINE mode requires -PipelineRunID."
    }
    if ([string]::IsNullOrWhiteSpace($RunType)) {
        throw "OFFLINE mode requires -RunType matching ^[A-Z0-9][A-Z0-9_-]*$."
    }
    if ($StartingCapital -le 0) {
        throw "StartingCapital must be positive."
    }
    $env:DSE_JEH_PIPELINE_RUN_ID = $PipelineRunID.Trim()
    $env:DSE_JEH_RUN_TYPE = $RunType
    $env:DSE_JEH_STARTING_CAPITAL = $StartingCapital.ToString([System.Globalization.CultureInfo]::InvariantCulture)
    $env:DSE_JEH_ALLOCATION_PCT = $AllocationPct.ToString([System.Globalization.CultureInfo]::InvariantCulture)
    $env:DSE_JEH_RISK_R = $RiskR.ToString([System.Globalization.CultureInfo]::InvariantCulture)
}
$env:DSE_JEH_STOP_FILE = $stopFile
Write-Host "Starting DSE_JEH_TransSat_1 ($Mode)"
$runtimeEndpoint = $env:DSE_JEH_SERVER_ADDRESS
if ([string]::IsNullOrWhiteSpace($runtimeEndpoint)) {
    $runtimeEndpoint = "127.0.0.1:50052"
}
Write-Host "Runtime endpoint: $runtimeEndpoint"
$process = Start-Process -FilePath $executable -NoNewWindow -PassThru
Set-Content -Path $pidFile -Value $process.Id -NoNewline
Write-Host "PID: $($process.Id)"
try {
    $process.WaitForExit()
    exit $process.ExitCode
}
finally {
    if ((Test-Path $pidFile) -and ((Get-Content $pidFile -Raw).Trim() -eq [string]$process.Id)) {
        Remove-Item $pidFile -Force
    }
    Remove-Item $stopFile -Force -ErrorAction SilentlyContinue
}