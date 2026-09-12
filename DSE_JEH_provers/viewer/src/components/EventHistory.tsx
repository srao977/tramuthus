import type { StrategyEvent } from "@/lib/types";

export default function EventHistory({ events, decisions }: { events: StrategyEvent[]; decisions: StrategyEvent[] }) {
  const combined = [...events, ...decisions].sort((left, right) => left.generator_sequence_no - right.generator_sequence_no || left.event_type.localeCompare(right.event_type));
  if (!combined.length) return <p className="no-events">No authoritative strategy or decision events for this entity.</p>;
  return <div className="event-table-wrap"><table className="event-table"><thead><tr><th>Sequence</th><th>Evidence Type</th><th>Phase</th><th>Prior Zone</th><th>Current Zone</th><th>Decision / Mode</th><th>Lineage</th></tr></thead><tbody>{combined.map((event) => <tr key={event.event_id} className={event.event_type === "DECISION_EVENT" ? "decision-row" : ""}><td>{event.generator_sequence_no}</td><td>{event.event_type}</td><td>{event.phase_angle_degrees.toFixed(4)}°</td><td>{event.previous_zone ?? "--"}</td><td>{event.current_zone}</td><td>{event.decision_kind || event.ranking_mode || "--"}</td><td title={event.source_identity}>{event.source_identity}</td></tr>)}</tbody></table></div>;
}