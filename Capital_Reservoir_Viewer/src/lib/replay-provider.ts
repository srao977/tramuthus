import "server-only";

import type { Document } from "mongodb";
import { mergeReservoirEvents, normalizeReservoirEvent, type CapitalReservoirEvent } from "./capital-reservoir";
import { reservoirCollection } from "./mongo";

export type ReservoirRun = {
  pipelineRunId: string;
  collectionRunId: string;
  runType: string;
  eventCount: number;
  firstProcessedAt: string;
  lastProcessedAt: string;
  complete: boolean;
};

export async function listRuns(): Promise<ReservoirRun[]> {
  const collection = await reservoirCollection();
  const rows = await collection.aggregate<Document>([
    { $group: {
      _id: "$pipeline_run_id",
      collectionRunId: { $first: "$collection_run_id" },
      runType: { $first: "$run_type" },
      eventCount: { $sum: 1 },
      firstProcessedAt: { $min: "$processed_at" },
      lastProcessedAt: { $max: "$processed_at" },
      endCount: { $sum: { $cond: [{ $eq: ["$event_type", "RUN_END"] }, 1, 0] } },
    } },
    { $sort: { lastProcessedAt: -1 } },
  ]).toArray();
  return rows.map((row) => ({
    pipelineRunId: String(row._id),
    collectionRunId: String(row.collectionRunId),
    runType: String(row.runType),
    eventCount: Number(row.eventCount),
    firstProcessedAt: new Date(row.firstProcessedAt as Date).toISOString(),
    lastProcessedAt: new Date(row.lastProcessedAt as Date).toISOString(),
    complete: Number(row.endCount) === 1,
  }));
}

export async function readRun(pipelineRunId: string): Promise<CapitalReservoirEvent[]> {
  const collection = await reservoirCollection();
  const documents = await collection.find(
    { pipeline_run_id: pipelineRunId },
    { projection: { _id: 0 }, sort: { event_sequence: 1 } },
  ).toArray();
  return mergeReservoirEvents(documents.map(normalizeReservoirEvent));
}