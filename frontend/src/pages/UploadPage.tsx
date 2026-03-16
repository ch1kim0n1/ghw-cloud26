import { AnimatePresence, motion, useReducedMotion } from "framer-motion";
import { ChangeEvent, DragEvent, FormEvent, useEffect, useRef, useState } from "react";
import { contentSwapVariants, publicSwapTransition } from "../components/publicMotion";
import { FloatingDecor } from "../components/FloatingDecor";
import { HeartIcon, SparkleIcon, UploadIcon } from "../components/PinkIcons";
import { ShowcaseUploadPanel } from "../components/upload/ShowcaseUploadPanel";
import { VideoUploadFlow } from "../components/upload/VideoUploadFlow";
import { WebsiteAdsFlow } from "../components/upload/WebsiteAdsFlow";
import type { VideoUploadFormState, WebsiteUploadFormState } from "../components/upload/types";
import { runtimeConfig } from "../config/runtime";
import { publicCopy } from "../content/publicCopy";
import { websiteAdsContent } from "../content/websiteAdsContent";
import { useJob } from "../hooks/useJob";
import { usePreview } from "../hooks/usePreview";
import { startAnalysis } from "../services/analysisApi";
import { createCampaign } from "../services/campaignsApi";
import { getPreviewDownloadUrl } from "../services/previewApi";
import { listProducts } from "../services/productsApi";
import { createWebsiteAd } from "../services/websiteAdsApi";
import { ApiError } from "../types/Api";
import type { Product } from "../types/Product";
import type { WebsiteAdJob } from "../types/WebsiteAd";

const initialVideoFormState: VideoUploadFormState = {
  campaignName: "",
  brandName: "",
  videoFile: null as File | null,
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
  const reducedMotion = useReducedMotion();
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
  const [productsLoading, setProductsLoading] = useState(false);
  const fileInputRef = useRef<HTMLInputElement>(null);

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
    setProductsLoading(true);
    void listProducts()
      .then((response) => {
        if (!cancelled) {
          setProducts(Array.isArray(response.products) ? response.products : []);
          setProductsLoading(false);
        }
      })
      .catch(() => {
        if (!cancelled) {
          setProducts([]);
          setProductsLoading(false);
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
  const activePanelKey = isShowcaseMode
    ? `showcase-${uploadMode}`
    : shouldShowVideoProgressState
      ? "video-progress"
      : shouldShowWebsiteResult
        ? "website-result"
        : `${uploadMode}-form`;

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
            <label className={uploadMode === "video" ? "mode-toggle__option mode-toggle__option--active" : "mode-toggle__option"}>
              <input type="radio" name="upload-mode" checked={uploadMode === "video"} onChange={() => resetFlow("video")} />
              Video ad
            </label>
            <label className={uploadMode === "website" ? "mode-toggle__option mode-toggle__option--active" : "mode-toggle__option"}>
              <input type="radio" name="upload-mode" checked={uploadMode === "website"} onChange={() => resetFlow("website")} />
              Website ad
            </label>
          </fieldset>
        </div>

        <AnimatePresence mode="wait" initial={false}>
          <motion.div
            key={activePanelKey}
            className="upload-panel-stack"
            initial={reducedMotion ? false : "hidden"}
            animate="show"
            exit={reducedMotion ? undefined : "exit"}
            variants={contentSwapVariants}
            transition={publicSwapTransition}
          >
            {isShowcaseMode ? <ShowcaseUploadPanel uploadMode={uploadMode} /> : null}

            <VideoUploadFlow
              active={uploadMode === "video"}
              createdJobId={createdJobId}
              error={error}
              fileInputRef={fileInputRef}
              isShowcaseMode={isShowcaseMode}
              job={job}
              jobError={jobError}
              jobLoading={jobLoading}
              onAssignFile={assignFile}
              onDrop={handleDrop}
              onFileSelection={handleFileSelection}
              onFormStateChange={setVideoFormState}
              onResetFlow={() => resetFlow("video")}
              onSetDragging={setDragging}
              onSubmit={handleVideoSubmit}
              preview={preview}
              previewDownloadUrl={previewDownloadUrl}
              previewError={previewError}
              statusMessage={statusMessage}
              submitting={submitting}
              videoFormState={videoFormState}
              dragging={dragging}
            />

            <WebsiteAdsFlow
              active={uploadMode === "website"}
              createdWebsiteAd={createdWebsiteAd}
              error={error}
              isShowcaseMode={isShowcaseMode}
              onFormStateChange={setWebsiteFormState}
              onResetFlow={() => resetFlow("website")}
              onSubmit={handleWebsiteSubmit}
              products={products}
              productsLoading={productsLoading}
              statusMessage={statusMessage}
              submitting={submitting}
              websiteFormState={websiteFormState}
            />
          </motion.div>
        </AnimatePresence>

        {error && ((uploadMode === "video" && !shouldShowVideoProgressState) || (uploadMode === "website" && !shouldShowWebsiteResult)) ? (
          <p className="form-message form-message--error">{error}</p>
        ) : null}
      </section>
    </div>
  );
}
