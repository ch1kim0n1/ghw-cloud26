import { Link, Navigate, Route, Routes } from "react-router-dom";
import { HealthStatusPanel } from "./components/HealthStatusPanel";
import { CreateCampaignPage } from "./pages/CreateCampaignPage";
import { JobPage } from "./pages/JobPage";
import { PreviewPage } from "./pages/PreviewPage";
import { ProductsPage } from "./pages/ProductsPage";

function App() {
  return (
    <div className="app-shell">
      <header className="hero">
        <div>
          <p className="eyebrow">CAFAI foundation</p>
          <h1>Cloud-assisted ad insertion dashboard</h1>
          <p className="hero-copy">
            Phase 0 keeps the full operator workflow visible while only the
            health endpoint is live. The rest of the surface is scaffolded for
            Phase 1+ implementation.
          </p>
        </div>
        <HealthStatusPanel />
      </header>

      <nav className="nav-bar" aria-label="Primary">
        <Link to="/products">Products</Link>
        <Link to="/campaigns/new">Create Campaign</Link>
        <Link to="/jobs/demo-job">Job</Link>
        <Link to="/preview/demo-job">Preview</Link>
      </nav>

      <main>
        <Routes>
          <Route path="/" element={<Navigate to="/products" replace />} />
          <Route path="/products" element={<ProductsPage />} />
          <Route path="/campaigns/new" element={<CreateCampaignPage />} />
          <Route path="/jobs/:jobId" element={<JobPage />} />
          <Route path="/preview/:jobId" element={<PreviewPage />} />
        </Routes>
      </main>
    </div>
  );
}

export default App;
