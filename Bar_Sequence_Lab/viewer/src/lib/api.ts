import type {
  CollectionRunSummary,
  GroupDescriptor,
  SymbolDescriptor,
  TrajectorySeries,
  TrajectoryType,
} from "./types";

async function readJson<T>(path: string): Promise<T> {
  const response = await fetch(path, { cache: "no-store" });
  const body = (await response.json()) as T & { error?: string };
  if (!response.ok) {
    throw new Error(body.error || `Request failed (${response.status})`);
  }
  return body;
}

export function fetchRuns() {
  return readJson<{ runs: CollectionRunSummary[] }>("/api/runs");
}

export function fetchGroups(run: string) {
  return readJson<{ groups: GroupDescriptor[] }>(`/api/runs/${encodeURIComponent(run)}/groups`);
}

export function fetchSymbols(run: string, group: string) {
  const q = group && group !== "ALL" ? `?group=${encodeURIComponent(group)}` : "";
  return readJson<{ symbols: SymbolDescriptor[] }>(
    `/api/runs/${encodeURIComponent(run)}/symbols${q}`,
  );
}

export function fetchTrajectory(run: string, symbol: string, type: TrajectoryType) {
  const params = new URLSearchParams({ symbol, type });
  return readJson<TrajectorySeries>(
    `/api/runs/${encodeURIComponent(run)}/trajectory?${params.toString()}`,
  );
}
