import type { Job } from "../types/Job";

interface JobStatusCardProps {
  job: Job | null;
  loading: boolean;
  error: string | null;
  onStartAnalysis: () => void;
  startDisabled: boolean;
  startPending: boolean;
}

export function JobStatusCard({
  job,
  loading,
  error,
  onStartAnalysis,
  startDisabled,
  startPending,
}: JobStatusCardProps) {
  return (
    <article className="card">
      <div className="list-block__header">
        <div>
          <p className="eyebrow">Job state</p>
          <h3>Analysis workflow</h3>
        </div>
        <button
          className="button-secondary"
          type="button"
          onClick={onStartAnalysis}
          disabled={startDisabled || startPending}
        >
          {startPending ? "Starting..." : "Start analysis"}
        </button>
      </div>

      {loading && !job ? <p>Loading job…</p> : null}
      {error ? <p className="muted">{error}</p> : null}
      {job ? (
        <dl className="job-metadata">
          <div>
            <dt>Status</dt>
            <dd>{job.status}</dd>
          </div>
          <div>
            <dt>Current stage</dt>
            <dd>{job.current_stage ?? "n/a"}</dd>
          </div>
          <div>
            <dt>Progress</dt>
            <dd>{job.progress_percent}%</dd>
          </div>
          <div>
            <dt>Started</dt>
            <dd>{job.started_at ?? "Not started"}</dd>
          </div>
          <div>
            <dt>Error</dt>
            <dd>{job.error_message ?? "None"}</dd>
          </div>
        </dl>
      ) : null}
    </article>
  );
}
