import { AnimatePresence, motion, useReducedMotion } from "framer-motion";
import type { ChangeEvent, DragEvent, FormEvent, RefObject } from "react";
import { Link } from "react-router-dom";
import { contentSwapVariants, publicLayoutTransition, publicSwapTransition } from "../publicMotion";
import { DownloadIcon, HeartIcon } from "../PinkIcons";
import { publicCopy } from "../../content/publicCopy";
import type { Job } from "../../types/Job";
import type { Preview } from "../../types/Preview";
import type { VideoUploadFormState } from "./types";

interface VideoUploadFlowProps {
  isShowcaseMode: boolean;
  active: boolean;
  videoFormState: VideoUploadFormState;
  dragging: boolean;
  submitting: boolean;
  error: string | null;
  createdJobId: string | null;
  job: Job | null;
  jobError: string | null;
  jobLoading: boolean;
  preview: Preview | null;
  previewError: string | null;
  previewDownloadUrl?: string;
  statusMessage: string | null;
  fileInputRef: RefObject<HTMLInputElement>;
  onSubmit: (event: FormEvent<HTMLFormElement>) => Promise<void>;
  onFormStateChange: (updater: (current: VideoUploadFormState) => VideoUploadFormState) => void;
  onAssignFile: (file: File | null) => void;
  onFileSelection: (event: ChangeEvent<HTMLInputElement>) => void;
  onDrop: (event: DragEvent<HTMLDivElement>) => void;
  onSetDragging: (dragging: boolean) => void;
  onResetFlow: () => void;
}

export function VideoUploadFlow({
  isShowcaseMode,
  active,
  videoFormState,
  dragging,
  submitting,
  error,
  createdJobId,
  job,
  jobError,
  jobLoading,
  preview,
  previewError,
  previewDownloadUrl,
  statusMessage,
  fileInputRef,
  onSubmit,
  onFormStateChange,
  onAssignFile,
  onFileSelection,
  onDrop,
  onSetDragging,
  onResetFlow,
}: VideoUploadFlowProps) {
  const reducedMotion = useReducedMotion();
  const shouldShowVideoProgressState = Boolean(createdJobId);
  const friendlyStatus = buildFriendlyStatus(job?.status, job?.current_stage, preview?.status);
  const progressSteps = buildUploadProgressSteps(job?.status, job?.current_stage, preview?.status);

  return (
    <>
      {!isShowcaseMode && active && !shouldShowVideoProgressState ? (
        <form className="upload-form" onSubmit={(event) => void onSubmit(event)}>
          <label className="cute-field">
            <span>Campaign name</span>
            <input
              value={videoFormState.campaignName}
              onChange={(event) => onFormStateChange((current) => ({ ...current, campaignName: event.target.value }))}
              placeholder="Cherry pixel dream"
              required
            />
          </label>

          <label className="cute-field">
            <span>Brand / Product name</span>
            <input
              value={videoFormState.brandName}
              onChange={(event) => onFormStateChange((current) => ({ ...current, brandName: event.target.value }))}
              placeholder="Cherry Pop"
              required
            />
          </label>

          <div
            className={`upload-dropzone${dragging ? " upload-dropzone--dragging" : ""}${videoFormState.videoFile ? " upload-dropzone--has-file" : ""}`}
            onDragEnter={(event) => {
              event.preventDefault();
              onSetDragging(true);
            }}
            onDragLeave={(event) => {
              event.preventDefault();
              onSetDragging(false);
            }}
            onDragOver={(event) => {
              event.preventDefault();
              onSetDragging(true);
            }}
            onDrop={onDrop}
          >
            <input
              ref={fileInputRef}
              aria-label={publicCopy.upload.dropzoneTitle}
              className="upload-dropzone__input"
              type="file"
              accept=".mp4,video/mp4"
              onChange={onFileSelection}
            />

            <div className="upload-dropzone__content">
              <div className="upload-dropzone__chips">
                <span className="voxel-chip voxel-chip--soft">{publicCopy.upload.dropzoneTitle}</span>
                <span className={`signal-pill${dragging ? " signal-pill--warm" : ""}`}>
                  {dragging ? "drop to attach" : videoFormState.videoFile ? "file ready" : "mp4 only"}
                </span>
              </div>
              <strong>{publicCopy.upload.dropzoneHint}</strong>
              <p>{publicCopy.upload.dropzoneSubhint}</p>

              <div className="upload-dropzone__actions">
                <button className="cute-button cute-button--secondary" type="button" onClick={() => fileInputRef.current?.click()}>
                  Browse files
                </button>
                {videoFormState.videoFile ? (
                  <button className="cute-button cute-button--secondary" type="button" onClick={() => onAssignFile(null)}>
                    {publicCopy.upload.resetLabel}
                  </button>
                ) : null}
              </div>
            </div>

            {videoFormState.videoFile ? (
              <div className="upload-dropzone__file">
                <span>{publicCopy.upload.selectedFileLabel}</span>
                <strong>{videoFormState.videoFile.name}</strong>
                <small>{formatFileSize(videoFormState.videoFile.size)}</small>
              </div>
            ) : null}
          </div>

          <div className="upload-form__actions">
            <button type="submit" className="cute-button" disabled={submitting}>
              {submitting ? "Uploading..." : publicCopy.upload.primaryCta}
            </button>
            <Link className="cute-button cute-button--secondary" to="/studio">
              <HeartIcon className="inline-icon" />
              Open studio
            </Link>
            <Link className="cute-link" to="/">
              {publicCopy.upload.secondaryCta}
            </Link>
          </div>
        </form>
      ) : null}

      {!isShowcaseMode && active && shouldShowVideoProgressState ? (
        <div className="upload-status-card">
          <div className="upload-status-card__top">
            <div>
              <span className={`status-pill status-pill--${statusTone(job?.status, preview?.status)}`}>
                {preview?.status === "completed" ? "completed" : job?.status ?? "queued"}
              </span>
              <h2>{publicCopy.upload.statusTitle}</h2>
            </div>
            <button type="button" className="cute-button cute-button--secondary" onClick={onResetFlow}>
              Upload another video
            </button>
          </div>

          <div className="upload-progress-rail" aria-label="Pipeline progress">
            {progressSteps.map((step) => (
              <motion.div
                key={step.label}
                className={`upload-progress-step upload-progress-step--${step.state}`}
                layout
                transition={publicLayoutTransition}
              >
                <small>{step.kicker}</small>
                <strong>{step.label}</strong>
              </motion.div>
            ))}
          </div>

          <AnimatePresence mode="wait" initial={false}>
            <motion.p
              key={statusMessage ?? friendlyStatus}
              className="upload-status-card__message"
              initial={reducedMotion ? false : { opacity: 0, y: 8 }}
              animate={{ opacity: 1, y: 0 }}
              exit={reducedMotion ? undefined : { opacity: 0, y: -6 }}
              transition={publicSwapTransition}
            >
              {statusMessage ?? friendlyStatus}
            </motion.p>
          </AnimatePresence>

          <div className="upload-status-grid">
            <motion.div layout transition={publicLayoutTransition}>
              <span>Stage</span>
              <strong>{job?.current_stage ?? "starting"}</strong>
            </motion.div>
            <motion.div layout transition={publicLayoutTransition}>
              <span>Progress</span>
              <strong>{job?.progress_percent ?? 0}%</strong>
            </motion.div>
            <motion.div layout transition={publicLayoutTransition}>
              <span>Preview</span>
              <strong>{preview?.status ?? "not ready yet"}</strong>
            </motion.div>
          </div>

          {error ? <p className="form-message form-message--error">{error}</p> : null}
          {jobError ? <p className="form-message form-message--error">{jobError}</p> : null}
          {previewError ? <p className="form-message form-message--error">{previewError}</p> : null}
          {jobLoading ? <p className="loading-inline">Refreshing pipeline status...</p> : null}

          <div className="upload-status-card__actions">
            {createdJobId ? <Link className="cute-link" to={`/jobs/${createdJobId}`}>{publicCopy.upload.reviewLink}</Link> : null}
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
      ) : null}
    </>
  );
}

