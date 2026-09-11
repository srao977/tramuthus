"use client";

import { useEffect, useRef } from "react";
import {
  dispose,
  init,
  type Chart,
  type Crosshair,
  type DeepPartial,
  type KLineData,
  type NeighborData,
  type Nullable,
  type Styles,
} from "klinecharts";
import {
  fittedBarSpace,
  formatBarIndexTick,
  klineToPhasePoint,
  phasePointAtCoordinate,
  phasePointsToKLines,
  type PhaseKLineData,
} from "@/lib/phase-adapter";
import type { PhasePoint } from "@/lib/types";
import { useTheme } from "./ThemeProvider";

type Props = {
  symbol: string;
  color: string;
  points: PhasePoint[];
  onSelect: (point: PhasePoint) => void;
};

function chartStyles(dark: boolean, color: string): DeepPartial<Styles> {
  return {
    grid: {
      horizontal: { color: dark ? "#33464c" : "#dde3e5", style: "dashed" },
      vertical: { color: dark ? "#293b41" : "#e8ecee", style: "dashed" },
    },
    candle: {
      type: "area",
      area: {
        lineSize: 2,
        lineColor: color,
        value: "close",
        smooth: false,
        backgroundColor: [
          { offset: 0, color: `${color}22` },
          { offset: 1, color: `${color}00` },
        ],
        point: { show: true, color, radius: 2, animation: false, rippleRadius: 0 },
      },
      tooltip: {
        showRule: "follow_cross",
        showType: "standard",
        title: { template: "{ticker}" },
        legend: {
          template: (data: NeighborData<Nullable<KLineData>>) => {
            const point = data.current as PhaseKLineData | null;
            if (!point) return [];
            return [
              { title: "seq ", value: String(point.generatorSequenceNo) },
              { title: "phase ", value: `${point.phaseAngleDegrees.toFixed(4)}°` },
              { title: "state ", value: point.validityState },
              { title: "series ", value: String(point.seriesSize) },
            ];
          },
        },
      },
      priceMark: { high: { show: false }, low: { show: false }, last: { show: true } },
    },
    xAxis: {
      axisLine: { color: dark ? "#42565c" : "#cbd3d6" },
      tickText: { color: dark ? "#91a4aa" : "#637075" },
    },
    yAxis: {
      axisLine: { color: dark ? "#42565c" : "#cbd3d6" },
      tickText: { color: dark ? "#91a4aa" : "#637075" },
    },
    crosshair: {
      horizontal: { line: { color: dark ? "#9eb0b5" : "#44545a" } },
      vertical: { line: { color: dark ? "#9eb0b5" : "#44545a" } },
    },
  };
}

/** Renders persisted numeric phase only; null initialization records never enter the chart. */
export default function PhaseChart({ symbol, color, points, onSelect }: Props) {
  const elementRef = useRef<HTMLDivElement>(null);
  const chartRef = useRef<Chart | null>(null);
  const pointsRef = useRef(points);
  const onSelectRef = useRef(onSelect);
  const { resolvedTheme } = useTheme();

  useEffect(() => {
    pointsRef.current = points;
    onSelectRef.current = onSelect;
  }, [points, onSelect]);

  useEffect(() => {
    const element = elementRef.current;
    if (!element || !symbol || points.length === 0) return;
    const chart = init(element, {
      locale: "en-US",
      styles: chartStyles(document.documentElement.classList.contains("dark"), color),
      formatter: { formatDate: ({ timestamp }) => formatBarIndexTick(timestamp) },
    });
    if (!chart) return;
    chartRef.current = chart;
    chart.setSymbol({ ticker: `${symbol} PHASE`, pricePrecision: 4, volumePrecision: 0 });
    chart.setPeriod({ type: "minute", span: 1 });
    chart.setOffsetRightDistance(10);
    chart.setLeftMinVisibleBarCount(0);
    chart.setRightMinVisibleBarCount(0);
    const fitObservableSeries = () => {
      chart.setBarSpace(fittedBarSpace(element.clientWidth, pointsRef.current.length));
    };
    fitObservableSeries();
    chart.setDataLoader({
      getBars({ type, callback }) {
        if (type !== "init" && type !== "update") {
          callback([], { backward: false, forward: false });
          return;
        }
        callback(phasePointsToKLines(pointsRef.current), { backward: false, forward: false });
      },
    });

    const crosshair = { current: undefined as Crosshair | undefined };
    const selectPointFromChart = (data?: KLineData, dataIndex?: number) => {
      const point =
        klineToPhasePoint(data, pointsRef.current) ??
        phasePointAtCoordinate(pointsRef.current, data?.timestamp) ??
        (typeof dataIndex === "number" ? pointsRef.current[dataIndex] ?? null : null);
      if (point) onSelectRef.current(point);
    };
    const onCrosshair = (data?: unknown) => {
      crosshair.current = data as Crosshair | undefined;
    };
    const onCandleClick = (data?: unknown) => {
      const payload = data as { kLineData?: KLineData; dataIndex?: number } | undefined;
      selectPointFromChart(payload?.kLineData, payload?.dataIndex);
    };
    const onClick = (event: MouseEvent) => {
      const live = chartRef.current;
      if (!live || pointsRef.current.length === 0) return;
      const rect = element.getBoundingClientRect();
      const converted = live.convertFromPixel([
        { x: event.clientX - rect.left, y: event.clientY - rect.top },
      ]);
      const convertedPoint = Array.isArray(converted) ? converted[0] : converted;
      const dataList = live.getDataList();
      const byIndex = typeof convertedPoint?.dataIndex === "number"
        ? dataList[convertedPoint.dataIndex] ?? pointsRef.current[convertedPoint.dataIndex]
        : undefined;
      selectPointFromChart(
        byIndex ?? crosshair.current?.kLineData,
        convertedPoint?.dataIndex ?? crosshair.current?.dataIndex,
      );
    };
    chart.subscribeAction("onCrosshairChange", onCrosshair);
    chart.subscribeAction("onCandleBarClick", onCandleClick);
    element.addEventListener("click", onClick, true);
    const resizeObserver = new ResizeObserver(() => {
      chart.resize();
      fitObservableSeries();
    });
    resizeObserver.observe(element);
    return () => {
      resizeObserver.disconnect();
      element.removeEventListener("click", onClick, true);
      chart.unsubscribeAction("onCrosshairChange", onCrosshair);
      chart.unsubscribeAction("onCandleBarClick", onCandleClick);
      dispose(chart);
      chartRef.current = null;
    };
  }, [symbol, color, points.length]);

  useEffect(() => {
    chartRef.current?.setStyles(chartStyles(resolvedTheme !== "light", color));
  }, [resolvedTheme, color]);

  useEffect(() => {
    chartRef.current?.resetData();
  }, [points]);

  return <div ref={elementRef} className="phase-chart-canvas" role="img" aria-label={`${symbol} phase angle by generator sequence number`} />;
}
