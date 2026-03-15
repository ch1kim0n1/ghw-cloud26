import { FormEvent, useEffect, useMemo, useState } from "react";
import { Link } from "react-router-dom";
import { useJob } from "../hooks/useJob";
import { usePreview } from "../hooks/usePreview";
import { DownloadIcon, HeartIcon, SparkleIcon, UploadIcon } from "../components/PinkIcons";
import { startAnalysis } from "../services/analysisApi";
import { createCampaign } from "../services/campaignsApi";
import { getPreviewDownloadUrl } from "../services/previewApi";
import { ApiError } from "../types/Api";

const initialFormState = {
  campaignName: "",
  brandName: "",
  videoFile: null as File | null,
};

export function UploadPage() {
  const [formState, setFormState] = useState(initialFormState);
  const [submitting, setSubmitting] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [createdJobId, setCreatedJobId] = useState<string | null>(null);
  const [statusMessage, setStatusMessage] = useState<string | null>(null);

  const { job, error: jobError, loading: jobLoading } = useJob(createdJobId ?? undefined, {
    poll: Boolean(createdJobId),
  });
  const { preview, error: previewError } = usePreview(createdJobId ?? undefined, {
    poll: Boolean(createdJobId),
  });

  const previewDownloadUrl =
    createdJobId && preview?.status === "completed" && preview.output_video_path
      ? getPreviewDownloadUrl(createdJobId)
      : undefined;

  const friendlyStatus = useMemo(() => buildFriendlyStatus(job?.status, job?.current_stage, preview?.status), [
    job?.status,
    job?.current_stage,
    preview?.status,
  ]);

  useEffect(() => {
    if (!job && !preview && !createdJobId) {
      return;
    }

    if (preview?.status === "completed") {
      setStatusMessage("Your video made it all the way to a completed preview.");
      return;
    }

    if (job?.status === "failed") {
      setStatusMessage("This upload needs a hidden review check before it can keep going.");
      return;
    }

    if (job?.current_stage === "slot_selection" || job?.current_stage === "line_review") {
      setStatusMessage("Your upload is ready for the hidden review screen if you want to keep iterating.");
      return;
    }

    if (createdJobId) {
      setStatusMessage("Your upload is being processed now. You can stay here while it works.");
    }
  }, [createdJobId, job, preview]);

  async function handleSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    setError(null);
    setStatusMessage(null);

    if (!formState.videoFile) {
      setError("Please choose a source video first.");
      return;
    }

    if (formState.brandName.trim() === "") {
      setError("Please add a brand or product name.");
      return;
    }

    setSubmitting(true);

    const formData = new FormData();
    formData.set("name", formState.campaignName.trim() || `${formState.brandName.trim()} upload`);
    formData.set("video_file", formState.videoFile);
    formData.set("target_ad_duration_seconds", "6");
    formData.set("product_name", formState.brandName.trim());

    try {
      const campaign = await createCampaign(formData);
      setCreatedJobId(campaign.job_id);
      await startAnalysis(campaign.job_id);
      setStatusMessage("Upload complete. Analysis started automatically.");
      setFormState(initialFormState);
    } catch (reason) {
      if (reason instanceof ApiError) {
        setError(reason.message);
      } else {
        setError("Could not upload your video right now.");
      }
    } finally {
      setSubmitting(false);
    }
  }

  function resetFlow() {
    setFormState(initialFormState);
    setSubmitting(false);
    setError(null);
    setCreatedJobId(null);
    setStatusMessage(null);
  }

  const shouldShowProgressState = Boolean(createdJobId);

  return (
    <div className="public-page public-page--upload">
      <section className="upload-card">
        <div className="upload-card__intro">
          <span className="showcase-pill showcase-pill--pink">
            <UploadIcon className="inline-icon" />
            Upload a new video
          </span>
          <h1>Drop in a clip and let the pipeline start working right away.</h1>
          <p>
            Just give it a name, add the brand, and upload the video. Everything else stays behind the curtain unless
            you want the hidden review screen.
          </p>
          <div className="upload-card__chips">
            <span>
              <SparkleIcon className="inline-icon" />
              One-step flow
            </span>
            <span>
              <HeartIcon className="inline-icon" />
              Cute public UI
            </span>
          </div>
        </div>

        {!shouldShowProgressState ? (
          <form className="upload-form" onSubmit={handleSubmit}>
            <label className="cute-field">
              <span>Campaign name</span>
              <input
                value={formState.campaignName}
                onChange={(event) => setFormState((current) => ({ ...current, campaignName: event.target.value }))}
                placeholder="Spring soda moment"
                required
              />
            </label>

            <label className="cute-field">
              <span>Brand / Product name</span>
              <input
                value={formState.brandName}
                onChange={(event) => setFormState((current) => ({ ...current, brandName: event.target.value }))}
                placeholder="Cherry Pop"
                required
              />
            </label>

            <label className="cute-field">
              <span>Source video</span>
              <input
                type="file"
                accept=".mp4,video/mp4"
                onChange={(event) =>
                  setFormState((current) => ({ ...current, videoFile: event.target.files?.[0] ?? null }))
                }
                required
              />
            </label>

            <div className="upload-form__actions">
              <button type="submit" className="cute-button" disabled={submitting}>
                {submitting ? "Uploading..." : "Start processing"}
              </button>
              <Link className="cute-link" to="/">
                Back to showcase
              </Link>
            </div>
          </form>
        ) : (
          <div className="upload-status-card">
            <div className="upload-status-card__top">
              <div>
                <span className={`status-pill status-pill--${statusTone(job?.status, preview?.status)}`}>
                  {preview?.status === "completed" ? "completed" : job?.status ?? "queued"}
                </span>
                <h2>Upload status</h2>
              </div>
              <button type="button" className="cute-button cute-button--secondary" onClick={resetFlow}>
                Upload another video
              </button>
            </div>

            <p className="upload-status-card__message">{statusMessage ?? friendlyStatus}</p>

            <div className="upload-status-grid">
              <div>
                <span>Stage</span>
                <strong>{job?.current_stage ?? "starting"}</strong>
              </div>
              <div>
                <span>Progress</span>
                <strong>{job?.progress_percent ?? 0}%</strong>
              </div>
              <div>
                <span>Preview</span>
                <strong>{preview?.status ?? "not ready yet"}</strong>
              </div>
            </div>

            {error ? <p className="form-message form-message--error">{error}</p> : null}
            {jobError ? <p className="form-message form-message--error">{jobError}</p> : null}
            {previewError ? <p className="form-message form-message--error">{previewError}</p> : null}
            {jobLoading ? <p className="muted">Refreshing status...</p> : null}

            <div className="upload-status-card__actions">
              {createdJobId ? (
                <Link className="cute-link" to={`/jobs/${createdJobId}`}>
                  Open hidden review screen
                </Link>
              ) : null}
              {createdJobId && preview?.status === "completed" ? (
                <Link className="cute-link" to={`/jobs/${createdJobId}/preview`}>
                  Open hidden preview page
                </Link>
              ) : null}
              {previewDownloadUrl ? (
                <a className="cute-link" href={previewDownloadUrl} download>
                  <DownloadIcon className="inline-icon" />
                  Download preview
                </a>
              ) : null}
            </div>
          </div>
        )}

        {error && !shouldShowProgressState ? <p className="form-message form-message--error">{error}</p> : null}
      </section>
    </div>
  );
}

