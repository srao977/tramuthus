// Read-only comparison for the bounded Dynamic Pipeline Execution Runs completed on 2026-09-14.
const collection = db.getSiblingDB("bar_sequence_db").getCollection("h_h_stage_emit_values");
const sourceCollection = db.getSiblingDB("bar_sequence_db").getCollection("bar_sequence");
const runs = {
  run_a: "DPE-GOVERNED-RUN-A-20260914-001",
  run_b: "DPE-GOVERNED-RUN-B-20260914-001",
};

function summarize(pipelineRunId, expectedRunType, expectedRisk) {
  const match = { pipeline_run_id: pipelineRunId };
  const sequence = collection.aggregate([
    { $match: match },
    { $group: { _id: null, min: { $min: "$sequence_no" }, max: { $max: "$sequence_no" } } },
  ]).toArray()[0] || null;
  const stages = collection.aggregate([
    { $match: match },
    { $group: { _id: "$stage", count: { $sum: 1 } } },
    { $sort: { _id: 1 } },
  ]).toArray();
  const emits = collection.aggregate([
    { $match: { pipeline_run_id: pipelineRunId, emit_type: { $ne: null } } },
    { $group: { _id: "$emit_type", count: { $sum: 1 } } },
    { $sort: { _id: 1 } },
  ]).toArray();
  return {
    pipeline_run_id: pipelineRunId,
    expected_run_type: expectedRunType,
    total_documents: collection.countDocuments(match),
    run_types: collection.distinct("run_type", match),
    mismatched_run_type_documents: collection.countDocuments({ pipeline_run_id: pipelineRunId, run_type: { $ne: expectedRunType } }),
    risk_r_values: collection.distinct("capital_allocation.risk_r", { pipeline_run_id: pipelineRunId, "capital_allocation.risk_r": { $exists: true } }),
    mismatched_risk_documents: collection.countDocuments({ pipeline_run_id: pipelineRunId, "capital_allocation.risk_r": { $exists: true, $ne: expectedRisk } }),
    collection_run_ids: collection.distinct("collection_run_id", match),
    distinct_symbols: collection.distinct("symbol", match).length,
    sequence_range: sequence,
    sequence_64_documents: collection.countDocuments({ pipeline_run_id: pipelineRunId, sequence_no: 64 }),
    missing_required_envelope_fields: collection.countDocuments({
      pipeline_run_id: pipelineRunId,
      $or: [
        { run_type: { $in: [null, ""] } },
        { collection_run_id: { $in: [null, ""] } },
        { symbol: { $in: [null, ""] } },
        { sequence_no: null },
        { stage: { $in: [null, ""] } },
        { stage_order: null },
        { event_time: null },
      ],
    }),
    stages,
    emits,
  };
}

function finalCapital(pipelineRunId) {
  return collection.aggregate([
    { $match: { pipeline_run_id: pipelineRunId, stage: "CAPITAL_RESULT" } },
    { $sort: { symbol: 1, sequence_no: -1, stage_order: -1 } },
    { $group: { _id: "$symbol", sequence_no: { $first: "$sequence_no" }, capital: { $first: "$capital_allocation" } } },
    { $project: { _id: 0, symbol: "$_id", sequence_no: 1, current_capital: "$capital.current_capital", realized_pnl: "$capital.realized_pnl", return_pct: "$capital.return_pct", active_quantity: "$capital.active_quantity", cash_remaining: "$capital.cash_remaining" } },
    { $sort: { symbol: 1 } },
  ]).toArray();
}

function portfolioSummary(rows) {
  return {
    symbols_with_post_classification_results: rows.length,
    total_current_capital: rows.reduce((total, row) => total + row.current_capital, 0),
    total_realized_pnl: rows.reduce((total, row) => total + row.realized_pnl, 0),
    active_positions: rows.filter((row) => row.active_quantity > 0).length,
  };
}

