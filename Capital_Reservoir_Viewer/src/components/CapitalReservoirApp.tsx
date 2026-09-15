"use client";

import { startTransition, useEffect, useMemo, useState } from "react";
import dynamic from "next/dynamic";
import { CirclePause, CirclePlay, RefreshCw, RotateCcw, StepForward, WalletCards } from "lucide-react";
import { deriveCapitalSummary, derivePipeStates, GOVERNED_SYMBOLS, mergeReservoirEvents, type CapitalReservoirEvent, type CapitalSummary, type PipeState } from "@/lib/capital-reservoir";

const ReservoirChart = dynamic(() => import("./ReservoirChart"), { ssr: false });

type Mode = "LIVE" | "REPLAY";
type Run = { pipelineRunId: string; collectionRunId: string; runType: string; eventCount: number; complete: boolean };
const speeds = [1, 5, 10, 0] as const;

function money(value: number | null | undefined): string {
  if (value === null || value === undefined) return "--";
  return new Intl.NumberFormat("en-US", { style: "currency", currency: "USD", maximumFractionDigits: 2 }).format(value);
}

function signedMoney(value: number | null | undefined): string {
  if (value === null || value === undefined) return "--";
  return `${value > 0 ? "+" : ""}${money(value)}`;
}

function time(value: string | null | undefined): string {
  return value ? new Date(value).toLocaleString() : "--";
}

function percent(value: number | null): string {
  return value === null ? "--" : `${value.toFixed(1)}%`;
}

function CapitalOverview({ summary, selectedSequence, totalEvents }: { summary: CapitalSummary; selectedSequence: number; totalEvents: number }) {
  const deployedWidth = Math.min(100, Math.max(0, summary.utilizationPct ?? 0));
  return <section className="capital-overview" aria-label="Common capital reservoir summary">
    <div className="capital-equation">
      <article className="capital-anchor"><span>INITIAL CAPITAL</span><strong>{money(summary.initialCapital)}</strong><small>One common reservoir system</small></article>
      <span className="equation-symbol">=</span>
      <div className="capital-locations">
        <article><span>RESERVOIR CASH</span><strong>{money(summary.reservoirCash)}</strong><small>{percent(summary.availablePct)} available</small></article>
        <span className="equation-symbol">+</span>
        <article><span>DEPLOYED MARKED CAPITAL</span><strong>{money(summary.deployedMarkedCapital)}</strong><small>{summary.deployedMarkedCapital === null ? "Published at RUN_END" : `${percent(summary.utilizationPct)} deployed`}</small></article>
      </div>
      <span className="equation-symbol">=</span>
      <article className="capital-total"><span>TOTAL MARKED CAPITAL</span><strong>{money(summary.totalMarkedCapital)}</strong><small>{summary.totalMarkedCapital === null ? "Published at RUN_END" : `Total P&L ${signedMoney(summary.totalPnl)}`}</small></article>
    </div>
    <div className="utilization-row">
      <div><span>RESERVOIR UTILIZATION</span><strong>{percent(summary.utilizationPct)}</strong></div>
      <div className="utilization-track" role="progressbar" aria-label="Reservoir capital deployed" aria-valuemin={0} aria-valuemax={100} aria-valuenow={summary.utilizationPct ?? undefined}><span style={{ width: `${deployedWidth}%` }} /></div>
      <div className="utilization-legend"><span><i className="available-key" />Available {percent(summary.availablePct)}</span><span><i className="deployed-key" />Deployed {percent(summary.utilizationPct)}</span></div>
    </div>
    <div className="economic-summary">
      <article><span>REALIZED P&amp;L</span><strong className={(summary.realizedPnl ?? 0) < 0 ? "negative" : "positive"}>{signedMoney(summary.realizedPnl)}</strong></article>
      <article><span>UNREALIZED P&amp;L</span><strong className={(summary.unrealizedPnl ?? 0) < 0 ? "negative" : "positive"}>{signedMoney(summary.unrealizedPnl)}</strong></article>
      <article><span>TOTAL P&amp;L</span><strong className={(summary.totalPnl ?? 0) < 0 ? "negative" : "positive"}>{signedMoney(summary.totalPnl)}</strong></article>
      <article><span>CAPITAL CURRENTLY DEPLOYED</span><strong>{summary.activeSymbols} symbols</strong></article>
      <article><span>SYMBOLS PARTICIPATED</span><strong>{summary.participatingSymbols} / 30</strong></article>
      <article><span>EVIDENCE</span><strong>#{selectedSequence || "--"} / {totalEvents}</strong></article>
    </div>
  </section>;
}

