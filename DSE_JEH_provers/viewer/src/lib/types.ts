/**
 * Viewer contracts for deterministic evidence written by the Go DSE_JEH proving runtime.
 * Inputs are persisted JSON/JSONL records; outputs are display-only typed values.
 * These types do not calculate phase, zones, ranking, transitions, or decisions.
 */
export type Zone = "HOP_ON" | "MOMENTUM_HOLD" | "HOP_OFF" | "DISREGARD";

export type RunMetadata = {
  run_id: string;
  mode: string;
  source_csv: string;
  collection_run_id: string;
  series_size: number;
  solver_name: string;
  solver_version: string;
  input_series_type: string;
  strategy_version: string;
  ranking_mode: string;
  source_csv_sha256: string;
  input_count: number;
  strategy_event_count: number;
  decision_event_count: number;
  duplicate_input_count: number;
  output_digest: string;
  replay_digest: string;
  replay_matched: boolean;
};

export type PhaseEvidence = {
  source_record_number: number;
  collection_run_id: string;
  partition_id: string;
  symbol: string;
  generator_sequence_no: number;
  series_size: number;
  solver_name: string;
  solver_version: string;
  input_series_type: string;
  phase_angle_degrees: number | null;
  phase_angle_text: string;
  phase_observable: boolean;
  validity_state: string;
  analysis_created_at: string;
};

export type StrategyEvent = {
  event_id: string;
  event_type: "PHASE_BECAME_OBSERVABLE" | "ZONE_ENTERED" | "ZONE_EXITED" | "HOP_ON_CANDIDATE" | "HOP_OFF_CANDIDATE" | "DECISION_EVENT";
  strategy_version: string;
  source_identity: string;
  collection_run_id: string;
  partition_id: string;
  symbol: string;
  generator_sequence_no: number;
  phase_angle_degrees: number;
  previous_phase_degrees?: number;
  current_zone: Zone;
  previous_zone?: Zone;
  solver_name: string;
  solver_version: string;
  validity_state: string;
  decision_kind?: string;
  ranking_mode?: string;
};

export type EntityState = {
  symbol: string;
  has_previous_phase: boolean;
  previous_phase_degrees?: number;
  current_phase_degrees: number;
  previous_zone?: Zone;
  current_zone: Zone;
  generator_sequence_no: number;
  source_identity: string;
  solver_name: string;
  solver_version: string;
  validity_state: string;
};

export type RunSummary = Pick<RunMetadata,
  "run_id" | "mode" | "collection_run_id" | "series_size" | "solver_version" |
  "strategy_version" | "ranking_mode" | "input_count" | "replay_matched"
>;

export type RunEvidence = {
  metadata: RunMetadata;
  inputs: PhaseEvidence[];
  strategyEvents: StrategyEvent[];
  decisionEvents: StrategyEvent[];
  states: Record<string, EntityState>;
};
