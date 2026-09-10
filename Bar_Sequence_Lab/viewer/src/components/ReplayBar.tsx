"use client";

import { Pause, Play, RotateCcw } from "lucide-react";
import { REPLAY_SPEEDS } from "@/lib/replay";

type Props = {
  playing: boolean;
  speed: number;
  cursor: number;
  maxSequence: number;
  onPlayPause: () => void;
  onRestart: () => void;
  onSpeed: (speed: number) => void;
  onSeek: (cursor: number) => void;
};

export default function ReplayBar({
  playing,
  speed,
  cursor,
  maxSequence,
  onPlayPause,
  onRestart,
  onSpeed,
  onSeek,
}: Props) {
  const pct = maxSequence > 0 ? Math.round((cursor / maxSequence) * 100) : 0;
  return (
    <section className="replay-bar" aria-label="Bar-index replay">
      <div className="replay-actions">
        <button type="button" className="refresh-button" onClick={onRestart}>
          <RotateCcw size={15} /> Restart
        </button>
        <button type="button" className="refresh-button" onClick={onPlayPause}>
          {playing ? <Pause size={15} /> : <Play size={15} />}
          {playing ? "Pause" : "Play"}
        </button>
        <div className="seg speeds" role="group" aria-label="Replay speed">
          {REPLAY_SPEEDS.map((value) => (
            <button
              key={value}
              type="button"
              className={speed === value ? "is-on" : ""}
              onClick={() => onSpeed(value)}
            >
              {value}x
            </button>
          ))}
        </div>
      </div>
      <label className="replay-slider">
        <span>
          Bar index {cursor} / {maxSequence} · {pct}%
        </span>
        <input
          type="range"
          min={0}
          max={Math.max(0, maxSequence)}
          value={cursor}
          disabled={maxSequence <= 0}
          onChange={(event) => onSeek(Number(event.target.value))}
        />
      </label>
      <p className="replay-note">
        BAR-INDEX REPLAY · each wave advances on its own generator_sequence_no · not market-time
        sync
      </p>
    </section>
  );
}
