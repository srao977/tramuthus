"use client";

import { useEffect, useMemo, useRef, useState } from "react";
import type { TrajectoryObservation, TrajectoryType } from "@/lib/types";

type Props = {
  observations: TrajectoryObservation[];
  trajectory: TrajectoryType;
  selectedSequence: number | null;
  onSelect: (obs: TrajectoryObservation) => void;
};

function formatY(value: number, trajectory: TrajectoryType) {
  if (trajectory === "volume") return value.toLocaleString("en-US", { maximumFractionDigits: 0 });
  return value.toLocaleString("en-US", { minimumFractionDigits: 2, maximumFractionDigits: 4 });
}

export default function WaveChart({ observations, trajectory, selectedSequence, onSelect }: Props) {
  const wrapRef = useRef<HTMLDivElement>(null);
  const [width, setWidth] = useState(640);
  const height = 360;
  const pad = { top: 18, right: 16, bottom: 36, left: 64 };

  useEffect(() => {
    const el = wrapRef.current;
    if (!el) return;
    const ro = new ResizeObserver((entries) => {
      const next = Math.floor(entries[0]?.contentRect.width ?? 640);
      setWidth(Math.max(280, next));
    });
    ro.observe(el);
    return () => ro.disconnect();
  }, []);

  const layout = useMemo(() => {
    if (observations.length === 0) return null;
    const xs = observations.map((o) => o.generatorSequenceNo);
    const ys = observations.map((o) => o.y);
    const minX = Math.min(...xs);
    const maxX = Math.max(...xs);
    const minY = Math.min(...ys);
    const maxY = Math.max(...ys);
    const xSpan = Math.max(1, maxX - minX);
    const yPad = maxY === minY ? Math.max(1, Math.abs(maxY) * 0.05 || 1) : (maxY - minY) * 0.08;
    const y0 = minY - yPad;
    const y1 = maxY + yPad;
    const innerW = Math.max(1, width - pad.left - pad.right);
    const innerH = Math.max(1, height - pad.top - pad.bottom);
    const xOf = (seq: number) => pad.left + ((seq - minX) / xSpan) * innerW;
    const yOf = (y: number) => pad.top + ((y1 - y) / (y1 - y0)) * innerH;
    const path = observations
      .map((o, i) => `${i === 0 ? "M" : "L"} ${xOf(o.generatorSequenceNo).toFixed(2)} ${yOf(o.y).toFixed(2)}`)
      .join(" ");
    const ticksX = 5;
    const ticksY = 4;
    const xTicks = Array.from({ length: ticksX }, (_, i) => minX + (xSpan * i) / (ticksX - 1));
    const yTicks = Array.from({ length: ticksY }, (_, i) => y0 + ((y1 - y0) * i) / (ticksY - 1));
    return { path, xOf, yOf, xTicks, yTicks, minX, maxX };
  }, [observations, width, height, pad.left, pad.right, pad.top, pad.bottom]);

  function nearest(clientX: number) {
    const svg = wrapRef.current?.querySelector("svg");
    if (!svg || !layout || observations.length === 0) return;
    const rect = svg.getBoundingClientRect();
    const px = ((clientX - rect.left) / rect.width) * width;
    let best = observations[0];
    let bestD = Infinity;
    for (const o of observations) {
      const d = Math.abs(layout.xOf(o.generatorSequenceNo) - px);
      if (d < bestD) {
        bestD = d;
        best = o;
      }
    }
    onSelect(best);
  }

  return (
    <div ref={wrapRef} className="wave-chart">
      {observations.length === 0 || !layout ? (
        <p className="empty">No observations for this symbol.</p>
      ) : (
        <svg
          role="img"
          aria-label={`${trajectory} wave by generator_sequence_no`}
          viewBox={`0 0 ${width} ${height}`}
          width="100%"
          height={height}
          onClick={(e) => nearest(e.clientX)}
          onKeyDown={(e) => {
            if (e.key !== "Enter" && e.key !== " ") return;
            const current = observations.find((o) => o.generatorSequenceNo === selectedSequence) ?? observations[0];
            onSelect(current);
          }}
          tabIndex={0}
        >
          <rect x="0" y="0" width={width} height={height} className="plot-bg" />
          {layout.yTicks.map((y) => (
            <g key={`y-${y}`}>
              <line
                x1={pad.left}
                x2={width - pad.right}
                y1={layout.yOf(y)}
                y2={layout.yOf(y)}
                className="grid"
              />
              <text x={pad.left - 8} y={layout.yOf(y) + 4} textAnchor="end" className="tick">
                {formatY(y, trajectory)}
              </text>
            </g>
          ))}
          {layout.xTicks.map((x) => (
            <g key={`x-${x}`}>
              <line
                y1={pad.top}
                y2={height - pad.bottom}
                x1={layout.xOf(x)}
                x2={layout.xOf(x)}
                className="grid"
              />
              <text x={layout.xOf(x)} y={height - 12} textAnchor="middle" className="tick">
                {Math.round(x)}
              </text>
            </g>
          ))}
          <text x={width / 2} y={height - 2} textAnchor="middle" className="axis-label">
            generator_sequence_no
          </text>
          <text
            x={16}
            y={20}
            className="axis-label"
            transform={`rotate(-90 16 20)`}
            textAnchor="end"
          >
            {trajectory === "volume" ? "volume" : "close"}
          </text>
          <path d={layout.path} className="wave-line" fill="none" />
          {observations.map((o) => {
            const selected = o.generatorSequenceNo === selectedSequence;
            return (
              <circle
                key={o.generatorSequenceNo}
                cx={layout.xOf(o.generatorSequenceNo)}
                cy={layout.yOf(o.y)}
                r={selected ? 5 : 3}
                className={selected ? "pt selected" : "pt"}
                onClick={(e) => {
                  e.stopPropagation();
                  onSelect(o);
                }}
              />
            );
          })}
        </svg>
      )}
    </div>
  );
}
