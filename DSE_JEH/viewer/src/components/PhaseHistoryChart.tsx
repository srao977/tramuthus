"use client";

import { useEffect, useRef } from "react";
import { dispose, init, type KLineData } from "klinecharts";
import type { PhaseEvidence } from "@/lib/types";

export default function PhaseHistoryChart({ symbol, inputs }: { symbol: string; inputs: PhaseEvidence[] }) {
  const elementRef = useRef<HTMLDivElement>(null);
  const observableCount = inputs.filter((input) => input.phase_observable && input.phase_angle_degrees !== null).length;

  useEffect(() => {
    const element = elementRef.current;
    const observable = inputs.flatMap((input) => input.phase_observable && input.phase_angle_degrees !== null ? [{
      timestamp: input.generator_sequence_no,
      open: input.phase_angle_degrees,
      high: input.phase_angle_degrees,
      low: input.phase_angle_degrees,
      close: input.phase_angle_degrees,
      volume: 0,
    } satisfies KLineData] : []);
    if (!element || !observable.length) return;
    const chart = init(element, {
      locale: "en-US",
      formatter: { formatDate: ({ timestamp }) => String(Math.round(timestamp)) },
      styles: {
        grid: { horizontal: { color: "#2b3c43", style: "dashed" }, vertical: { color: "#26363c", style: "dashed" } },
        candle: { type: "area", area: { lineColor: "#e3b341", lineSize: 2, value: "close", backgroundColor: [{ offset: 0, color: "rgba(227,179,65,.22)" }, { offset: 1, color: "rgba(227,179,65,0)" }] }, tooltip: { showRule: "follow_cross", showType: "standard" }, priceMark: { high: { show: false }, low: { show: false } } },
        xAxis: { tickText: { color: "#8fa0a5" }, axisLine: { color: "#42545a" } },
        yAxis: { tickText: { color: "#8fa0a5" }, axisLine: { color: "#42545a" } },
      },
    });
    if (!chart) return;
    chart.setSymbol({ ticker: `${symbol} PHASE`, pricePrecision: 4, volumePrecision: 0 });
    chart.setPeriod({ type: "minute", span: 1 });
    chart.setOffsetRightDistance(8);
    chart.setBarSpace(Math.max(4, Math.min(18, (element.clientWidth - 70) / observable.length)));
    chart.setDataLoader({ getBars({ type, callback }) { callback(type === "init" || type === "update" ? observable : [], { backward: false, forward: false }); } });
    const observer = new ResizeObserver(() => chart.resize());
    observer.observe(element);
    return () => { observer.disconnect(); dispose(chart); };
  }, [symbol, inputs]);

  if (!observableCount) return <div className="chart-loading">No observable phase history for {symbol}. Initialization remains unplotted.</div>;
  return <div ref={elementRef} className="history-chart" role="img" aria-label={`${symbol} phase angle by generator sequence number`} />;
}