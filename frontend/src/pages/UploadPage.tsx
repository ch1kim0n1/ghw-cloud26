import { ChangeEvent, DragEvent, FormEvent, useEffect, useMemo, useRef, useState } from "react";
import { Link } from "react-router-dom";
import { FloatingDecor } from "../components/FloatingDecor";
import { DownloadIcon, HeartIcon, SparkleIcon, UploadIcon } from "../components/PinkIcons";
import { publicCopy } from "../content/publicCopy";
import { useJob } from "../hooks/useJob";
import { usePreview } from "../hooks/usePreview";
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
  const [dragging, setDragging] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [createdJobId, setCreatedJobId] = useState<string | null>(null);
  const [statusMessage, setStatusMessage] = useState<string | null>(null);
  const fileInputRef = useRef<HTMLInputElement | null>(null);

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
      setStatusMessage(publicCopy.upload.statusCompleted);
      return;
    }

    if (job?.status === "failed") {
      setStatusMessage("This run needs an operator pass before the final preview can recover.");
      return;
    }

    if (job?.current_stage === "slot_selection" || job?.current_stage === "line_review") {
      setStatusMessage("Your clip is ready for studio review if you want to keep polishing the result.");
      return;
    }

    if (createdJobId) {
      setStatusMessage(publicCopy.upload.statusQueued);
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
      setStatusMessage("Upload complete. Scene analysis started automatically.");
      setFormState(initialFormState);
      if (fileInputRef.current) {
        fileInputRef.current.value = "";
      }
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

  function assignFile(file: File | null) {
    setFormState((current) => ({ ...current, videoFile: file }));
    setDragging(false);
    setError(null);
  }

  function handleFileSelection(event: ChangeEvent<HTMLInputElement>) {
    assignFile(event.target.files?.[0] ?? null);
  }

  function handleDrop(event: DragEvent<HTMLDivElement>) {
    event.preventDefault();
    const file = event.dataTransfer.files?.[0] ?? null;
    assignFile(file);
  }

  function resetFlow() {
    setFormState(initialFormState);
    setSubmitting(false);
    setError(null);
    setCreatedJobId(null);
    setStatusMessage(null);
    setDragging(false);
    if (fileInputRef.current) {
      fileInputRef.current.value = "";
    }
  }

  const shouldShowProgressState = Boolean(createdJobId);

  return (
    <div className="public-page public-page--upload">
      <section className="upload-card voxel-panel">
        <FloatingDecor ids={["cloud", "bow", "flower"]} variant="upload" />

        <div className="upload-card__intro">
          <span className="voxel-chip">
            <UploadIcon className="inline-icon" />
            {publicCopy.upload.eyebrow}
          </span>
          <h1>{publicCopy.upload.title}</h1>
          <p>{publicCopy.upload.lede}</p>

          <div className="upload-card__chips">
            {publicCopy.upload.chips.map((chip, index) => (
              <span key={chip}>
                {index === 0 ? <SparkleIcon className="inline-icon" /> : <HeartIcon className="inline-icon" />}
                {chip}
              </span>
            ))}
          </div>
        </div>

        {!shouldShowProgressState ? (
          <form className="upload-form" onSubmit={handleSubmit}>
            <label className="cute-field">
              <span>Campaign name</span>
              <input
                value={formState.campaignName}
                onChange={(event) => setFormState((current) => ({ ...current, campaignName: event.target.value }))}
                placeholder="Cherry pixel dream"
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

            <div
              className={`upload-dropzone${dragging ? " upload-dropzone--dragging" : ""}${formState.videoFile ? " upload-dropzone--has-file" : ""}`}
              onDragEnter={(event) => {
                event.preventDefault();
                setDragging(true);
              }}
              onDragLeave={(event) => {
                event.preventDefault();
                setDragging(false);
              }}
              onDragOver={(event) => {
                event.preventDefault();
                setDragging(true);
              }}
              onDrop={handleDrop}
            >
              <input
                ref={fileInputRef}
                aria-label={publicCopy.upload.dropzoneTitle}
                className="upload-dropzone__input"
                type="file"
                accept=".mp4,video/mp4"
                onChange={handleFileSelection}
              />

              <div className="upload-dropzone__content">
                <span className="voxel-chip voxel-chip--soft">{publicCopy.upload.dropzoneTitle}</span>
                <strong>{publicCopy.upload.dropzoneHint}</strong>
                <p>{publicCopy.upload.dropzoneSubhint}</p>

                <div className="upload-dropzone__actions">
                  <button
                    className="cute-button cute-button--secondary"
                    type="button"
                    onClick={() => fileInputRef.current?.click()}
                  >
                    Browse files
                  </button>
                  {formState.videoFile ? (
                    <button className="cute-button cute-button--secondary" type="button" onClick={() => assignFile(null)}>
                      {publicCopy.upload.resetLabel}
                    </button>
                  ) : null}
                </div>
              </div>

              {formState.videoFile ? (
                <div className="upload-dropzone__file">
                  <span>{publicCopy.upload.selectedFileLabel}</span>
                  <strong>{formState.videoFile.name}</strong>
                  <small>{formatFileSize(formState.videoFile.size)}</small>
                </div>
              ) : null}
            </div>

            <div className="upload-form__actions">
              <button type="submit" className="cute-button" disabled={submitting}>
                {submitting ? "Uploading..." : publicCopy.upload.primaryCta}
              </button>
              <Link className="cute-button cute-button--secondary" to="/products">
                <HeartIcon className="inline-icon" />
                {publicCopy.upload.productCta}
              </Link>
              <Link className="cute-link" to="/">
                {publicCopy.upload.secondaryCta}
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
                <h2>{publicCopy.upload.statusTitle}</h2>
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
            {jobLoading ? <p className="muted">Refreshing pipeline status...</p> : null}

            <div className="upload-status-card__actions">
              {createdJobId ? (
                <Link className="cute-link" to={`/jobs/${createdJobId}`}>
                  {publicCopy.upload.reviewLink}
                </Link>
              ) : null}
              {createdJobId && preview?.status === "completed" ? (
                <Link className="cute-link" to={`/jobs/${createdJobId}/preview`}>
                  {publicCopy.upload.previewLink}
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
    return publicCopy.upload.statusCompleted;
  }

  if (jobStatus === "failed") {
    return "Something slipped during processing, so this clip needs a quick studio pass.";
  }

  if (currentStage === "ready_for_analysis") {
    return "The upload is set. Scene analysis is about to start.";
  }

  if (currentStage === "analysis_submission" || currentStage === "analysis_poll") {
    return "CAFAI is reading the scene and looking for a believable insert window.";
  }

  if (currentStage === "slot_selection") {
    return "The insert candidates are ready for studio review.";
  }

  if (currentStage === "line_review") {
    return "The product line is being polished before generation.";
  }

  if (jobStatus === "generating") {
    return "The branded bridge clip is generating now.";
  }

  if (jobStatus === "stitching" || previewStatus === "stitching") {
    return "The final preview is being stitched together.";
  }

  return "Your upload is queued and waiting for the next pipeline step.";
}

function statusTone(jobStatus?: string, previewStatus?: string) {
  if (previewStatus === "completed") {
    return "success";
  }

  if (jobStatus === "failed" || previewStatus === "failed") {
    return "error";
  }

  return "progress";
}

function formatFileSize(bytes: number) {
  if (bytes < 1024 * 1024) {
    return `${Math.max(1, Math.round(bytes / 1024))} KB`;
  }

  return `${(bytes / (1024 * 1024)).toFixed(1)} MB`;
}
