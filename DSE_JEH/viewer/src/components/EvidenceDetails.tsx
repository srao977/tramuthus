import type { entityEvidence } from "@/lib/presentation";
import type { RunEvidence } from "@/lib/types";

type EntityEvidence = ReturnType<typeof entityEvidence>;

export default function EvidenceDetails({ run, symbol, evidence }: { run: RunEvidence; symbol: string; evidence: EntityEvidence }) {
  const state = evidence.state;
  const latest = [...evidence.inputs].reverse().find((input) => state ? input.generator_sequence_no === state.generator_sequence_no : true);
  const transition = [...evidence.events].reverse().find((event) => event.event_type === "ZONE_ENTERED");
  const candidate = [...evidence.events].reverse().find((event) => event.event_type === "HOP_ON_CANDIDATE" || event.event_type === "HOP_OFF_CANDIDATE");
  const decision = evidence.decisions.at(-1);
  const rows: Array<[string, string]> = [
    ["Symbol", symbol], ["Current Phase", state ? `${state.current_phase_degrees.toFixed(4)}°` : "--"], ["Current Zone", state?.current_zone ?? "INITIALIZING"], ["Previous Phase", state?.has_previous_phase ? `${state.previous_phase_degrees?.toFixed(4)}°` : "--"], ["Previous Zone", state?.previous_zone ?? "--"], ["Sequence", String(state?.generator_sequence_no ?? latest?.generator_sequence_no ?? "--")], ["Validity", state?.validity_state ?? latest?.validity_state ?? "--"], ["Solver", latest ? `${latest.solver_name} · ${latest.solver_version}` : `${run.metadata.solver_name} · ${run.metadata.solver_version}`], ["Input Series", latest?.input_series_type ?? run.metadata.input_series_type], ["Collection Run", latest?.collection_run_id ?? run.metadata.collection_run_id], ["Partition", latest?.partition_id ?? "--"], ["Analysis Time", latest?.analysis_created_at ?? "--"], ["Last Transition", transition ? `${transition.event_type} @ ${transition.generator_sequence_no}` : "--"], ["Candidate Event", candidate?.event_type ?? "--"], ["Decision Evidence", decision?.decision_kind ?? "--"], ["Source Identity", state?.source_identity ?? "--"], ["Phase Velocity", "NOT YET DEFINED"],
  ];
  return <dl className="evidence-grid">{rows.map(([label, value]) => <div key={label}><dt>{label}</dt><dd>{value}</dd></div>)}</dl>;
}