function PipeDetails({ states, summary, selectedSymbol, onSelect }: { states: PipeState[]; summary: CapitalSummary; selectedSymbol: string; onSelect(symbol: string): void }) {
  const selected = states.find((pipe) => pipe.symbol === selectedSymbol);
  const cycle = selected?.currentCycle ?? selected?.completedCycles.at(-1) ?? null;
  const cumulativeRealizedPnl = selected?.completedCycles.reduce((total, completed) => total + (completed.realizedPnl ?? 0), 0) ?? null;
  return (
    <section className="panel pipe-panel">
      <div className="panel-heading"><div><span className="eyebrow">COMMON RESERVOIR TOPOLOGY</span><h2>One capital system · 30 bidirectional symbol pipes</h2></div><span className="selection-chip">{selectedSymbol}</span></div>
      <div className="reservoir-node"><div><WalletCards size={18} /><span>COMMON CAPITAL RESERVOIR</span><strong>{money(summary.reservoirCash)}</strong></div><p>Capital deploys through confirmed BUY and returns through confirmed SELL. Each pipe is a connection, not a separate cash account.</p></div>
      <div className="pipe-connector" aria-hidden="true" />
      <div className="pipe-grid">
        {states.map((pipe) => <button type="button" key={pipe.symbol} className={`pipe-cell ${pipe.symbol === selectedSymbol ? "selected" : ""} ${pipe.activeQuantity ? "active" : ""}`} onClick={() => onSelect(pipe.symbol)} title={`${pipe.symbol}: ${pipe.activeQuantity} active shares`}><strong>{pipe.symbol}</strong><span>{pipe.activeQuantity ? "DEPLOYED" : "IDLE"}</span></button>)}
      </div>
      <div className="table-scroll"><table><thead><tr><th>Pipe</th><th>Capital State</th><th>Quantity</th><th>Capital Deployed</th><th>Capital Returned</th><th>Realized P&amp;L</th><th>Last Price</th><th>Last Cause</th><th>Event</th></tr></thead><tbody>
        {states.map((pipe) => <tr key={pipe.symbol} className={pipe.symbol === selectedSymbol ? "selected-row" : ""} onClick={() => onSelect(pipe.symbol)}><td><strong>{pipe.symbol}</strong></td><td><span className={`status-tag ${pipe.activeQuantity ? "on" : ""}`}>{pipe.activeQuantity ? "DEPLOYED" : "IDLE"}</span></td><td>{pipe.activeQuantity.toLocaleString()}</td><td>{money(pipe.capitalDeployed)}</td><td>{money(pipe.capitalReturned)}</td><td className={(pipe.realizedPnl ?? 0) < 0 ? "negative" : pipe.realizedPnl !== null ? "positive" : ""}>{signedMoney(pipe.realizedPnl)}</td><td>{money(pipe.lastExecutionPrice)}</td><td>{pipe.lastCause ?? "--"}</td><td>{pipe.lastEventSequence || "--"}</td></tr>)}
      </tbody></table></div>
      <div className="position-ledger">
        <div className="position-title"><div><span className="eyebrow">SELECTED POSITION</span><h2>{selectedSymbol} · {selected?.activeQuantity ? "ACTIVE" : "IDLE"}</h2></div><span>{selected?.completedCycles.length ?? 0} completed cycle{selected?.completedCycles.length === 1 ? "" : "s"}</span></div>
        <section><h3>Entry</h3><dl><div><dt>Cause</dt><dd>{cycle?.entryCause ?? "--"}</dd></div><div><dt>Capital deployed</dt><dd>{money(cycle?.capitalDeployed)}</dd></div><div><dt>Quantity</dt><dd>{cycle?.quantity.toLocaleString() ?? "--"}</dd></div><div><dt>Entry price</dt><dd>{money(cycle?.entryPrice)}</dd></div><div><dt>Entry event</dt><dd>{cycle ? `#${cycle.entryEventSequence}` : "--"}</dd></div><div><dt>Trigger time</dt><dd>{time(cycle?.triggerTime)}</dd></div><div><dt>Execution time</dt><dd>{time(cycle?.entryExecutionTime)}</dd></div></dl></section>
        <section><h3>Current</h3><dl><div><dt>Current quantity</dt><dd>{selected?.activeQuantity.toLocaleString() ?? "--"}</dd></div><div><dt>Last execution price</dt><dd>{money(selected?.lastExecutionPrice)}</dd></div><div><dt>Cumulative reservoir flow</dt><dd className={(selected?.cumulativeReservoirFlow ?? 0) < 0 ? "negative" : (selected?.cumulativeReservoirFlow ?? 0) > 0 ? "positive" : ""}>{selected?.lastEventSequence ? signedMoney(selected.cumulativeReservoirFlow) : "--"}</dd></div></dl></section>
        <section><h3>Exit</h3><dl><div><dt>Capital returned</dt><dd>{money(cycle?.capitalReturned)}</dd></div><div><dt>Exit cause</dt><dd>{cycle?.exitCause ?? "--"}</dd></div><div><dt>Exit event</dt><dd>{cycle?.exitEventSequence ? `#${cycle.exitEventSequence}` : "--"}</dd></div><div><dt>Exit execution time</dt><dd>{time(cycle?.exitExecutionTime)}</dd></div></dl></section>
        <section><h3>Result</h3><dl><div><dt>Realized P&amp;L</dt><dd className={(cycle?.realizedPnl ?? 0) < 0 ? "negative" : cycle?.realizedPnl !== null && cycle?.realizedPnl !== undefined ? "positive" : ""}>{signedMoney(cycle?.realizedPnl)}</dd></div><div><dt>Cumulative realized P&amp;L</dt><dd className={(cumulativeRealizedPnl ?? 0) < 0 ? "negative" : selected?.completedCycles.length ? "positive" : ""}>{selected?.completedCycles.length ? signedMoney(cumulativeRealizedPnl) : "--"}</dd></div></dl></section>
      </div>
    </section>
  );
}

