# Bar Sequence Lab Generator Provenance

**Date:** 2026-09-10  
**Component:** Bar_Sequence_Lab/generator  
**Tramuthus HEAD at implementation:** `70d2e6afdc2981bcd049976d46e7e54b28565f9d`  
**Standalone Fin_FeedSat_1 git HEAD (read-only reference tree):** `3d763a15d07945a9077dc6bae576027f6503fef8`

Copied/adapted Alpaca live-ingestion mechanics only. After adaptation the lab module is independent. There is **no** Go `replace` or import of `fin_feedsat_1`. Fin science (Price/Volume/model/gap-fill/gRPC/CSV/drop-on-slow-subscriber) was not copied.

## Fin files used as reference

| Path | Use |
| --- | --- |
| Fin_FeedSat_1/internal/marketfeed/alpaca_ws.go | Adapt WS dial, JSON auth, subscribe `bars`+`updatedBars`, ping/pong heartbeat, reconnect 100ms–30s, read loop |
| Fin_FeedSat_1/internal/marketfeed/decode.go | Adapt array decode, control decode, T=b/T=u, OHLC/volume/timestamp parse |
| Fin_FeedSat_1/internal/marketfeed/source.go | Adapt source id concept `ALPACA_IEX` / `ALPACA_TEST` only |
| Fin_FeedSat_1/internal/config/config.go | Adapt URLs, heartbeat/reconnect constants, MaxSymbolsBasicPlan=30 operational, comma-separated symbol split, ALPACA_API_KEY / ALPACA_API_SECRET or ALPACA_SECRET_KEY |
| Fin_FeedSat_1/cmd/fin-feedsat-server/main.go | Terminal identity/startup/shutdown reporting quality (not gRPC/model) |
| C:\Users\chino\Fin_FeedSat_1\scripts\Start-FinFeedSatIngestion.ps1 | READ-ONLY analog for lab PowerShell startup (no group A/B/C in Fin; lab-owned groups) |

## Copied vs adapted

- **Copied conceptually:** Alpaca WS URL, auth JSON `action/key/secret`, subscribe bars+updatedBars, T=b and T=u, reconnect backoff, heartbeat, credential env names, comma-separated symbol lists.
- **Adapted / lab-owned:** observation type, accepted-arrival sequence, partitions A/B/C, three-buffer pipeline, JSONL writers, PowerShell remaining-args groups, no Fin pipeline drop-on-full, no CSV, no REST gap fill, no gRPC.
- **Not copied:** Price Engine, Volume Engine, adaptive model, stage-transition science, subscriber queue drop, Alpaca REST, CSV source, protobuf.

## Integrity statements

- `Fin_FeedSat_1/**` was not modified (tramuthus copy or standalone tree).
- `Bar_Sequence_Lab/viewer` was not modified.
- DSE documents and code were not modified.
- Approved V0.1 plan was not rewritten. JSONL is a proving-mode persistence deviation from the plan’s Mongo initial store; Mongo was not installed, created, or connected.
