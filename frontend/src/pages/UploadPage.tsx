import { ChangeEvent, DragEvent, FormEvent, useEffect, useMemo, useRef, useState } from "react";
import { Link } from "react-router-dom";
import { FloatingDecor } from "../components/FloatingDecor";
import { DownloadIcon, HeartIcon, SparkleIcon, UploadIcon } from "../components/PinkIcons";
import { runtimeConfig } from "../config/runtime";
import { publicCopy } from "../content/publicCopy";
import { websiteAdsContent } from "../content/websiteAdsContent";
import { useJob } from "../hooks/useJob";
import { usePreview } from "../hooks/usePreview";
import { startAnalysis } from "../services/analysisApi";
import { createCampaign } from "../services/campaignsApi";
import { buildApiUrl } from "../services/apiClient";
import { getPreviewDownloadUrl } from "../services/previewApi";
import { listProducts } from "../services/productsApi";
import { createWebsiteAd } from "../services/websiteAdsApi";
import { ApiError } from "../types/Api";
import type { Product } from "../types/Product";
import type { WebsiteAdJob } from "../types/WebsiteAd";

const initialVideoFormState = {
  campaignName: "",
  brandName: "",
  videoFile: null as File | null,
};

type WebsiteUploadFormState = {
  productMode: "existing" | "custom";
  productId: string;
  productName: string;
  productDescription: string;
  articleHeadline: string;
  articleBody: string;
  brandStyle: string;
};

const initialWebsiteFormState: WebsiteUploadFormState = {
  productMode: "custom" as "existing" | "custom",
  productId: "",
  productName: "",
  productDescription: "",
  articleHeadline: "",
  articleBody: "",
  brandStyle: websiteAdsContent.styleOptions[0],
};

