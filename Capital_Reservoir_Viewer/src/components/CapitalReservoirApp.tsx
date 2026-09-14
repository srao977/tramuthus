"use client";

import { startTransition, useEffect, useMemo, useState } from "react";
import dynamic from "next/dynamic";
import { Activity, CirclePause, CirclePlay, Gauge, RefreshCw, RotateCcw, StepForward, WalletCards } from "lucide-react";
import { derivePipeStates, GOVERNED_SYMBOLS, mergeReservoirEvents, type CapitalReservoirEvent, type PipeState } from "@/lib/capital-reservoir";

const ReservoirChart = dynamic(() => import("./ReservoirChart"), { ssr: false });

type Mode = "LIVE" | "REPLAY";
type Run = { pipelineRunId: string; collectionRunId: string; runType: string; eventCount: number; complete: boolean };
const speeds = [1, 5, 10, 0] as const;

function money(value: number | null | undefined): string {
  if (value === null || value === undefined) return "--";
  return new Intl.NumberFormat("en-US", { style: "currency", currency: "USD", maximumFractionDigits: 2 }).format(value);
}

function PipeDetails({ states, selectedSequence, selectedSymbol, onSelect }: { states: PipeState[]; selectedSequence: number; selectedSymbol: string; onSelect(symbol: string): void }) {
  return (
    <section className="panel pipe-panel">
      <div className="panel-heading"><div><span className="eyebrow">RELATED DETAILS</span><h2>30 pipes at event #{selectedSequence || "--"}</h2></div><span className="selection-chip">{selectedSymbol}</span></div>
      <div className="pipe-grid">
        {states.map((pipe) => <button type="button" key={pipe.symbol} className={`pipe-cell ${pipe.symbol === selectedSymbol ? "selected" : ""} ${pipe.activeQuantity ? "active" : ""}`} onClick={() => onSelect(pipe.symbol)} title={`${pipe.symbol}: ${pipe.activeQuantity} active shares`}><strong>{pipe.symbol}</strong><span>{pipe.activeQuantity ? `${pipe.activeQuantity.toLocaleString()} sh` : "idle"}</span></button>)}
      </div>
      <div className="table-scroll"><table><thead><tr><th>Pipe</th><th>Status</th><th>Quantity</th><th>Deployed</th><th>Last price</th><th>Net flow</th><th>Cause</th><th>Event</th></tr></thead><tbody>
        {states.map((pipe) => <tr key={pipe.symbol} className={pipe.symbol === selectedSymbol ? "selected-row" : ""} onClick={() => onSelect(pipe.symbol)}><td><strong>{pipe.symbol}</strong></td><td><span className={`status-tag ${pipe.activeQuantity ? "on" : ""}`}>{pipe.activeQuantity ? "ACTIVE" : "IDLE"}</span></td><td>{pipe.activeQuantity.toLocaleString()}</td><td>{money(pipe.deployedCapital)}</td><td>{money(pipe.lastExecutionPrice)}</td><td className={pipe.netFlow < 0 ? "negative" : pipe.netFlow > 0 ? "positive" : ""}>{money(pipe.netFlow)}</td><td>{pipe.lastCause ?? "--"}</td><td>{pipe.lastEventSequence || "--"}</td></tr>)}
      </tbody></table></div>
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
  const current = detailEvents.at(-1);
  const ending = [...detailEvents].reverse().find((event) => event.eventType === "RUN_END");
  const selectedEvent = detailEvents.find((event) => event.eventSequence === selectedSequence);
  const activeCount = pipeStates.filter((pipe) => pipe.activeQuantity > 0).length;

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
    <section className="metrics"><article><WalletCards /><span>Reservoir cash</span><strong>{money(current?.reservoirAfter)}</strong></article><article><Gauge /><span>Initial reservoir</span><strong>{money(detailEvents[0]?.initialReservoir)}</strong></article><article><Activity /><span>Active pipes</span><strong>{activeCount} / 30</strong></article><article><CirclePlay /><span>Event sequence</span><strong>{selectedSequence || "--"} / {events.at(-1)?.eventSequence ?? 0}</strong></article><article><CirclePause /><span>Total P&amp;L</span><strong className={(ending?.totalPnl ?? 0) < 0 ? "negative" : "positive"}>{money(ending?.totalPnl)}</strong></article></section>
    <section className="charts-grid"><article className="panel chart-panel"><div className="panel-heading"><div><span className="eyebrow">COMMON LEDGER</span><h2>Reservoir continuity</h2></div><span>event sequence → cash</span></div>{visibleEvents.length ? <ReservoirChart events={visibleEvents} kind="reservoir" selectedSymbol={selectedSymbol} onSequence={setHoveredSequence} /> : <div className="empty-state">Waiting for Capital Reservoir events.</div>}</article><article className="panel chart-panel"><div className="panel-heading"><div><span className="eyebrow">BIDIRECTIONAL PIPE</span><h2>{selectedSymbol} signed flow</h2></div><span>inflow + / outflow −</span></div>{visibleEvents.length ? <ReservoirChart events={visibleEvents} kind="flow" selectedSymbol={selectedSymbol} onSequence={setHoveredSequence} /> : <div className="empty-state">Select a run with execution flow.</div>}</article></section>
    <PipeDetails states={pipeStates} selectedSequence={selectedSequence} selectedSymbol={selectedSymbol} onSelect={setSelectedSymbol} />
    <section className="panel inspector"><div className="panel-heading"><div><span className="eyebrow">EVENT INSPECTOR</span><h2>{selectedEvent ? `${selectedEvent.eventType} · #${selectedEvent.eventSequence}` : "No event selected"}</h2></div><span>{selectedEvent?.processedAt ? new Date(selectedEvent.processedAt).toLocaleString() : "Move a chart crosshair"}</span></div><pre>{selectedEvent ? JSON.stringify(selectedEvent, null, 2) : "Crosshair selection resolves the reservoir and all 30 pipes at one authoritative event sequence."}</pre></section>
  </main>;
}