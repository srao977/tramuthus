"use client";

import dynamic from "next/dynamic";
import { useEffect, useMemo, useRef, useState } from "react";
import {
  BookOpen,
  Database,
  GitBranch,
  Layers,
  Moon,
  Radio,
  ReceiptText,
  Server,
  Sun,
  Waves,
} from "lucide-react";
import { useTheme } from "next-themes";
import { fetchGroups, fetchRuns, fetchSymbols, fetchTrajectory } from "@/lib/api";
import {
  BASE_MS_PER_BAR,
  WAVE_SLOT_COLORS,
  WAVE_SLOT_COUNT,
  clampCursor,
  maxSequenceOf,
  pickDefaultWaveSymbols,
  reconcileSlotsWithAvailable,
  replaceWaveSymbol,
  waveComplete,
} from "@/lib/replay";
import type {
  CollectionRunSummary,
  GroupDescriptor,
  SymbolDescriptor,
  TrajectoryObservation,
  TrajectorySeries,
  TrajectoryType,
} from "@/lib/types";
import ObservationDetails from "./ObservationDetails";
import ReplayBar from "./ReplayBar";

const WaveLaneChart = dynamic(() => import("./WaveLaneChart"), {
  ssr: false,
  loading: () => <div className="chart-canvas empty-lane">Loading chart…</div>,
});

type LabView = "waves" | "observation";

function formatRaw(value: number | undefined, trajectory: TrajectoryType) {
  if (value === undefined) return "--";
  if (trajectory === "volume") return value.toLocaleString("en-US", { maximumFractionDigits: 0 });
  return value.toLocaleString("en-US", { minimumFractionDigits: 2, maximumFractionDigits: 4 });
}