function buildFriendlyStatus(jobStatus?: string, currentStage?: string, previewStatus?: string) {
  if (previewStatus === "completed") {
    return "Your preview is ready.";
  }

  if (jobStatus === "failed") {
    return "Something went wrong and this upload needs a hidden review check.";
  }

  if (currentStage === "ready_for_analysis") {
    return "Upload is ready. Analysis is about to start.";
  }

  if (currentStage === "analysis_submission" || currentStage === "analysis_poll") {
    return "We are scanning the scene to find a sweet spot for the branded moment.";
  }

  if (currentStage === "slot_selection") {
    return "The system found candidate moments. This is the point where the hidden review screen can step in.";
  }

  if (currentStage === "line_review") {
    return "The line review step is waiting in the hidden review screen.";
  }

  if (currentStage === "generation_submission" || currentStage === "generation_poll") {
    return "The inserted moment is being generated now.";
  }

  if (currentStage === "render_poll" || previewStatus === "stitching" || previewStatus === "pending") {
    return "The final preview is being stitched together.";
  }

  if (jobStatus === "queued") {
    return "Your upload is queued up and ready to move.";
  }

  return "Your upload is on its way through the pipeline.";
}

function statusTone(jobStatus?: string, previewStatus?: string) {
  if (previewStatus === "completed") {
    return "success";
  }
  if (jobStatus === "failed") {
    return "error";
  }
  return "progress";
}
