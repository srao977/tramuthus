/**
 * Presentation-only selectors and polar geometry for persisted DSE_JEH evidence.
 * Inputs are Go-authored states/events; outputs are counts, filtering, and SVG coordinates.
 * This module never classifies a zone, calculates velocity, ranks, or generates decisions.
 */
import type { EntityState, PhaseEvidence, RunEvidence, StrategyEvent, Zone } from "./types";

export const ZONE_LABELS: Record<Zone, string> = {
  HOP_ON: "HOP_ON candidate zone",
  MOMENTUM_HOLD: "MOMENTUM_HOLD",
  HOP_OFF: "HOP_OFF candidate zone",
  DISREGARD: "DISREGARD",
};

export const ZONE_COLORS: Record<Zone, string> = {
  HOP_ON: "#26a269",
  MOMENTUM_HOLD: "#d6a21d",
  HOP_OFF: "#d45b52",
  DISREGARD: "#57717a",
};

export type Point = { x: number; y: number };

/** Maps normalized phase to 0° top and clockwise-positive SVG coordinates. */
export function polarPoint(angle: number, center: number, radius: number): Point {
  const radians = angle * Math.PI / 180;
  return {
    x: center + radius * Math.sin(radians),
    y: center - radius * Math.cos(radians),
  };
}

export function currentCounts(run: RunEvidence) {
  const symbols = new Set(run.inputs.map((input) => input.symbol));
  const states = Object.values(run.states);
  const zoneCounts: Record<Zone, number> = { HOP_ON: 0, MOMENTUM_HOLD: 0, HOP_OFF: 0, DISREGARD: 0 };
  for (const state of states) zoneCounts[state.current_zone]++;
  return {
    total: symbols.size,
    observable: states.length,
    initializing: symbols.size - states.length,
    zones: zoneCounts,
  };
}

export function initializingSymbols(run: RunEvidence): string[] {
  const all = new Set(run.inputs.map((input) => input.symbol));
  for (const symbol of Object.keys(run.states)) all.delete(symbol);
  return [...all].sort();
}

export function entityEvidence(run: RunEvidence, symbol: string): {
  state?: EntityState;
  inputs: PhaseEvidence[];
  events: StrategyEvent[];
  decisions: StrategyEvent[];
} {
  return {
    state: run.states[symbol],
    inputs: run.inputs.filter((input) => input.symbol === symbol),
    events: run.strategyEvents.filter((event) => event.symbol === symbol),
    decisions: run.decisionEvents.filter((event) => event.symbol === symbol),
  };
}
