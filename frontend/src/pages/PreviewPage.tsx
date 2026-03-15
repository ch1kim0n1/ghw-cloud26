import { Link, useParams } from "react-router-dom";
import { PreviewPlayer } from "../components/PreviewPlayer";
import { Reveal } from "../components/Reveal";
import { SectionHeading } from "../components/SectionHeading";
import { usePreview } from "../hooks/usePreview";
import { getPreviewDownloadUrl, getPreviewStreamUrl } from "../services/previewApi";
import { latestDemoExample } from "../content/demoContent";

export function PreviewPage() {
  const { jobId } = useParams();
  const { preview, error, loading } = usePreview(jobId);

  const downloadUrl = jobId ? getPreviewDownloadUrl(jobId) : undefined;
  const streamUrl = jobId ? getPreviewStreamUrl(jobId) : undefined;
  const isCompleted = preview?.status === "completed" && Boolean(preview.output_video_path);

  return (
    <div className="preview-page">
      <Reveal as="section" className="hero-panel hero-panel--preview">
        <div className="hero-panel__copy">
          <p className="eyebrow">Review surface</p>
          <h1>Preview the stitched scene in a dedicated presentation frame.</h1>
          <p className="hero-panel__lede">
            This route keeps the focus on playback while still surfacing render status, selected slot, and download
            readiness for live demos.
          </p>
          <div className="hero-actions">
            <Link className="button-link button-link--secondary" to={`/jobs/${jobId ?? latestDemoExample.jobId}`}>
              Enter studio job
            </Link>
            {isCompleted && downloadUrl ? (
              <a className="button-link" href={downloadUrl} download>
                Download preview
              </a>
            ) : null}
          </div>
        </div>

        <div className="preview-header__status">
          <div className="signal-pill signal-pill--warm">{loading ? "Loading preview..." : preview?.status ?? "waiting"}</div>
          <p>{error ?? renderPreviewSummary(preview?.status)}</p>
        </div>
      </Reveal>

      <div className="preview-layout">
        <Reveal as="section" className="preview-layout__player" delay={0.08}>
          <PreviewPlayer videoUrl={isCompleted ? streamUrl : undefined} ready={isCompleted} />
        </Reveal>

        <Reveal as="section" className="preview-layout__meta panel" delay={0.12}>
          <SectionHeading eyebrow="Render state" title={`Preview status ${jobId ?? "demo-job"}`} compact />
          <p className="preview-layout__summary">{loading ? "Loading preview..." : error ?? renderPreviewSummary(preview?.status)}</p>

          {preview ? (
            <dl className="job-metadata">
              <div>
                <dt>Status</dt>
                <dd>{preview.status}</dd>
              </div>
              <div>
                <dt>Slot</dt>
                <dd>{preview.slot_id}</dd>
              </div>
              <div>
                <dt>Duration</dt>
                <dd>{preview.duration_seconds ? `${preview.duration_seconds.toFixed(1)}s` : "n/a"}</dd>
              </div>
              <div>
                <dt>Retries</dt>
                <dd>{preview.render_retry_count ?? 0}</dd>
              </div>
            </dl>
          ) : null}

          {preview?.error_message ? <p className="form-message form-message--error">{preview.error_message}</p> : null}
          {preview?.output_video_path ? <p className="muted"><strong>Local preview:</strong> {preview.output_video_path}</p> : null}

          <div className="form-actions">
            <Link className="button-secondary" to={`/jobs/${jobId}`}>
              Back to job
            </Link>
            {isCompleted && downloadUrl ? (
              <a className="button-secondary" href={downloadUrl} download>
                Download preview
              </a>
            ) : null}
          </div>
        </Reveal>
      </div>
    </div>
  );
}

function renderPreviewSummary(status?: string): string {
  switch (status) {
    case "pending":
      return "Preview render is queued.";
    case "stitching":
      return "Preview render is in progress.";
    case "completed":
      return "Preview render is complete.";
    case "failed":
      return "Preview render failed.";
    default:
      return "No preview has been started yet.";
  }
}
