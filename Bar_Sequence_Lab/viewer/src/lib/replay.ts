/**
 * Viewer-only BAR-INDEX replay helpers.
 * Does not alter Mongo observations, sequence numbers, or y values.
 * Per-symbol generator_sequence_no. Not synchronized market-time replay.
 */

export const WAVE_SLOT_COUNT = 4;

/** Presentation-slot colors only. Not Alpha/Beta/Gamma/Delta science. */
export const WAVE_SLOT_COLORS = ["#087f5b", "#3b82f6", "#e8b931", "#c2413b"] as const;

export const DEFAULT_WAVE_SYMBOLS = ["AAPL", "MSFT", "SPY", "JPM"] as const;

export const REPLAY_SPEEDS = [0.5, 1, 2, 4] as const;

/** Wall-clock ms per bar-index step at 1x. Presentation only. */
export const BASE_MS_PER_BAR = 140;

export function pickDefaultWaveSymbols(available: string[]): string[] {
  const unique = [...new Set(available.filter(Boolean))];
  const picked: string[] = [];
  for (const symbol of DEFAULT_WAVE_SYMBOLS) {
    if (unique.includes(symbol) && picked.length < WAVE_SLOT_COUNT) {
      picked.push(symbol);
    }
  }
  for (const symbol of unique) {
    if (picked.length >= WAVE_SLOT_COUNT) break;
    if (!picked.includes(symbol)) picked.push(symbol);
  }
  while (picked.length < WAVE_SLOT_COUNT) picked.push("");
  return picked.slice(0, WAVE_SLOT_COUNT);
}

export function replaceWaveSymbol(slots: string[], index: number, next: string): string[] {
  if (index < 0 || index >= WAVE_SLOT_COUNT) return slots.slice(0, WAVE_SLOT_COUNT);
  const copy = slots.slice(0, WAVE_SLOT_COUNT);
  while (copy.length < WAVE_SLOT_COUNT) copy.push("");
  copy[index] = next;
  return copy;
}

export function reconcileSlotsWithAvailable(slots: string[], available: string[]): string[] {
  const unique = [...new Set(available.filter(Boolean))];
  const next = slots.slice(0, WAVE_SLOT_COUNT);
  while (next.length < WAVE_SLOT_COUNT) next.push("");
  const used = new Set<string>();
  return next.map((symbol) => {
    if (symbol && unique.includes(symbol) && !used.has(symbol)) {
      used.add(symbol);
      return symbol;
    }
    const fallback = unique.find((candidate) => !used.has(candidate)) ?? "";
    if (fallback) used.add(fallback);
    return fallback;
  });
}

export function maxSequenceOf(lengths: number[]): number {
  return lengths.reduce((max, n) => (n > max ? n : max), 0);
}

/**
 * Prefix of a stored wave at bar-index cursor.
 * Observations with generator_sequence_no > cursor are hidden, not invented.
 * A shorter wave whose max seq <= cursor is returned in full.
 */
export function visiblePrefix<T extends { generatorSequenceNo: number }>(
  observations: T[],
  cursor: number,
): T[] {
  if (cursor <= 0 || observations.length === 0) return [];
  return observations.filter((row) => row.generatorSequenceNo <= cursor);
}

export function waveComplete(maxSequence: number, cursor: number): boolean {
  return maxSequence > 0 && cursor >= maxSequence;
}

export function clampCursor(cursor: number, maxSequence: number): number {
  if (maxSequence <= 0) return 0;
  if (cursor < 0) return 0;
  if (cursor > maxSequence) return maxSequence;
  return cursor;
}
