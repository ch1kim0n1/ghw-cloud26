import { FormEvent, useEffect, useState } from "react";
import { createCampaign } from "../services/campaignsApi";
import { listProducts } from "../services/productsApi";
import { ApiError } from "../types/Api";
import type { Campaign } from "../types/Campaign";
import type { Product } from "../types/Product";

type ProductMode = "existing" | "inline";

const initialInlineProduct = {
  name: "",
  description: "",
  category: "",
  contextKeywords: "",
  sourceUrl: "",
  imageFile: null as File | null,
};

export function CampaignForm() {
  const [products, setProducts] = useState<Product[]>([]);
  const [productsLoading, setProductsLoading] = useState(true);
  const [productsError, setProductsError] = useState<string | null>(null);
  const [campaignError, setCampaignError] = useState<string | null>(null);
  const [success, setSuccess] = useState<Campaign | null>(null);
  const [saving, setSaving] = useState(false);
  const [productMode, setProductMode] = useState<ProductMode>("existing");
  const [name, setName] = useState("");
  const [videoFile, setVideoFile] = useState<File | null>(null);
  const [targetAdDurationSeconds, setTargetAdDurationSeconds] = useState("6");
  const [productId, setProductId] = useState("");
  const [inlineProduct, setInlineProduct] = useState(initialInlineProduct);

  useEffect(() => {
    void loadProducts();
  }, []);

  async function loadProducts() {
    setProductsLoading(true);
    try {
      const response = await listProducts();
      const nextProducts = Array.isArray(response.products) ? response.products : [];
      setProducts(nextProducts);
      setProductsError(null);
      if (nextProducts.length === 0) {
        setProductMode("inline");
      } else if (!productId) {
        setProductId(nextProducts[0].id);
      }
    } catch (reason) {
      setProductsError(reason instanceof ApiError ? reason.message : "Unable to load products.");
      setProductMode("inline");
    } finally {
      setProductsLoading(false);
    }
  }

  async function handleSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    setCampaignError(null);
    setSuccess(null);

    if (!videoFile) {
      setCampaignError("A source video is required.");
      return;
    }
    if (productMode === "existing" && !productId) {
      setCampaignError("Select an existing product or switch to inline product creation.");
      return;
    }

    const formData = new FormData();
    formData.set("name", name);
    formData.set("video_file", videoFile);
    if (targetAdDurationSeconds.trim()) {
      formData.set("target_ad_duration_seconds", targetAdDurationSeconds);
    }

    if (productMode === "existing") {
      formData.set("product_id", productId);
    } else {
      formData.set("product_name", inlineProduct.name);
      if (inlineProduct.description.trim()) formData.set("product_description", inlineProduct.description);
      if (inlineProduct.category.trim()) formData.set("product_category", inlineProduct.category);
      if (inlineProduct.contextKeywords.trim()) {
        formData.set("product_context_keywords", inlineProduct.contextKeywords);
      }
      if (inlineProduct.sourceUrl.trim()) formData.set("product_source_url", inlineProduct.sourceUrl);
      if (inlineProduct.imageFile) formData.set("product_image_file", inlineProduct.imageFile);
    }

    setSaving(true);
    try {
      const campaign = await createCampaign(formData);
      setSuccess(campaign);
      setCampaignError(null);
      setName("");
      setVideoFile(null);
      setTargetAdDurationSeconds("6");
      setInlineProduct(initialInlineProduct);
    } catch (reason) {
      setCampaignError(reason instanceof ApiError ? reason.message : "Unable to create campaign.");
    } finally {
      setSaving(false);
    }
  }

  return (
    <section className="panel">
      <p className="eyebrow">Phase 1 live</p>
      <h2>Campaign Intake</h2>
      <p>Create a campaign, upload the source video, and leave the job ready for explicit analysis start.</p>

      <form className="form-grid" onSubmit={handleSubmit}>
        <label className="field">
          <span>Campaign name</span>
          <input name="name" value={name} onChange={(event) => setName(event.target.value)} required />
        </label>

        <div className="field-row">
          <label className="field">
            <span>Source video (H.264 MP4)</span>
            <input
              name="video_file"
              type="file"
              accept=".mp4,video/mp4"
              onChange={(event) => setVideoFile(event.target.files?.[0] ?? null)}
              required
            />
          </label>

          <label className="field">
            <span>Target ad duration (seconds)</span>
            <input
              name="target_ad_duration_seconds"
              type="number"
              min="1"
              max="8"
              value={targetAdDurationSeconds}
              onChange={(event) => setTargetAdDurationSeconds(event.target.value)}
            />
          </label>
        </div>

        <fieldset className="mode-toggle">
          <legend>Product source</legend>
          <label>
            <input
              type="radio"
              name="product_mode"
              checked={productMode === "existing"}
              onChange={() => setProductMode("existing")}
              disabled={products.length === 0}
            />
            Use existing product
          </label>
          <label>
            <input
              type="radio"
              name="product_mode"
              checked={productMode === "inline"}
              onChange={() => setProductMode("inline")}
            />
            Create inline product
          </label>
        </fieldset>

        {productMode === "existing" ? (
          <label className="field">
            <span>Existing product</span>
            <select value={productId} onChange={(event) => setProductId(event.target.value)} disabled={productsLoading}>
              {products.length === 0 ? <option value="">No products available</option> : null}
              {products.map((product) => (
                <option key={product.id} value={product.id}>
                  {product.name}
                </option>
              ))}
            </select>
          </label>
        ) : (
          <div className="inline-product">
            <div className="field-row">
              <label className="field">
                <span>Product name</span>
                <input
                  value={inlineProduct.name}
                  onChange={(event) =>
                    setInlineProduct((current) => ({ ...current, name: event.target.value }))
                  }
                />
              </label>
              <label className="field">
                <span>Product category</span>
                <input
                  value={inlineProduct.category}
                  onChange={(event) =>
                    setInlineProduct((current) => ({ ...current, category: event.target.value }))
                  }
                />
              </label>
            </div>

            <label className="field">
              <span>Product description</span>
              <textarea
                rows={3}
                value={inlineProduct.description}
                onChange={(event) =>
                  setInlineProduct((current) => ({ ...current, description: event.target.value }))
                }
              />
            </label>

            <div className="field-row">
              <label className="field">
                <span>Context keywords</span>
                <input
                  value={inlineProduct.contextKeywords}
                  onChange={(event) =>
                    setInlineProduct((current) => ({ ...current, contextKeywords: event.target.value }))
                  }
                  placeholder="drink, kitchen, refreshment"
                />
              </label>
              <label className="field">
                <span>Source URL</span>
                <input
                  type="url"
                  value={inlineProduct.sourceUrl}
                  onChange={(event) =>
                    setInlineProduct((current) => ({ ...current, sourceUrl: event.target.value }))
                  }
                />
              </label>
            </div>

            <label className="field">
              <span>Product image</span>
              <input
                type="file"
                accept=".png,.jpg,.jpeg,image/png,image/jpeg"
                onChange={(event) =>
                  setInlineProduct((current) => ({ ...current, imageFile: event.target.files?.[0] ?? null }))
                }
              />
            </label>
          </div>
        )}

        <div className="form-actions">
          <button type="submit" disabled={saving}>
            {saving ? "Creating campaign..." : "Create campaign"}
          </button>
        </div>
      </form>

      {productsError ? <p className="form-message form-message--error">{productsError}</p> : null}
      {campaignError ? <p className="form-message form-message--error">{campaignError}</p> : null}

      {success ? (
        <section className="success-summary">
          <p className="eyebrow">Campaign created</p>
          <h3>{success.name}</h3>
          <p>
            Job <strong>{success.job_id}</strong> is <strong>{success.status}</strong> and waiting at{" "}
            <strong>{success.current_stage}</strong>.
          </p>
          <p className="muted">Analysis has not started yet.</p>
          <p className="muted">
            Video: {success.video_filename} • FPS {success.source_fps ?? "n/a"} • Duration{" "}
            {success.duration_seconds?.toFixed(2) ?? "n/a"}s
          </p>
        </section>
      ) : null}
    </section>
  );
}
