# Bar Sequence Lab Phase Angle Series Generator

Standalone Go batch analysis of existing MongoDB Bar Sequences. It discovers
symbols in partitions A/B/C, selects each ordered sequence prefix, computes a
price-only Ehlers dominant-cycle phase series, and upserts lean derived evidence.
The authoritative `bar_sequence` collection is read only.

## Initialize MongoDB

Run `scripts/Initialize-BarSequencePhaseAngleMongo.js` explicitly in mongosh or
MongoDB Compass before the first experiment. The batch program checks for both
required indexes and never initializes schema itself.

## Run

```powershell
.\scripts\Start-PhaseAngleSeriesGenerator.ps1 -SERIES_SIZE 6
.\scripts\Start-PhaseAngleSeriesGenerator.ps1 -SERIES_SIZE 12
.\scripts\Start-PhaseAngleSeriesGenerator.ps1 -SERIES_SIZE 24
.\scripts\Start-PhaseAngleSeriesGenerator.ps1 -SERIES_SIZE 48
```

Mongo settings use `BAR_SEQ_LAB_MONGO_URI`, `BAR_SEQ_LAB_MONGO_DB`,
`BAR_SEQ_LAB_RAW_COLLECTION`, and `BAR_SEQ_LAB_PHASE_COLLECTION`, with local Lab
defaults. `SERIES_SIZE` is always explicit and positive. Each invocation starts a
fresh solver per symbol and supplies no more than that many ordered raw bars.

## Scientific Boundary

Input is median price `(high + low) / 2` over `generator_sequence_no`. Volume is
excluded. The native Go solver follows TA-Lib `HT_DCPHASE` with a 63-bar lookback
and normalizes observable output to `[0, 360)`. Earlier bars persist with a BSON
null angle and `INITIALIZING`; no Hop or strategy interpretation is performed.