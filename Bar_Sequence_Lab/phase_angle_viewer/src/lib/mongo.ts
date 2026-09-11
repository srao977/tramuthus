import "server-only";

import { MongoClient, type Collection, type Document } from "mongodb";

export type PhaseMongoConfig = {
  uri: string;
  db: string;
  rawCollection: string;
  phaseCollection: string;
};

export function phaseMongoConfig(): PhaseMongoConfig {
  return {
    uri: process.env.BAR_SEQ_LAB_MONGO_URI?.trim() || "mongodb://127.0.0.1:27017",
    db: process.env.BAR_SEQ_LAB_MONGO_DB?.trim() || "bar_sequence_db",
    rawCollection: process.env.BAR_SEQ_LAB_RAW_COLLECTION?.trim() || "bar_sequence",
    phaseCollection:
      process.env.BAR_SEQ_LAB_PHASE_COLLECTION?.trim() || "bar_sequence_phase_angle_series",
  };
}

declare global {
  var __barSeqPhaseMongo: { client: MongoClient; promise: Promise<MongoClient> } | undefined;
}

async function getClient(): Promise<MongoClient> {
  const { uri } = phaseMongoConfig();
  if (!globalThis.__barSeqPhaseMongo) {
    const client = new MongoClient(uri, { serverSelectionTimeoutMS: 4000 });
    globalThis.__barSeqPhaseMongo = { client, promise: client.connect() };
  }
  return globalThis.__barSeqPhaseMongo.promise;
}

async function getCollection(name: string): Promise<Collection<Document>> {
  const client = await getClient();
  return client.db(phaseMongoConfig().db).collection(name);
}

export async function getPhaseCollection(): Promise<Collection<Document>> {
  return getCollection(phaseMongoConfig().phaseCollection);
}

export async function getRawCollection(): Promise<Collection<Document>> {
  return getCollection(phaseMongoConfig().rawCollection);
}

export function mongoPublicError(error: unknown): string {
  const message = error instanceof Error ? error.message : "MongoDB unavailable";
  if (/ECONNREFUSED|timed out|Server selection/i.test(message)) {
    return "MongoDB is unavailable at the configured host. The viewer did not write data.";
  }
  return "Failed to read phase-angle evidence.";
}
