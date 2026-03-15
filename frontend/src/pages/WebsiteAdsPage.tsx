import { FormEvent, useEffect, useState } from "react";
import { Reveal } from "../components/Reveal";
import { websiteAdsContent } from "../content/websiteAdsContent";
import { publicCopy } from "../content/publicCopy";
import { buildApiUrl } from "../services/apiClient";
import { listProducts } from "../services/productsApi";
import { createWebsiteAd, listWebsiteAds } from "../services/websiteAdsApi";
import { ApiError } from "../types/Api";
import type { Product } from "../types/Product";
import type { WebsiteAdJob } from "../types/WebsiteAd";

type WebsiteAdsFormState = {
  mode: "existing" | "custom";
  productId: string;
  productName: string;
  productDescription: string;
  articleHeadline: string;
  articleBody: string;
  brandStyle: string;
};

const initialFormState: WebsiteAdsFormState = {
  mode: "existing" as "existing" | "custom",
  productId: "",
  productName: "",
  productDescription: "",
  articleHeadline: "",
  articleBody: "",
  brandStyle: websiteAdsContent.styleOptions[0],
};

export function WebsiteAdsPage() {
  const [products, setProducts] = useState<Product[]>([]);
  const [jobs, setJobs] = useState<WebsiteAdJob[]>([]);
  const [loading, setLoading] = useState(true);
  const [submitting, setSubmitting] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [success, setSuccess] = useState<string | null>(null);
  const [formState, setFormState] = useState(initialFormState);

  useEffect(() => {
    void loadPageData();
  }, []);

  async function loadPageData() {
    setLoading(true);
    try {
      const [productsResponse, jobsResponse] = await Promise.all([listProducts(), listWebsiteAds()]);
      setProducts(Array.isArray(productsResponse.products) ? productsResponse.products : []);
      setJobs(Array.isArray(jobsResponse.jobs) ? jobsResponse.jobs : []);
      setError(null);
    } catch (reason) {
      setError(reason instanceof ApiError ? reason.message : "Unable to load website ad tools.");
    } finally {
      setLoading(false);
    }
  }

  async function handleSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    setSubmitting(true);
    setError(null);
    setSuccess(null);

    try {
      const payload =
        formState.mode === "existing"
          ? {
              product_id: formState.productId,
              article_headline: formState.articleHeadline,
              article_body: formState.articleBody,
              brand_style: formState.brandStyle,
            }
          : {
              product_name: formState.productName,
              product_description: formState.productDescription,
              article_headline: formState.articleHeadline,
              article_body: formState.articleBody,
              brand_style: formState.brandStyle,
            };

      const job = await createWebsiteAd(payload);
      setJobs((current) => [job, ...current]);
      setSuccess(`Generated website ads for ${job.product_name}.`);
      setFormState((current) => ({
        ...initialFormState,
        mode: current.mode,
        productId: current.productId,
      }));
    } catch (reason) {
      setError(
        reason instanceof ApiError
          ? reason.message
          : "Unable to generate website ads right now.",
      );
    } finally {
      setSubmitting(false);
    }
  }

  return (
    <div className="studio-page">
      <Reveal as="section" className="website-ads-hero voxel-panel">
        <div className="section-heading section-heading--voxel">
          <p className="eyebrow">{publicCopy.websiteAds.eyebrow}</p>
          <h1>{publicCopy.websiteAds.title}</h1>
          <p>{publicCopy.websiteAds.lede}</p>
        </div>

        <div className="status-strip">
          {websiteAdsContent.chips.map((chip) => (
            <span key={chip}>{chip}</span>
          ))}
        </div>
      </Reveal>

      <div className="page-grid page-grid--studio">
        <section className="panel website-ads-form-panel">
          <p className="eyebrow">Generate a fresh pair</p>
          <h2>Website Ad Generator</h2>
          <p>Use an existing product or type one inline, then give the model enough article context to build static placements.</p>

          <form className="form-grid" onSubmit={handleSubmit}>
            <fieldset className="mode-toggle">
              <legend>Product source</legend>
              <label>
                <input
                  type="radio"
                  name="product-mode"
                  checked={formState.mode === "existing"}
                  onChange={() => setFormState((current) => ({ ...current, mode: "existing" }))}
                />
                Existing product
              </label>
              <label>
                <input
                  type="radio"
                  name="product-mode"
                  checked={formState.mode === "custom"}
                  onChange={() => setFormState((current) => ({ ...current, mode: "custom" }))}
                />
                Custom product
              </label>
            </fieldset>

            {formState.mode === "existing" ? (
              <label className="field">
                <span>Saved product</span>
                <select
                  value={formState.productId}
                  onChange={(event) => setFormState((current) => ({ ...current, productId: event.target.value }))}
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
                <label className="field">
                  <span>Product name</span>
                  <input
                    value={formState.productName}
                    onChange={(event) => setFormState((current) => ({ ...current, productName: event.target.value }))}
                    required
                  />
                </label>
                <label className="field">
                  <span>Product description</span>
                  <input
                    value={formState.productDescription}
                    onChange={(event) =>
                      setFormState((current) => ({ ...current, productDescription: event.target.value }))
                    }
                    required
                  />
                </label>
              </div>
            )}

            <label className="field">
              <span>Article headline</span>
              <input
                value={formState.articleHeadline}
                onChange={(event) => setFormState((current) => ({ ...current, articleHeadline: event.target.value }))}
                placeholder="The Colosseum: Engineering Marvel of Ancient Rome"
                required
              />
            </label>

            <label className="field">
              <span>Article context</span>
              <textarea
                value={formState.articleBody}
                onChange={(event) => setFormState((current) => ({ ...current, articleBody: event.target.value }))}
                rows={8}
                placeholder="Paste the article summary, body, or editorial context you want the ad to visually match."
                required
              />
            </label>

            <label className="field">
              <span>Visual direction</span>
              <select
                value={formState.brandStyle}
                onChange={(event) => setFormState((current) => ({ ...current, brandStyle: event.target.value }))}
              >
                {websiteAdsContent.styleOptions.map((option) => (
                  <option key={option} value={option}>
                    {option}
                  </option>
                ))}
              </select>
            </label>

            <div className="form-actions">
              <button type="submit" disabled={submitting || loading}>
                {submitting ? "Generating ads..." : "Generate website ads"}
              </button>
              <button type="button" className="button-secondary" onClick={() => void loadPageData()} disabled={loading}>
                Refresh wall
              </button>
            </div>
          </form>

          {error ? <p className="form-message form-message--error">{error}</p> : null}
          {success ? <p className="form-message form-message--success">{success}</p> : null}
        </section>

        <section className="panel panel--supporting">
          <p className="eyebrow">What the backend does</p>
          <h2>Real generation, not a mock gallery</h2>
          <p>
            The backend stores a generated horizontal banner plus a vertical sidebar version for each job, then serves
            the images back into the site through the same API host as the rest of the demo.
          </p>
          <p>
            If generation fails, the page will surface the backend error directly. The most common setup issue is a
            missing Hugging Face token.
          </p>
        </section>
      </div>

      <section className="voxel-panel website-ads-gallery">
        <div className="section-heading section-heading--voxel">
          <p className="eyebrow">Recent generations</p>
          <h2>Banner wall</h2>
          <p>The newest generated pairs stay here so the feature can work as both a tool and a proof surface.</p>
        </div>

        {loading ? <p>Loading website ad jobs...</p> : null}
        {!loading && jobs.length === 0 ? <p>No website ad jobs yet.</p> : null}

        <div className="website-ads-grid">
          {jobs.map((job) => (
            <article className="website-ad-card" key={job.id}>
              <div className="website-ad-card__header">
                <div>
                  <p className="eyebrow">{job.brand_style || "playful editorial"}</p>
                  <h3>{job.product_name}</h3>
                </div>
                <span className="voxel-chip voxel-chip--soft">{job.status}</span>
              </div>

              <p className="website-ad-card__headline">{job.article_headline}</p>

              <div className="website-ad-card__assets">
                {job.banner_image_url ? (
                  <figure className="website-ad-card__asset website-ad-card__asset--banner">
                    <img src={buildApiUrl(job.banner_image_url)} alt={`${job.product_name} horizontal banner`} />
                    <figcaption>1200 x 628 banner</figcaption>
                  </figure>
                ) : null}

                {job.vertical_image_url ? (
                  <figure className="website-ad-card__asset website-ad-card__asset--vertical">
                    <img src={buildApiUrl(job.vertical_image_url)} alt={`${job.product_name} vertical banner`} />
                    <figcaption>300 x 600 vertical</figcaption>
                  </figure>
                ) : null}
              </div>

              <details className="studio-disclosure">
                <summary>Prompt + context</summary>
                <div className="studio-disclosure__content">
                  <p>{job.prompt}</p>
                  <p className="muted">{job.article_body}</p>
                </div>
              </details>
            </article>
          ))}
        </div>
      </section>
    </div>
  );
}
