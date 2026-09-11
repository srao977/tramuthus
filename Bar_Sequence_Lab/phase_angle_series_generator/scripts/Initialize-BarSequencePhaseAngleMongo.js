// Bar Sequence Lab phase-angle derived collection initialization.
// Target: bar_sequence_db.bar_sequence_phase_angle_series
// Creates the collection only when absent and creates two indexes idempotently.
// Never drops/deletes/inserts data, modifies raw bar_sequence, or adds validation.

const DB_NAME = "bar_sequence_db";
const COLL_NAME = "bar_sequence_phase_angle_series";

db = db.getSiblingDB(DB_NAME);
print("[Bar Sequence Lab] database: " + db.getName());
print("[Bar Sequence Lab] derived collection: " + COLL_NAME);

if (db.getCollectionNames().indexOf(COLL_NAME) === -1) {
  db.createCollection(COLL_NAME);
  print("Created collection '" + COLL_NAME + "'.");
} else {
  print("Collection '" + COLL_NAME + "' already exists.");
}

const phase = db.getCollection(COLL_NAME);
phase.createIndex(
  {
    collection_run_id: 1,
    symbol: 1,
    generator_sequence_no: 1,
    series_size: 1,
    solver_name: 1,
    solver_version: 1
  },
  { name: "ux_phase_run_symbol_sequence_size_solver", unique: true }
);
print("Verified unique analytical identity index.");

phase.createIndex(
  {
    collection_run_id: 1,
    partition_id: 1,
    series_size: 1,
    solver_name: 1,
    solver_version: 1,
    symbol: 1,
    generator_sequence_no: 1
  },
  { name: "ix_phase_run_partition_size_solver_symbol_sequence" }
);
print("Verified viewer/query index.");

print("");
print("Indexes on " + COLL_NAME + ":");
printjson(phase.getIndexes());
print("Current derived document count: " + phase.countDocuments({}));
print("Phase-angle Mongo initialization complete. Raw bar_sequence was not accessed.");