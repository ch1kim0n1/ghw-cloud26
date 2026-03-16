import { useState } from "react";
import { Link } from "react-router-dom";
import { Reveal } from "../components/Reveal";
import { usePollingRequest } from "../hooks/usePollingRequest";
import { listJobs } from "../services/jobsApi";
import { ApiError } from "../types/Api";
import type { JobSummary } from "../types/Job";

export function StudioPage() {
  const [jobs, setJobs] = useState<JobSummary[]>([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const { refresh } = usePollingRequest(
    async (showLoading, isCancelled) => {
      if (showLoading) {
        setLoading(true);
      }

      try {
        const response = await listJobs();
        if (isCancelled()) {
          return;
        }
        setJobs(response.jobs ?? []);
        setError(null);
      } catch (reason) {
        if (isCancelled()) {
          return;
        }
        setError(reason instanceof ApiError ? reason.message : "Unable to load recent jobs.");
      } finally {
        if (!isCancelled()) {
          setLoading(false);
        }
      }
    },
    [],
    { poll: true, pollIntervalMs: 5000 },
  );

  return (
    <div className="studio-page studio-page--overview">
      <Reveal as="section" className="studio-hero studio-hero--job">
        <p className="eyebrow">Operator studio</p>
        <h2>Recent CAFAI jobs</h2>
        <p>
          The studio keeps the real workflow discoverable: upload a video, start analysis, review slots, and keep
          each active run one click away.
        </p>
        <div className="status-strip">
          <span>{loading ? "Refreshing jobs..." : `${jobs.length} recent run(s)`}</span>
          <span>{error ?? "Video workflow is the primary MVP lane"}</span>
        </div>
        <div className="form-actions">
          <Link to="/upload">Start a new upload</Link>
          <button className="button-secondary" type="button" onClick={refresh} disabled={loading}>
            {loading ? "Refreshing..." : "Refresh jobs"}
          </button>
        </div>
      </Reveal>

      <section className="panel panel--studio">
        <div className="list-block__header">
          <div>
            <p className="eyebrow">Workflow queue</p>
            <h2>Job Studio</h2>
          </div>
        </div>
        {error ? <p className="form-message form-message--error">{error}</p> : null}
        {!loading && jobs.length === 0 ? <p>No jobs yet. Start from the upload flow.</p> : null}
        <div className="card-grid">
          {jobs.map((job) => (
            <article key={job.id} className="card">
              <div className="list-block__header">
                <div>
                  <p className="eyebrow">{job.status}</p>
                  <h3>{job.id}</h3>
                </div>
                <Link className="button-secondary" to={`/jobs/${job.id}`}>
                  Open studio
                </Link>
              </div>

              <dl className="job-metadata">
                <div>
                  <dt>Campaign</dt>
                  <dd>{job.campaign_id}</dd>
                </div>
                <div>
                  <dt>Stage</dt>
                  <dd>{job.current_stage ?? "n/a"}</dd>
                </div>
                <div>
                  <dt>Progress</dt>
                  <dd>{job.progress_percent}%</dd>
                </div>
                <div>
                  <dt>Created</dt>
                  <dd>{job.created_at}</dd>
                </div>
              </dl>

              {job.error_code ? <p className="form-message form-message--error">Error code: {job.error_code}</p> : null}
            </article>
          ))}
        </div>
      </section>
    </div>
  );
}
