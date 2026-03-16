import type { Slot } from "../../types/Slot";

interface JobSelectedSlotSummaryProps {
  selectedSlot: Slot;
}

export function JobSelectedSlotSummary({ selectedSlot }: JobSelectedSlotSummaryProps) {
  return (
    <section className="panel panel--studio">
      <div className="list-block__header">
        <div>
          <p className="eyebrow">Selected slot</p>
          <h2>Generation handoff</h2>
        </div>
      </div>
      <p>Slot ID: {selectedSlot.id}</p>
      <p>
        Anchor frames: {selectedSlot.anchor_start_frame} to {selectedSlot.anchor_end_frame}
      </p>
      {selectedSlot.generated_clip_path ? <p>Generated clip: {selectedSlot.generated_clip_path}</p> : null}
      {selectedSlot.generated_audio_path ? <p>Generated audio: {selectedSlot.generated_audio_path}</p> : null}
    </section>
  );
}
