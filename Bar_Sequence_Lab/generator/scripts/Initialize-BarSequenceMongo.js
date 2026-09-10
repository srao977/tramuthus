// Bar Sequence Lab — MongoDB V0.1 initialization
// Target: bar_sequence_db.bar_sequence
//
// Safe to run more than once.
// Does NOT drop the database or collection.
// Does NOT delete documents.
// Does NOT insert sample/dummy observations.
// Does NOT add a JSON Schema validator.
// Does NOT create a second raw-bar collection.
//
// Intended document fields (from generator Observation / JSONL; not inserted here):
//   collection_run_id, partition_id, symbol, generator_sequence_no,
//   source_event_time, source_timestamp_text, received_time, persisted_time,
//   interval, open, high, low, close, volume, event_count,
//   source_id, alpaca_message_type, payload_hash,
//   source_time_regression, duplicate_arrival
//
// generator_sequence_no = accepted-arrival order PER SYMBOL PER collection_run_id.
// Duplicates are flagged (duplicate_arrival) and still receive the next sequence
// number. This index is therefore NOT unique: a uniqueness constraint must not
// reject raw arrival evidence while duplicate disposition (OE-08) remains open.

const DB_NAME = "bar_sequence_db";
const COLL_NAME = "bar_sequence";

db = db.getSiblingDB(DB_NAME);

print("[Bar Sequence Lab] database: " + db.getName());
print("[Bar Sequence Lab] collection: " + COLL_NAME);

const existing = db.getCollectionNames();
if (existing.indexOf(COLL_NAME) === -1) {
  print("ERROR: collection '" + COLL_NAME + "' does not exist.");
  print("Create it in MongoDB Compass first. Aborting without creating collections, dropping, or inserting.");
  throw new Error("missing collection " + COLL_NAME);
}

const coll = db.getCollection(COLL_NAME);

print("");
print("Creating/verifying indexes...");

// Primary Maths / raw-sequence path: one symbol's accepted-arrival sequence
// for one collection run. Also covers prefix queries on collection_run_id
// and on {collection_run_id, symbol}. NOT unique (see header).
coll.createIndex(
  { collection_run_id: 1, symbol: 1, generator_sequence_no: 1 },
  { name: "ix_run_symbol_sequence" }
);
print("  verified ix_run_symbol_sequence { collection_run_id: 1, symbol: 1, generator_sequence_no: 1 } (not unique)");

// Partition-scoped audit analog of partition_{A|B|C}.jsonl.
// Not a prefix of ix_run_symbol_sequence (symbol sits between run and seq).
coll.createIndex(
  { collection_run_id: 1, partition_id: 1 },
  { name: "ix_run_partition" }
);
print("  verified ix_run_partition { collection_run_id: 1, partition_id: 1 }");

print("");
print("Indexes on " + COLL_NAME + ":");
printjson(coll.getIndexes());

print("");
print("Current document count: " + coll.countDocuments({}));
print("");
print("Bar Sequence Lab Mongo initialization complete.");
print("Generator persistence is NOT connected. JSONL audit path is unchanged.");
