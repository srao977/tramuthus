"use client";

import { useEffect, useRef, useState } from "react";
import {
  Activity,
  Gauge,
  Moon,
  Server,
  Sun,
  Waves,
} from "lucide-react";
import {
  fetchExperiments,
  fetchPhaseSeries,
  fetchRawObservation,
  fetchSymbols,
} from "@/lib/api";
import type {
  ExperimentMetadata,
  ExperimentQuery,
  PhasePoint,
  PhaseSeries,
  PhaseSymbol,
  RawObservation,
} from "@/lib/types";
import { DEFAULT_EXPERIMENT } from "@/lib/types";
import PartitionPhasePanel from "./PartitionPhasePanel";
import { useTheme } from "./ThemeProvider";

const PARTITIONS = [
  { id: "A", color: "#087f5b", preferred: "GOOGL" },
  { id: "B", color: "#c48a12", preferred: "SPY" },
  { id: "C", color: "#3273a8", preferred: "JPM" },
];

function experimentKey(experiment: ExperimentQuery): string {
  return [
    experiment.collectionRunId,
    experiment.seriesSize,
    experiment.solverName,
    experiment.solverVersion,
    experiment.inputSeriesType,
  ].join("|");
}

function pickExperiment(experiments: ExperimentMetadata[]): ExperimentMetadata | undefined {
  return experiments.find((item) => experimentKey(item) === experimentKey(DEFAULT_EXPERIMENT)) ?? experiments[0];
}

