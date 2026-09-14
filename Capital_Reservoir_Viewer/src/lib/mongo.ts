import "server-only";

import { MongoClient, type Collection, type Document } from "mongodb";

function config() {
  return {
    uri: process.env.CAPITAL_RESERVOIR_MONGO_URI?.trim() || "mongodb://127.0.0.1:27017",
    database: process.env.CAPITAL_RESERVOIR_MONGO_DB?.trim() || "bar_sequence_db",
    collection: process.env.CAPITAL_RESERVOIR_MONGO_COLLECTION?.trim() || "capital_reservoir_events",
  };
}

declare global {
  var __capitalReservoirMongo: { client: MongoClient; promise: Promise<MongoClient> } | undefined;
}

async function client(): Promise<MongoClient> {
  if (!globalThis.__capitalReservoirMongo) {
    const instance = new MongoClient(config().uri, { serverSelectionTimeoutMS: 4000 });
    globalThis.__capitalReservoirMongo = { client: instance, promise: instance.connect() };
  }
  return globalThis.__capitalReservoirMongo.promise;
}

export async function reservoirCollection(): Promise<Collection<Document>> {
  const connected = await client();
  const { database, collection } = config();
  return connected.db(database).collection(collection);
}

export function publicMongoError(error: unknown): string {
  const message = error instanceof Error ? error.message : "MongoDB unavailable";
  return /ECONNREFUSED|timed out|Server selection/i.test(message)
    ? "MongoDB is unavailable. The viewer made no changes."
    : "Capital Reservoir replay could not be read.";
}