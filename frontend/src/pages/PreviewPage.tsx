import { Link, useParams } from "react-router-dom";
import { PreviewPlayer } from "../components/PreviewPlayer";
import { usePreview } from "../hooks/usePreview";
import { getPreviewDownloadUrl, getPreviewStreamUrl } from "../services/previewApi";

export function PreviewPage() {
  const { jobId } = useParams();
  const { preview, error, loading } = usePreview(jobId);

  const downloadUrl = jobId ? getPreviewDownloadUrl(jobId) : undefined;
  const streamUrl = jobId ? getPreviewStreamUrl(jobId) : undefined;
  const isCompleted = preview?.status === "completed" && Boolean(preview.output_video_path);

  return (
    <div className="page-grid">
      <PreviewPlayer videoUrl={isCompleted ? streamUrl : undefined} ready={isCompleted} />
      <section className="panel">
        <p className="eyebrow">Render state</p>
        <h2>Preview status {jobId ?? "demo-job"}</h2>
        <p>
          {loading ? "Loading preview..." : error ?? renderPreviewSummary(preview?.status)}
        </p>

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
        {preview?.output_video_path ? <p><strong>Local preview:</strong> {preview.output_video_path}</p> : null}

        <div className="form-actions">
          <Link className="button-secondary" to={`/jobs/${jobId}`}>Back to job</Link>
          {isCompleted && downloadUrl ? (
            <a href={downloadUrl} download>
              Download preview
            </a>
          ) : null}
        </div>
      </section>
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
