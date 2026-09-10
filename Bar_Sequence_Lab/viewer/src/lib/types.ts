/**
 * Portable viewer types for Bar Sequence Lab V0.1.
 * Independent of Fin_FeedSat_1_Viewer and DSE_TransSat_1_viewer.
 * The ordered bar sequence IS the wave. x = generator_sequence_no.
 */

export type TrajectoryType = "price" | "volume";

export type CollectionRunSummary = {
  collectionRunId: string;
  observationCount: number;
  symbolCount: number;
  partitions: string[];
};

export type GroupDescriptor = {
  partitionId: string;
  symbolCount: number;
  observationCount: number;
};

export type SymbolDescriptor = {
  symbol: string;
  partitionId: string;
  observationCount: number;
  minSequence: number;
  maxSequence: number;
};

export type TrajectoryDescriptor = {
  type: TrajectoryType;
  label: string;
  yField: "close" | "volume";
};

export type TrajectoryObservation = {
  generatorSequenceNo: number;
  symbol: string;
  partitionId: string;
  collectionRunId: string;
  y: number;
  open: number;
  high: number;
  low: number;
  close: number;
  volume: number;
  interval: string;
  sourceEventTime: string;
  sourceTimestampText: string;
  receivedTime: string;
  persistedTime: string;
  sourceId: string;
  alpacaMessageType: string;
  payloadHash: string;
  sourceTimeRegression: boolean;
  duplicateArrival: boolean;
  eventCount: number;
};

export type SequenceSummary = {
  symbol: string;
  partitionId: string;
  collectionRunId: string;
  observationCount: number;
  minSequence: number;
  maxSequence: number;
  contiguousFromOne: boolean;
};

export type TrajectorySeries = {
  collectionRunId: string;
  symbol: string;
  partitionId: string;
  trajectory: TrajectoryType;
  observations: TrajectoryObservation[];
  summary: SequenceSummary;
};

export const TRAJECTORIES: TrajectoryDescriptor[] = [
  { type: "price", label: "Price", yField: "close" },
  { type: "volume", label: "Volume", yField: "volume" },
];
