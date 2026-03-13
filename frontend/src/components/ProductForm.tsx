import { FormEvent, useEffect, useState } from "react";
import { createProduct, listProducts } from "../services/productsApi";
import { ApiError } from "../types/Api";
import type { Product } from "../types/Product";

const initialFormState = {
  name: "",
  description: "",
  category: "",
  contextKeywords: "",
  sourceUrl: "",
  imageFile: null as File | null,
};

export function ProductForm() {
  const [products, setProducts] = useState<Product[]>([]);
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [success, setSuccess] = useState<string | null>(null);
  const [formState, setFormState] = useState(initialFormState);

  useEffect(() => {
    void loadProducts();
  }, []);

  async function loadProducts() {
    setLoading(true);
    try {
      const response = await listProducts();
      setProducts(Array.isArray(response.products) ? response.products : []);
      setError(null);
    } catch (reason) {
      setError(reason instanceof ApiError ? reason.message : "Unable to load products.");
    } finally {
      setLoading(false);
    }
  }

  async function handleSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    setSaving(true);
    setError(null);
    setSuccess(null);

    const formData = new FormData();
    formData.set("name", formState.name);
    if (formState.description.trim()) formData.set("description", formState.description);
    if (formState.category.trim()) formData.set("category", formState.category);
    if (formState.contextKeywords.trim()) formData.set("context_keywords", formState.contextKeywords);
    if (formState.sourceUrl.trim()) formData.set("source_url", formState.sourceUrl);
    if (formState.imageFile) formData.set("image_file", formState.imageFile);

    try {
      const product = await createProduct(formData);
      setProducts((current) => [product, ...current]);
      setFormState(initialFormState);
      setSuccess(`Created ${product.name}.`);
    } catch (reason) {
      setError(reason instanceof ApiError ? reason.message : "Unable to create product.");
    } finally {
      setSaving(false);
    }
  }

  return (
    <section className="panel">
      <p className="eyebrow">Phase 1 live</p>
      <h2>Product Catalog</h2>
      <p>Create reusable products with source metadata and an optional product image.</p>

      <form className="form-grid" onSubmit={handleSubmit}>
        <label className="field">
          <span>Name</span>
          <input
            name="name"
            value={formState.name}
            onChange={(event) => setFormState((current) => ({ ...current, name: event.target.value }))}
            required
          />
        </label>

        <label className="field">
          <span>Description</span>
          <textarea
            name="description"
            value={formState.description}
            onChange={(event) => setFormState((current) => ({ ...current, description: event.target.value }))}
            rows={3}
          />
        </label>

        <div className="field-row">
          <label className="field">
            <span>Category</span>
            <input
              name="category"
              value={formState.category}
              onChange={(event) => setFormState((current) => ({ ...current, category: event.target.value }))}
            />
          </label>

          <label className="field">
            <span>Context keywords</span>
            <input
              name="context_keywords"
              value={formState.contextKeywords}
              onChange={(event) =>
                setFormState((current) => ({ ...current, contextKeywords: event.target.value }))
              }
              placeholder="drink, water, refreshment"
            />
          </label>
        </div>

        <label className="field">
          <span>Source URL</span>
          <input
            name="source_url"
            type="url"
            value={formState.sourceUrl}
            onChange={(event) => setFormState((current) => ({ ...current, sourceUrl: event.target.value }))}
            placeholder="https://example.com/product"
          />
        </label>

        <label className="field">
          <span>Image file (PNG or JPG)</span>
          <input
            name="image_file"
            type="file"
            accept=".png,.jpg,.jpeg,image/png,image/jpeg"
            onChange={(event) =>
              setFormState((current) => ({ ...current, imageFile: event.target.files?.[0] ?? null }))
            }
          />
        </label>

        <div className="form-actions">
          <button type="submit" disabled={saving}>
            {saving ? "Creating product..." : "Create product"}
          </button>
        </div>
      </form>

      {error ? <p className="form-message form-message--error">{error}</p> : null}
      {success ? <p className="form-message form-message--success">{success}</p> : null}

      <div className="list-block">
        <div className="list-block__header">
          <h3>Saved products</h3>
          <button type="button" className="button-secondary" onClick={() => void loadProducts()} disabled={loading}>
            Refresh
          </button>
        </div>

        {loading ? <p>Loading products...</p> : null}
        {!loading && products.length === 0 ? <p>No products yet.</p> : null}

        <div className="card-grid">
          {products.map((product) => (
            <article key={product.id} className="card">
              <p className="eyebrow">{product.category || "Uncategorized"}</p>
              <h3>{product.name}</h3>
              <p>{product.description || "No description provided."}</p>
              <p className="muted">Keywords: {product.context_keywords?.join(", ") || "None"}</p>
              <p className="muted">Source: {product.source_url || "Uploaded image only"}</p>
            </article>
          ))}
        </div>
      </div>
    </section>
  );
}
