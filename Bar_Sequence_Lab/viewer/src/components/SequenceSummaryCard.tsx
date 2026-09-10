"use client";

import type { SequenceSummary, TrajectoryType } from "@/lib/types";

type Props = {
  summary: SequenceSummary | null;
  trajectory: TrajectoryType;
};

export default function SequenceSummaryCard({ summary, trajectory }: Props) {
  if (!summary) {
    return (
      <section className="panel summary">
        <h2>Sequence</h2>
        <p className="muted">Select a run and symbol.</p>
      </section>
    );
  }
  return (
    <section className="panel summary">
      <h2>Sequence</h2>
      <ul>
        <li>
          <span>observations</span>
          <strong>{summary.observationCount}</strong>
        </li>
        <li>
          <span>min seq</span>
          <strong>{summary.minSequence}</strong>
        </li>
        <li>
          <span>max seq</span>
          <strong>{summary.maxSequence}</strong>
        </li>
        <li>
          <span>1..N</span>
          <strong>{summary.contiguousFromOne ? "yes" : "no"}</strong>
        </li>
        <li>
          <span>partition</span>
          <strong>{summary.partitionId}</strong>
        </li>
        <li>
          <span>trajectory</span>
          <strong>{trajectory}</strong>
        </li>
      </ul>
    </section>
  );
}
