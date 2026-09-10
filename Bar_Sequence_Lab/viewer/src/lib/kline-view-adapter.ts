/**
 * KlineChartsViewAdapter
 *
 * klinecharts is a rendering mechanism only. It does NOT redefine the
 * analytical axis. For Bar Sequence Lab:
 *
 *   x semantics = generator_sequence_no  (BAR-INDEX REPLAY)
 *
 * The library requires a time-like `timestamp` field. This adapter maps
 * ordered bar index into that field. It does NOT use source_event_time
 * as the chart coordinate and does NOT fabricate market timestamps.
 *
 * Raw Mongo observations are never mutated. Extra KLineData keys carry
 * provenance for inspection. Future derived overlays (Ehlers / Hop)
 * must be separate overlay layers, not replacements of this mapping.
 */

import type { KLineData } from "klinecharts";
import type { TrajectoryObservation, TrajectoryType } from "./types";

export type LabKLineData = KLineData & {
  generatorSequenceNo: number;
  sourceEventTime: string;
  rawOpen: number;
  rawHigh: number;
  rawLow: number;
  rawClose: number;
  rawVolume: number;
  symbol: string;
  partitionId: string;
  collectionRunId: string;
  interval: string;
  receivedTime: string;
  persistedTime: string;
  sourceId: string;
  payloadHash: string;
  sourceTimeRegression: boolean;
  duplicateArrival: boolean;
  eventCount: number;
  trajectory: TrajectoryType;
};

/** Chart coordinate equals bar index. Not milliseconds. Not wall clock. */
export function sequenceToChartTimestamp(generatorSequenceNo: number): number {
  return generatorSequenceNo;
}

export function chartTimestampToSequence(timestamp: number | undefined | null): number {
  if (typeof timestamp !== "number" || !Number.isFinite(timestamp)) return 0;
  return timestamp;
}

export function observationToKLine(
  observation: TrajectoryObservation,
  trajectory: TrajectoryType,
): LabKLineData {
  const plotted = trajectory === "volume" ? observation.volume : observation.close;
  return {
    timestamp: sequenceToChartTimestamp(observation.generatorSequenceNo),
    open: trajectory === "volume" ? plotted : observation.open,
    high: trajectory === "volume" ? plotted : observation.high,
    low: trajectory === "volume" ? plotted : observation.low,
    close: plotted,
    volume: observation.volume,
    generatorSequenceNo: observation.generatorSequenceNo,
    sourceEventTime: observation.sourceEventTime,
    rawOpen: observation.open,
    rawHigh: observation.high,
    rawLow: observation.low,
    rawClose: observation.close,
    rawVolume: observation.volume,
    symbol: observation.symbol,
    partitionId: observation.partitionId,
    collectionRunId: observation.collectionRunId,
    interval: observation.interval,
    receivedTime: observation.receivedTime,
    persistedTime: observation.persistedTime,
    sourceId: observation.sourceId,
    payloadHash: observation.payloadHash,
    sourceTimeRegression: observation.sourceTimeRegression,
    duplicateArrival: observation.duplicateArrival,
    eventCount: observation.eventCount,
    trajectory,
  };
}

export function observationsToKLines(
  observations: TrajectoryObservation[],
  trajectory: TrajectoryType,
): LabKLineData[] {
  return observations.map((row) => observationToKLine(row, trajectory));
}

export function formatBarIndexTick(timestamp: number): string {
  return String(Math.round(timestamp));
}

export function observationAtChartTimestamp(
  observations: TrajectoryObservation[],
  timestamp: number | undefined | null,
): TrajectoryObservation | null {
  const seq = chartTimestampToSequence(timestamp ?? undefined);
  if (!seq) return null;
  return observations.find((row) => row.generatorSequenceNo === seq) ?? null;
}

export function klineToObservation(
  data: LabKLineData | KLineData | undefined,
  fallback: TrajectoryObservation | null,
  observations: TrajectoryObservation[] = [],
): TrajectoryObservation | null {
  if (!data) return fallback;
  const fromStore = observationAtChartTimestamp(observations, data.timestamp);
  if (fromStore) return fromStore;
  const lab = data as LabKLineData;
  if (typeof lab.generatorSequenceNo !== "number") return fallback;
  return {
    generatorSequenceNo: lab.generatorSequenceNo,
    symbol: lab.symbol,
    partitionId: lab.partitionId,
    collectionRunId: lab.collectionRunId,
    y: lab.close,
    open: lab.rawOpen,
    high: lab.rawHigh,
    low: lab.rawLow,
    close: lab.rawClose,
    volume: lab.rawVolume,
    interval: lab.interval,
    sourceEventTime: lab.sourceEventTime,
    sourceTimestampText: fallback?.sourceTimestampText ?? "",
    receivedTime: lab.receivedTime,
    persistedTime: lab.persistedTime,
    sourceId: lab.sourceId,
    alpacaMessageType: fallback?.alpacaMessageType ?? "",
    payloadHash: lab.payloadHash,
    sourceTimeRegression: lab.sourceTimeRegression,
    duplicateArrival: lab.duplicateArrival,
    eventCount: lab.eventCount,
  };
}
