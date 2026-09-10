import assert from "node:assert/strict";
import test from "node:test";
import { asNumber, newestCollectionRunId, sequencesContiguousFromOne } from "./numbers.ts";

test("newest collection_run_id is lexicographic max, not ObjectId", () => {
  assert.equal(
    newestCollectionRunId(["20260910T190625Z-1", "20260910T191246Z-1"]),
    "20260910T191246Z-1",
  );
});

test("contiguous 1..N", () => {
  assert.equal(sequencesContiguousFromOne([1, 2, 3, 51]), false);
  assert.equal(sequencesContiguousFromOne([1, 2, 3]), true);
  assert.equal(sequencesContiguousFromOne([]), false);
});

test("asNumber handles Long-like objects", () => {
  assert.equal(asNumber(51), 51);
  assert.equal(asNumber({ $numberLong: "51" }), 51);
});
