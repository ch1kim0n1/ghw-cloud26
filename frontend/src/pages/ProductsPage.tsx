import { ProductForm } from "../components/ProductForm";

export function ProductsPage() {
  return (
    <div className="page-grid">
      <ProductForm />
      <section className="panel">
        <p className="eyebrow">Phase 1 contract</p>
        <h2>Reusable product ingest is live</h2>
        <p>
          Products can be created with metadata plus either a retailer URL or an uploaded PNG/JPG. The catalog stays local in SQLite and the uploaded assets are stored on disk for MVP debugging.
        </p>
      </section>
    </div>
  );
}