export default function CapitalReservoirApp() {
  const [mode, setMode] = useState<Mode>("REPLAY");
  const [runs, setRuns] = useState<Run[]>([]);
  const [runId, setRunId] = useState("");
  const [events, setEvents] = useState<CapitalReservoirEvent[]>([]);
  const [cursor, setCursor] = useState(0);
  const [hoveredSequence, setHoveredSequence] = useState(0);
  const [selectedSymbol, setSelectedSymbol] = useState<string>(GOVERNED_SYMBOLS[0]);
  const [playing, setPlaying] = useState(false);
  const [speed, setSpeed] = useState<(typeof speeds)[number]>(1);
  const [liveGeneration, setLiveGeneration] = useState(0);
  const [status, setStatus] = useState("Discovering replay runs");
  const [error, setError] = useState("");

  useEffect(() => {
    fetch("/api/runs", { cache: "no-store" }).then(async (response) => {
      const body = await response.json();
      if (!response.ok) throw new Error(body.error);
      setRuns(body.runs);
      if (body.runs[0]) setRunId(body.runs[0].pipelineRunId);
      setStatus(body.runs.length ? "Replay catalog ready" : "No persisted runs");
    }).catch((reason) => { setError(reason.message); setStatus("Replay unavailable"); });
  }, []);

  useEffect(() => {
    if (mode !== "REPLAY" || !runId) return;
    fetch(`/api/runs/${encodeURIComponent(runId)}/events`, { cache: "no-store" }).then(async (response) => {
      const body = await response.json();
      if (!response.ok) throw new Error(body.error);
      startTransition(() => {
        setEvents(body.events);
        setCursor(body.events.length);
        setHoveredSequence(body.events.at(-1)?.eventSequence ?? 0);
      });
      setStatus(`${body.events.length} authoritative events loaded`);
      setError("");
    }).catch((reason) => { setError(reason.message); setStatus("Replay unavailable"); });
  }, [mode, runId]);

  useEffect(() => {
    if (mode !== "LIVE") return;
    const source = new EventSource("/api/live");
    source.addEventListener("ready", () => { setStatus("Live bridge connected"); setError(""); });
    source.addEventListener("reservoir", (message) => {
      const event = JSON.parse((message as MessageEvent).data) as CapitalReservoirEvent;
      setEvents((current) => {
        const merged = mergeReservoirEvents([...current, event]);
        setCursor(merged.length); setHoveredSequence(event.eventSequence);
        return merged;
      });
    });
    source.addEventListener("fault", (message) => { setError(JSON.parse((message as MessageEvent).data).message); setStatus("Live runtime unavailable"); });
    source.onerror = () => setStatus("Live bridge reconnecting");
    return () => source.close();
  }, [mode, liveGeneration]);

  useEffect(() => {
    if (!playing || mode !== "REPLAY" || cursor >= events.length) return;
    const delay = speed === 0 ? 24 : Math.max(40, 700 / speed);
    const timer = window.setInterval(() => setCursor((value) => {
      const next = Math.min(events.length, value + 1);
      setHoveredSequence(events[next - 1]?.eventSequence ?? 0);
      if (next >= events.length) setPlaying(false);
      return next;
    }), delay);
    return () => window.clearInterval(timer);
  }, [playing, mode, cursor, events, speed]);

  const visibleEvents = useMemo(() => events.slice(0, mode === "LIVE" ? events.length : cursor), [events, cursor, mode]);
  const selectedSequence = Math.min(hoveredSequence || visibleEvents.at(-1)?.eventSequence || 0, visibleEvents.at(-1)?.eventSequence || 0);
  const detailEvents = useMemo(() => visibleEvents.filter((event) => event.eventSequence <= selectedSequence), [visibleEvents, selectedSequence]);
  const pipeStates = useMemo(() => derivePipeStates(detailEvents, [...GOVERNED_SYMBOLS]), [detailEvents]);
  const capitalSummary = useMemo(() => deriveCapitalSummary(detailEvents, [...GOVERNED_SYMBOLS]), [detailEvents]);
  const selectedEvent = detailEvents.find((event) => event.eventSequence === selectedSequence);

  function changeMode(nextMode: Mode) {
    if (nextMode === mode) return;
    setPlaying(false);
    setEvents([]);
    setCursor(0);
    setHoveredSequence(0);
    setError("");
    setStatus(nextMode === "LIVE" ? "Connecting to DSE_JEH gRPC bridge" : "Loading ordered reservoir events");
    setMode(nextMode);
  }

  return <main className="app-shell">
    <header className="topbar"><div className="brand"><div className="brand-mark"><WalletCards size={20} /></div><div><h1>Capital Reservoir</h1><p>DSE_JEH common-capital evidence viewer</p></div></div><div className="top-status"><span className={`signal ${error ? "fault" : ""}`} /><div><strong>{mode} · READ ONLY</strong><span>{status}</span></div></div></header>
    <section className="control-band"><div className="segmented" aria-label="Viewer mode">{(["LIVE", "REPLAY"] as Mode[]).map((value) => <button type="button" key={value} className={mode === value ? "active" : ""} onClick={() => changeMode(value)}>{value}</button>)}</div>
      {mode === "REPLAY" && <><label className="run-picker"><span>PIPELINE RUN</span><select value={runId} onChange={(event) => { setStatus("Loading ordered reservoir events"); setRunId(event.target.value); }}>{runs.map((run) => <option key={run.pipelineRunId} value={run.pipelineRunId}>{run.pipelineRunId} · {run.runType} · {run.complete ? "complete" : "open"}</option>)}</select></label><div className="replay-controls"><button title={playing ? "Pause" : "Play"} aria-label={playing ? "Pause" : "Play"} onClick={() => setPlaying(!playing)}>{playing ? <CirclePause /> : <CirclePlay />}</button><button title="Restart" aria-label="Restart" onClick={() => { setPlaying(false); setCursor(0); setHoveredSequence(0); }}><RotateCcw /></button><button title="Step one event" aria-label="Step one event" onClick={() => { const next = Math.min(events.length, cursor + 1); setCursor(next); setHoveredSequence(events[next - 1]?.eventSequence ?? 0); }}><StepForward /></button>{speeds.map((value) => <button className={`speed ${speed === value ? "active" : ""}`} key={value} onClick={() => setSpeed(value)}>{value === 0 ? "MAX" : `${value}x`}</button>)}</div></>}
      {mode === "LIVE" && <button className="icon-command" title="Reconnect live stream" aria-label="Reconnect live stream" onClick={() => { setStatus("Reconnecting to DSE_JEH gRPC bridge"); setLiveGeneration((value) => value + 1); }}><RefreshCw /></button>}
    </section>
    {error && <div className="error-banner"><strong>Connection fault</strong><span>{error}</span></div>}
    <CapitalOverview summary={capitalSummary} selectedSequence={selectedSequence} totalEvents={events.at(-1)?.eventSequence ?? 0} />
    <section className="charts-grid"><article className="panel chart-panel"><div className="panel-heading"><div><span className="eyebrow">COMMON LEDGER</span><h2>Reservoir continuity</h2></div><span>event sequence → cash</span></div>{visibleEvents.length ? <ReservoirChart events={visibleEvents} kind="reservoir" selectedSymbol={selectedSymbol} onSequence={setHoveredSequence} /> : <div className="empty-state">Waiting for Capital Reservoir events.</div>}</article><article className="panel chart-panel"><div className="panel-heading"><div><span className="eyebrow">RESERVOIR PERSPECTIVE</span><h2>{selectedSymbol} signed reservoir flow</h2></div><span>BUY outflow &lt; 0 · SELL inflow &gt; 0</span></div>{visibleEvents.length ? <ReservoirChart events={visibleEvents} kind="flow" selectedSymbol={selectedSymbol} onSequence={setHoveredSequence} /> : <div className="empty-state">Select a run with execution flow.</div>}</article></section>
    <PipeDetails states={pipeStates} summary={capitalSummary} selectedSymbol={selectedSymbol} onSelect={setSelectedSymbol} />
    <section className="panel inspector"><div className="panel-heading"><div><span className="eyebrow">EVENT INSPECTOR</span><h2>{selectedEvent ? `${selectedEvent.eventType} · #${selectedEvent.eventSequence}` : "No event selected"}</h2></div><span>{selectedEvent?.processedAt ? new Date(selectedEvent.processedAt).toLocaleString() : "Move a chart crosshair"}</span></div><pre>{selectedEvent ? JSON.stringify(selectedEvent, null, 2) : "Crosshair selection resolves the reservoir and all 30 pipes at one authoritative event sequence."}</pre></section>
  </main>;
}