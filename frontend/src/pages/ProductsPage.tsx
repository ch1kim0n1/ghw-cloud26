import { ProductForm } from "../components/ProductForm";
import { Reveal } from "../components/Reveal";

export function ProductsPage() {
  return (
    <div className="studio-page">
      <Reveal as="section" className="studio-hero">
        <p className="eyebrow">Studio catalog</p>
        <h1>Maintain the product library behind the demo surface.</h1>
        <p>
          The catalog remains available for operators, but it now sits inside a calmer studio shell instead of
          competing with the presentation pages.
        </p>
      </Reveal>

      <div className="page-grid page-grid--studio">
        <ProductForm />
        <section className="panel panel--supporting">
          <p className="eyebrow">Operator note</p>
          <h2>Reusable product ingest is live</h2>
          <p>
            Products can be created with metadata plus either a retailer URL or an uploaded PNG/JPG. The catalog stays
            local in SQLite and uploaded assets remain on disk for MVP debugging and repeatable demos.
          </p>
        </section>
      </div>
    </div>
  );
}
