"use client";

import dynamic from "next/dynamic";
import { useRouter } from "next/navigation";
import { useState } from "react";
import { Activity, CircleDot, RefreshCw, ShieldCheck } from "lucide-react";
import { currentCounts, entityEvidence, initializingSymbols, ZONE_LABELS } from "@/lib/presentation";
import type { RunEvidence, RunSummary } from "@/lib/types";
import EvidenceDetails from "./EvidenceDetails";
import EventHistory from "./EventHistory";
import PolarPhaseDiagram from "./PolarPhaseDiagram";

const PhaseHistoryChart = dynamic(() => import("./PhaseHistoryChart"), {
  ssr: false,
  loading: () => <div className="chart-loading">Loading phase-history surface...</div>,
});

export default function Dashboard({ runs, run }: { runs: RunSummary[]; run: RunEvidence }) {
  const router = useRouter();
  const symbols = [...new Set(run.inputs.map((input) => input.symbol))].sort();
  const [selectedSymbol, setSelectedSymbol] = useState(Object.keys(run.states).sort()[0] ?? symbols[0] ?? "");
  const selected = entityEvidence(run, selectedSymbol);
  const counts = currentCounts(run);
  const initializing = initializingSymbols(run);

  return (
    <main className="app-shell">
      <header className="topbar">
        <div className="brand"><span className="brand-mark"><CircleDot size={22} /></span><div><h1>DSE_JEH</h1><p>CSV proving evidence · observational only</p></div></div>
        <div className="run-controls">
          <label>Proving / Replay Run<select value={run.metadata.run_id} onChange={(event) => router.push(`/?run=${encodeURIComponent(event.target.value)}`)}>{runs.map((item) => <option key={item.run_id} value={item.run_id}>{item.run_id} · {item.collection_run_id}</option>)}</select></label>
          <button type="button" className="icon-button" onClick={() => router.refresh()} title="Refresh evidence" aria-label="Refresh evidence"><RefreshCw size={17} /></button>
        </div>
      </header>

      <section className="status-strip"><div><ShieldCheck size={16} /><strong>{run.metadata.replay_matched ? "REPLAY MATCHED" : "REPLAY MISMATCH"}</strong></div><span>Run {run.metadata.run_id}</span><span>Source {run.metadata.collection_run_id}</span><span>Series {run.metadata.series_size}</span><span>{run.metadata.ranking_mode}</span></section>
      <section className="metric-grid" aria-label="Current state counts"><Metric label="Total" value={counts.total} /><Metric label="Observable" value={counts.observable} tone="green" /><Metric label="Initializing" value={counts.initializing} tone="muted" /><Metric label="HOP_ON" value={counts.zones.HOP_ON} tone="green" /><Metric label="MOMENTUM_HOLD" value={counts.zones.MOMENTUM_HOLD} tone="amber" /><Metric label="HOP_OFF" value={counts.zones.HOP_OFF} tone="red" /><Metric label="DISREGARD" value={counts.zones.DISREGARD} tone="muted" /></section>

      <section className="primary-grid">
        <article className="panel polar-panel"><PanelHeading eyebrow="CURRENT CIRCULAR STATE" title="Polar Phase Diagram" detail={`${counts.observable} observable entities`} /><PolarPhaseDiagram states={run.states} selectedSymbol={selectedSymbol} onSelect={setSelectedSymbol} /><div className="initializing-list"><strong>Initializing ({initializing.length})</strong><span>{initializing.length ? initializing.join(" · ") : "None"}</span></div></article>
        <article className="panel entity-panel"><PanelHeading eyebrow="SELECTED ENTITY" title={selectedSymbol || "No entity"} detail={selected.state ? ZONE_LABELS[selected.state.current_zone] : "INITIALIZING"} /><label className="symbol-select">Entity<select value={selectedSymbol} onChange={(event) => setSelectedSymbol(event.target.value)}>{symbols.map((symbol) => <option key={symbol}>{symbol}</option>)}</select></label><EvidenceDetails run={run} symbol={selectedSymbol} evidence={selected} /></article>
      </section>

      <section className="panel history-panel"><PanelHeading eyebrow="LINEAR HISTORY VIEW" title={`${selectedSymbol} Phase History`} detail="generator_sequence_no → phase_angle_degrees" /><PhaseHistoryChart symbol={selectedSymbol} inputs={selected.inputs} /><p className="science-note">Phase is circular: 0° equals 360°. Wraps remain visible; initialization values are not fabricated.</p></section>
      <section className="panel events-panel"><PanelHeading eyebrow="AUTHORITATIVE GO OUTPUT" title="Transition / Candidate / Decision Evidence" detail={`${selected.events.length + selected.decisions.length} selected-entity events`} /><EventHistory events={selected.events} decisions={selected.decisions} /></section>
      <footer><Activity size={14} /> Phase velocity: NOT YET DEFINED · Decision evidence is not broker execution.</footer>
    </main>
  );
}

function Metric({ label, value, tone = "" }: { label: string; value: number; tone?: string }) {
  return <div className={`metric ${tone}`}><span>{label}</span><strong>{value}</strong></div>;
}

function PanelHeading({ eyebrow, title, detail }: { eyebrow: string; title: string; detail: string }) {
  return <header className="panel-heading"><div><p>{eyebrow}</p><h2>{title}</h2></div><span>{detail}</span></header>;
}