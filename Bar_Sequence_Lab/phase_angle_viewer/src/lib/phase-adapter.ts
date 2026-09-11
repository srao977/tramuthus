import type { KLineData } from "klinecharts";
import type { PhasePoint } from "./types";

export type PhaseKLineData = KLineData & PhasePoint;

/** Maps the analytical bar index directly to the chart coordinate. */
export function sequenceToChartCoordinate(generatorSequenceNo: number): number {
  return generatorSequenceNo;
}

/** Converts only persisted numeric observable phase records into chart data. */
export function phasePointsToKLines(points: PhasePoint[]): PhaseKLineData[] {
  return points.map((point) => ({
    timestamp: sequenceToChartCoordinate(point.generatorSequenceNo),
    open: point.phaseAngleDegrees,
    high: point.phaseAngleDegrees,
    low: point.phaseAngleDegrees,
    close: point.phaseAngleDegrees,
    volume: 0,
    ...point,
  }));
}

export function phasePointAtCoordinate(
  points: PhasePoint[],
  coordinate: number | null | undefined,
): PhasePoint | null {
  if (typeof coordinate !== "number" || !Number.isFinite(coordinate)) return null;
  return points.find((point) => point.generatorSequenceNo === coordinate) ?? null;
}

export function klineToPhasePoint(
  data: PhaseKLineData | KLineData | undefined,
  points: PhasePoint[],
): PhasePoint | null {
  if (!data) return null;
  const fromCoordinate = phasePointAtCoordinate(points, data.timestamp);
  if (fromCoordinate) return fromCoordinate;
  const phaseData = data as Partial<PhaseKLineData>;
  if (typeof phaseData.generatorSequenceNo !== "number") return null;
  return phasePointAtCoordinate(points, phaseData.generatorSequenceNo);
}

export function formatBarIndexTick(coordinate: number): string {
  return String(Math.round(coordinate));
}

export function fittedBarSpace(chartWidth: number, pointCount: number): number {
  if (!Number.isFinite(chartWidth) || chartWidth <= 0 || pointCount <= 0) return 6;
  const plottingWidth = Math.max(chartWidth - 76, 1);
  return Math.max(4, Math.min(24, plottingWidth / pointCount));
}
