import { Link } from "react-router-dom";
import { JobStatusCard } from "../JobStatusCard";
import type { HealthResponse } from "../../types/Api";
import type { Job, JobLog } from "../../types/Job";
import type { Preview } from "../../types/Preview";
import type { Slot } from "../../types/Slot";

interface JobOperationsGridProps {
  job: Job | null;
  jobLoading: boolean;
  jobError: string | null;
  onStartAnalysis: () => void;
  canStartAnalysis: boolean;
  startPending: boolean;
  logs: JobLog[];
  logsError: string | null;
  notionDashboardUrl?: string;
  health: HealthResponse | null;
  healthError: string | null;
  healthLoading: boolean;
  refreshHealth: () => void;
  allSlotsRejected: boolean;
  repickPending: boolean;
  onRepick: () => void;
  preview: Preview | null;
  previewLoading: boolean;
  previewError: string | null;
  canRenderPreview: boolean;
  renderPending: boolean;
  onRenderPreview: () => void;
  previewDownloadUrl?: string;
  selectedSlot: Slot | null;
  jobId?: string;
}

export function JobOperationsGrid({
  job,
  jobLoading,
  jobError,
  onStartAnalysis,
  canStartAnalysis,
  startPending,
  logs,
  logsError,
  notionDashboardUrl,
  health,
  healthError,
  healthLoading,
  refreshHealth,
  allSlotsRejected,
  repickPending,
  onRepick,
  preview,
  previewLoading,
  previewError,
  canRenderPreview,
  renderPending,
  onRenderPreview,
  previewDownloadUrl,
  selectedSlot,
  jobId,
}: JobOperationsGridProps) {
  return (
    <div className="card-grid">
      <JobStatusCard
        job={job}
        loading={jobLoading}
        error={jobError}
        onStartAnalysis={onStartAnalysis}
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
            <p className="eyebrow">Notion audit</p>
            <h3>Live integration status</h3>
          </div>
          <button className="button-secondary" type="button" onClick={refreshHealth} disabled={healthLoading}>
            {healthLoading ? "Refreshing..." : "Refresh"}
          </button>
        </div>
        {healthError ? <p className="form-message form-message--error">{healthError}</p> : null}
        <p>Audit sink: {healthLoading && !health ? "loading" : health?.audit?.enabled ? "enabled" : "disabled"}</p>
        <p>Status: {healthLoading && !health ? "loading" : health?.audit?.status ?? "unknown"}</p>
        <p className="muted">{health?.audit?.details ?? "Audit status details are not available yet."}</p>
        <p className="muted">Backend health: {health?.status ?? "unknown"}</p>
        {notionDashboardUrl ? (
          <div className="form-actions">
            <a href={notionDashboardUrl} target="_blank" rel="noreferrer">
              Open Notion dashboard
            </a>
          </div>
        ) : null}
      </section>

      <section className="card">
        <div className="list-block__header">
          <div>
            <p className="eyebrow">Slot actions</p>
            <h3>Re-pick control</h3>
          </div>
          <button className="button-secondary" type="button" onClick={onRepick} disabled={!allSlotsRejected || repickPending}>
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
          <button className="button-secondary" type="button" onClick={onRenderPreview} disabled={!canRenderPreview}>
            {renderPending ? "Starting..." : preview?.status === "failed" ? "Retry render" : "Render preview"}
          </button>
        </div>
        <p className="muted">
          {previewLoading ? "Loading preview state..." : previewError ?? previewSummary(preview?.status, selectedSlot?.status)}
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
