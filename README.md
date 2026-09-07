# Fin_FeedSat_1

Fin_FeedSat_1 is the first Finance Domain FeedSat derived from the authoritative QuanTRAM baseline. It preserves the existing QuanTRAM scientific/runtime behavior and adds only the minimal satellite boundary identity required for HACCAM participation: `FeedSat` / `PRODUCER` / `Finance`.

The runtime identity is configured through `FIN_FEEDSAT_ID`, `FIN_FEEDSAT_ROLE`, `FIN_FEEDSAT_SUBSCRIBER_TYPE`, and `FIN_FEEDSAT_DOMAIN_PLANE`. Defaults are `Fin_FeedSat_1`, `FeedSat`, `PRODUCER`, and `Finance`.

Quantitative Trading Adaptive Model. Increment 1 remains the Go data-ingestion layer: Alpaca IEX (or test) WebSocket, REST gap-fill, optional CSV replay, and gRPC bar/health APIs.

Do not commit Alpaca keys. Use environment variables. Rotate any key that appeared in local notes or PDFs.

## Generate, test, build

```powershell
buf lint
buf generate
go test ./...
go build ./cmd/quantram-server ./cmd/quantram-ingest-client
```

## CSV validation (anytime)

```powershell
$env:FIN_FEEDSAT_SOURCE="csv"
$env:FIN_FEEDSAT_CSV_PATH="AAPL_1min_firstratedata.csv"
$env:FIN_FEEDSAT_SYMBOLS="AAPL"
$env:FIN_FEEDSAT_MODEL="off"
$env:FIN_FEEDSAT_PRICING="off"
go run ./cmd/quantram-server
```

In another terminal:

```powershell
go run ./cmd/quantram-ingest-client -operation stream -symbols AAPL -max-bars 5
go run ./cmd/quantram-ingest-client -operation health
go run ./cmd/quantram-ingest-client -operation window -symbols AAPL
go run ./cmd/quantram-ingest-client -operation decisions -symbols AAPL -max-bars 5
```

`StreamDecisions` requires `FIN_FEEDSAT_MODEL=adaptive`. Default is `off` (`FailedPrecondition`).

`StreamPriceEvents` requires `FIN_FEEDSAT_PRICING=expm` **and** `FIN_FEEDSAT_MODEL=adaptive`. Default pricing is `off` (`FailedPrecondition`). Unknown `FIN_FEEDSAT_PRICING` fails startup. Price Engine color needs 45 consecutive accepted eligible minutes after a cold start. After the 2 Sep continuity fix, a skipped IEX no-trade minute is an irregular interval, not `INPUT_GAP`. `INPUT_GAP` is reserved for proven missing eligible bars.

`SemanticService` (`GetTerm`, `ListTerms`, `GetSemanticContract`) is read-only. Contract: `internal/semantics/data/quantram_semantics_v1.json` version `1.0`. Authoring source: `internal/semantics/catalog/v1.go`. Tooling: `go run ./cmd/quantram-semantics validate|audit|build`. See [Semantic Contract V1](docs/design/QuanTRAM_SEMANTIC_CONTRACT_V1_090226.md).

StageTransition V1.1 is sideways publication when meaningful P-01–P-04 state changes. Bar-driven P-03/P-04 events carry a value copy of the accepted `domain.Bar`. Diagnostic subscriber writes `./stage_transitions.txt` (git-ignored). Optional `FIN_FEEDSAT_STAGE_TRANSITION_LOG`. See [Stage Transition Publication](docs/design/QuanTRAM_STAGE_TRANSITION_PUBLICATION_V1_2026-09-04.md).

## Alpaca test feed (outside regular hours)

```powershell
$env:FIN_FEEDSAT_SOURCE="alpaca"
$env:FIN_FEEDSAT_FEED="test"
$env:FIN_FEEDSAT_SYMBOLS="FAKEPACA"
$env:ALPACA_API_KEY="..."
$env:ALPACA_API_SECRET="..."
go run ./cmd/quantram-server
```

```powershell
go run ./cmd/quantram-ingest-client -operation source
go run ./cmd/quantram-ingest-client -operation stream -symbols FAKEPACA -max-bars 3 -timeout 4m
```

## Alpaca IEX (regular hours, Basic plan)

```powershell
$env:FIN_FEEDSAT_SOURCE="alpaca"
$env:FIN_FEEDSAT_FEED="iex"
$env:FIN_FEEDSAT_SYMBOLS="AAPL"
$env:ALPACA_API_KEY="..."
$env:ALPACA_API_SECRET="..."
go run ./cmd/quantram-server
```

Paper trading is not started in this increment.