function buildFriendlyStatus(jobStatus?: string, currentStage?: string, previewStatus?: string) {
  if (previewStatus === "completed") {
    return publicCopy.upload.statusCompleted;
  }

  if (jobStatus === "failed") {
    return "This run needs an operator pass before the final preview can recover.";
  }

  if (currentStage === "slot_selection" || currentStage === "line_review") {
    return "Your clip is ready for studio review if you want to keep polishing the result.";
  }

  return publicCopy.upload.statusQueued;
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

function buildUploadProgressSteps(jobStatus?: string, currentStage?: string, previewStatus?: string) {
  const isCompleted = previewStatus === "completed";
  const isReviewReady = currentStage === "slot_selection" || currentStage === "line_review";
  const isAnalyzing = Boolean(currentStage) && !isReviewReady && !isCompleted;
  const queueState = isCompleted || isReviewReady || isAnalyzing || jobStatus ? "done" : "active";
  const analyzingState = isCompleted || isReviewReady ? "done" : isAnalyzing ? "active" : "idle";
  const reviewState = isCompleted ? "done" : isReviewReady ? "active" : "idle";
  const previewState = isCompleted ? "done" : isReviewReady ? "active" : "idle";

  return [
    { kicker: "step 1", label: "Queued", state: queueState },
    { kicker: "step 2", label: "Analyzing", state: analyzingState },
    { kicker: "step 3", label: "Review ready", state: reviewState },
    { kicker: "step 4", label: "Preview", state: previewState },
  ] as const;
}

function formatFileSize(bytes: number) {
  if (bytes < 1024 * 1024) {
    return `${(bytes / 1024).toFixed(1)} KB`;
  }
  return `${(bytes / (1024 * 1024)).toFixed(1)} MB`;
}
