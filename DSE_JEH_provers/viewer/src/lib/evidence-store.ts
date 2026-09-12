import "server-only";

/**
 * Read-only repository for deterministic files emitted by the Go proving command.
 * Input is DSE_JEH_EVIDENCE_ROOT (default ../output); output is authoritative run evidence.
 * It performs no strategy classification, phase transformation, ranking, or decision logic.
 */
import { readdir, readFile } from "node:fs/promises";
import path from "node:path";
import type { RunEvidence, RunMetadata, RunSummary } from "./types";

const RUN_ID_PATTERN = /^csv-prove-[a-f0-9]{16}$/;

function evidenceRoot(): string {
  return path.resolve(process.env.DSE_JEH_EVIDENCE_ROOT ?? path.join(process.cwd(), "..", "output"));
}

async function readJSON<T>(filePath: string): Promise<T> {
  return JSON.parse(await readFile(filePath, "utf8")) as T;
}

async function readJSONL<T>(filePath: string): Promise<T[]> {
  const text = await readFile(filePath, "utf8");
  return text.split(/\r?\n/).filter(Boolean).map((line) => JSON.parse(line) as T);
}

function runDirectory(runId: string): string {
  if (!RUN_ID_PATTERN.test(runId)) throw new Error("Invalid proving run identity");
  return path.join(evidenceRoot(), runId);
}

/** Enumerates complete proving runs by their Go-authored metadata. */
export async function listRuns(): Promise<RunSummary[]> {
  let entries;
  try {
    entries = await readdir(evidenceRoot(), { withFileTypes: true });
  } catch (error) {
    if ((error as NodeJS.ErrnoException).code === "ENOENT") return [];
    throw error;
  }
  const metadata = await Promise.all(entries
    .filter((entry) => entry.isDirectory() && RUN_ID_PATTERN.test(entry.name))
    .map(async (entry) => readJSON<RunMetadata>(path.join(evidenceRoot(), entry.name, "run_metadata.json"))));
  return metadata
    .map((run) => ({
      run_id: run.run_id,
      mode: run.mode,
      collection_run_id: run.collection_run_id,
      series_size: run.series_size,
      solver_version: run.solver_version,
      strategy_version: run.strategy_version,
      ranking_mode: run.ranking_mode,
      input_count: run.input_count,
      replay_matched: run.replay_matched,
    }))
    .sort((left, right) => right.run_id.localeCompare(left.run_id));
}

/** Loads the five separated evidence artifacts for one selected run. */
export async function loadRun(runId: string): Promise<RunEvidence> {
  const directory = runDirectory(runId);
  const [metadata, inputs, strategyEvents, decisionEvents, states] = await Promise.all([
    readJSON<RunMetadata>(path.join(directory, "run_metadata.json")),
    readJSONL<RunEvidence["inputs"][number]>(path.join(directory, "admitted_phase_input_evidence.jsonl")),
    readJSONL<RunEvidence["strategyEvents"][number]>(path.join(directory, "strategy_state_evidence.jsonl")),
    readJSONL<RunEvidence["decisionEvents"][number]>(path.join(directory, "decision_evidence.jsonl")),
    readJSON<RunEvidence["states"]>(path.join(directory, "final_entity_states.json")),
  ]);
  if (metadata.run_id !== runId) throw new Error("Run directory and metadata identity do not match");
  return { metadata, inputs, strategyEvents, decisionEvents, states };
}