export default function LabApp() {
  const { resolvedTheme, setTheme } = useTheme();
  const [activeView, setActiveView] = useState<LabView>("waves");
  const [runs, setRuns] = useState<CollectionRunSummary[]>([]);
  const [runId, setRunId] = useState("");
  const [groups, setGroups] = useState<GroupDescriptor[]>([]);
  const [group, setGroup] = useState("ALL");
  const [symbols, setSymbols] = useState<SymbolDescriptor[]>([]);
  const [slots, setSlots] = useState<string[]>(["", "", "", ""]);
  const [trajectory, setTrajectory] = useState<TrajectoryType>("price");
  const [seriesBySymbol, setSeriesBySymbol] = useState<Record<string, TrajectorySeries>>({});
  const [selected, setSelected] = useState<TrajectoryObservation | null>(null);
  const [error, setError] = useState("");
  const [loading, setLoading] = useState(true);
  const [playing, setPlaying] = useState(false);
  const [speed, setSpeed] = useState(1);
  const [cursor, setCursor] = useState(0);
  const [replayEpoch, setReplayEpoch] = useState("");
  const slotsInitializedForRun = useRef("");

  useEffect(() => {
    let cancelled = false;
    (async () => {
      try {
        const { runs: next } = await fetchRuns();
        if (cancelled) return;
        setRuns(next);
        setRunId(next[0]?.collectionRunId ?? "");
        setError("");
      } catch (err) {
        if (!cancelled) setError(err instanceof Error ? err.message : "Failed to load runs");
      } finally {
        if (!cancelled) setLoading(false);
      }
    })();
    return () => {
      cancelled = true;
    };
  }, []);

  useEffect(() => {
    if (!runId) return;
    let cancelled = false;
    (async () => {
      try {
        const { groups: nextGroups } = await fetchGroups(runId);
        if (cancelled) return;
        setGroups(nextGroups);
        setGroup("ALL");
        slotsInitializedForRun.current = "";
        setError("");
      } catch (err) {
        if (!cancelled) setError(err instanceof Error ? err.message : "Failed to load groups");
      }
    })();
    return () => {
      cancelled = true;
    };
  }, [runId]);

  useEffect(() => {
    if (!runId) return;
    let cancelled = false;
    (async () => {
      try {
        const { symbols: nextSymbols } = await fetchSymbols(runId, group);
        if (cancelled) return;
        setSymbols(nextSymbols);
        const names = nextSymbols.map((row) => row.symbol);
        setSlots((current) => {
          if (slotsInitializedForRun.current !== runId) {
            slotsInitializedForRun.current = runId;
            return pickDefaultWaveSymbols(names);
          }
          return reconcileSlotsWithAvailable(current, names);
        });
      } catch (err) {
        if (!cancelled) setError(err instanceof Error ? err.message : "Failed to load symbols");
      }
    })();
    return () => {
      cancelled = true;
    };
  }, [runId, group]);

  const selectedSymbols = useMemo(() => [...new Set(slots.filter(Boolean))], [slots]);

  useEffect(() => {
    if (!runId || selectedSymbols.length === 0) return;
    let cancelled = false;
    (async () => {
      try {
        const loaded = await Promise.all(
          selectedSymbols.map((symbol) => fetchTrajectory(runId, symbol, trajectory)),
        );
        if (cancelled) return;
        const next: Record<string, TrajectorySeries> = {};
        for (const series of loaded) next[series.symbol] = series;
        setSeriesBySymbol(next);
        setError("");
      } catch (err) {
        if (!cancelled) setError(err instanceof Error ? err.message : "Failed to load trajectories");
      }
    })();
    return () => {
      cancelled = true;
    };
  }, [runId, selectedSymbols, trajectory]);

  const maxSequence = useMemo(
    () => maxSequenceOf(slots.map((symbol) => seriesBySymbol[symbol]?.summary.maxSequence ?? 0)),
    [slots, seriesBySymbol],
  );

  const selectedKey = selectedSymbols.join("|");
  const epoch = `${runId}|${selectedKey}|${trajectory}`;
  const epochMatches = replayEpoch === epoch;
  // Until the user starts replay, show the full stored prefix so lanes are not blank.
  const displayCursor = epochMatches ? clampCursor(cursor, maxSequence) : maxSequence;
  const displayPlaying = epochMatches ? playing : false;

  useEffect(() => {
    if (!displayPlaying || maxSequence <= 0) return;
    const interval = window.setInterval(() => {
      setReplayEpoch(epoch);
      setCursor((current) => {
        const next = clampCursor(current, maxSequence);
        if (next >= maxSequence) {
          setPlaying(false);
          return maxSequence;
        }
        return next + 1;
      });
    }, Math.max(24, BASE_MS_PER_BAR / speed));
    return () => window.clearInterval(interval);
  }, [displayPlaying, speed, maxSequence, epoch]);

  const groupButtons = useMemo(() => {
    const ids = groups.map((g) => g.partitionId).filter(Boolean);
    return ["ALL", ...ids.filter((id) => id !== "ALL")];
  }, [groups]);

  const currentRun = runs.find((run) => run.collectionRunId === runId);
  const mongoOnline = Boolean(runId) && !error;

  function selectObservation(obs: TrajectoryObservation) {
    const series = seriesBySymbol[obs.symbol];
    const full =
      series?.observations.find((row) => row.generatorSequenceNo === obs.generatorSequenceNo) ?? obs;
    setSelected(full);
    setActiveView("observation");
    window.requestAnimationFrame(() => {
      document.getElementById("observation-drawer")?.scrollIntoView({
        behavior: "smooth",
        block: "nearest",
      });
    });
  }

  return (
    <main className="dashboard-shell">
      <header className="topbar">
        <div className="brand-block">
          <div className="brand-mark">
            <Waves size={21} strokeWidth={2.4} />
          </div>
          <div>
            <h1>Bar Sequence Lab</h1>
            <p>Raw bar-sequence waves · bar-index replay</p>
          </div>
        </div>
        <div className="header-actions">
          <button
            className="theme-toggle"
            type="button"
            onClick={() => setTheme(resolvedTheme === "dark" ? "light" : "dark")}
            aria-label="Toggle color theme"
            title="Toggle color theme"
          >
            <Sun className="theme-sun" size={17} />
            <Moon className="theme-moon" size={17} />
          </button>
          <div className="connection-cluster">
            <span className={`live-dot ${mongoOnline ? "is-on" : ""}`} />
            <div>
              <strong>{mongoOnline ? "MONGO READ-ONLY" : "MONGO OFFLINE"}</strong>
              <span>LabMongoProvider · bar_sequence_db</span>
            </div>
          </div>
        </div>
      </header>

      <aside className="rail">
        <div className="rail-label">Workspace</div>
        <nav aria-label="Lab views" className="view-menu">
          <button
            className={activeView === "waves" ? "view-button is-active" : "view-button"}
            onClick={() => setActiveView("waves")}
            type="button"
            aria-current={activeView === "waves" ? "page" : undefined}
          >
            <Layers size={17} />
            <span>Multi-Wave Replay</span>
          </button>
          <button
            className={activeView === "observation" ? "view-button is-active" : "view-button"}
            onClick={() => {
              setActiveView("observation");
              window.requestAnimationFrame(() => {
                document.getElementById("observation-drawer")?.scrollIntoView({
                  behavior: "smooth",
                  block: "nearest",
                });
              });
            }}
            type="button"
            aria-current={activeView === "observation" ? "page" : undefined}
          >
            <BookOpen size={17} />
            <span>Observation Details</span>
          </button>
          <button className="view-button is-disabled" type="button" disabled title="Not implemented. Raw replay baseline only.">
            <GitBranch size={17} />
            <span>Ehlers / Phase</span>
            <small>Later</small>
          </button>
          <button className="view-button is-disabled" type="button" disabled title="Not implemented. Raw replay baseline only.">
            <ReceiptText size={17} />
            <span>Hop / Strategy</span>
            <small>Later</small>
          </button>
        </nav>
        <div className="rail-status">
          <Server size={16} />
          <span>BAR-INDEX REPLAY</span>
        </div>
      </aside>

      <section className="workspace">
        {error ? (
          <div className="health-banner health-grpc_unavailable" role="status">
            <div className="health-summary">
              <span className="health-indicator" />
              <div>
                <strong>Viewer online · Mongo read failed</strong>
                <span>{error}</span>
              </div>
            </div>
          </div>
        ) : (
          <div className="health-banner" role="status">
            <div className="health-summary">
              <span className="health-indicator" />
              <div>
                <strong>{loading ? "Loading collection runs" : "Viewer online · raw stored sequences"}</strong>
                <span>
                  {currentRun
                    ? `${currentRun.collectionRunId} · ${currentRun.symbolCount} symbols · ${currentRun.observationCount} observations`
                    : "Waiting for LabMongoProvider"}
                </span>
              </div>
            </div>
            <div className="health-timing">
              <span>read-only</span>
              <small>no writes · no maths</small>
            </div>
          </div>
        )}

        <div className="instrument-heading">
          <div>
            <div className="eyebrow">COLLECTION RUN · GENERATOR_SEQUENCE_NO</div>
            <div className="quote-line">
              <h2>Four-wave canvas</h2>
              <strong>
                {displayCursor}/{maxSequence}
              </strong>
              <span>bar index</span>
            </div>
          </div>
          <div className={`stream-pill ${displayPlaying ? "stream-live" : "stream-waiting"}`}>
            <Radio size={14} />
            {displayPlaying ? "Playing" : displayCursor >= maxSequence && maxSequence > 0 ? "Complete" : "Paused"}
          </div>
        </div>

        <div className="lab-controls">
          <label>
            Collection run
            <select value={runId} onChange={(e) => setRunId(e.target.value)} disabled={!runs.length}>
              {runs.map((run) => (
                <option key={run.collectionRunId} value={run.collectionRunId}>
                  {run.collectionRunId} · {run.symbolCount} sym · {run.observationCount} obs
                </option>
              ))}
            </select>
          </label>
          <div className="seg" role="group" aria-label="Group">
            {groupButtons.map((id) => (
              <button key={id} type="button" className={group === id ? "is-on" : ""} onClick={() => setGroup(id)}>
                {id}
              </button>
            ))}
          </div>
          <div className="seg" role="group" aria-label="Trajectory">
            <button type="button" className={trajectory === "price" ? "is-on" : ""} onClick={() => setTrajectory("price")}>
              Price
            </button>
            <button type="button" className={trajectory === "volume" ? "is-on" : ""} onClick={() => setTrajectory("volume")}>
              Volume
            </button>
          </div>
        </div>

        <div className="wave-slot-row" role="group" aria-label="Wave presentation slots">
          {slots.map((symbol, index) => (
            <label key={`slot-${index}`}>
              Wave {index + 1}
              <select value={symbol} onChange={(e) => setSlots(replaceWaveSymbol(slots, index, e.target.value))}>
                <option value="">—</option>
                {symbols.map((row) => (
                  <option key={`${index}-${row.symbol}`} value={row.symbol}>
                    {row.symbol} · {row.partitionId} · {row.observationCount}
                  </option>
                ))}
              </select>
            </label>
          ))}
        </div>

        <div className="wave-grid" aria-label="Independently scaled klinecharts lanes">
          {Array.from({ length: WAVE_SLOT_COUNT }, (_, index) => {
            const symbol = slots[index] ?? "";
            const series = symbol ? seriesBySymbol[symbol] : undefined;
            const lastVisible = series?.observations.filter((row) => row.generatorSequenceNo <= displayCursor).at(-1);
            const complete = waveComplete(series?.summary.maxSequence ?? 0, displayCursor);
            return (
              <article key={`lane-${index}`} className="wave-lane">
                <header>
                  <i style={{ background: WAVE_SLOT_COLORS[index] }} aria-hidden />
                  <div>
                    <strong>{symbol || `Slot ${index + 1}`}</strong>
                    <small>
                      presentation slot {index + 1}
                      {complete && series ? " · complete" : ""}
                      {series?.summary.maxSequence === maxSequence && maxSequence > 0 ? " · longest" : ""}
                    </small>
                  </div>
                  <span>
                    {trajectory === "volume" ? "raw volume" : "raw close"}{" "}
                    {formatRaw(trajectory === "volume" ? lastVisible?.volume : lastVisible?.close, trajectory)}
                  </span>
                </header>
                <section className="chart-section lane-chart">
                  <WaveLaneChart
                    symbol={symbol}
                    slotIndex={index}
                    color={WAVE_SLOT_COLORS[index]}
                    observations={series?.observations ?? []}
                    trajectory={trajectory}
                    cursor={displayCursor}
                    onSelect={selectObservation}
                  />
                </section>
                <p className="lane-meta">
                  {series
                    ? `${series.observations.filter((row) => row.generatorSequenceNo <= displayCursor).length}/${series.summary.observationCount} obs · seq 1..${series.summary.maxSequence} · partition ${series.partitionId}`
                    : "No series loaded"}
                </p>
              </article>
            );
          })}
        </div>

        <ReplayBar
          playing={displayPlaying}
          speed={speed}
          cursor={displayCursor}
          maxSequence={maxSequence}
          onPlayPause={() => {
            if (!epochMatches) {
              setReplayEpoch(epoch);
              setCursor(0);
              setPlaying(true);
              return;
            }
            if (displayCursor >= maxSequence && maxSequence > 0) setCursor(0);
            setPlaying(!displayPlaying);
          }}
          onRestart={() => {
            setReplayEpoch(epoch);
            setCursor(0);
            setPlaying(false);
          }}
          onSpeed={setSpeed}
          onSeek={(next) => {
            setReplayEpoch(epoch);
            setPlaying(false);
            setCursor(clampCursor(next, maxSequence));
          }}
        />

        <section className="operations-band">
          <div className="ops-heading">
            <div>
              <div className="eyebrow">WAVE STATUS</div>
              <h3>Independent raw trajectories</h3>
            </div>
            <span>x = generator_sequence_no</span>
          </div>
          <div className="ops-grid">
            {slots.map((symbol, index) => {
              const series = symbol ? seriesBySymbol[symbol] : undefined;
              return (
                <div className="status-cell" key={`status-${index}`}>
                  <span className="status-icon">
                    <Database size={18} />
                  </span>
                  <div>
                    <small>
                      Slot {index + 1} · {symbol || "empty"}
                    </small>
                    <strong>
                      {series
                        ? `${series.summary.observationCount} obs · ${waveComplete(series.summary.maxSequence, displayCursor) ? "complete" : "advancing"}`
                        : "no data"}
                    </strong>
                  </div>
                  <span
                    className={
                      series && waveComplete(series.summary.maxSequence, displayCursor)
                        ? "state-mark good"
                        : "state-mark"
                    }
                  />
                </div>
              );
            })}
          </div>
        </section>

        <section id="observation-drawer" className="observation-drawer is-open" aria-live="polite">
          <h2>Observation details</h2>
          <ObservationDetails observation={selected} trajectory={trajectory} />
        </section>
      </section>
    </main>
  );
}
