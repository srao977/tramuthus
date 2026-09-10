import "server-only";

import type { Document } from "mongodb";
import { getBarSequenceCollection } from "./mongo";
import { asIso, asNumber, newestCollectionRunId, sequencesContiguousFromOne } from "./numbers";
import type { BarSequenceDataProvider } from "./provider";
import type {
  CollectionRunSummary,
  GroupDescriptor,
  SequenceSummary,
  SymbolDescriptor,
  TrajectoryObservation,
  TrajectoryType,
} from "./types";

type RawObservation = {
  collection_run_id?: unknown;
  partition_id?: unknown;
  symbol?: unknown;
  generator_sequence_no?: unknown;
  source_event_time?: unknown;
  source_timestamp_text?: unknown;
  received_time?: unknown;
  persisted_time?: unknown;
  interval?: unknown;
  open?: unknown;
  high?: unknown;
  low?: unknown;
  close?: unknown;
  volume?: unknown;
  event_count?: unknown;
  source_id?: unknown;
  alpaca_message_type?: unknown;
  payload_hash?: unknown;
  source_time_regression?: unknown;
  duplicate_arrival?: unknown;
};

function asString(value: unknown): string {
  return typeof value === "string" ? value : value == null ? "" : String(value);
}

function mapObservation(doc: RawObservation, trajectory: TrajectoryType): TrajectoryObservation {
  const close = asNumber(doc.close);
  const volume = asNumber(doc.volume);
  return {
    generatorSequenceNo: asNumber(doc.generator_sequence_no),
    symbol: asString(doc.symbol),
    partitionId: asString(doc.partition_id),
    collectionRunId: asString(doc.collection_run_id),
    y: trajectory === "volume" ? volume : close,
    open: asNumber(doc.open),
    high: asNumber(doc.high),
    low: asNumber(doc.low),
    close,
    volume,
    interval: asString(doc.interval),
    sourceEventTime: asIso(doc.source_event_time),
    sourceTimestampText: asString(doc.source_timestamp_text),
    receivedTime: asIso(doc.received_time),
    persistedTime: asIso(doc.persisted_time),
    sourceId: asString(doc.source_id),
    alpacaMessageType: asString(doc.alpaca_message_type),
    payloadHash: asString(doc.payload_hash),
    sourceTimeRegression: Boolean(doc.source_time_regression),
    duplicateArrival: Boolean(doc.duplicate_arrival),
    eventCount: asNumber(doc.event_count),
  };
}

function summarize(obs: TrajectoryObservation[]): SequenceSummary {
  const seqs = obs.map((row) => row.generatorSequenceNo);
  const first = obs[0];
  return {
    symbol: first?.symbol ?? "",
    partitionId: first?.partitionId ?? "",
    collectionRunId: first?.collectionRunId ?? "",
    observationCount: obs.length,
    minSequence: seqs.length ? Math.min(...seqs) : 0,
    maxSequence: seqs.length ? Math.max(...seqs) : 0,
    contiguousFromOne: sequencesContiguousFromOne(seqs),
  };
}

export const labMongoProvider: BarSequenceDataProvider = {
  async listRuns() {
    const coll = await getBarSequenceCollection();
    const rows = await coll
      .aggregate<Document>([
        {
          $group: {
            _id: "$collection_run_id",
            observationCount: { $sum: 1 },
            symbols: { $addToSet: "$symbol" },
            partitions: { $addToSet: "$partition_id" },
          },
        },
      ])
      .toArray();

    const summaries: CollectionRunSummary[] = rows.map((row) => ({
      collectionRunId: asString(row._id),
      observationCount: asNumber(row.observationCount),
      symbolCount: Array.isArray(row.symbols) ? row.symbols.length : 0,
      partitions: Array.isArray(row.partitions)
        ? row.partitions.map((p) => asString(p)).sort()
        : [],
    }));

    summaries.sort((a, b) => (a.collectionRunId < b.collectionRunId ? 1 : -1));
    const newest = newestCollectionRunId(summaries.map((s) => s.collectionRunId));
    if (newest && summaries[0]?.collectionRunId !== newest) {
      summaries.sort((a, b) => (a.collectionRunId < b.collectionRunId ? 1 : -1));
    }
    return summaries;
  },

  async listGroups(collectionRunId) {
    const coll = await getBarSequenceCollection();
    const rows = await coll
      .aggregate<Document>([
        { $match: { collection_run_id: collectionRunId } },
        {
          $group: {
            _id: "$partition_id",
            observationCount: { $sum: 1 },
            symbols: { $addToSet: "$symbol" },
          },
        },
        { $sort: { _id: 1 } },
      ])
      .toArray();

    return rows.map((row): GroupDescriptor => ({
      partitionId: asString(row._id),
      observationCount: asNumber(row.observationCount),
      symbolCount: Array.isArray(row.symbols) ? row.symbols.length : 0,
    }));
  },

  async listSymbols(collectionRunId, group) {
    const coll = await getBarSequenceCollection();
    const match: Document = { collection_run_id: collectionRunId };
    if (group && group !== "ALL") match.partition_id = group;
    const rows = await coll
      .aggregate<Document>([
        { $match: match },
        {
          $group: {
            _id: { symbol: "$symbol", partition: "$partition_id" },
            observationCount: { $sum: 1 },
            minSequence: { $min: "$generator_sequence_no" },
            maxSequence: { $max: "$generator_sequence_no" },
          },
        },
        { $sort: { "_id.symbol": 1 } },
      ])
      .toArray();

    return rows.map((row): SymbolDescriptor => {
      const id = (row._id ?? {}) as { symbol?: unknown; partition?: unknown };
      return {
        symbol: asString(id.symbol),
        partitionId: asString(id.partition),
        observationCount: asNumber(row.observationCount),
        minSequence: asNumber(row.minSequence),
        maxSequence: asNumber(row.maxSequence),
      };
    });
  },

  async getTrajectory(collectionRunId, symbol, trajectory) {
    const coll = await getBarSequenceCollection();
    const docs = await coll
      .find(
        { collection_run_id: collectionRunId, symbol },
        {
          projection: {
            _id: 0,
            collection_run_id: 1,
            partition_id: 1,
            symbol: 1,
            generator_sequence_no: 1,
            source_event_time: 1,
            source_timestamp_text: 1,
            received_time: 1,
            persisted_time: 1,
            interval: 1,
            open: 1,
            high: 1,
            low: 1,
            close: 1,
            volume: 1,
            event_count: 1,
            source_id: 1,
            alpaca_message_type: 1,
            payload_hash: 1,
            source_time_regression: 1,
            duplicate_arrival: 1,
          },
          sort: { generator_sequence_no: 1 },
        },
      )
      .toArray();

    const observations = docs
      .map((doc) => mapObservation(doc as RawObservation, trajectory))
      .sort((a, b) => a.generatorSequenceNo - b.generatorSequenceNo);

    return {
      collectionRunId,
      symbol,
      partitionId: observations[0]?.partitionId ?? "",
      trajectory,
      observations,
      summary: summarize(observations),
    };
  },
};
