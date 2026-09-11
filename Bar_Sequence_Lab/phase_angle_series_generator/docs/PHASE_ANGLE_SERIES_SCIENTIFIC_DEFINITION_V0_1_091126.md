# Phase Angle Series Scientific Definition

**Date:** 2026-09-11  
**Version:** V0.1  
**Status:** APPROVED FOR BAR SEQUENCE LAB EXPERIMENT  
**Scope:** `Bar_Sequence_Lab / phase_angle_series_generator`  
**DSE Promotion:** NOT AUTHORIZED

## Executive Summary

Lab V0.1 computes John F. Ehlers Dominant Cycle Phase from existing raw Bar
Sequences. The analytical axis is `generator_sequence_no`; the ordered Bar
Sequence is itself the wave. Each `SERIES_SIZE` is an independent experiment,
calculated from a fresh solver receiving only that sequence prefix.

## Frozen Scientific Decisions

1. Input is price only: `Median Price = (High + Low) / 2`.
2. Volume cannot alter price smoothing, I, Q, period, or phase.
3. Solver behavior follows TA-Lib `HT_DCPHASE`, a recognized Ehlers Hilbert
   transform and homodyne-discriminator implementation.
4. The process uses the four-bar weighted smoother, Hilbert coefficients
   `0.0962` and `0.5769`, distinct I/Q paths, period smoothing, and period bounds
   of 6 through 50 bars.
5. Default TA-Lib unstable period is zero, producing a 63-bar lookback. Bars
   1 through 63 are `INITIALIZING`; the first observable output is zero-based
   index 63, corresponding to the 64th selected observation.
6. Stored angles are normalized deterministically into `0 <= angle < 360`.
7. Initialization is BSON null, `phase_observable=false`, and
   `validity_state="INITIALIZING"`; zero is never an unavailable placeholder.
8. Strategy eligibility, Hop semantics, and phase zones are out of scope.

## Source Material Notes

The supplied earlier *Strategy & Implementation Document: Hop-On Hop-Off Phase
Engine* established useful concepts: event-frequency/bar-index space, median
price, four-bar smoothing, Ehlers coefficients, asynchronous ordered bars, and
angular normalization. Its sample Go calculation is not present in this checkout
and is not adopted. In particular, identical FIR expressions for I and Q and any
volume scaling of Q are explicitly rejected.

The independent production-behavior reference is TA-Lib `HT_DCPHASE`, sourced
from the TA-Lib project implementation and regression suite. TA-Lib identifies
John F. Ehlers, *Rocket Science for Traders: Digital Signal Processing
Applications* (ISBN 0471405671). The Go runtime has no TA-Lib or Python dependency.

## Algorithm

For each selected raw bar $n$:

$$P[n] = \frac{High[n] + Low[n]}{2}$$

TA-Lib-compatible processing applies its four-sample weighted smoother, parity
separated Hilbert detrender and quadrature filters, phase-advanced `jI`/`jQ`,
smoothed `I2`/`Q2`, homodyne period discrimination, dominant-period projection,
and dominant-cycle phase calculation. This is not the rejected construction
where Q is a scalar multiple of I.

TA-Lib emits a full 360-degree span represented approximately as `[-45,315]`.
Lab persistence applies modulo normalization only; it does not smooth or reshape
angles for display.

## SERIES_SIZE And Identity

`SERIES_SIZE` is the count of existing ordered observations supplied to one
fresh symbol solver. Short sequences use only available observations. Gaps are
preserved; bars are never fabricated, repeated, interpolated, or renumbered.

Analytical identity is:

```text
collection_run_id + symbol + generator_sequence_no + series_size
+ solver_name + solver_version
```

Identical reruns upsert. Different sizes and solver versions coexist.

## Storage Boundary

Raw source: `bar_sequence_db.bar_sequence`. Derived destination:
`bar_sequence_db.bar_sequence_phase_angle_series`. Derived records contain the
join key, experiment/solver identity, input-series type, phase/null state, and
analysis creation time. They do not duplicate OHLCV, timestamps, source IDs,
payload hashes, or other raw provenance.

Indexes are `ux_phase_run_symbol_sequence_size_solver` (unique) and
`ix_phase_run_partition_size_solver_symbol_sequence` (query-oriented). The
initialization script creates them explicitly and idempotently without deleting
or modifying records.

## Validation And Limitations

Unit tests cover median price, angle normalization, null initialization, the
63-bar transition, deterministic output, and distinct I/Q diagnostics. A frozen
numerical comparison vector from an independently executed TA-Lib build is still
required before promoting V0.1 outside this Lab. SERIES_SIZE=6 is expected to
produce only initialization evidence because it is below the recognized lookback.

This experiment does not amend DSE authority, calculate strategy eligibility,
implement Hop actions, or provide a viewer. Successful Lab evidence may be
proposed later through controlled DSE governance.