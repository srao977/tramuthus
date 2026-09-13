$ErrorActionPreference = "Stop"
$root = Split-Path -Parent $PSScriptRoot
$executable = [System.IO.Path]::GetFullPath((Join-Path $root "bin/dse-jeh-transsat-1.exe"))
$runtimeDirectory = Join-Path $root "bin/runtime"
$pidFile = Join-Path $runtimeDirectory "dse-jeh-transsat-1.pid"
$stopFile = Join-Path $runtimeDirectory "dse-jeh-transsat-1.stop"

if (-not (Test-Path $pidFile)) {
    throw "No DSE_JEH_TransSat_1 PID file exists at $pidFile."
}
$runtimePID = (Get-Content $pidFile -Raw).Trim()
if ($runtimePID -notmatch '^\d+$') {
    throw "Invalid DSE_JEH_TransSat_1 PID file content."
}
$process = Get-Process -Id ([int]$runtimePID) -ErrorAction SilentlyContinue
if ($null -eq $process) {
    Remove-Item $pidFile -Force
    throw "DSE_JEH_TransSat_1 PID $runtimePID is no longer running; removed stale PID file."
}
$actualExecutable = [System.IO.Path]::GetFullPath($process.Path)
if (-not $actualExecutable.Equals($executable, [System.StringComparison]::OrdinalIgnoreCase)) {
    throw "PID $runtimePID belongs to '$actualExecutable', not DSE_JEH_TransSat_1. Refusing to stop it."
}

Set-Content -Path $stopFile -Value $runtimePID -NoNewline
Write-Host "Graceful stop requested for DSE_JEH_TransSat_1 PID $runtimePID."