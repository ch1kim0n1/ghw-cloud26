import { ProductForm } from "../components/ProductForm";

export function ProductsPage() {
  return (
    <div className="page-grid">
      <ProductForm />
      <section className="panel">
        <p className="eyebrow">Current state</p>
        <h2>No products yet</h2>
        <p>
          Product listing is intentionally empty in Phase 0. The backend route
          exists and returns a documented placeholder error until Phase 1
          implements catalog ingest.
        </p>
      </section>
    </div>
  );
}
