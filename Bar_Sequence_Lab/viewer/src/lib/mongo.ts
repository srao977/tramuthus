import "server-only";

import { MongoClient, type Collection, type Document } from "mongodb";

export type LabMongoConfig = {
  uri: string;
  db: string;
  collection: string;
};

export function labMongoConfig(): LabMongoConfig {
  return {
    uri: process.env.BAR_SEQ_LAB_MONGO_URI?.trim() || "mongodb://127.0.0.1:27017",
    db: process.env.BAR_SEQ_LAB_MONGO_DB?.trim() || "bar_sequence_db",
    collection: process.env.BAR_SEQ_LAB_MONGO_COLLECTION?.trim() || "bar_sequence",
  };
}

declare global {
  var __barSeqLabMongo: { client: MongoClient; promise: Promise<MongoClient> } | undefined;
}

async function getClient(): Promise<MongoClient> {
  const { uri } = labMongoConfig();
  if (!globalThis.__barSeqLabMongo) {
    const client = new MongoClient(uri, { serverSelectionTimeoutMS: 4000 });
    globalThis.__barSeqLabMongo = { client, promise: client.connect() };
  }
  return globalThis.__barSeqLabMongo.promise;
}

export async function getBarSequenceCollection(): Promise<Collection<Document>> {
  const client = await getClient();
  const { db, collection } = labMongoConfig();
  return client.db(db).collection(collection);
}

export function mongoPublicError(error: unknown): string {
  const message = error instanceof Error ? error.message : "MongoDB unavailable";
  if (/ECONNREFUSED|timed out|Server selection/i.test(message)) {
    return "MongoDB is unavailable at the configured host. Viewer is read-only and did not write.";
  }
  return "Failed to read bar-sequence observations.";
}