export default function PhaseAngleApp() {
  const { resolvedTheme, setTheme } = useTheme();
  const [experiments, setExperiments] = useState<ExperimentMetadata[]>([]);
  const [experiment, setExperiment] = useState<ExperimentMetadata | null>(null);
  const [symbols, setSymbols] = useState<PhaseSymbol[]>([]);
  const [selectedByPartition, setSelectedByPartition] = useState<Record<string, string>>({});
  const [seriesByPartition, setSeriesByPartition] = useState<Record<string, PhaseSeries>>({});
  const [loadingPartitions, setLoadingPartitions] = useState<Record<string, boolean>>({});
  const [selectedPoints, setSelectedPoints] = useState<Record<string, PhasePoint>>({});
  const [rawObservations, setRawObservations] = useState<Record<string, RawObservation>>({});
  const [rawLoading, setRawLoading] = useState<Record<string, boolean>>({});
  const rawRequestKeys = useRef<Record<string, string>>({});
  const [error, setError] = useState("");
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    let cancelled = false;
    fetchExperiments()
      .then((items) => {
        if (cancelled) return;
        setExperiments(items);
        setExperiment(pickExperiment(items) ?? null);
      })
      .catch((cause) => {
        if (!cancelled) setError(cause instanceof Error ? cause.message : "Failed to load experiments");
      })
      .finally(() => {
        if (!cancelled) setLoading(false);
      });
    return () => { cancelled = true; };
  }, []);

  useEffect(() => {
    if (!experiment) return;
    let cancelled = false;
    fetchSymbols(experiment)
      .then((items) => {
        if (cancelled) return;
        setSymbols(items);
        const next: Record<string, string> = {};
        for (const partition of PARTITIONS) {
          const available = items.filter((item) => item.partitionId === partition.id);
          next[partition.id] = available.find((item) => item.symbol === partition.preferred)?.symbol ?? available[0]?.symbol ?? "";
        }
        setLoadingPartitions(Object.fromEntries(PARTITIONS.map((partition) => [partition.id, true])));
        setSelectedByPartition(next);
        setError("");
      })
      .catch((cause) => {
        if (!cancelled) setError(cause instanceof Error ? cause.message : "Failed to load symbols");
      });
    return () => { cancelled = true; };
  }, [experiment]);

  // Load all three selected series independently while keeping each response symbol-scoped.
  useEffect(() => {
    if (!experiment) return;
    let cancelled = false;
    const requested = PARTITIONS.flatMap((partition) => {
      const symbol = selectedByPartition[partition.id];
      return symbol ? [{ partitionId: partition.id, symbol }] : [];
    });
    if (requested.length === 0) return;
    Promise.all(requested.map((item) => fetchPhaseSeries(experiment, item.partitionId, item.symbol)))
      .then((loaded) => {
        if (cancelled) return;
        setSeriesByPartition(Object.fromEntries(loaded.map((series) => [series.partitionId, series])));
        setError("");
      })
      .catch((cause) => {
        if (!cancelled) setError(cause instanceof Error ? cause.message : "Failed to load phase series");
      })
      .finally(() => {
        if (!cancelled) setLoadingPartitions({});
      });
    return () => { cancelled = true; };
  }, [experiment, selectedByPartition]);

  function selectPoint(point: PhasePoint) {
    const partitionId = point.partitionId;
    const requestKey = `${point.collectionRunId}|${point.symbol}|${point.generatorSequenceNo}`;
    rawRequestKeys.current[partitionId] = requestKey;
    setSelectedPoints((current) => ({ ...current, [partitionId]: point }));
    setRawObservations((current) => {
      const next = { ...current };
      delete next[partitionId];
      return next;
    });
    setRawLoading((current) => ({ ...current, [partitionId]: true }));
    fetchRawObservation(point.collectionRunId, point.symbol, point.generatorSequenceNo)
      .then((observation) => {
        if (rawRequestKeys.current[partitionId] !== requestKey) return;
        setRawObservations((current) => ({ ...current, [partitionId]: observation }));
      })
      .catch((cause) => setError(cause instanceof Error ? cause.message : "Raw lookup failed"))
      .finally(() => {
        if (rawRequestKeys.current[partitionId] === requestKey) {
          setRawLoading((current) => ({ ...current, [partitionId]: false }));
        }
      });
  }

  const online = Boolean(experiment) && !error;

  return (
    <main className="dashboard-shell">
      <header className="topbar">
        <div className="brand-block">
          <div className="brand-mark"><Waves size={21} strokeWidth={2.4} /></div>
          <div><h1>Bar Sequence Lab</h1><p>Persisted phase evidence · analytical bar index</p></div>
        </div>
        <div className="header-actions">
          <button className="icon-button" type="button" onClick={() => setTheme(resolvedTheme === "dark" ? "light" : "dark")} aria-label="Toggle color theme" title="Toggle color theme">
            <Sun className="theme-sun" size={17} /><Moon className="theme-moon" size={17} />
          </button>
          <div className="connection-cluster">
            <span className={`live-dot ${online ? "is-on" : ""}`} />
            <div><strong>{online ? "MONGO READ-ONLY" : "MONGO OFFLINE"}</strong><span>bar_sequence_phase_angle_series</span></div>
          </div>
        </div>
      </header>

      <aside className="rail">
        <div className="rail-label">Workspace</div>
        <nav aria-label="Lab views" className="view-menu">
          <button className="view-button is-active" type="button" aria-current="page"><Activity size={17} /><span>Phase Angle</span></button>
        </nav>
        <div className="rail-status"><Server size={16} /><span>BAR-INDEX PHASE</span></div>
      </aside>

      <section className="workspace">
        <div className={`health-banner ${error ? "is-error" : ""}`} role="status">
          <div className="health-summary"><span className="health-indicator" /><div><strong>{error || (loading ? "Loading analytical experiment" : "Viewer online · persisted phase evidence")}</strong><span>{experiment ? `${experiment.recordCount.toLocaleString()} records · ${experiment.observableCount.toLocaleString()} observable · ${experiment.initializingCount.toLocaleString()} initializing` : "Waiting for MongoDB"}</span></div></div>
          <div className="health-timing"><span>presentation only</span><small>no writes · no calculations</small></div>
        </div>

        <div className="instrument-heading">
          <div><div className="eyebrow">JOHN EHLERS DOMINANT CYCLE PHASE · V0.1 OUTPUT</div><div className="quote-line"><h2>Phase Angle Analysis</h2><span>0°–360°</span></div></div>
          <div className="evidence-pill"><Gauge size={15} /> persisted evidence</div>
        </div>

        <section className="experiment-band" aria-label="Experiment identity">
          <label>Experiment
            <select value={experiment ? experimentKey(experiment) : ""} onChange={(event) => {
              rawRequestKeys.current = {};
              setSelectedPoints({});
              setRawObservations({});
              setRawLoading({});
              setExperiment(experiments.find((item) => experimentKey(item) === event.target.value) ?? null);
            }}>
              {experiments.map((item) => <option key={experimentKey(item)} value={experimentKey(item)}>{item.collectionRunId} · size {item.seriesSize} · {item.solverVersion}</option>)}
            </select>
          </label>
          <div><small>Series size</small><strong>{experiment?.seriesSize ?? "--"}</strong></div>
          <div><small>Input</small><strong>{experiment?.inputSeriesType ?? "--"}</strong></div>
          <div><small>Solver</small><strong>{experiment?.solverName ?? "--"}</strong></div>
          <div><small>Version</small><strong>{experiment?.solverVersion ?? "--"}</strong></div>
        </section>

        <div className="partition-stack">
          {PARTITIONS.map((partition) => (
            <PartitionPhasePanel
              key={partition.id}
              partitionId={partition.id}
              color={partition.color}
              symbols={symbols.filter((item) => item.partitionId === partition.id)}
              selectedSymbol={selectedByPartition[partition.id] ?? ""}
              series={seriesByPartition[partition.id]}
              loading={Boolean(loadingPartitions[partition.id])}
              selectedPoint={selectedPoints[partition.id] ?? null}
              rawObservation={rawObservations[partition.id] ?? null}
              rawLoading={Boolean(rawLoading[partition.id])}
              onSymbolChange={(symbol) => {
                delete rawRequestKeys.current[partition.id];
                setSelectedPoints((current) => {
                  const next = { ...current };
                  delete next[partition.id];
                  return next;
                });
                setRawObservations((current) => {
                  const next = { ...current };
                  delete next[partition.id];
                  return next;
                });
                setRawLoading((current) => ({ ...current, [partition.id]: false }));
                setLoadingPartitions((current) => ({ ...current, [partition.id]: true }));
                setSelectedByPartition((current) => ({ ...current, [partition.id]: symbol }));
              }}
              onSelectPoint={selectPoint}
            />
          ))}
        </div>
      </section>
    </main>
  );
}
