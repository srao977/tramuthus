"use client";

import type { TrajectoryObservation, TrajectoryType } from "@/lib/types";

type Props = {
  observation: TrajectoryObservation | null;
  trajectory: TrajectoryType;
};

function Row({ label, value }: { label: string; value: string | number | boolean }) {
  return (
    <div className="detail-row">
      <dt>{label}</dt>
      <dd>{String(value)}</dd>
    </div>
  );
}

export default function ObservationDetails({ observation, trajectory }: Props) {
  if (!observation) {
    return (
      <section className="panel details">
        <h2>Observation</h2>
        <p className="muted">Click a wave point. Hover shows the chart readout; click fills this drawer from the stored observation.</p>
      </section>
    );
  }
  return (
    <section className="panel details">
      <h2>Observation</h2>
      <dl>
        <Row label="generator_sequence_no" value={observation.generatorSequenceNo} />
        <Row label="symbol" value={observation.symbol} />
        <Row label="partition_id" value={observation.partitionId} />
        <Row label="collection_run_id" value={observation.collectionRunId} />
        <Row label="trajectory" value={trajectory} />
        <Row label="y" value={observation.y} />
        <Row label="open" value={observation.open} />
        <Row label="high" value={observation.high} />
        <Row label="low" value={observation.low} />
        <Row label="close" value={observation.close} />
        <Row label="volume" value={observation.volume} />
        <Row label="interval" value={observation.interval} />
        <Row label="source_event_time" value={observation.sourceEventTime} />
        <Row label="source_timestamp_text" value={observation.sourceTimestampText} />
        <Row label="received_time" value={observation.receivedTime} />
        <Row label="persisted_time" value={observation.persistedTime} />
        <Row label="source_id" value={observation.sourceId} />
        <Row label="alpaca_message_type" value={observation.alpacaMessageType} />
        <Row label="payload_hash" value={observation.payloadHash} />
        <Row label="source_time_regression" value={observation.sourceTimeRegression} />
        <Row label="duplicate_arrival" value={observation.duplicateArrival} />
        <Row label="event_count" value={observation.eventCount} />
      </dl>
    </section>
  );
}
