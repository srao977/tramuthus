# Fin_FeedSat_1 Startup Guide V1

**Date:** 2026-09-09  
**Audience:** Non-technical financial analysts and operators  
**Script:** `scripts/Start-FinFeedSatIngestion.ps1`

## Overview

This guide explains how to start Fin_FeedSat_1 ingestion from one PowerShell window, confirm that Alpaca market data is being received, understand the normal messages printed by the system, and stop or restart the service safely.

The startup script now manages both parts of the local session:

1. It starts the Fin_FeedSat_1 Go server.
2. It waits for the feed to become healthy.
3. It checks service readiness.
4. It runs the selected monitoring operation in the same terminal.
5. It stops the server and its child processes when the session ends.

No second terminal is required for normal operation.

## Before Starting

Open PowerShell in the repository root:

```powershell
Set-Location C:\Users\chino\Fin_FeedSat_1
```

For the Alpaca live path, the following environment variables must already be set in the current PowerShell session:

- `ALPACA_API_KEY`
- `ALPACA_API_SECRET`

`ALPACA_SECRET_KEY` is accepted as an alternative name for the secret.

The variables must be available in the same terminal where the script is started. If they were created with `setx`, open a new PowerShell window before running the script. The script checks that credentials exist but does not print them.

The Go command must also be available on `PATH`:

```powershell
go version
```

## Normal Live Startup

For the regular Alpaca IEX feed and several symbols, run:

```powershell
.\scripts\Start-FinFeedSatIngestion.ps1 `
  -Feed iex `
  -Symbols AAPL,MSFT,NVDA,AMZN
```

The default source is Alpaca. The default port is `50051`.

Normal mode continuously receives live data. The script uses a long client timeout so the session remains active until the operator stops it.

To run one symbol:

```powershell
.\scripts\Start-FinFeedSatIngestion.ps1 -Feed iex -Symbols AAPL
```

The configuration supports up to 30 symbols.

## Smoke Test

A smoke test validates startup without leaving the server running:

```powershell
.\scripts\Start-FinFeedSatIngestion.ps1 `
  -Feed iex `
  -Symbols AAPL `
  -SmokeTest
```

The smoke test:

1. Starts the server.
2. Waits up to 30 seconds for a healthy feed state.
3. Checks readiness.
4. Receives one bar by default.
5. Prints `Smoke test passed.` when successful.
6. Stops the server automatically.

A successful run ends with output similar to:

```text
source=ALPACA_IEX state=FEED_STATE_HEALTHY symbols=AAPL last_error=""
ready=true observe=true infer=false message="observable; inference gated on data quality"
symbol=AAPL source=ALPACA_IEX ... final=true backfill=false quality=QUALITY_STATUS_COMPLETE
received=1
Smoke test passed.
```

## What Successful Startup Looks Like

The following messages are normal and indicate that the service is running:

```text
runtime identity satellite_id=Fin_FeedSat_1 ...
semantic contract 1.0 (107 terms)
starting Fin_FeedSat_1 ingestion gRPC server ... port=50051 source=alpaca feed=iex symbols=[AAPL]
alpaca subscribed type=subscription bars=[AAPL]
```

The most important confirmation is:

```text
source=ALPACA_IEX state=FEED_STATE_HEALTHY symbols=AAPL last_error=""
```

For multiple symbols, the `symbols=` field should list all requested symbols. Alpaca delivers bars asynchronously, so messages for AAPL, MSFT, NVDA, and AMZN will not necessarily appear at the same instant.

A received bar looks similar to:

```text
symbol=AAPL source=ALPACA_IEX ts=2026-09-09T15:07:00Z ... final=true backfill=false quality=QUALITY_STATUS_COMPLETE
```

This confirms that a finalized, complete bar was received from Alpaca. `backfill=false` means it came from the live stream rather than historical recovery.

## Readiness and Model Messages

Readiness may display:

```text
ready=true observe=true infer=false message="observable; inference gated on data quality"
```

This means the ingestion and observation paths are working. `infer=false` is a model-quality gate; it does not mean that Alpaca ingestion failed.

During adaptive model startup, the following is normal:

```text
model skip symbol=MSFT reason=INITIALIZING interval=...
```

`INITIALIZING` means that symbol is building its first 15 accepted eligible minutes. The 16th accepted eligible bar can produce an actionable `BUY`, `SELL`, or `HOLD` result. Each symbol warms up independently, and symbols may produce these messages at different times.

Other normal model messages include:

```text
model skip symbol=AAPL reason=INFER_OFF interval=...
model skip symbol=AAPL reason=INITIALIZING interval=...
volume maturing symbol=AAPL ...
```

These are downstream processing or warm-up messages, not evidence that the feed is disconnected.

## Viewing Model Decisions

The default operation displays incoming bars. To display model events instead, stop the current run and restart it with:

```powershell
.\scripts\Start-FinFeedSatIngestion.ps1 `
  -Feed iex `
  -Symbols AAPL,MSFT,NVDA,AMZN `
  -Operation decisions
```

