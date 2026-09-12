"use client";

import { polarPoint, ZONE_COLORS, ZONE_LABELS } from "@/lib/presentation";
import type { EntityState } from "@/lib/types";

export default function PolarPhaseDiagram({ states, selectedSymbol, onSelect }: { states: Record<string, EntityState>; selectedSymbol: string; onSelect: (symbol: string) => void }) {
  const center = 250;
  const radius = 174;

  return (
    <div className="polar-wrap"><svg viewBox="0 0 500 500" role="img" aria-label="Current observable entity phase, zero degrees top and clockwise">
      <circle cx={center} cy={center} r={210} className="polar-field" />
      <path d="M250 40 A210 210 0 0 1 460 250 L250 250 Z" className="zone-sector hold" /><path d="M460 250 A210 210 0 0 1 250 460 L250 250 Z" className="zone-sector hopoff" /><path d="M250 460 A210 210 0 0 1 40 250 L250 250 Z" className="zone-sector disregard" /><path d="M40 250 A210 210 0 0 1 250 40 L250 250 Z" className="zone-sector hopon" />
      <circle cx={center} cy={center} r={radius} className="phase-orbit" /><line x1="250" y1="28" x2="250" y2="472" className="polar-axis" /><line x1="28" y1="250" x2="472" y2="250" className="polar-axis" />
      <text x="250" y="20" textAnchor="middle" className="degree-label">0 / 360</text><text x="480" y="254" textAnchor="middle" className="degree-label">90</text><text x="250" y="493" textAnchor="middle" className="degree-label">180</text><text x="20" y="254" textAnchor="middle" className="degree-label">270</text>
      <text x="356" y="94" textAnchor="middle" className="sector-label">MOMENTUM_HOLD</text><text x="388" y="406" textAnchor="middle" className="sector-label">HOP_OFF</text><text x="112" y="406" textAnchor="middle" className="sector-label">DISREGARD</text><text x="112" y="94" textAnchor="middle" className="sector-label">HOP_ON</text>
      {Object.values(states).sort((left, right) => left.symbol.localeCompare(right.symbol)).map((state) => {
        const point = polarPoint(state.current_phase_degrees, center, radius);
        const selected = state.symbol === selectedSymbol;
        return <g key={state.symbol} transform={`translate(${point.x.toFixed(6)} ${point.y.toFixed(6)})`} className={`entity-point ${selected ? "selected" : ""}`} onClick={() => onSelect(state.symbol)} role="button" tabIndex={0} onKeyDown={(event) => { if (event.key === "Enter" || event.key === " ") onSelect(state.symbol); }} aria-label={`${state.symbol}, ${state.current_phase_degrees.toFixed(4)} degrees, ${ZONE_LABELS[state.current_zone]}`}><circle r={selected ? 8 : 6} fill={ZONE_COLORS[state.current_zone]} /><text y={point.y < center ? -11 : 18} textAnchor="middle">{state.symbol}</text></g>;
      })}
      <circle cx={center} cy={center} r="3" className="polar-center" />
    </svg></div>
  );
}