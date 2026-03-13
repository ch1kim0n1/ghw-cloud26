import type { Slot } from "../types/Slot";

interface SlotCardProps {
  slot: Slot;
  onSelect?: (slotId: string) => void;
  onReject: (slotId: string) => void;
  selectDisabled: boolean;
  rejectDisabled: boolean;
}

export function SlotCard({ slot, onSelect, onReject, selectDisabled, rejectDisabled }: SlotCardProps) {
  const canSelect = slot.status === "proposed" && onSelect !== undefined;
  const canReject = slot.status === "proposed";

  return (
    <article className="card slot-card">
      <div className="list-block__header">
        <div>
          <p className="eyebrow">Rank {slot.rank}</p>
          <h3>{slot.status}</h3>
        </div>
        <div className="form-actions">
          <button type="button" onClick={() => onSelect?.(slot.id)} disabled={!canSelect || selectDisabled}>
            Select
          </button>
          <button
            className="button-secondary"
            type="button"
            onClick={() => onReject(slot.id)}
            disabled={!canReject || rejectDisabled}
          >
            Reject
          </button>
        </div>
      </div>

      <p>{slot.reasoning}</p>

      <dl className="job-metadata">
        <div>
          <dt>Anchor frames</dt>
          <dd>
            {slot.anchor_start_frame}
            {" -> "}
            {slot.anchor_end_frame}
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
        {slot.product_line_mode ? (
          <div>
            <dt>Line mode</dt>
            <dd>{slot.product_line_mode}</dd>
          </div>
        ) : null}
      </dl>

      {slot.suggested_product_line ? <p><strong>Suggested line:</strong> {slot.suggested_product_line}</p> : null}
      {slot.final_product_line ? <p><strong>Final line:</strong> {slot.final_product_line}</p> : null}
      {slot.generated_clip_path ? <p><strong>Generated clip:</strong> {slot.generated_clip_path}</p> : null}
      {slot.generated_audio_path ? <p><strong>Generated audio:</strong> {slot.generated_audio_path}</p> : null}
      {slot.generation_error ? <p className="form-message form-message--error">{slot.generation_error}</p> : null}
    </article>
  );
}
