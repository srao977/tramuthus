#Requires -Version 5.1
<#
.SYNOPSIS
    Start Bar_Sequence_Lab generator with selected partition groups.

.DESCRIPTION
    Lab-owned analog of Fin_FeedSat_1 Start-FinFeedSatIngestion.ps1 (read-only reference).
    Remaining arguments select groups A/B/C. No groups prints usage and exits.
    Does not default to A+B+C. Does not start Fin_FeedSat_1. Does not print secrets.

.EXAMPLE
    .\scripts\Start-BarSequenceLab.ps1 A

.EXAMPLE
    .\scripts\Start-BarSequenceLab.ps1 -Feed iex A B

.EXAMPLE
    .\scripts\Start-BarSequenceLab.ps1 A B C
#>
[CmdletBinding(PositionalBinding = $false)]
param(
    [ValidateSet("iex", "test")]
    [string]$Feed = "test",

    [string]$DataDir = "data",

    [int]$MaxBars = 0,

    [ValidateRange(0, [int]::MaxValue)]
    [int]$TargetBarsPerSymbol = 0,

    [string]$Duration = "",

    [Parameter(Position = 0, ValueFromRemainingArguments = $true)]
    [string[]]$Groups
)

$ErrorActionPreference = "Stop"

$usage = @"
Usage:
  .\scripts\Start-BarSequenceLab.ps1 [options] <groups>

Groups (required, one or more, no default of A+B+C):
  A
  B
  C
  A B
  A C
  B C
  A B C

Options:
  -Feed iex|test     Alpaca feed (default test)
  -DataDir <path>    JSONL output root (default data)
  -MaxBars <n>       Stop after n accepted bars (0 = unlimited)
    -TargetBarsPerSymbol <n>
                                         Stop when every configured symbol reaches n accepted bars (0 = disabled)
  -Duration <dur>    Go duration, e.g. 2m (optional)

Examples:
  .\scripts\Start-BarSequenceLab.ps1 A
  .\scripts\Start-BarSequenceLab.ps1 -Feed iex A B
  .\scripts\Start-BarSequenceLab.ps1 A B C

Does not start Fin_FeedSat_1. JSONL audit is always on.
Mongo is opt-in via BAR_SEQ_LAB_MONGO_ENABLED=true (URI/DB required). Ctrl+C drains and flushes JSONL and Mongo.
"@

if (-not $Groups -or $Groups.Count -eq 0) {
    Write-Host $usage
    exit 1
}

$allowed = @("A", "B", "C")
$selected = New-Object System.Collections.Generic.List[string]
$seen = @{}
foreach ($raw in $Groups) {
    foreach ($token in ($raw -split "[,\s;]+")) {
        $g = $token.Trim().ToUpperInvariant()
        if (-not $g) { continue }
        if ($allowed -notcontains $g) {
            Write-Error "Invalid group '$token'. Allowed: A B C"
            Write-Host $usage
            exit 1
        }
        if (-not $seen.ContainsKey($g)) {
            $seen[$g] = $true
            [void]$selected.Add($g)
        }
    }
}

if ($selected.Count -eq 0) {
    Write-Host $usage
    exit 1
}

$generatorRoot = Split-Path -Parent $PSScriptRoot
Set-Location $generatorRoot

if (-not (Get-Command go -ErrorAction SilentlyContinue)) {
    throw "Go is not on PATH. Install Go or open a shell where 'go version' works."
}

$hasKey = [string]::IsNullOrWhiteSpace($env:ALPACA_API_KEY) -eq $false
$hasSecret = ([string]::IsNullOrWhiteSpace($env:ALPACA_API_SECRET) -eq $false) -or
    ([string]::IsNullOrWhiteSpace($env:ALPACA_SECRET_KEY) -eq $false)
if (-not $hasKey -or -not $hasSecret) {
    throw "Set ALPACA_API_KEY and ALPACA_API_SECRET (or ALPACA_SECRET_KEY) in this session first."
}

$env:BAR_SEQ_LAB_FEED = $Feed
$env:BAR_SEQ_LAB_DATA_DIR = $DataDir
$env:BAR_SEQ_LAB_SELECTED_GROUPS = ($selected -join ",")
$env:BAR_SEQ_LAB_MAX_BARS = "$MaxBars"
$env:BAR_SEQ_LAB_TARGET_BARS_PER_SYMBOL = "$TargetBarsPerSymbol"
if ($Duration) {
    $env:BAR_SEQ_LAB_DURATION = $Duration
}

$mongoOn = $env:BAR_SEQ_LAB_MONGO_ENABLED
if ($mongoOn -match '^(1|true|yes|on)$') {
    Write-Host "Bar_Sequence_Lab generator  feed=$Feed  groups=$($selected -join ',')  data_dir=$DataDir  persistence=jsonl+mongo"
} else {
    Write-Host "Bar_Sequence_Lab generator  feed=$Feed  groups=$($selected -join ',')  data_dir=$DataDir  persistence=jsonl (mongo disabled)"
}
Write-Host "Does not print credentials. Does not start Fin_FeedSat_1."

if ($env:BAR_SEQ_LAB_PARSE_ONLY -eq "1") {
    Write-Host "PARSE_OK feed=$Feed groups=$($selected -join ',') data_dir=$DataDir max_bars=$MaxBars target_bars_per_symbol=$TargetBarsPerSymbol duration=$Duration"
    exit 0
}

& go run ./cmd/generator
if ($LASTEXITCODE -ne 0) {
    throw "generator failed with exit $LASTEXITCODE"
}
