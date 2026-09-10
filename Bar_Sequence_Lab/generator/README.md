# Bar Sequence Lab generator

Experimental raw-bar collector. Not DSE. Not Fin_FeedSat_1.

Persistence: **local JSONL audit is always on**. MongoDB raw persistence is **opt-in** (`BAR_SEQ_LAB_MONGO_ENABLED=true`) into existing `bar_sequence_db.bar_sequence`. JSONL is not removed when Mongo is enabled.

## Start (PowerShell)

From `Bar_Sequence_Lab/generator`:

```powershell
.\scripts\Start-BarSequenceLab.ps1 A
.\scripts\Start-BarSequenceLab.ps1 A B
.\scripts\Start-BarSequenceLab.ps1 -Feed iex A B C
```

No groups prints usage and exits. There is no default of A+B+C.

Requires `ALPACA_API_KEY` and `ALPACA_API_SECRET` (or `ALPACA_SECRET_KEY`) in the session. The script does not print secrets and does not start Fin.

Mongo (opt-in):

- `BAR_SEQ_LAB_MONGO_ENABLED=true`
- `BAR_SEQ_LAB_MONGO_URI` (host/auth only; do not print)
- `BAR_SEQ_LAB_MONGO_DB=bar_sequence_db`
- `BAR_SEQ_LAB_MONGO_COLLECTION=bar_sequence` (optional; this default)

## Group lists

Lab-owned (comma-separated, Fin-style):

- `BAR_SEQ_LAB_GROUP_A` (test-feed default `FAKEPACA`; iex default `AAPL`)
- `BAR_SEQ_LAB_GROUP_B` (test-feed default empty; iex default `MSFT`)
- `BAR_SEQ_LAB_GROUP_C` (test-feed default empty; iex default `NVDA`)

Duplicate symbols across **selected** groups fail startup. Empty selected groups fail startup.

## JSONL layout

`data\<collection_run_id>\partition_{A|B|C}.jsonl` for **active** partitions only.
