const run = "20260910T191246Z-1";
const coll = db.getSiblingDB("bar_sequence_db").bar_sequence;
print("mongo_count_run=" + coll.countDocuments({ collection_run_id: run }));
print("mongo_total=" + coll.countDocuments({}));
print("distinct_symbol_count=" + coll.distinct("symbol", { collection_run_id: run }).length);
print("--- by partition ---");
coll.aggregate([
  { $match: { collection_run_id: run } },
  { $group: { _id: "$partition_id", n: { $sum: 1 } } },
  { $sort: { _id: 1 } }
]).forEach(function (d) { print(d._id + "=" + d.n); });
print("--- by symbol ---");
coll.aggregate([
  { $match: { collection_run_id: run } },
  { $group: {
      _id: "$symbol",
      n: { $sum: 1 },
      minseq: { $min: "$generator_sequence_no" },
      maxseq: { $max: "$generator_sequence_no" },
      partition: { $min: "$partition_id" }
  } },
  { $sort: { _id: 1 } }
]).forEach(function (d) {
  print(d._id + " partition=" + d.partition + " n=" + d.n + " minseq=" + d.minseq + " maxseq=" + d.maxseq);
});
print("--- sequence gaps ---");
const symbols = coll.distinct("symbol", { collection_run_id: run }).sort();
symbols.forEach(function (sym) {
  const seqs = coll.find({ collection_run_id: run, symbol: sym }, { generator_sequence_no: 1, _id: 0 })
    .sort({ generator_sequence_no: 1 })
    .map(function (d) { return Number(d.generator_sequence_no); });
  const gaps = [];
  for (let i = 0; i < seqs.length; i++) {
    if (seqs[i] !== i + 1) {
      gaps.push("expected=" + (i + 1) + " got=" + seqs[i]);
      break;
    }
  }
  if (gaps.length) {
    print("GAP " + sym + " count=" + seqs.length + " " + gaps.join(";"));
  } else {
    print("OK " + sym + " 1.." + seqs.length);
  }
});
