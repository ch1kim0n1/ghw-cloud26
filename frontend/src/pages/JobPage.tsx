import { useEffect, useState } from "react";
import { useParams } from "react-router-dom";
import { JobOperationsGrid } from "../components/job/JobOperationsGrid";
import { JobSelectedSlotSummary } from "../components/job/JobSelectedSlotSummary";
import { JobSlotReviewPanel } from "../components/job/JobSlotReviewPanel";
import { JobWorkflowHero } from "../components/job/JobWorkflowHero";
import { ProductLineEditor, type ProductLineMode } from "../components/ProductLineEditor";
import { useHealth } from "../hooks/useHealth";
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
  const notionDashboardUrl = import.meta.env.VITE_NOTION_DASHBOARD_URL?.trim();
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
  const { health, error: healthError, loading: healthLoading, refresh: refreshHealth } = useHealth({ poll: true });

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
    refreshHealth();
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
    <div className="studio-page studio-page--job">
      <JobWorkflowHero
        actionError={actionError}
        actionMessage={actionMessage}
        detectedContentLanguage={detectedContentLanguage}
        jobError={jobError}
        jobId={jobId}
        jobLoading={jobLoading}
        logCount={logs.length}
        logsError={logsError}
        logsLoading={logsLoading}
        notionDashboardUrl={notionDashboardUrl}
        slotCount={slots.length}
        slotsError={slotsError}
        slotsLoading={slotsLoading}
      />

      <JobOperationsGrid
        allSlotsRejected={allSlotsRejected}
        canRenderPreview={canRenderPreview}
        canStartAnalysis={canStartAnalysis}
        health={health}
        healthError={healthError}
        healthLoading={healthLoading}
        job={job}
        jobError={jobError}
        jobId={jobId}
        jobLoading={jobLoading}
        logs={logs}
        logsError={logsError}
        notionDashboardUrl={notionDashboardUrl}
        onRenderPreview={handleRenderPreview}
        onRepick={handleRepick}
        onStartAnalysis={handleStartAnalysis}
        preview={preview}
        previewDownloadUrl={previewDownloadUrl}
        previewError={previewError}
        previewLoading={previewLoading}
        refreshHealth={refreshHealth}
        renderPending={renderPending}
        repickPending={repickPending}
        selectedSlot={selectedSlot}
        startPending={startPending}
      />

      <JobSlotReviewPanel
        canSelectSlots={canSelectSlots}
        manualEndSeconds={manualEndSeconds}
        manualImportAudioPath={manualImportAudioPath}
        manualImportClipPath={manualImportClipPath}
        manualImportEndSeconds={manualImportEndSeconds}
        manualImportPending={manualImportPending}
        manualImportStartSeconds={manualImportStartSeconds}
        manualSelectPending={manualSelectPending}
        manualStartSeconds={manualStartSeconds}
        noAutoSlotFound={noAutoSlotFound}
        onManualEndSecondsChange={(event) => setManualEndSeconds(event.target.value)}
        onManualImport={handleManualImport}
        onManualImportAudioPathChange={(event) => setManualImportAudioPath(event.target.value)}
        onManualImportClipPathChange={(event) => setManualImportClipPath(event.target.value)}
        onManualImportEndSecondsChange={(event) => setManualImportEndSeconds(event.target.value)}
        onManualImportStartSecondsChange={(event) => setManualImportStartSeconds(event.target.value)}
        onManualSelect={handleManualSelect}
        onManualStartSecondsChange={(event) => setManualStartSeconds(event.target.value)}
        onReject={handleReject}
        onSelect={handleSelect}
        rejectingSlotId={rejectingSlotId}
        selectedSlot={selectedSlot}
        selectingSlotId={selectingSlotId}
        slots={slots}
        slotsError={slotsError}
        slotsLoading={slotsLoading}
      />

      {selectedSlot ? <JobSelectedSlotSummary selectedSlot={selectedSlot} /> : null}

      {selectedSlot ? <ProductLineEditor slot={selectedSlot} pending={generatePending} onGenerate={handleGenerate} /> : null}
    </div>
  );
}
