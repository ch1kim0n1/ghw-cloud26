import type { FormEvent } from "react";
import { Link } from "react-router-dom";
import { motion } from "framer-motion";
import { publicLayoutTransition } from "../publicMotion";
import { HeartIcon } from "../PinkIcons";
import { buildApiUrl } from "../../services/apiClient";
import { websiteAdsContent } from "../../content/websiteAdsContent";
import type { Product } from "../../types/Product";
import type { WebsiteAdJob } from "../../types/WebsiteAd";
import type { WebsiteUploadFormState } from "./types";

interface WebsiteAdsFlowProps {
  isShowcaseMode: boolean;
  active: boolean;
  submitting: boolean;
  products: Product[];
  productsLoading: boolean;
  error: string | null;
  createdWebsiteAd: WebsiteAdJob | null;
  statusMessage: string | null;
  websiteFormState: WebsiteUploadFormState;
  onSubmit: (event: FormEvent<HTMLFormElement>) => Promise<void>;
  onFormStateChange: (updater: (current: WebsiteUploadFormState) => WebsiteUploadFormState) => void;
  onResetFlow: () => void;
}

export function WebsiteAdsFlow({
  isShowcaseMode,
  active,
  submitting,
  products,
  productsLoading,
  error,
  createdWebsiteAd,
  statusMessage,
  websiteFormState,
  onSubmit,
  onFormStateChange,
  onResetFlow,
}: WebsiteAdsFlowProps) {
  const shouldShowWebsiteResult = Boolean(createdWebsiteAd);

  return (
    <>
      {!isShowcaseMode && active && !shouldShowWebsiteResult ? (
        <form className="upload-form upload-form--website" onSubmit={(event) => void onSubmit(event)}>
          <div className="success-summary">
            <p><strong>Experimental lane.</strong> Website ads stay available, but the primary CAFAI MVP is the video workflow.</p>
            <p>This request is synchronous today, so generation can take several seconds before results return.</p>
          </div>

          <fieldset className="mode-toggle">
            <legend>Product source</legend>
            <label
              className={
                websiteFormState.productMode === "custom" ? "mode-toggle__option mode-toggle__option--active" : "mode-toggle__option"
              }
            >
              <input
                type="radio"
                name="website-product-mode"
                checked={websiteFormState.productMode === "custom"}
                onChange={() => onFormStateChange((current) => ({ ...current, productMode: "custom" }))}
              />
              Custom product
            </label>
            <label
              className={
                websiteFormState.productMode === "existing" ? "mode-toggle__option mode-toggle__option--active" : "mode-toggle__option"
              }
            >
              <input
                type="radio"
                name="website-product-mode"
                checked={websiteFormState.productMode === "existing"}
                onChange={() => onFormStateChange((current) => ({ ...current, productMode: "existing" }))}
              />
              Saved product
            </label>
          </fieldset>

          {websiteFormState.productMode === "existing" ? (
            productsLoading ? (
              <div className="loading-surface" aria-hidden="true">
                <div className="loading-surface__bar" />
                <div className="loading-surface__bar loading-surface__bar--short" />
              </div>
            ) : (
              <label className="cute-field">
                <span>Saved product</span>
                <select
                  value={websiteFormState.productId}
                  onChange={(event) => onFormStateChange((current) => ({ ...current, productId: event.target.value }))}
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
            )
          ) : (
            <div className="field-row">
              <label className="cute-field">
                <span>Product name</span>
                <input
                  value={websiteFormState.productName}
                  onChange={(event) => onFormStateChange((current) => ({ ...current, productName: event.target.value }))}
                  placeholder="Cherry Pop"
                  required
                />
              </label>
              <label className="cute-field">
                <span>Product description</span>
                <input
                  value={websiteFormState.productDescription}
                  onChange={(event) =>
                    onFormStateChange((current) => ({ ...current, productDescription: event.target.value }))
                  }
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
              onChange={(event) => onFormStateChange((current) => ({ ...current, articleHeadline: event.target.value }))}
              placeholder="Teaching Cultural, Historical, and Religious Landscapes with the Anime Demon Slayer"
              required
            />
          </label>

          <label className="cute-field">
            <span>Article context</span>
            <textarea
              rows={7}
              value={websiteFormState.articleBody}
              onChange={(event) => onFormStateChange((current) => ({ ...current, articleBody: event.target.value }))}
              placeholder="Paste the article summary or relevant body text to drive the website ad pipeline."
              required
            />
          </label>

          <label className="cute-field">
            <span>Visual direction</span>
            <select
              value={websiteFormState.brandStyle}
              onChange={(event) => onFormStateChange((current) => ({ ...current, brandStyle: event.target.value }))}
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
              {submitting ? "Generating experimental ads..." : "Run website-ad experiment"}
            </button>
            <Link className="cute-button cute-button--secondary" to="/website-ads">
              <HeartIcon className="inline-icon" />
              Open website ads gallery
            </Link>
            <Link className="cute-link" to="/studio">
              Open operator studio
            </Link>
          </div>

          {submitting ? <p className="loading-inline">Generating banner and sidebar assets. This can take several seconds.</p> : null}
          {error ? <p className="form-message form-message--error">{error}</p> : null}
        </form>
      ) : null}

      {!isShowcaseMode && active && shouldShowWebsiteResult && createdWebsiteAd ? (
        <div className="upload-status-card upload-status-card--website">
          <div className="upload-status-card__top">
            <div>
              <span className="status-pill status-pill--success">completed</span>
              <h2>Experimental ad set ready</h2>
            </div>
            <button type="button" className="cute-button cute-button--secondary" onClick={onResetFlow}>
              Create another website ad
            </button>
          </div>

          <div className="upload-progress-rail upload-progress-rail--website" aria-label="Website ad progress">
            <motion.div className="upload-progress-step upload-progress-step--done" layout transition={publicLayoutTransition}>
              <small>brief</small>
              <strong>Context parsed</strong>
            </motion.div>
            <motion.div className="upload-progress-step upload-progress-step--done" layout transition={publicLayoutTransition}>
              <small>creative</small>
              <strong>Assets rendered</strong>
            </motion.div>
            <motion.div className="upload-progress-step upload-progress-step--active" layout transition={publicLayoutTransition}>
              <small>review</small>
              <strong>Gallery ready</strong>
            </motion.div>
          </div>

          <p className="upload-status-card__message">
            {statusMessage ?? "Your experimental website banner pair is ready for review on this page."}
          </p>

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
    </>
  );
}
