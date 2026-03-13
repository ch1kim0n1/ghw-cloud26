import { Link, Navigate, Route, Routes } from "react-router-dom";
import { HealthStatusPanel } from "./components/HealthStatusPanel";
import { CreateCampaignPage } from "./pages/CreateCampaignPage";
import { JobPage } from "./pages/JobPage";
import { ProductsPage } from "./pages/ProductsPage";

function App() {
  return (
    <div className="app-shell">
      <header className="hero">
        <div>
          <p className="eyebrow">CAFAI phase 3</p>
          <h1>Cloud-assisted ad insertion dashboard</h1>
          <p className="hero-copy">
            Product ingest, campaign intake, explicit analysis start, slot
            review, product line review, and CAFAI generation are live. Real
            runs require Azure Video Indexer, Azure OpenAI, and Azure Machine
            Learning configuration.
          </p>
        </div>
        <HealthStatusPanel />
      </header>

      <nav className="nav-bar" aria-label="Primary">
        <Link to="/products">Products</Link>
        <Link to="/campaigns/new">Create Campaign</Link>
        <Link to="/jobs/demo-job">Job</Link>
      </nav>

      <main>
        <Routes>
          <Route path="/" element={<Navigate to="/products" replace />} />
          <Route path="/products" element={<ProductsPage />} />
          <Route path="/campaigns/new" element={<CreateCampaignPage />} />
          <Route path="/jobs/:jobId" element={<JobPage />} />
        </Routes>
      </main>
    </div>
  );
}

export default App;
