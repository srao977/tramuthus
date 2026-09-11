"use client";

import dynamic from "next/dynamic";
import type { PhasePoint, PhaseSeries, PhaseSymbol, RawObservation } from "@/lib/types";
import RawObservationDetails from "./RawObservationDetails";

const PhaseChart = dynamic(() => import("./PhaseChart"), {
  ssr: false,
  loading: () => <div className="chart-empty">Loading chart surface...</div>,
});

type Props = {
  partitionId: string;
  color: string;
  symbols: PhaseSymbol[];
  selectedSymbol: string;
  series: PhaseSeries | undefined;
  loading: boolean;
  selectedPoint: PhasePoint | null;
  rawObservation: RawObservation | null;
  rawLoading: boolean;
  onSymbolChange: (symbol: string) => void;
  onSelectPoint: (point: PhasePoint) => void;
};

export default function PartitionPhasePanel({
  partitionId,
  color,
  symbols,
  selectedSymbol,
  series,
  loading,
  selectedPoint,
  rawObservation,
  rawLoading,
  onSymbolChange,
  onSelectPoint,
}: Props) {
  const descriptor = symbols.find((item) => item.symbol === selectedSymbol);
  return (
    <section className="partition-panel" aria-labelledby={`partition-${partitionId}`}>
      <header className="partition-header">
        <div className="partition-title">
          <i style={{ background: color }} />
          <div>
            <span>Partition {partitionId}</span>
            <strong id={`partition-${partitionId}`}>{selectedSymbol || "No symbol"}</strong>
          </div>
        </div>
        <label>
          Symbol
          <select value={selectedSymbol} onChange={(event) => onSymbolChange(event.target.value)}>
            {symbols.map((item) => (
              <option value={item.symbol} key={item.symbol}>
                {item.symbol} · {item.observableCount} observable
              </option>
            ))}
          </select>
        </label>
      </header>

      <div className="series-stats">
        <span><b>{descriptor?.recordCount ?? 0}</b> selected</span>
        <span><b>{descriptor?.observableCount ?? 0}</b> observable</span>
        <span><b>{descriptor?.initializingCount ?? 0}</b> initializing</span>
        <span><b>{descriptor?.maxSequence ?? 0}</b> max sequence</span>
      </div>

      <div className="chart-frame">
        <div className="axis-caption axis-y">phase_angle_degrees</div>
        {loading ? (
          <div className="chart-empty">Loading {selectedSymbol} phase evidence...</div>
        ) : series && series.points.length > 0 ? (
          <PhaseChart symbol={selectedSymbol} color={color} points={series.points} onSelect={onSelectPoint} />
        ) : (
          <div className="chart-empty">
            <strong>No observable phase values for this experiment.</strong>
            <span>Initialization records remain null and are not plotted.</span>
          </div>
        )}
        <div className="axis-caption axis-x">generator_sequence_no</div>
      </div>
      <p className="panel-note">Numeric observable records only · stored angle convention 0° ≤ phase &lt; 360°</p>
      <section className="partition-observation" aria-labelledby={`observation-${partitionId}`}>
        <header>
          <div>
            <span>Phase + raw lineage</span>
            <strong id={`observation-${partitionId}`}>
              Observation Details{selectedPoint ? ` · ${selectedPoint.symbol} sequence ${selectedPoint.generatorSequenceNo}` : ""}
            </strong>
          </div>
          <small>Partition {partitionId}</small>
        </header>
        <div className="partition-observation-scroll" tabIndex={0}>
          <RawObservationDetails phase={selectedPoint} raw={rawObservation} loading={rawLoading} />
        </div>
      </section>
    </section>
  );
}
