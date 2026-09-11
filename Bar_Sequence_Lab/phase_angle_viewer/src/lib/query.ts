import { DEFAULT_EXPERIMENT, type ExperimentQuery } from "./types";

const SYMBOL_RE = /^[A-Z0-9.-]{1,15}$/;
const PARTITION_RE = /^[A-Z0-9_-]{1,12}$/;

export function experimentFromSearch(search: URLSearchParams): ExperimentQuery {
  const seriesSize = Number(search.get("seriesSize") ?? DEFAULT_EXPERIMENT.seriesSize);
  return {
    collectionRunId: (search.get("collectionRunId") ?? DEFAULT_EXPERIMENT.collectionRunId).trim(),
    seriesSize,
    solverName: (search.get("solverName") ?? DEFAULT_EXPERIMENT.solverName).trim(),
    solverVersion: (search.get("solverVersion") ?? DEFAULT_EXPERIMENT.solverVersion).trim(),
    inputSeriesType: (search.get("inputSeriesType") ?? DEFAULT_EXPERIMENT.inputSeriesType).trim(),
  };
}

export function validateExperiment(experiment: ExperimentQuery): string | null {
  if (!experiment.collectionRunId) return "collectionRunId is required";
  if (!Number.isInteger(experiment.seriesSize) || experiment.seriesSize <= 0) {
    return "seriesSize must be a positive integer";
  }
  if (!experiment.solverName || !experiment.solverVersion || !experiment.inputSeriesType) {
    return "solverName, solverVersion, and inputSeriesType are required";
  }
  return null;
}

export function validSymbol(value: string): boolean {
  return SYMBOL_RE.test(value);
}

export function validPartition(value: string): boolean {
  return PARTITION_RE.test(value);
}

export function experimentSearch(experiment: ExperimentQuery): URLSearchParams {
  return new URLSearchParams({
    collectionRunId: experiment.collectionRunId,
    seriesSize: String(experiment.seriesSize),
    solverName: experiment.solverName,
    solverVersion: experiment.solverVersion,
    inputSeriesType: experiment.inputSeriesType,
  });
}