This displays per-symbol events such as `INITIALIZING`, `INFER_OFF`, and actionable decisions.

Other supported operations are `health`, `ready`, `source`, `window`, `gapfill`, and `stream`.

## Stopping the Server

Press `Ctrl+C` once in the PowerShell window running the script.

The script will stop the client and terminate the Go server process tree. The PowerShell prompt should return. A few buffered log lines may appear while the process is closing; wait for the prompt before restarting.

Do not close the window repeatedly or start another copy while the first process is still shutting down.

## Restarting the Server

After the PowerShell prompt returns, run the startup command again:

```powershell
.\scripts\Start-FinFeedSatIngestion.ps1 `
  -Feed iex `
  -Symbols AAPL,MSFT,NVDA,AMZN
```

Each restart begins a new adaptive-model warm-up. Seeing `INITIALIZING` again after a restart is expected.

## Common Failure Messages

### Port 50051 is already in use

Message:

```text
listen on port 50051: listen tcp :50051: bind: Only one usage of each socket address ...
```

This means another Fin_FeedSat_1 server is still using the port. First wait for the previous PowerShell window to return to its prompt. If the old process remains, run this cleanup command once:

```powershell
Get-CimInstance Win32_Process |
  Where-Object { $_.CommandLine -match "fin-feedsat-server" } |
  ForEach-Object { taskkill.exe /PID $_.ProcessId /T /F }
```

Then start the script again. Do not run two ingestion servers on the same port.

As a temporary alternative, use another port:

```powershell
.\scripts\Start-FinFeedSatIngestion.ps1 `
  -Feed iex `
  -Symbols AAPL `
  -Port 50052
```

### Go is not on PATH

Message:

```text
Go is not on PATH. Install Go or open a shell where 'go version' works.
```

Install Go or open a PowerShell session where `go version` succeeds, then rerun the script.

### Missing Alpaca credentials

Message:

```text
Set ALPACA_API_KEY and ALPACA_API_SECRET (or ALPACA_SECRET_KEY) in this session first.
```

Set the required variables in the current PowerShell session. Do not put credentials in this guide, the script, or source control.

### Server did not become healthy

Message:

```text
server did not become HEALTHY within 30 seconds
```

Check the preceding server output. Common causes are an unavailable network connection, invalid Alpaca credentials, an occupied port, or an Alpaca connection failure. Confirm that the selected feed is `iex`, the market connection is available, and no older server is running.

### Alpaca connection failure or recovery

The server may log connection or recovery messages while attempting to reconnect. Confirm the current state with a smoke test or the `source` operation after the connection has had time to recover. A healthy result has:

```text
state=FEED_STATE_HEALTHY
```

### Invalid symbols or parameters

The script rejects unsupported values for source, feed, and operation. Symbols must contain at least one non-empty value. Correct the command and rerun it.

## Local CSV Test

To test the local ingestion path without Alpaca credentials or live market data:

```powershell
.\scripts\Start-FinFeedSatIngestion.ps1 -Source csv -SmokeTest
```

The default CSV file is `AAPL_1min_firstratedata.csv` in the repository root. CSV mode is useful for local pipeline validation but does not confirm live Alpaca connectivity.

## Operator Checklist

Before starting:

- Confirm the current directory is `C:\Users\chino\Fin_FeedSat_1`.
- Confirm credentials exist in the current PowerShell session.
- Confirm `go version` works.
- Confirm no older ingestion server is still running.

After starting:

- Confirm `FEED_STATE_HEALTHY`.
- Confirm the requested symbols appear in the source status.
- Confirm finalized bars show `source=ALPACA_IEX`, `final=true`, and `quality=QUALITY_STATUS_COMPLETE`.
- Treat `INITIALIZING`, `INFER_OFF`, and `volume maturing` as model or data-quality status messages unless feed health also reports a failure.

When finished:

- Press `Ctrl+C` once.
- Wait for the PowerShell prompt to return.
- Restart only after the previous session has stopped.
