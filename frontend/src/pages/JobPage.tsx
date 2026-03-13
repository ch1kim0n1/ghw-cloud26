import { useEffect, useState } from "react";
import { useParams } from "react-router-dom";
import { JobStatusCard } from "../components/JobStatusCard";
import { SlotCard } from "../components/SlotCard";
import { useJob } from "../hooks/useJob";
import { useJobLogs } from "../hooks/useJobLogs";
import { useSlots } from "../hooks/useSlots";
import { startAnalysis } from "../services/analysisApi";
import { rejectSlot, repickSlots } from "../services/slotsApi";
import { ApiError } from "../types/Api";
import type { Job } from "../types/Job";

function isTerminalJob(job: Job | null): boolean {
  if (!job) {
    return false;
  }
  return job.status === "completed" || job.status === "failed";
}

export function JobPage() {
  const { jobId } = useParams();
  const [actionError, setActionError] = useState<string | null>(null);
  const [actionMessage, setActionMessage] = useState<string | null>(null);
  const [startPending, setStartPending] = useState(false);
  const [rejectingSlotId, setRejectingSlotId] = useState<string | null>(null);
  const [repickPending, setRepickPending] = useState(false);
  const [jobPollingEnabled, setJobPollingEnabled] = useState(Boolean(jobId));

  const { job, error: jobError, loading: jobLoading, refresh: refreshJob } = useJob(jobId, {
    poll: jobPollingEnabled,
  });

  useEffect(() => {
    setJobPollingEnabled(Boolean(jobId) && !isTerminalJob(job));
  }, [job, jobId]);

  const shouldPollSlots = Boolean(jobId && job && job.status === "analyzing" && job.current_stage !== "slot_selection");
  const { slots, error: slotsError, loading: slotsLoading, refresh: refreshSlots } = useSlots(jobId, {
    poll: shouldPollSlots,
  });
  const { logs, error: logsError, loading: logsLoading, refresh: refreshLogs } = useJobLogs(jobId, {
    poll: jobPollingEnabled,
  });

  const allSlotsRejected = slots.length > 0 && slots.every((slot) => slot.status === "rejected");
  const canStartAnalysis = Boolean(job && job.status === "queued" && job.current_stage === "ready_for_analysis");

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
      refreshJob();
      refreshSlots();
      refreshLogs();
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
      refreshJob();
      refreshSlots();
      refreshLogs();
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
      refreshJob();
      refreshSlots();
      refreshLogs();
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

  return (
    <div className="page-grid">
      <section className="panel">
        <p className="eyebrow">Job workflow</p>
        <h2>Job dashboard {jobId ?? "demo-job"}</h2>
        <p>
          Start analysis explicitly, follow worker progress, review proposed
          insertion slots, and request a re-pick once every current candidate
          has been rejected.
        </p>
        <div className="status-strip">
          <span>{jobLoading ? "Loading job..." : jobError ?? job?.status ?? "Job unavailable"}</span>
          <span>{slotsLoading ? "Loading slots..." : slotsError ?? `${slots.length} slot(s)`}</span>
          <span>{logsLoading ? "Loading logs..." : logsError ?? `${logs.length} log entry(ies)`}</span>
        </div>
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
            <button
              className="button-secondary"
              type="button"
              onClick={handleRepick}
              disabled={!allSlotsRejected || repickPending}
            >
              {repickPending ? "Requesting..." : "Re-pick slots"}
            </button>
          </div>
          <p className="muted">
            Re-pick is only available after all currently proposed slots have
            been rejected.
          </p>
          <p>
            Current gate: {allSlotsRejected ? "all slots rejected" : "waiting for more rejections"}
          </p>
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
        {!slotsLoading && slots.length === 0 ? <p>No slots available yet.</p> : null}
        <div className="card-grid">
          {slots.map((slot) => (
            <SlotCard
              key={slot.id}
              slot={slot}
              onReject={handleReject}
              rejectDisabled={rejectingSlotId !== null && rejectingSlotId !== slot.id}
            />
          ))}
        </div>
      </section>
    </div>
  );
}
