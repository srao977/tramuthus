import type {
  CollectionRunSummary,
  GroupDescriptor,
  SymbolDescriptor,
  TrajectorySeries,
  TrajectoryType,
} from "./types";

/** Portable data-provider contract. Lab Mongo is one adapter. */
export type BarSequenceDataProvider = {
  listRuns(): Promise<CollectionRunSummary[]>;
  listGroups(collectionRunId: string): Promise<GroupDescriptor[]>;
  listSymbols(collectionRunId: string, group?: string): Promise<SymbolDescriptor[]>;
  getTrajectory(
    collectionRunId: string,
    symbol: string,
    trajectory: TrajectoryType,
  ): Promise<TrajectorySeries>;
};
