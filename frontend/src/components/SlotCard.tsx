import type { Slot } from "../types/Slot";

interface SlotCardProps {
  slot: Slot;
  onReject: (slotId: string) => void;
  rejectDisabled: boolean;
}

export function SlotCard({ slot, onReject, rejectDisabled }: SlotCardProps) {
  return (
    <article className="card slot-card">
      <div className="list-block__header">
        <div>
          <p className="eyebrow">Rank {slot.rank}</p>
          <h3>{slot.status}</h3>
        </div>
        <button
          className="button-secondary"
          type="button"
          onClick={() => onReject(slot.id)}
          disabled={rejectDisabled || slot.status !== "proposed"}
        >
          Reject
        </button>
      </div>

      <p>{slot.reasoning}</p>

      <dl className="job-metadata">
        <div>
          <dt>Anchor frames</dt>
          <dd>
            {slot.anchor_start_frame} → {slot.anchor_end_frame}
          </dd>
        </div>
        <div>
          <dt>Score</dt>
          <dd>{slot.score.toFixed(3)}</dd>
        </div>
        <div>
          <dt>Quiet window</dt>
          <dd>{slot.quiet_window_seconds.toFixed(1)}s</dd>
        </div>
        <div>
          <dt>Scene</dt>
          <dd>{slot.scene_id}</dd>
        </div>
      </dl>
    </article>
  );
}
