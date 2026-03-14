import { useEffect, useState } from "react";
import { Link, useParams } from "react-router-dom";
import { JobStatusCard } from "../components/JobStatusCard";
import { ProductLineEditor, type ProductLineMode } from "../components/ProductLineEditor";
import { SlotCard } from "../components/SlotCard";
import { useJob } from "../hooks/useJob";
import { useJobLogs } from "../hooks/useJobLogs";
import { usePreview } from "../hooks/usePreview";
import { useSlots } from "../hooks/useSlots";
import { startAnalysis } from "../services/analysisApi";
import { getPreviewDownloadUrl, renderPreview } from "../services/previewApi";
import {
  generateSlot,
  manualImportSlot,
  manualSelectSlot,
  rejectSlot,
  repickSlots,
  selectSlot,
} from "../services/slotsApi";
import { ApiError } from "../types/Api";
import type { Job } from "../types/Job";
import type { Slot } from "../types/Slot";

function isTerminalJob(job: Job | null): boolean {
  if (!job) {
    return false;
  }
  return job.status === "completed" || job.status === "failed";
}

function isPhaseThreeSettled(slot: Slot | null): boolean {
  return slot?.status === "generated";
}

export function JobPage() {
  const { jobId } = useParams();
  const [actionError, setActionError] = useState<string | null>(null);
  const [actionMessage, setActionMessage] = useState<string | null>(null);
  const [startPending, setStartPending] = useState(false);
  const [selectingSlotId, setSelectingSlotId] = useState<string | null>(null);
  const [rejectingSlotId, setRejectingSlotId] = useState<string | null>(null);
  const [repickPending, setRepickPending] = useState(false);
  const [generatePending, setGeneratePending] = useState(false);
  const [renderPending, setRenderPending] = useState(false);
  const [manualStartSeconds, setManualStartSeconds] = useState("");
  const [manualEndSeconds, setManualEndSeconds] = useState("");
  const [manualSelectPending, setManualSelectPending] = useState(false);
  const [manualImportClipPath, setManualImportClipPath] = useState("");
  const [manualImportAudioPath, setManualImportAudioPath] = useState("");
  const [manualImportStartSeconds, setManualImportStartSeconds] = useState("");
  const [manualImportEndSeconds, setManualImportEndSeconds] = useState("");
  const [manualImportPending, setManualImportPending] = useState(false);
  const [jobPollingEnabled, setJobPollingEnabled] = useState(Boolean(jobId));

  const { job, error: jobError, loading: jobLoading, refresh: refreshJob } = useJob(jobId, {
    poll: jobPollingEnabled,
  });

  const shouldPollSlots = Boolean(
    jobId &&
    job &&
    (
      (job.status === "analyzing" &&
        (job.current_stage === "analysis_submission" || job.current_stage === "analysis_poll")) ||
      job.status === "generating"
    ),
  );
  const { slots, error: slotsError, loading: slotsLoading, refresh: refreshSlots } = useSlots(jobId, {
    poll: shouldPollSlots,
  });
  const { logs, error: logsError, loading: logsLoading, refresh: refreshLogs } = useJobLogs(jobId, {
    poll: jobPollingEnabled,
  });
  const { preview, error: previewError, loading: previewLoading, refresh: refreshPreview } = usePreview(jobId, {
    poll: job?.status === "stitching",
  });

  const selectedSlot =
    slots.find((slot) => slot.id === job?.selected_slot_id) ??
    slots.find((slot) => ["selected", "generating", "generated", "failed"].includes(slot.status)) ??
    null;

  useEffect(() => {
    setJobPollingEnabled(Boolean(jobId) && !isTerminalJob(job) && !isPhaseThreeSettled(selectedSlot));
  }, [job, jobId, selectedSlot]);

  const allSlotsRejected =
    job?.current_stage === "slot_selection" && slots.length > 0 && slots.every((slot) => slot.status === "rejected");
  const canStartAnalysis = Boolean(job && job.status === "queued" && job.current_stage === "ready_for_analysis");
  const canSelectSlots = Boolean(job && job.status === "analyzing" && job.current_stage === "slot_selection");
  const noAutoSlotFound = job?.error_code === "NO_SUITABLE_SLOT_FOUND";
  const detectedContentLanguage =
    typeof job?.metadata?.content_language === "string" ? job.metadata.content_language : "en";

  async function refreshAll() {
    refreshJob();
    refreshSlots();
    refreshLogs();
    refreshPreview();
  }

  async function handleStartAnalysis() {
    if (!jobId) {
      return;
    }

    setActionError(null);
    setActionMessage(null);
    setStartPending(true);

    try {
      const response = await startAnalysis(jobId);
      setActionMessage(response.message);
      await refreshAll();
    } catch (reason: unknown) {
      if (reason instanceof ApiError) {
        setActionError(reason.message);
      } else {
        setActionError("Unable to start analysis.");
      }
    } finally {
      setStartPending(false);
    }
  }

  async function handleSelect(slotId: string) {
    if (!jobId) {
      return;
    }

    setActionError(null);
    setActionMessage(null);
    setSelectingSlotId(slotId);

    try {
      const response = await selectSlot(jobId, slotId);
      setActionMessage(String(response.message ?? "slot selected and product line prepared"));
      await refreshAll();
    } catch (reason: unknown) {
      if (reason instanceof ApiError) {
        setActionError(reason.message);
      } else {
        setActionError("Unable to select slot.");
      }
    } finally {
      setSelectingSlotId(null);
    }
  }

  async function handleReject(slotId: string) {
    if (!jobId) {
      return;
    }

    setActionError(null);
    setActionMessage(null);
    setRejectingSlotId(slotId);

    try {
      const response = await rejectSlot(jobId, slotId, "Rejected in dashboard review");
      setActionMessage(String(response.message ?? "slot rejected"));
      await refreshAll();
    } catch (reason: unknown) {
      if (reason instanceof ApiError) {
        setActionError(reason.message);
      } else {
        setActionError("Unable to reject slot.");
      }
    } finally {
      setRejectingSlotId(null);
    }
  }

  async function handleRepick() {
    if (!jobId) {
      return;
    }

    setActionError(null);
    setActionMessage(null);
    setRepickPending(true);

    try {
      const response = await repickSlots(jobId);
      setActionMessage(String(response.message ?? "re-pick requested"));
      await refreshAll();
    } catch (reason: unknown) {
      if (reason instanceof ApiError) {
        setActionError(reason.message);
      } else {
        setActionError("Unable to request re-pick.");
      }
    } finally {
      setRepickPending(false);
    }
  }

  async function handleGenerate(mode: ProductLineMode, customProductLine: string) {
    if (!jobId || !selectedSlot) {
      return;
    }

    setActionError(null);
    setActionMessage(null);
    setGeneratePending(true);

    try {
      const response = await generateSlot(jobId, selectedSlot.id, {
        product_line_mode: mode,
        custom_product_line: customProductLine,
      });
      setActionMessage(String(response.message ?? "cafai generation started"));
      await refreshAll();
    } catch (reason: unknown) {
      if (reason instanceof ApiError) {
        setActionError(reason.message);
      } else {
        setActionError("Unable to start CAFAI generation.");
      }
    } finally {
      setGeneratePending(false);
    }
  }

  async function handleRenderPreview() {
    if (!jobId || !selectedSlot) {
      return;
    }

    setActionError(null);
    setActionMessage(null);
    setRenderPending(true);

    try {
      const response = await renderPreview(jobId, selectedSlot.id);
      setActionMessage(String(response.message ?? "preview render started"));
      await refreshAll();
    } catch (reason: unknown) {
      if (reason instanceof ApiError) {
        setActionError(reason.message);
      } else {
        setActionError("Unable to start preview render.");
      }
    } finally {
      setRenderPending(false);
    }
  }

  async function handleManualSelect() {
    if (!jobId) {
      return;
    }

    const startSeconds = Number(manualStartSeconds);
    const endSeconds = Number(manualEndSeconds);
    if (!Number.isFinite(startSeconds) || !Number.isFinite(endSeconds)) {
      setActionError("Manual slot selection requires numeric start and end times.");
      return;
    }

    setActionError(null);
    setActionMessage(null);
    setManualSelectPending(true);

    try {
      const response = await manualSelectSlot(jobId, {
        start_seconds: startSeconds,
        end_seconds: endSeconds,
      });
      setActionMessage(String(response.message ?? "manual slot selected and product line prepared"));
      setManualStartSeconds("");
      setManualEndSeconds("");
      await refreshAll();
    } catch (reason: unknown) {
      if (reason instanceof ApiError) {
        setActionError(reason.message);
      } else {
        setActionError("Unable to select a manual slot.");
      }
    } finally {
      setManualSelectPending(false);
    }
  }

  async function handleManualImport() {
    if (!jobId) {
      return;
    }

    const clipPath = manualImportClipPath.trim();
    if (clipPath === "") {
      setActionError("Manual generation import requires the generated clip path.");
      return;
    }

    const startSeconds = manualImportStartSeconds.trim();
    const endSeconds = manualImportEndSeconds.trim();
    const shouldSendTimes = !selectedSlot;
    const parsedStartSeconds = startSeconds === "" ? Number.NaN : Number(startSeconds);
    const parsedEndSeconds = endSeconds === "" ? Number.NaN : Number(endSeconds);

    if (shouldSendTimes && (!Number.isFinite(parsedStartSeconds) || !Number.isFinite(parsedEndSeconds))) {
      setActionError("Manual generation import requires numeric start and end times when no slot is selected.");
      return;
    }

    setActionError(null);
    setActionMessage(null);
    setManualImportPending(true);

    try {
      const response = await manualImportSlot(jobId, {
        slot_id: selectedSlot?.id,
        start_seconds: shouldSendTimes ? parsedStartSeconds : undefined,
        end_seconds: shouldSendTimes ? parsedEndSeconds : undefined,
        generated_clip_path: clipPath,
        generated_audio_path: manualImportAudioPath.trim() || undefined,
      });
      setActionMessage(String(response.message ?? "manual generated clip imported"));
      setManualImportAudioPath("");
      if (!selectedSlot) {
        setManualImportStartSeconds("");
        setManualImportEndSeconds("");
      }
      await refreshAll();
    } catch (reason: unknown) {
      if (reason instanceof ApiError) {
        setActionError(reason.message);
      } else {
        setActionError("Unable to import the generated clip.");
      }
    } finally {
      setManualImportPending(false);
    }
  }

  const canRenderPreview = Boolean(
    selectedSlot &&
    selectedSlot.status === "generated" &&
    preview?.status !== "pending" &&
    preview?.status !== "stitching" &&
    !renderPending,
  );
  const previewDownloadUrl =
    jobId && preview?.status === "completed" && preview.output_video_path
      ? getPreviewDownloadUrl(jobId)
      : undefined;

  return (
    <div className="page-grid">
      <section className="panel">
        <p className="eyebrow">Job workflow</p>
        <h2>Job dashboard {jobId ?? "demo-job"}</h2>
        <p>
          Start analysis explicitly, review the ranked insertion slots, prepare one product line, and start CAFAI
          generation for the selected candidate.
        </p>
        <p className="muted">
          Demo runs can bypass live generation by importing a locally generated MP4 into the selected slot and then
          continuing through preview render.
        </p>
        <div className="status-strip">
          <span>{jobLoading ? "Loading job..." : jobError ?? job?.status ?? "Job unavailable"}</span>
          <span>{slotsLoading ? "Loading slots..." : slotsError ?? `${slots.length} slot(s)`}</span>
          <span>{logsLoading ? "Loading logs..." : logsError ?? `${logs.length} log entry(ies)`}</span>
        </div>
        <p>Detected content language: {detectedContentLanguage.toUpperCase()}</p>
        {actionError ? <p className="form-message form-message--error">{actionError}</p> : null}
        {actionMessage ? <p className="form-message form-message--success">{actionMessage}</p> : null}
      </section>

      <div className="card-grid">
        <JobStatusCard
          job={job}
          loading={jobLoading}
          error={jobError}
          onStartAnalysis={handleStartAnalysis}
          startDisabled={!canStartAnalysis}
          startPending={startPending}
        />

        <section className="card">
          <div className="list-block__header">
            <div>
              <p className="eyebrow">Logs</p>
              <h3>Operational timeline</h3>
            </div>
          </div>
          {logsError ? <p className="muted">{logsError}</p> : null}
          {logs.length === 0 ? <p>No job logs yet.</p> : null}
          <div className="log-list">
            {logs.map((log) => (
              <div key={`${log.timestamp}-${log.message}`} className="log-item">
                <strong>{log.event_type}</strong>
                <span>{log.stage_name ?? "n/a"}</span>
                <p>{log.message}</p>
              </div>
            ))}
          </div>
        </section>

        <section className="card">
          <div className="list-block__header">
            <div>
              <p className="eyebrow">Slot actions</p>
              <h3>Re-pick control</h3>
            </div>
            <button className="button-secondary" type="button" onClick={handleRepick} disabled={!allSlotsRejected || repickPending}>
              {repickPending ? "Requesting..." : "Re-pick slots"}
            </button>
          </div>
          <p className="muted">Re-pick is only available after all currently proposed slots have been rejected.</p>
          <p>Current gate: {allSlotsRejected ? "all slots rejected" : "waiting for more rejections"}</p>
        </section>

        <section className="card">
          <div className="list-block__header">
            <div>
              <p className="eyebrow">Preview render</p>
              <h3>Phase 4</h3>
            </div>
            <button className="button-secondary" type="button" onClick={handleRenderPreview} disabled={!canRenderPreview}>
              {renderPending ? "Starting..." : preview?.status === "failed" ? "Retry render" : "Render preview"}
            </button>
          </div>
          <p className="muted">
            {previewLoading
              ? "Loading preview state..."
              : previewError ?? previewSummary(preview?.status, selectedSlot?.status)}
          </p>
          {preview?.error_message ? <p className="form-message form-message--error">{preview.error_message}</p> : null}
          {preview ? <p>Preview status: {preview.status}</p> : null}
          {preview ? <p>Retry count: {preview.render_retry_count ?? 0}</p> : null}
          <div className="form-actions">
            {preview ? <Link to={`/jobs/${jobId}/preview`}>Open preview</Link> : null}
            {previewDownloadUrl ? (
              <a href={previewDownloadUrl} download>
                Download preview
              </a>
            ) : null}
          </div>
        </section>
      </div>

      <section className="panel">
        <div className="list-block__header">
          <div>
            <p className="eyebrow">Slot review</p>
            <h2>Current candidates</h2>
          </div>
        </div>
        {slotsError ? <p className="form-message form-message--error">{slotsError}</p> : null}
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
                  <input
                    type="number"
                    min="0"
                    step="0.1"
                    value={manualStartSeconds}
                    onChange={(event) => setManualStartSeconds(event.target.value)}
                  />
                </label>
                <label className="field">
                  <span>End seconds</span>
                  <input
                    type="number"
                    min="0"
                    step="0.1"
                    value={manualEndSeconds}
                    onChange={(event) => setManualEndSeconds(event.target.value)}
                  />
                </label>
              </div>
              <div className="form-actions">
                <button type="button" onClick={handleManualSelect} disabled={manualSelectPending}>
                  {manualSelectPending ? "Selecting..." : "Select manual slot"}
                </button>
              </div>
            </div>
          </section>
        ) : null}
        {canSelectSlots || selectedSlot ? (
          <>
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
                    onChange={(event) => setManualImportClipPath(event.target.value)}
                  />
                </label>
                <label className="field">
                  <span>Generated audio path (optional)</span>
                  <input
                    type="text"
                    placeholder="/absolute/path/to/generated.wav"
                    value={manualImportAudioPath}
                    onChange={(event) => setManualImportAudioPath(event.target.value)}
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
                        onChange={(event) => setManualImportStartSeconds(event.target.value)}
                      />
                    </label>
                    <label className="field">
                      <span>Import end seconds</span>
                      <input
                        type="number"
                        min="0"
                        step="0.1"
                        value={manualImportEndSeconds}
                        onChange={(event) => setManualImportEndSeconds(event.target.value)}
                      />
                    </label>
                  </div>
                ) : null}
                <div className="form-actions">
                  <button type="button" onClick={handleManualImport} disabled={manualImportPending}>
                    {manualImportPending ? "Importing..." : "Import generated clip"}
                  </button>
                </div>
              </div>
            </section>
          </>
        ) : null}
        {!slotsLoading && slots.length === 0 ? <p>No slots available yet.</p> : null}
        <div className="card-grid">
          {slots.map((slot) => (
            <SlotCard
              key={slot.id}
              slot={slot}
              onSelect={handleSelect}
              onReject={handleReject}
              selectDisabled={!canSelectSlots || (selectingSlotId !== null && selectingSlotId !== slot.id)}
              rejectDisabled={!canSelectSlots || (rejectingSlotId !== null && rejectingSlotId !== slot.id)}
            />
          ))}
        </div>
      </section>

      {selectedSlot ? (
        <section className="panel">
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
      ) : null}

      {selectedSlot ? <ProductLineEditor slot={selectedSlot} pending={generatePending} onGenerate={handleGenerate} /> : null}
    </div>
  );
}

function previewSummary(previewStatus?: string, slotStatus?: string): string {
  if (previewStatus === "pending") {
    return "Preview render is queued.";
  }
  if (previewStatus === "stitching") {
    return "Preview render is in progress.";
  }
  if (previewStatus === "completed") {
    return "Preview render completed.";
  }
  if (previewStatus === "failed") {
    return "Preview render failed. Retry is available.";
  }
  if (slotStatus === "generated") {
    return "Generated slot is ready for preview rendering.";
  }
  return "Preview rendering becomes available after CAFAI generation succeeds.";
}
