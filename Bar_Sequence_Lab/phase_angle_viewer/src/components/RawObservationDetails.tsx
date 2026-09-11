import type { PhasePoint, RawObservation } from "@/lib/types";

type Props = {
  phase: PhasePoint | null;
  raw: RawObservation | null;
  loading: boolean;
};

function show(value: unknown): string {
  if (value === undefined || value === null || value === "") return "--";
  if (typeof value === "boolean") return value ? "true" : "false";
  return String(value);
}

export default function RawObservationDetails({ phase, raw, loading }: Props) {
  if (!phase) return <p className="empty-copy">Select an observable point to inspect its analytical identity and raw lineage.</p>;
  const rows: Array<[string, unknown]> = [
    ["symbol", phase.symbol],
    ["partition_id", phase.partitionId],
    ["generator_sequence_no", phase.generatorSequenceNo],
    ["phase_angle_degrees", phase.phaseAngleDegrees],
    ["validity_state", phase.validityState],
    ["phase_observable", phase.phaseObservable],
    ["collection_run_id", phase.collectionRunId],
    ["series_size", phase.seriesSize],
    ["solver_name", phase.solverName],
    ["solver_version", phase.solverVersion],
    ["input_series_type", phase.inputSeriesType],
    ["open", raw?.open],
    ["high", raw?.high],
    ["low", raw?.low],
    ["close", raw?.close],
    ["volume", raw?.volume],
    ["event_count", raw?.eventCount],
    ["source_event_time", raw?.sourceEventTime],
    ["received_time", raw?.receivedTime],
    ["persisted_time", raw?.persistedTime],
    ["source_id", raw?.sourceId],
    ["alpaca_message_type", raw?.alpacaMessageType],
    ["payload_hash", raw?.payloadHash],
    ["source_time_regression", raw?.sourceTimeRegression],
    ["duplicate_arrival", raw?.duplicateArrival],
  ];
  return (
    <div className="details-grid" aria-busy={loading}>
      {rows.map(([label, value]) => (
        <div className="detail-pair" key={label}>
          <dt>{label}</dt>
          <dd>{loading && label === "open" ? "Loading raw observation..." : show(value)}</dd>
        </div>
      ))}
    </div>
  );
}
