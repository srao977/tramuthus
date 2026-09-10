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
import { useTheme } from "next-themes";
import {
  formatBarIndexTick,
  klineToObservation,
  observationAtChartTimestamp,
  observationsToKLines,
  type LabKLineData,
} from "@/lib/kline-view-adapter";
import { visiblePrefix } from "@/lib/replay";
import type { TrajectoryObservation, TrajectoryType } from "@/lib/types";

type Props = {
  symbol: string;
  slotIndex: number;
  color: string;
  observations: TrajectoryObservation[];
  trajectory: TrajectoryType;
  cursor: number;
  onSelect: (obs: TrajectoryObservation) => void;
};

function labChartStyles(dark: boolean, color: string): DeepPartial<Styles> {
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
          { offset: 0, color: `${color}33` },
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
            const current = data.current as LabKLineData | null;
            if (!current) return [];
            const seq = current.generatorSequenceNo ?? current.timestamp;
            const rawY =
              current.trajectory === "volume" ? current.rawVolume : current.rawClose;
            return [
              { title: "seq ", value: String(seq) },
              {
                title: current.trajectory === "volume" ? "vol " : "close ",
                value: String(rawY),
              },
            ];
          },
        },
      },
      priceMark: {
        high: { show: false },
        low: { show: false },
        last: { show: true },
      },
    },
    xAxis: {
      axisLine: { color: dark ? "#42565c" : "#cbd3d6" },
      tickText: { color: dark ? "#91a4aa" : "#637075" },
    },
    yAxis: {
      axisLine: { color: dark ? "#42565c" : "#cbd3d6" },
      tickText: { color: dark ? "#91a4aa" : "#637075" },
    },
    separator: { color: dark ? "#42565c" : "#cbd3d6", activeBackgroundColor: dark ? "#60767d" : "#aab7ba" },
    crosshair: {
      horizontal: { line: { color: dark ? "#9eb0b5" : "#44545a" } },
      vertical: { line: { color: dark ? "#9eb0b5" : "#44545a" } },
    },
  };
}

export default function WaveLaneChart({
  symbol,
  slotIndex,
  color,
  observations,
  trajectory,
  cursor,
  onSelect,
}: Props) {
  const elementRef = useRef<HTMLDivElement>(null);
  const chartRef = useRef<Chart | null>(null);
  const observationsRef = useRef(observations);
  const trajectoryRef = useRef(trajectory);
  const cursorRef = useRef(cursor);
  const onSelectRef = useRef(onSelect);
  const { resolvedTheme } = useTheme();

  useEffect(() => {
    observationsRef.current = observations;
    trajectoryRef.current = trajectory;
    cursorRef.current = cursor;
    onSelectRef.current = onSelect;
  }, [observations, trajectory, cursor, onSelect]);

  useEffect(() => {
    const element = elementRef.current;
    if (!element || !symbol) return;
    const dark = resolvedTheme !== "light";
    const chart = init(element, {
      locale: "en-US",
      styles: labChartStyles(dark, color),
      formatter: {
        formatDate: ({ timestamp }) => formatBarIndexTick(timestamp),
      },
    });
    if (!chart) return;
    chartRef.current = chart;

    chart.setSymbol({
      ticker: symbol,
      pricePrecision: trajectoryRef.current === "volume" ? 0 : 4,
      volumePrecision: 0,
    });
    chart.setPeriod({ type: "minute", span: 1 });
    chart.setOffsetRightDistance(8);

    chart.setDataLoader({
      getBars({ type, callback }) {
        if (type !== "init" && type !== "update") {
          callback([], { backward: false, forward: false });
          return;
        }
        const prefix = visiblePrefix(observationsRef.current, cursorRef.current);
        callback(observationsToKLines(prefix, trajectoryRef.current), {
          backward: false,
          forward: false,
        });
      },
    });

    const lastCrosshairRef = { current: undefined as Crosshair | undefined };

    function selectObservationFromChart(data?: KLineData, dataIndex?: number) {
      const prefix = visiblePrefix(observationsRef.current, cursorRef.current);
      if (prefix.length === 0) return;
      const mapped =
        klineToObservation(data, null, prefix) ??
        observationAtChartTimestamp(prefix, data?.timestamp) ??
        (typeof dataIndex === "number" ? prefix[dataIndex] ?? null : null);
      if (mapped) onSelectRef.current(mapped);
    }

    const onCrosshair = (data?: unknown) => {
      lastCrosshairRef.current = data as Crosshair | undefined;
    };
    const onCandleClick = (data?: unknown) => {
      const payload = data as { kLineData?: KLineData; dataIndex?: number } | undefined;
      selectObservationFromChart(payload?.kLineData, payload?.dataIndex);
    };
    const onDomClick = (event: MouseEvent) => {
      const live = chartRef.current;
      const prefix = visiblePrefix(observationsRef.current, cursorRef.current);
      if (!live || prefix.length === 0) return;
      const rect = element.getBoundingClientRect();
      const converted = live.convertFromPixel([
        { x: event.clientX - rect.left, y: event.clientY - rect.top },
      ]);
      const point = Array.isArray(converted) ? converted[0] : converted;
      const dataList = live.getDataList();
      const byIndex =
        typeof point?.dataIndex === "number" ? dataList[point.dataIndex] ?? prefix[point.dataIndex] : undefined;
      const cross = lastCrosshairRef.current;
      selectObservationFromChart(
        byIndex ?? cross?.kLineData,
        point?.dataIndex ?? cross?.dataIndex,
      );
    };

    chart.subscribeAction("onCrosshairChange", onCrosshair);
    chart.subscribeAction("onCandleBarClick", onCandleClick);
    element.addEventListener("click", onDomClick);

    const resizeObserver = new ResizeObserver(() => chart.resize());
    resizeObserver.observe(element);

    return () => {
      resizeObserver.disconnect();
      element.removeEventListener("click", onDomClick);
      chart.unsubscribeAction("onCrosshairChange", onCrosshair);
      chart.unsubscribeAction("onCandleBarClick", onCandleClick);
      dispose(chart);
      chartRef.current = null;
    };
  }, [symbol, color, resolvedTheme, slotIndex, trajectory]);

  useEffect(() => {
    chartRef.current?.resetData();
  }, [cursor, observations, trajectory]);

  if (!symbol) {
    return <div className="chart-canvas empty-lane">Select a symbol for this presentation slot.</div>;
  }

  return <div ref={elementRef} className="chart-canvas" role="img" aria-label={`${symbol} bar-index wave`} />;
}
