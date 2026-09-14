"use client";

import { useEffect, useRef } from "react";
import { dispose, init, type Chart, type Crosshair, type DeepPartial, type KLineData, type Styles } from "klinecharts";
import type { CapitalReservoirEvent } from "@/lib/capital-reservoir";

type Props = {
  events: CapitalReservoirEvent[];
  kind: "reservoir" | "flow";
  selectedSymbol: string;
  onSequence: (sequence: number) => void;
};

type ReservoirKLine = KLineData & { eventSequence: number };
const MINUTE = 60_000;

function chartStyles(kind: Props["kind"]): DeepPartial<Styles> {
  return {
    grid: {
      horizontal: { color: "#243237", style: "dashed" },
      vertical: { color: "#202c30", style: "dashed" },
    },
    candle: kind === "reservoir" ? {
      type: "area",
      area: {
        lineColor: "#57d6b1", lineSize: 2, value: "close", smooth: false,
        backgroundColor: [{ offset: 0, color: "rgba(87,214,177,.24)" }, { offset: 1, color: "rgba(87,214,177,0)" }],
        point: { show: true, color: "#b8ffea", radius: 2, animation: false, rippleRadius: 0 },
      },
    } : {
      type: "candle_solid",
      bar: { upColor: "#57d6b1", downColor: "#ff6b62", noChangeColor: "#66767b", upBorderColor: "#57d6b1", downBorderColor: "#ff6b62", noChangeBorderColor: "#66767b", upWickColor: "#57d6b1", downWickColor: "#ff6b62", noChangeWickColor: "#66767b" },
    },
    xAxis: { axisLine: { color: "#425257" }, tickText: { color: "#90a0a4" } },
    yAxis: { axisLine: { color: "#425257" }, tickText: { color: "#90a0a4" } },
    crosshair: {
      horizontal: { line: { color: "#e4f2ee" } },
      vertical: { line: { color: "#e4f2ee" } },
    },
  };
}

function chartData(events: CapitalReservoirEvent[], kind: Props["kind"], symbol: string): ReservoirKLine[] {
  return events.map((event) => {
    const value = kind === "reservoir" ? event.reservoirAfter : event.symbol === symbol ? (event.signedFlowAmount ?? 0) : 0;
    return {
      timestamp: event.eventSequence * MINUTE,
      eventSequence: event.eventSequence,
      open: kind === "flow" ? 0 : value,
      high: Math.max(0, value),
      low: Math.min(0, value),
      close: value,
      volume: Math.abs(event.signedFlowAmount ?? 0),
    };
  });
}

export default function ReservoirChart({ events, kind, selectedSymbol, onSequence }: Props) {
  const hostRef = useRef<HTMLDivElement>(null);
  const chartRef = useRef<Chart | null>(null);
  const eventsRef = useRef(events);
  const symbolRef = useRef(selectedSymbol);
  const onSequenceRef = useRef(onSequence);

  useEffect(() => {
    eventsRef.current = events;
    symbolRef.current = selectedSymbol;
    onSequenceRef.current = onSequence;
  }, [events, selectedSymbol, onSequence]);

  useEffect(() => {
    const host = hostRef.current;
    if (!host) return;
    const chart = init(host, {
      locale: "en-US",
      styles: chartStyles(kind),
      formatter: { formatDate: ({ timestamp }) => `#${Math.round(timestamp / MINUTE)}` },
    });
    if (!chart) return;
    chartRef.current = chart;
    chart.setSymbol({ ticker: kind === "reservoir" ? "COMMON RESERVOIR" : `${selectedSymbol} SIGNED FLOW`, pricePrecision: 2, volumePrecision: 0 });
    chart.setPeriod({ type: "minute", span: 1 });
    chart.setOffsetRightDistance(12);
    chart.setDataLoader({
      getBars({ type, callback }) {
        if (type !== "init" && type !== "update") {
          callback([], { backward: false, forward: false });
          return;
        }
        callback(chartData(eventsRef.current, kind, symbolRef.current), { backward: false, forward: false });
      },
    });
    const onCrosshair = (payload?: unknown) => {
      const crosshair = payload as Crosshair | undefined;
      const line = crosshair?.kLineData as ReservoirKLine | undefined;
      const indexedEvent = typeof crosshair?.dataIndex === "number" ? eventsRef.current[crosshair.dataIndex] : undefined;
      const sequence = line?.eventSequence || (line?.timestamp ? Math.round(line.timestamp / MINUTE) : 0) || indexedEvent?.eventSequence || 0;
      if (sequence) onSequenceRef.current(sequence);
    };
    const onChartMove = (event: MouseEvent) => {
      const rect = host.getBoundingClientRect();
      const converted = chart.convertFromPixel([{ x: event.clientX - rect.left, y: event.clientY - rect.top }]);
      const point = Array.isArray(converted) ? converted[0] : converted;
      const selected = typeof point?.dataIndex === "number" ? eventsRef.current[point.dataIndex] : undefined;
      if (selected) onSequenceRef.current(selected.eventSequence);
    };
    chart.subscribeAction("onCrosshairChange", onCrosshair);
    host.addEventListener("mousemove", onChartMove);
    const resize = new ResizeObserver(() => chart.resize());
    resize.observe(host);
    return () => {
      resize.disconnect();
      chart.unsubscribeAction("onCrosshairChange", onCrosshair);
      host.removeEventListener("mousemove", onChartMove);
      dispose(chart);
      chartRef.current = null;
    };
  }, [kind, selectedSymbol]);

  useEffect(() => {
    chartRef.current?.resetData();
  }, [events, kind, selectedSymbol]);

  return <div className="chart-host" ref={hostRef} role="img" aria-label={kind === "reservoir" ? "Common reservoir by event sequence" : `${selectedSymbol} signed flow by event sequence`} />;
}