/** Presentation contracts for persisted phase evidence and lineage lookups. */
export type ExperimentQuery = {
  collectionRunId: string;
  seriesSize: number;
  solverName: string;
  solverVersion: string;
  inputSeriesType: string;
};

export type ExperimentMetadata = ExperimentQuery & {
  recordCount: number;
  observableCount: number;
  initializingCount: number;
  symbolCount: number;
  partitions: string[];
};

export type PhaseSymbol = {
  partitionId: string;
  symbol: string;
  recordCount: number;
  observableCount: number;
  initializingCount: number;
  minSequence: number;
  maxSequence: number;
};

export type PhasePoint = {
  collectionRunId: string;
  partitionId: string;
  symbol: string;
  generatorSequenceNo: number;
  seriesSize: number;
  solverName: string;
  solverVersion: string;
  inputSeriesType: string;
  phaseAngleDegrees: number;
  phaseObservable: true;
  validityState: string;
};

export type PhaseSeries = {
  experiment: ExperimentQuery;
  partitionId: string;
  symbol: string;
  recordCount: number;
  observableCount: number;
  initializingCount: number;
  points: PhasePoint[];
};

export type RawObservation = {
  collectionRunId: string;
  partitionId: string;
  symbol: string;
  generatorSequenceNo: number;
  interval?: string;
  open?: number;
  high?: number;
  low?: number;
  close?: number;
  volume?: number;
  eventCount?: number;
  sourceEventTime?: string;
  sourceTimestampText?: string;
  receivedTime?: string;
  persistedTime?: string;
  sourceId?: string;
  alpacaMessageType?: string;
  payloadHash?: string;
  sourceTimeRegression?: boolean;
  duplicateArrival?: boolean;
};

export const DEFAULT_EXPERIMENT: ExperimentQuery = {
  collectionRunId: "20260911T161623Z-1",
  seriesSize: 120,
  solverName: "EHLERS_DOMINANT_CYCLE_PHASE",
  solverVersion: "V0.1",
  inputSeriesType: "MEDIAN_PRICE",
};
