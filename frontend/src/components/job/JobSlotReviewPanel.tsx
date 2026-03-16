import type { ChangeEvent } from "react";
import { SlotCard } from "../SlotCard";
import type { Slot } from "../../types/Slot";

interface JobSlotReviewPanelProps {
  slots: Slot[];
  slotsLoading: boolean;
  slotsError: string | null;
  canSelectSlots: boolean;
  selectedSlot: Slot | null;
  noAutoSlotFound: boolean;
  selectingSlotId: string | null;
  rejectingSlotId: string | null;
  manualStartSeconds: string;
  manualEndSeconds: string;
  manualSelectPending: boolean;
  manualImportClipPath: string;
  manualImportAudioPath: string;
  manualImportStartSeconds: string;
  manualImportEndSeconds: string;
  manualImportPending: boolean;
  onSelect: (slotId: string) => void;
  onReject: (slotId: string) => void;
  onManualSelect: () => void;
  onManualImport: () => void;
  onManualStartSecondsChange: (event: ChangeEvent<HTMLInputElement>) => void;
  onManualEndSecondsChange: (event: ChangeEvent<HTMLInputElement>) => void;
  onManualImportClipPathChange: (event: ChangeEvent<HTMLInputElement>) => void;
  onManualImportAudioPathChange: (event: ChangeEvent<HTMLInputElement>) => void;
  onManualImportStartSecondsChange: (event: ChangeEvent<HTMLInputElement>) => void;
  onManualImportEndSecondsChange: (event: ChangeEvent<HTMLInputElement>) => void;
}

export function JobSlotReviewPanel({
  slots,
  slotsLoading,
  slotsError,
  canSelectSlots,
  selectedSlot,
  noAutoSlotFound,
  selectingSlotId,
  rejectingSlotId,
  manualStartSeconds,
  manualEndSeconds,
  manualSelectPending,
  manualImportClipPath,
  manualImportAudioPath,
  manualImportStartSeconds,
  manualImportEndSeconds,
  manualImportPending,
  onSelect,
  onReject,
  onManualSelect,
  onManualImport,
  onManualStartSecondsChange,
  onManualEndSecondsChange,
  onManualImportClipPathChange,
  onManualImportAudioPathChange,
  onManualImportStartSecondsChange,
  onManualImportEndSecondsChange,
}: JobSlotReviewPanelProps) {
  return (
    <section className="panel panel--studio">
      <div className="list-block__header">
        <div>
          <p className="eyebrow">Slot review</p>
          <h2>Current candidates</h2>
        </div>
      </div>
      {slotsError ? <p className="form-message form-message--error">{slotsError}</p> : null}
      {canSelectSlots || selectedSlot ? (
        <details className="studio-disclosure" open={noAutoSlotFound}>
          <summary>Advanced recovery and manual overrides</summary>
          <div className="studio-disclosure__content">
            {canSelectSlots ? (
              <section className="card">
                <div className="list-block__header">
                  <div>
                    <p className="eyebrow">Manual override</p>
                    <h3>Pick a slot by time</h3>
                  </div>
                </div>
                <p className="muted">
                  {noAutoSlotFound
                    ? "Automatic ranking found no suitable slot. Manual selection is the primary recovery path."
                    : "Use manual selection when you want to override the automatic slot proposals."}
                </p>
                <div className="form-grid">
                  <div className="field-row">
                    <label className="field">
                      <span>Start seconds</span>
                      <input type="number" min="0" step="0.1" value={manualStartSeconds} onChange={onManualStartSecondsChange} />
                    </label>
                    <label className="field">
                      <span>End seconds</span>
                      <input type="number" min="0" step="0.1" value={manualEndSeconds} onChange={onManualEndSecondsChange} />
                    </label>
                  </div>
                  <div className="form-actions">
                    <button type="button" onClick={onManualSelect} disabled={manualSelectPending}>
                      {manualSelectPending ? "Selecting..." : "Select manual slot"}
                    </button>
                  </div>
                </div>
              </section>
            ) : null}

            <section className="card">
              <div className="list-block__header">
                <div>
                  <p className="eyebrow">Manual generation import</p>
                  <h3>Use a locally generated clip</h3>
                </div>
              </div>
              <p className="muted">
                {selectedSlot
                  ? "Import the generated bridge clip into the currently selected slot, then render the preview normally."
                  : "If no slot is selected yet, provide the generated clip plus anchor times inside one analyzed scene."}
              </p>
              <div className="form-grid">
                <label className="field">
                  <span>Generated clip path</span>
                  <input
                    type="text"
                    placeholder="/absolute/path/to/generated.mp4"
                    value={manualImportClipPath}
                    onChange={onManualImportClipPathChange}
                  />
                </label>
                <label className="field">
                  <span>Generated audio path (optional)</span>
                  <input
                    type="text"
                    placeholder="/absolute/path/to/generated.wav"
                    value={manualImportAudioPath}
                    onChange={onManualImportAudioPathChange}
                  />
                </label>
                {!selectedSlot ? (
                  <div className="field-row">
                    <label className="field">
                      <span>Import start seconds</span>
                      <input
                        type="number"
                        min="0"
                        step="0.1"
                        value={manualImportStartSeconds}
                        onChange={onManualImportStartSecondsChange}
                      />
                    </label>
                    <label className="field">
                      <span>Import end seconds</span>
                      <input
                        type="number"
                        min="0"
                        step="0.1"
                        value={manualImportEndSeconds}
                        onChange={onManualImportEndSecondsChange}
                      />
                    </label>
                  </div>
                ) : null}
                <div className="form-actions">
                  <button type="button" onClick={onManualImport} disabled={manualImportPending}>
                    {manualImportPending ? "Importing..." : "Import generated clip"}
                  </button>
                </div>
              </div>
            </section>
          </div>
        </details>
      ) : null}
      {!slotsLoading && slots.length === 0 ? <p>No slots available yet.</p> : null}
      <div className="card-grid">
        {slots.map((slot) => (
          <SlotCard
            key={slot.id}
            slot={slot}
            onSelect={onSelect}
            onReject={onReject}
            selectDisabled={!canSelectSlots || (selectingSlotId !== null && selectingSlotId !== slot.id)}
            rejectDisabled={!canSelectSlots || (rejectingSlotId !== null && rejectingSlotId !== slot.id)}
          />
        ))}
      </div>
    </section>
  );
}
