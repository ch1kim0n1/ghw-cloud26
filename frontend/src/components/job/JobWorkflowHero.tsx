import { Reveal } from "../Reveal";

interface JobWorkflowHeroProps {
  jobId?: string;
  jobLoading: boolean;
  jobError: string | null;
  slotsLoading: boolean;
  slotsError: string | null;
  slotCount: number;
  logsLoading: boolean;
  logsError: string | null;
  logCount: number;
  detectedContentLanguage: string;
  notionDashboardUrl?: string;
  actionError: string | null;
  actionMessage: string | null;
}

export function JobWorkflowHero({
  jobId,
  jobLoading,
  jobError,
  slotsLoading,
  slotsError,
  slotCount,
  logsLoading,
  logsError,
  logCount,
  detectedContentLanguage,
  notionDashboardUrl,
  actionError,
  actionMessage,
}: JobWorkflowHeroProps) {
  return (
    <Reveal as="section" className="studio-hero studio-hero--job">
      <p className="eyebrow">Job workflow</p>
      <h2>Job studio {jobId ?? "demo-job"}</h2>
      <p>
        Start analysis explicitly, review the ranked insertion slots, prepare one product line, and start CAFAI
        generation for the selected candidate.
      </p>
      <p className="muted">
        Demo runs can bypass live generation by importing a locally generated MP4 into the selected slot and then
        continuing through preview render.
      </p>
      <div className="status-strip">
        <span>{jobLoading ? "Loading job..." : jobError ?? "Job ready"}</span>
        <span>{slotsLoading ? "Loading slots..." : slotsError ?? `${slotCount} slot(s)`}</span>
        <span>{logsLoading ? "Loading logs..." : logsError ?? `${logCount} log entry(ies)`}</span>
      </div>
      <p>Detected content language: {detectedContentLanguage.toUpperCase()}</p>
      {notionDashboardUrl ? (
        <p>
          <a href={notionDashboardUrl} target="_blank" rel="noreferrer">
            View in Notion
          </a>
        </p>
      ) : null}
      {actionError ? <p className="form-message form-message--error">{actionError}</p> : null}
      {actionMessage ? <p className="form-message form-message--success">{actionMessage}</p> : null}
    </Reveal>
  );
}