function actionCounts(pipelineRunId) {
  const rows = collection.aggregate([
    { $match: { pipeline_run_id: pipelineRunId, stage: "STATE_ACTION" } },
    { $group: { _id: { symbol: "$symbol", emit_type: "$emit_type" }, count: { $sum: 1 } } },
  ]).toArray();
  const bySymbol = {};
  for (const row of rows) {
    bySymbol[row._id.symbol] ||= {};
    bySymbol[row._id.symbol][row._id.emit_type] = row.count;
  }
  return bySymbol;
}

function comparison(runA, runB) {
  const capitalA = Object.fromEntries(finalCapital(runA).map((row) => [row.symbol, row]));
  const capitalB = Object.fromEntries(finalCapital(runB).map((row) => [row.symbol, row]));
  const actionsA = actionCounts(runA);
  const actionsB = actionCounts(runB);
  return [...new Set([...Object.keys(capitalA), ...Object.keys(capitalB)])].sort().map((symbol) => ({
    symbol,
    run_a: { ...capitalA[symbol], actions: actionsA[symbol] || {} },
    run_b: { ...capitalB[symbol], actions: actionsB[symbol] || {} },
    current_capital_difference_b_minus_a: capitalB[symbol].current_capital - capitalA[symbol].current_capital,
  }));
}

const finalCapitalRunA = finalCapital(runs.run_a);
const finalCapitalRunB = finalCapital(runs.run_b);
const collectionPipelineRunIds = collection.distinct("pipeline_run_id").sort();
const sourceSymbols = sourceCollection.distinct("symbol", { collection_run_id: "20260911T161623Z-1" }).sort();
const tracedSymbols = collection.distinct("symbol", { pipeline_run_id: runs.run_a }).sort();

const report = {
  source_collection_run_id: "20260911T161623Z-1",
  collection_pipeline_run_ids: collectionPipelineRunIds,
  source_symbols_without_post_classification_trace: sourceSymbols.filter((symbol) => !tracedSymbols.includes(symbol)),
  run_a: summarize(runs.run_a, "RUN_A", 0),
  run_b: summarize(runs.run_b, "RUN_B", 1),
  portfolio_run_a: portfolioSummary(finalCapitalRunA),
  portfolio_run_b: portfolioSummary(finalCapitalRunB),
  per_symbol_comparison: comparison(runs.run_a, runs.run_b),
  final_capital_run_a: finalCapitalRunA,
  final_capital_run_b: finalCapitalRunB,
};

const expectedPipelineRunIds = [runs.run_a, runs.run_b].sort();
if (EJSON.stringify(collectionPipelineRunIds) !== EJSON.stringify(expectedPipelineRunIds)) {
  throw new Error(`Trace collection does not contain exactly the two authoritative pipeline runs: ${EJSON.stringify(collectionPipelineRunIds)}`);
}
for (const run of [report.run_a, report.run_b]) {
  if (run.total_documents === 0 || run.mismatched_run_type_documents !== 0 || run.mismatched_risk_documents !== 0 ||
      run.missing_required_envelope_fields !== 0 || run.collection_run_ids.length !== 1 ||
      run.collection_run_ids[0] !== report.source_collection_run_id) {
    throw new Error(`Authoritative run verification failed: ${EJSON.stringify(run)}`);
  }
}

if (typeof summaryOnly !== "undefined" && summaryOnly) {
  print(EJSON.stringify({
    source_collection_run_id: report.source_collection_run_id,
    collection_pipeline_run_ids: report.collection_pipeline_run_ids,
    source_symbols_without_post_classification_trace: report.source_symbols_without_post_classification_trace,
    run_a: report.run_a,
    run_b: report.run_b,
    portfolio_run_a: report.portfolio_run_a,
    portfolio_run_b: report.portfolio_run_b,
    per_symbol_rows: report.per_symbol_comparison.length,
  }, null, 2, { relaxed: true }));
} else {
  print(EJSON.stringify(report, null, 2, { relaxed: true }));
}
