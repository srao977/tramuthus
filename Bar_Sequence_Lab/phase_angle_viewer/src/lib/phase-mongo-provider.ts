import "server-only";

import type { Document, Filter } from "mongodb";
import { getPhaseCollection, getRawCollection } from "./mongo";
import type {
  ExperimentMetadata,
  ExperimentQuery,
  PhasePoint,
  PhaseSeries,
  PhaseSymbol,
  RawObservation,
} from "./types";

function asString(value: unknown): string {
  return typeof value === "string" ? value : value == null ? "" : String(value);
}

function asNumber(value: unknown): number {
  if (typeof value === "number") return value;
  if (value && typeof value === "object" && "toString" in value) {
    const parsed = Number(String(value));
    return Number.isFinite(parsed) ? parsed : 0;
  }
  const parsed = Number(value);
  return Number.isFinite(parsed) ? parsed : 0;
}

function asIso(value: unknown): string | undefined {
  if (value instanceof Date) return value.toISOString();
  if (typeof value === "string") return value;
  return undefined;
}

function experimentFilter(experiment: ExperimentQuery): Filter<Document> {
  return {
    collection_run_id: experiment.collectionRunId,
    series_size: experiment.seriesSize,
    solver_name: experiment.solverName,
    solver_version: experiment.solverVersion,
    input_series_type: experiment.inputSeriesType,
  };
}

function mapExperimentId(id: Document, row: Document): ExperimentMetadata {
  return {
    collectionRunId: asString(id.collectionRunId),
    seriesSize: asNumber(id.seriesSize),
    solverName: asString(id.solverName),
    solverVersion: asString(id.solverVersion),
    inputSeriesType: asString(id.inputSeriesType),
    recordCount: asNumber(row.recordCount),
    observableCount: asNumber(row.observableCount),
    initializingCount: asNumber(row.initializingCount),
    symbolCount: Array.isArray(row.symbols) ? row.symbols.length : 0,
    partitions: Array.isArray(row.partitions)
      ? row.partitions.map(asString).sort()
      : [],
  };
}

export const phaseMongoProvider = {
  async listExperiments(): Promise<ExperimentMetadata[]> {
    const collection = await getPhaseCollection();
    const rows = await collection
      .aggregate<Document>([
        {
          $group: {
            _id: {
              collectionRunId: "$collection_run_id",
              seriesSize: "$series_size",
              solverName: "$solver_name",
              solverVersion: "$solver_version",
              inputSeriesType: "$input_series_type",
            },
            recordCount: { $sum: 1 },
            observableCount: { $sum: { $cond: ["$phase_observable", 1, 0] } },
            initializingCount: { $sum: { $cond: ["$phase_observable", 0, 1] } },
            symbols: { $addToSet: "$symbol" },
            partitions: { $addToSet: "$partition_id" },
          },
        },
        { $sort: { "_id.collectionRunId": -1, "_id.seriesSize": -1 } },
      ])
      .toArray();
    return rows.map((row) => mapExperimentId((row._id ?? {}) as Document, row));
  },

  async listSymbols(experiment: ExperimentQuery): Promise<PhaseSymbol[]> {
    const collection = await getPhaseCollection();
    const rows = await collection
      .aggregate<Document>([
        { $match: experimentFilter(experiment) },
        {
          $group: {
            _id: { partitionId: "$partition_id", symbol: "$symbol" },
            recordCount: { $sum: 1 },
            observableCount: { $sum: { $cond: ["$phase_observable", 1, 0] } },
            initializingCount: { $sum: { $cond: ["$phase_observable", 0, 1] } },
            minSequence: { $min: "$generator_sequence_no" },
            maxSequence: { $max: "$generator_sequence_no" },
          },
        },
        { $sort: { "_id.partitionId": 1, "_id.symbol": 1 } },
      ])
      .toArray();
    return rows.map((row) => {
      const id = (row._id ?? {}) as Document;
      return {
        partitionId: asString(id.partitionId),
        symbol: asString(id.symbol),
        recordCount: asNumber(row.recordCount),
        observableCount: asNumber(row.observableCount),
        initializingCount: asNumber(row.initializingCount),
        minSequence: asNumber(row.minSequence),
        maxSequence: asNumber(row.maxSequence),
      };
    });
  },

  async getSeries(
    experiment: ExperimentQuery,
    partitionId: string,
    symbol: string,
  ): Promise<PhaseSeries> {
    const collection = await getPhaseCollection();
    const filter = { ...experimentFilter(experiment), partition_id: partitionId, symbol };
    const docs = await collection
      .find(filter, {
        projection: {
          _id: 0,
          collection_run_id: 1,
          partition_id: 1,
          symbol: 1,
          generator_sequence_no: 1,
          series_size: 1,
          solver_name: 1,
          solver_version: 1,
          input_series_type: 1,
          phase_angle_degrees: 1,
          phase_observable: 1,
          validity_state: 1,
        },
        sort: { generator_sequence_no: 1 },
      })
      .toArray();

    const points: PhasePoint[] = docs.flatMap((doc) => {
      if (doc.phase_observable !== true || typeof doc.phase_angle_degrees !== "number") return [];
      return [{
        collectionRunId: asString(doc.collection_run_id),
        partitionId: asString(doc.partition_id),
        symbol: asString(doc.symbol),
        generatorSequenceNo: asNumber(doc.generator_sequence_no),
        seriesSize: asNumber(doc.series_size),
        solverName: asString(doc.solver_name),
        solverVersion: asString(doc.solver_version),
        inputSeriesType: asString(doc.input_series_type),
        phaseAngleDegrees: doc.phase_angle_degrees,
        phaseObservable: true as const,
        validityState: asString(doc.validity_state),
      }];
    });

    return {
      experiment,
      partitionId,
      symbol,
      recordCount: docs.length,
      observableCount: points.length,
      initializingCount: docs.length - points.length,
      points,
    };
  },

  async getRawObservation(
    collectionRunId: string,
    symbol: string,
    generatorSequenceNo: number,
  ): Promise<RawObservation | null> {
    const collection = await getRawCollection();
    const doc = await collection.findOne(
      { collection_run_id: collectionRunId, symbol, generator_sequence_no: generatorSequenceNo },
      { projection: { _id: 0 } },
    );
    if (!doc) return null;
    return {
      collectionRunId: asString(doc.collection_run_id),
      partitionId: asString(doc.partition_id),
      symbol: asString(doc.symbol),
      generatorSequenceNo: asNumber(doc.generator_sequence_no),
      interval: asString(doc.interval) || undefined,
      open: doc.open == null ? undefined : asNumber(doc.open),
      high: doc.high == null ? undefined : asNumber(doc.high),
      low: doc.low == null ? undefined : asNumber(doc.low),
      close: doc.close == null ? undefined : asNumber(doc.close),
      volume: doc.volume == null ? undefined : asNumber(doc.volume),
      eventCount: doc.event_count == null ? undefined : asNumber(doc.event_count),
      sourceEventTime: asIso(doc.source_event_time),
      sourceTimestampText: asString(doc.source_timestamp_text) || undefined,
      receivedTime: asIso(doc.received_time),
      persistedTime: asIso(doc.persisted_time),
      sourceId: asString(doc.source_id) || undefined,
      alpacaMessageType: asString(doc.alpaca_message_type) || undefined,
      payloadHash: asString(doc.payload_hash) || undefined,
      sourceTimeRegression: Boolean(doc.source_time_regression),
      duplicateArrival: Boolean(doc.duplicate_arrival),
    };
  },
};
