import { experimentSearch } from "./query";
import type {
  ExperimentMetadata,
  ExperimentQuery,
  PhaseSeries,
  PhaseSymbol,
  RawObservation,
} from "./types";

async function readJson<T>(response: Response): Promise<T> {
  const body = (await response.json()) as T & { error?: string };
  if (!response.ok) throw new Error(body.error || `Request failed (${response.status})`);
  return body;
}

export async function fetchExperiments(): Promise<ExperimentMetadata[]> {
  const response = await fetch("/api/phase/experiments", { cache: "no-store" });
  return (await readJson<{ experiments: ExperimentMetadata[] }>(response)).experiments;
}

export async function fetchSymbols(experiment: ExperimentQuery): Promise<PhaseSymbol[]> {
  const response = await fetch(`/api/phase/symbols?${experimentSearch(experiment)}`, {
    cache: "no-store",
  });
  return (await readJson<{ symbols: PhaseSymbol[] }>(response)).symbols;
}

export async function fetchPhaseSeries(
  experiment: ExperimentQuery,
  partitionId: string,
  symbol: string,
): Promise<PhaseSeries> {
  const search = experimentSearch(experiment);
  search.set("partitionId", partitionId);
  search.set("symbol", symbol);
  return readJson<PhaseSeries>(
    await fetch(`/api/phase/series?${search}`, { cache: "no-store" }),
  );
}

export async function fetchRawObservation(
  collectionRunId: string,
  symbol: string,
  generatorSequenceNo: number,
): Promise<RawObservation> {
  const search = new URLSearchParams({
    collectionRunId,
    symbol,
    generatorSequenceNo: String(generatorSequenceNo),
  });
  const response = await fetch(`/api/raw-observation?${search}`, { cache: "no-store" });
  return (await readJson<{ observation: RawObservation }>(response)).observation;
}