export function UploadPage() {
  const isShowcaseMode = runtimeConfig.showcaseMode;
  const [uploadMode, setUploadMode] = useState<"video" | "website">("video");
  const [videoFormState, setVideoFormState] = useState(initialVideoFormState);
  const [websiteFormState, setWebsiteFormState] = useState(initialWebsiteFormState);
  const [submitting, setSubmitting] = useState(false);
  const [dragging, setDragging] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [createdJobId, setCreatedJobId] = useState<string | null>(null);
  const [createdWebsiteAd, setCreatedWebsiteAd] = useState<WebsiteAdJob | null>(null);
  const [statusMessage, setStatusMessage] = useState<string | null>(null);
  const [products, setProducts] = useState<Product[]>([]);
  const fileInputRef = useRef<HTMLInputElement | null>(null);

  const { job, error: jobError, loading: jobLoading } = useJob(createdJobId ?? undefined, {
    poll: Boolean(createdJobId),
  });
  const { preview, error: previewError } = usePreview(createdJobId ?? undefined, {
    poll: Boolean(createdJobId),
  });

  useEffect(() => {
    if (isShowcaseMode || uploadMode !== "website" || products.length > 0) {
      return;
    }

    let cancelled = false;
    void listProducts()
      .then((response) => {
        if (!cancelled) {
          setProducts(Array.isArray(response.products) ? response.products : []);
        }
      })
      .catch(() => {
        if (!cancelled) {
          setProducts([]);
        }
      });

    return () => {
      cancelled = true;
    };
  }, [isShowcaseMode, products.length, uploadMode]);

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
    if (uploadMode !== "video") {
      return;
    }

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
  }, [createdJobId, job, preview, uploadMode]);

  async function handleVideoSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();

    if (isShowcaseMode) {
      return;
    }

    setError(null);
    setStatusMessage(null);

    if (!videoFormState.videoFile) {
      setError("Please choose a source video first.");
      return;
    }

    if (videoFormState.brandName.trim() === "") {
      setError("Please add a brand or product name.");
      return;
    }

    setSubmitting(true);

    const formData = new FormData();
    formData.set("name", videoFormState.campaignName.trim() || `${videoFormState.brandName.trim()} upload`);
    formData.set("video_file", videoFormState.videoFile);
    formData.set("target_ad_duration_seconds", "6");
    formData.set("product_name", videoFormState.brandName.trim());

    try {
      const campaign = await createCampaign(formData);
      setCreatedJobId(campaign.job_id);
      setCreatedWebsiteAd(null);
      await startAnalysis(campaign.job_id);
      setStatusMessage("Upload complete. Scene analysis started automatically.");
      setVideoFormState(initialVideoFormState);
      if (fileInputRef.current) {
        fileInputRef.current.value = "";
      }
    } catch (reason) {
      setError(reason instanceof ApiError ? reason.message : "Could not upload your video right now.");
    } finally {
      setSubmitting(false);
    }
  }

  async function handleWebsiteSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();

    if (isShowcaseMode) {
      return;
    }

    setError(null);
    setStatusMessage(null);
    setSubmitting(true);

    try {
      const payload =
        websiteFormState.productMode === "existing"
          ? {
              product_id: websiteFormState.productId,
              article_headline: websiteFormState.articleHeadline,
              article_body: websiteFormState.articleBody,
              brand_style: websiteFormState.brandStyle,
            }
          : {
              product_name: websiteFormState.productName,
              product_description: websiteFormState.productDescription,
              article_headline: websiteFormState.articleHeadline,
              article_body: websiteFormState.articleBody,
              brand_style: websiteFormState.brandStyle,
            };

      const websiteAd = await createWebsiteAd(payload);
      setCreatedWebsiteAd(websiteAd);
      setCreatedJobId(null);
      setWebsiteFormState(initialWebsiteFormState);
      setStatusMessage("Website ad set generated.");
    } catch (reason) {
      setError(reason instanceof ApiError ? reason.message : "Could not generate website ads right now.");
    } finally {
      setSubmitting(false);
    }
  }

  function assignFile(file: File | null) {
    setVideoFormState((current) => ({ ...current, videoFile: file }));
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

  function resetFlow(nextMode?: "video" | "website") {
    setVideoFormState(initialVideoFormState);
    setWebsiteFormState(initialWebsiteFormState);
    setSubmitting(false);
    setError(null);
    setCreatedJobId(null);
    setCreatedWebsiteAd(null);
    setStatusMessage(null);
    setDragging(false);
    if (nextMode) {
      setUploadMode(nextMode);
    }
    if (fileInputRef.current) {
      fileInputRef.current.value = "";
    }
  }

  const shouldShowVideoProgressState = uploadMode === "video" && Boolean(createdJobId);
  const shouldShowWebsiteResult = uploadMode === "website" && Boolean(createdWebsiteAd);

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

          <fieldset className="upload-mode-toggle mode-toggle">
            <legend>Pipeline mode</legend>
            <label>
              <input type="radio" name="upload-mode" checked={uploadMode === "video"} onChange={() => resetFlow("video")} />
              Video ad
            </label>
            <label>
              <input type="radio" name="upload-mode" checked={uploadMode === "website"} onChange={() => resetFlow("website")} />
              Website ad
            </label>
          </fieldset>
        </div>

        {isShowcaseMode ? (
          <ShowcaseUploadPanel uploadMode={uploadMode} />
        ) : uploadMode === "video" && !shouldShowVideoProgressState ? (
          <form className="upload-form" onSubmit={handleVideoSubmit}>
            <label className="cute-field">
              <span>Campaign name</span>
              <input
                value={videoFormState.campaignName}
                onChange={(event) => setVideoFormState((current) => ({ ...current, campaignName: event.target.value }))}
                placeholder="Cherry pixel dream"
                required
              />
            </label>

            <label className="cute-field">
              <span>Brand / Product name</span>
              <input
                value={videoFormState.brandName}
                onChange={(event) => setVideoFormState((current) => ({ ...current, brandName: event.target.value }))}
                placeholder="Cherry Pop"
                required
              />
            </label>

            <div
              className={`upload-dropzone${dragging ? " upload-dropzone--dragging" : ""}${videoFormState.videoFile ? " upload-dropzone--has-file" : ""}`}
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
                  <button className="cute-button cute-button--secondary" type="button" onClick={() => fileInputRef.current?.click()}>
                    Browse files
                  </button>
                  {videoFormState.videoFile ? (
                    <button className="cute-button cute-button--secondary" type="button" onClick={() => assignFile(null)}>
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
              <Link className="cute-button cute-button--secondary" to="/products">
                <HeartIcon className="inline-icon" />
                {publicCopy.upload.productCta}
              </Link>
              <Link className="cute-link" to="/">
                {publicCopy.upload.secondaryCta}
              </Link>
            </div>
          </form>
        ) : null}

        {shouldShowVideoProgressState ? (
          <div className="upload-status-card">
            <div className="upload-status-card__top">
              <div>
                <span className={`status-pill status-pill--${statusTone(job?.status, preview?.status)}`}>
                  {preview?.status === "completed" ? "completed" : job?.status ?? "queued"}
                </span>
                <h2>{publicCopy.upload.statusTitle}</h2>
              </div>
              <button type="button" className="cute-button cute-button--secondary" onClick={() => resetFlow("video")}>
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

        {uploadMode === "website" && !shouldShowWebsiteResult ? (
          <form className="upload-form upload-form--website" onSubmit={handleWebsiteSubmit}>
            <fieldset className="mode-toggle">
              <legend>Product source</legend>
              <label>
                <input
                  type="radio"
                  name="website-product-mode"
                  checked={websiteFormState.productMode === "custom"}
                  onChange={() => setWebsiteFormState((current) => ({ ...current, productMode: "custom" }))}
                />
                Custom product
              </label>
              <label>
                <input
                  type="radio"
                  name="website-product-mode"
                  checked={websiteFormState.productMode === "existing"}
                  onChange={() => setWebsiteFormState((current) => ({ ...current, productMode: "existing" }))}
                />
                Saved product
              </label>
            </fieldset>

            {websiteFormState.productMode === "existing" ? (
              <label className="cute-field">
                <span>Saved product</span>
                <select
                  value={websiteFormState.productId}
                  onChange={(event) => setWebsiteFormState((current) => ({ ...current, productId: event.target.value }))}
                  required
                >
                  <option value="">Select one product</option>
                  {products.map((product) => (
                    <option key={product.id} value={product.id}>
                      {product.name}
                    </option>
                  ))}
                </select>
              </label>
            ) : (
              <div className="field-row">
                <label className="cute-field">
                  <span>Product name</span>
                  <input
                    value={websiteFormState.productName}
                    onChange={(event) => setWebsiteFormState((current) => ({ ...current, productName: event.target.value }))}
                    placeholder="Cherry Pop"
                    required
                  />
                </label>
                <label className="cute-field">
                  <span>Product description</span>
                  <input
                    value={websiteFormState.productDescription}
                    onChange={(event) => setWebsiteFormState((current) => ({ ...current, productDescription: event.target.value }))}
                    placeholder="Sparkling soda for bright little desk breaks"
                    required
                  />
                </label>
              </div>
            )}

            <label className="cute-field">
              <span>Article headline</span>
              <input
                value={websiteFormState.articleHeadline}
                onChange={(event) => setWebsiteFormState((current) => ({ ...current, articleHeadline: event.target.value }))}
                placeholder="Teaching Cultural, Historical, and Religious Landscapes with the Anime Demon Slayer"
                required
              />
            </label>

            <label className="cute-field">
              <span>Article context</span>
              <textarea
                rows={7}
                value={websiteFormState.articleBody}
                onChange={(event) => setWebsiteFormState((current) => ({ ...current, articleBody: event.target.value }))}
                placeholder="Paste the article summary or relevant body text to drive the website ad pipeline."
                required
              />
            </label>

            <label className="cute-field">
              <span>Visual direction</span>
              <select
                value={websiteFormState.brandStyle}
                onChange={(event) => setWebsiteFormState((current) => ({ ...current, brandStyle: event.target.value }))}
              >
                {websiteAdsContent.styleOptions.map((option) => (
                  <option key={option} value={option}>
                    {option}
                  </option>
                ))}
              </select>
            </label>

            <div className="upload-form__actions">
              <button type="submit" className="cute-button" disabled={submitting}>
                {submitting ? "Generating..." : "Start website ad pipeline"}
              </button>
              <Link className="cute-button cute-button--secondary" to="/website-ads">
                <HeartIcon className="inline-icon" />
                Open website ads gallery
              </Link>
              <Link className="cute-link" to="/">
                {publicCopy.upload.secondaryCta}
              </Link>
            </div>
          </form>
        ) : null}

        {shouldShowWebsiteResult && createdWebsiteAd ? (
          <div className="upload-status-card upload-status-card--website">
            <div className="upload-status-card__top">
              <div>
                <span className="status-pill status-pill--success">completed</span>
                <h2>Website ad set ready</h2>
              </div>
              <button type="button" className="cute-button cute-button--secondary" onClick={() => resetFlow("website")}>
                Create another website ad
              </button>
            </div>

            <p className="upload-status-card__message">{statusMessage ?? "Your website banner pair is ready for review."}</p>

            <div className="upload-status-grid">
              <div>
                <span>Product</span>
                <strong>{createdWebsiteAd.product_name}</strong>
              </div>
              <div>
                <span>Status</span>
                <strong>{createdWebsiteAd.status}</strong>
              </div>
              <div>
                <span>Style</span>
                <strong>{createdWebsiteAd.brand_style || "default"}</strong>
              </div>
            </div>

            <div className="website-upload-preview">
              {createdWebsiteAd.banner_image_url ? (
                <figure>
                  <img src={buildApiUrl(createdWebsiteAd.banner_image_url)} alt={`${createdWebsiteAd.product_name} banner`} />
                  <figcaption>Horizontal banner</figcaption>
                </figure>
              ) : null}
              {createdWebsiteAd.vertical_image_url ? (
                <figure>
                  <img src={buildApiUrl(createdWebsiteAd.vertical_image_url)} alt={`${createdWebsiteAd.product_name} vertical banner`} />
                  <figcaption>Vertical banner</figcaption>
                </figure>
              ) : null}
            </div>

            <div className="upload-status-card__actions">
              <Link className="cute-link" to="/website-ads">
                Open website ads showcase
              </Link>
            </div>
          </div>
        ) : null}

        {error && ((uploadMode === "video" && !shouldShowVideoProgressState) || (uploadMode === "website" && !shouldShowWebsiteResult)) ? (
          <p className="form-message form-message--error">{error}</p>
        ) : null}
      </section>
    </div>
  );
}

function ShowcaseUploadPanel({ uploadMode }: { uploadMode: "video" | "website" }) {
  const showcaseExamples = websiteAdsContent.examples;

  return (
    <div className="upload-showcase-card">
      <div className="upload-showcase-card__intro">
        <span className="status-pill status-pill--progress">showcase build</span>
        <h2>{uploadMode === "video" ? "Video pipeline preview" : "Website ad pipeline preview"}</h2>
        <p>
          This GitHub Pages deployment is a static showcase. The live generation backend is disabled here, so this page explains the
          inputs and points to the completed examples instead of submitting real jobs.
        </p>
      </div>

      {uploadMode === "video" ? (
        <div className="upload-showcase-grid">
          <article className="upload-showcase-block">
            <h3>What the live video flow asks for</h3>
            <ul>
              <li>Campaign name</li>
              <li>Brand or product name</li>
              <li>One MP4 source clip</li>
            </ul>
          </article>

          <article className="upload-showcase-block">
            <h3>What happens in the real backend</h3>
            <ul>
              <li>Scene analysis and candidate slot detection</li>
              <li>Operator review or manual override</li>
              <li>Bridge generation and final preview stitching</li>
            </ul>
          </article>
        </div>
      ) : (
        <div className="upload-showcase-grid">
          <article className="upload-showcase-block">
            <h3>What the live website-ad flow asks for</h3>
            <ul>
              <li>Saved product or inline product info</li>
              <li>Article headline and article context</li>
              <li>Visual direction for the creative</li>
            </ul>
          </article>

          <article className="upload-showcase-block">
            <h3>What the backend generates</h3>
            <ul>
              <li>One horizontal banner at 1200x628</li>
              <li>One vertical sidebar ad at 300x600</li>
              <li>Stored assets and reviewable gallery output</li>
            </ul>
          </article>
        </div>
      )}

      {uploadMode === "website" ? (
        <div className="upload-showcase-examples">
          {showcaseExamples.map((example) => (
            <article className="upload-showcase-example" key={example.id}>
              <img src={example.previewImage} alt={`${example.title} injected placement preview`} />
              <div>
                <strong>{example.label}: {example.title}</strong>
                <p>{example.note}</p>
              </div>
            </article>
          ))}
        </div>
      ) : null}

      <div className="upload-form__actions">
        <Link className="cute-button" to={uploadMode === "video" ? "/gallery" : "/website-ads"}>
          {uploadMode === "video" ? "Open video gallery" : "Open website ads gallery"}
        </Link>
        <Link className="cute-button cute-button--secondary" to="/">
          <HeartIcon className="inline-icon" />
          Back to home
        </Link>
      </div>
    </div>
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

function formatFileSize(bytes: number) {
  if (bytes < 1024 * 1024) {
    return `${(bytes / 1024).toFixed(1)} KB`;
  }
  return `${(bytes / (1024 * 1024)).toFixed(1)} MB`;
}
