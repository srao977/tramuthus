// DSE_JEH_TransSat_1 MongoDB persistence setup and verification.
// Target: bar_sequence_db.h_h_stage_emit_values
// Source verification only: bar_sequence_db.bar_sequence
//
// Safe to run repeatedly in MongoDB Compass Shell / mongosh.
// This script never drops or clears a collection, never modifies bar_sequence,
// never inserts sample/run data, and never generates a pipeline_run_id.

(function setupStageEmitPersistence() {
  "use strict";

  const DB_NAME = "bar_sequence_db";
  const SOURCE_COLLECTION = "bar_sequence";
  const TARGET_COLLECTION = "h_h_stage_emit_values";
  const issues = [];
  const warnings = [];

  const targetDb = db.getSiblingDB(DB_NAME);

  function fail(message) {
    issues.push(message);
    print("FAIL: " + message);
  }

  function warn(message) {
    warnings.push(message);
    print("WARN: " + message);
  }

  function canonicalize(value) {
    if (Array.isArray(value)) {
      return value.map(canonicalize);
    }
    if (value !== null && typeof value === "object") {
      const normalized = {};
      Object.keys(value).sort().forEach(function (key) {
        normalized[key] = canonicalize(value[key]);
      });
      return normalized;
    }
    return value;
  }

  function equivalentDocument(left, right) {
    return JSON.stringify(canonicalize(left)) === JSON.stringify(canonicalize(right));
  }

  function hasOwnFields(value) {
    return value !== null && typeof value === "object" && Object.keys(value).length > 0;
  }

  function isCompatibleIndex(index, expectedKey) {
    return equivalentDocument(index.key, expectedKey) &&
      Boolean(index.unique) === false &&
      Boolean(index.sparse) === false &&
      Boolean(index.hidden) === false &&
      index.partialFilterExpression === undefined &&
      index.expireAfterSeconds === undefined &&
      index.collation === undefined &&
      index.wildcardProjection === undefined;
  }

  print("[DSE_JEH_TransSat_1] MongoDB setup / verification");
  print("Database: " + targetDb.getName());
  print("Source collection: " + SOURCE_COLLECTION + " (read-only)");
  print("Target collection: " + TARGET_COLLECTION);
  print("");

  const collectionNames = targetDb.getCollectionNames();
  if (collectionNames.indexOf(SOURCE_COLLECTION) === -1) {
    fail("Required source collection '" + SOURCE_COLLECTION + "' does not exist in " + DB_NAME + ".");
    print("No collection was created and no index or validator change was attempted.");
    print("FINAL RESULT: FAIL");
    return;
  }

  print("bar_sequence presence: PRESENT");
  const source = targetDb.getCollection(SOURCE_COLLECTION);

  // persisted_time is the repository's Mongo write timestamp. Select the run
  // having the greatest valid persisted_time. This identifies the most recently
  // persisted available source run without relying on event order across symbols.
  // If legacy rows have no valid persisted_time, collection_run_id values use the
  // repository's sortable UTC run-ID convention (YYYYMMDDTHHMMSSZ-N), so a
  // descending ID fallback is used only for that legacy case.
  let latestRun = null;
  const latestByPersistedTime = source.aggregate([
    {
      $match: {
        collection_run_id: { $type: "string", $ne: "" },
        persisted_time: { $type: "date" }
      }
    },
    {
      $group: {
        _id: "$collection_run_id",
        latest_persisted_time: { $max: "$persisted_time" }
      }
    },
    { $sort: { latest_persisted_time: -1, _id: -1 } },
    { $limit: 1 }
  ]).toArray();

  if (latestByPersistedTime.length === 1) {
    latestRun = latestByPersistedTime[0]._id;
    print("Latest collection_run_id rule: greatest valid persisted_time per collection_run_id");
    print("Latest source persisted_time: " + latestByPersistedTime[0].latest_persisted_time);
  } else {
    const runIds = source.distinct("collection_run_id", {
      collection_run_id: { $type: "string", $ne: "" }
    });
    runIds.sort();
    if (runIds.length > 0) {
      latestRun = runIds[runIds.length - 1];
      warn("No valid persisted_time was found; used descending sortable collection_run_id fallback.");
      print("Latest collection_run_id rule: descending sortable UTC collection_run_id fallback");
    }
  }

  if (latestRun === null) {
    fail("bar_sequence contains no non-empty collection_run_id to verify.");
    print("No collection, index, or validator change was attempted.");
    print("FINAL RESULT: FAIL");
    return;
  } else {
    const latestRunFilter = { collection_run_id: latestRun };
    const latestRunCount = source.countDocuments(latestRunFilter);
    const latestRunSymbols = source.distinct("symbol", latestRunFilter).filter(function (symbol) {
      return typeof symbol === "string" && symbol.length > 0;
    });

    print("Most recent collection_run_id: " + latestRun);
    print("Records in most recent collection_run_id: " + latestRunCount);
    print("Distinct symbols in most recent collection_run_id: " + latestRunSymbols.length);
  }

  print("");
  let targetExists = collectionNames.indexOf(TARGET_COLLECTION) !== -1;
  if (!targetExists) {
    try {
      targetDb.createCollection(TARGET_COLLECTION);
      targetExists = true;
      print("Created collection '" + TARGET_COLLECTION + "'.");
    } catch (error) {
      fail("Could not create '" + TARGET_COLLECTION + "': " + error.message);
    }
  } else {
    print("Collection '" + TARGET_COLLECTION + "' already exists; continuing verification.");
  }

  if (!targetExists) {
    print("FINAL RESULT: FAIL");
    return;
  }

  const target = targetDb.getCollection(TARGET_COLLECTION);
  const expectedValidator = {
    $jsonSchema: {
      bsonType: "object",
      required: [
        "pipeline_run_id",
        "run_type",
        "collection_run_id",
        "symbol",
        "sequence_no",
        "stage",
        "stage_order",
        "event_time"
      ],
      properties: {
        pipeline_run_id: { bsonType: "string", minLength: 1 },
        run_type: { bsonType: "string", minLength: 1, pattern: "^[A-Z0-9][A-Z0-9_-]*$" },
        collection_run_id: { bsonType: "string", minLength: 1 },
        symbol: { bsonType: "string", minLength: 1 },
        sequence_no: { bsonType: ["int", "long"] },
        stage: { bsonType: "string", minLength: 1 },
        stage_order: { bsonType: ["int", "long"] },
        event_time: { bsonType: "date" },
        emit_type: { bsonType: ["string", "null"] },
        input: { bsonType: "object" },
        output: { bsonType: "object" },
        capital_allocation: { bsonType: "object" }
      }
    }
  };

  print("");
  print("Validator verification...");
  let collectionInfo = null;
  try {
    const infoResult = targetDb.runCommand({
      listCollections: 1,
      filter: { name: TARGET_COLLECTION }
    });
    if (infoResult.ok === 1 && infoResult.cursor.firstBatch.length === 1) {
      collectionInfo = infoResult.cursor.firstBatch[0];
    } else {
      fail("Could not read collection metadata for validator verification.");
    }
  } catch (error) {
    fail("Could not inspect the existing validator: " + error.message);
  }

  if (collectionInfo !== null) {
    const options = collectionInfo.options || {};
    const currentValidator = options.validator || {};
    const hasValidator = hasOwnFields(currentValidator);

    if (!hasValidator) {
      try {
        const validatorResult = targetDb.runCommand({
          collMod: TARGET_COLLECTION,
          validator: expectedValidator,
          validationLevel: "strict",
          validationAction: "error"
        });
        if (validatorResult.ok === 1) {
          print("Installed permissive common-envelope validator.");
        } else {
          fail("Validator installation did not succeed: " + JSON.stringify(validatorResult));
        }
      } catch (error) {
        fail("Could not install validator: " + error.message);
      }
    // MongoDB defaults omitted validationLevel/action to strict/error, so an
    // exact validator with omitted options is semantically compatible.
    } else if (
      equivalentDocument(currentValidator, expectedValidator) &&
      (options.validationLevel === undefined || options.validationLevel === "strict") &&
      (options.validationAction === undefined || options.validationAction === "error")
    ) {
      print("Existing validator matches the expected common-envelope validator.");
    } else {
      warn("Existing validator differs from the expected validator; it was not replaced.");
      print("Existing validator/options:");
      printjson({
        validator: currentValidator,
        validationLevel: options.validationLevel,
        validationAction: options.validationAction
      });
      print("Expected validator/options:");
      printjson({
        validator: expectedValidator,
        validationLevel: "strict",
        validationAction: "error"
      });
      fail("Human review is required before changing the existing validator.");
    }
  }

  const expectedIndexes = [
    {
      key: { pipeline_run_id: 1, symbol: 1, sequence_no: 1, stage_order: 1 },
      name: "ix_pipeline_symbol_sequence_stage"
    },
    {
      key: { collection_run_id: 1, pipeline_run_id: 1 },
      name: "ix_source_pipeline"
    },
    {
      key: { run_type: 1, pipeline_run_id: 1 },
      name: "ix_run_type_pipeline"
    },
    {
      key: { pipeline_run_id: 1, emit_type: 1 },
      name: "ix_pipeline_emit_type"
    },
    {
      key: { pipeline_run_id: 1, event_time: 1 },
      name: "ix_pipeline_event_time"
    }
  ];

  // The first index also covers { pipeline_run_id, symbol } prefix lookups, so a
  // separate redundant run/symbol index is intentionally not created.
  print("");
  print("Index verification...");
  expectedIndexes.forEach(function (expected) {
    const existingIndexes = target.getIndexes();
    const sameName = existingIndexes.find(function (index) {
      return index.name === expected.name;
    });
    const equivalent = existingIndexes.find(function (index) {
      return isCompatibleIndex(index, expected.key);
    });

    if (sameName !== undefined && !isCompatibleIndex(sameName, expected.key)) {
      warn("Index '" + expected.name + "' has incompatible keys or options and was not replaced.");
      printjson(sameName);
      fail("Human review is required for incompatible index '" + expected.name + "'.");
      return;
    }
    if (equivalent !== undefined) {
      print("Verified " + expected.name + " via existing index '" + equivalent.name + "'.");
      return;
    }

    try {
      target.createIndex(expected.key, { name: expected.name });
      print("Created " + expected.name + " " + JSON.stringify(expected.key) + ".");
    } catch (error) {
      fail("Could not create index '" + expected.name + "': " + error.message);
    }
  });

  print("");
  print("Resulting indexes on " + TARGET_COLLECTION + ":");
  printjson(target.getIndexes());
  print("Current target document count: " + target.countDocuments({}));
  print("");
  print("Safety summary:");
  print("  bar_sequence remained read-only.");
  print("  No pipeline_run_id was generated.");
  print("  No source or sample data was copied or inserted.");
  print("  No collection was dropped, cleared, updated, or replaced.");
  print("Warnings: " + warnings.length);
  print("Failures: " + issues.length);
  print("FINAL RESULT: " + (issues.length === 0 ? "PASS" : "FAIL"));
})();
