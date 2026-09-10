const run = "20260910T191246Z-1";
const coll = db.getSiblingDB("bar_sequence_db").bar_sequence;
print("mongo_count_run=" + coll.countDocuments({ collection_run_id: run }));
print("mongo_total=" + coll.countDocuments({}));
print("distinct_symbols=" + coll.distinct("symbol", { collection_run_id: run }).sort().join(","));
print("distinct_partitions=" + coll.distinct("partition_id", { collection_run_id: run }).sort().join(","));
print("--- by partition ---");
coll.aggregate([
  { $match: { collection_run_id: run } },
  { $group: { _id: "$partition_id", n: { $sum: 1 } } },
  { $sort: { _id: 1 } }
]).forEach(function (d) { print(d._id + "=" + d.n); });
print("--- by symbol ---");
coll.aggregate([
  { $match: { collection_run_id: run } },
  { $group: { _id: "$symbol", n: { $sum: 1 }, maxseq: { $max: "$generator_sequence_no" } } },
  { $sort: { _id: 1 } }
]).forEach(function (d) { print(d._id + " n=" + d.n + " maxseq=" + d.maxseq); });
