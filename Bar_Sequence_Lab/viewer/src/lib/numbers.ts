/** Coerce Mongo Long / number / string to a finite JS number. */

export function asNumber(value: unknown, fallback = 0): number {
  if (typeof value === "number" && Number.isFinite(value)) return value;
  if (typeof value === "bigint") return Number(value);
  if (typeof value === "string" && value.trim() !== "") {
    const n = Number(value);
    if (Number.isFinite(n)) return n;
  }
  if (value && typeof value === "object") {
    const rec = value as { $numberLong?: string; low?: number; high?: number; toString?: () => string };
    if (typeof rec.$numberLong === "string") {
      const n = Number(rec.$numberLong);
      if (Number.isFinite(n)) return n;
    }
    if (typeof rec.low === "number" && typeof rec.high === "number") {
      return rec.high * 2 ** 32 + (rec.low >>> 0);
    }
    if (typeof rec.toString === "function") {
      const n = Number(rec.toString());
      if (Number.isFinite(n)) return n;
    }
  }
  return fallback;
}

export function asIso(value: unknown): string {
  if (!value) return "";
  if (value instanceof Date) return value.toISOString();
  if (typeof value === "string") return value;
  if (typeof value === "object" && "toISOString" in value && typeof (value as Date).toISOString === "function") {
    try {
      return (value as Date).toISOString();
    } catch {
      return String(value);
    }
  }
  return String(value);
}

export function newestCollectionRunId(ids: string[]): string | undefined {
  if (ids.length === 0) return undefined;
  return [...ids].sort((a, b) => (a < b ? 1 : a > b ? -1 : 0))[0];
}

export function sequencesContiguousFromOne(seqs: number[]): boolean {
  if (seqs.length === 0) return false;
  const ordered = [...seqs].sort((a, b) => a - b);
  for (let i = 0; i < ordered.length; i += 1) {
    if (ordered[i] !== i + 1) return false;
  }
  return true;
}
