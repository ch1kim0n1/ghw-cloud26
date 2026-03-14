import { Link, Route, Routes } from "react-router-dom";
import { CreateCampaignPage } from "./pages/CreateCampaignPage";
import { HomePage } from "./pages/HomePage";
import { JobPage } from "./pages/JobPage";
import { PreviewPage } from "./pages/PreviewPage";
import { ProductsPage } from "./pages/ProductsPage";
import { ResultsPage } from "./pages/ResultsPage";

function App() {
  return (
    <div className="app-shell">
      <header className="top-bar">
        <Link className="top-bar__brand" to="/">
          <span className="top-bar__brand-mark" />
          <span>CAFAI</span>
        </Link>
        <nav className="top-bar__nav" aria-label="Primary">
          <Link to="/results">Results</Link>
          <Link to="/jobs/job_9de1cbb7-ec84-4e2c-99f7-0d2dc6f21e0a">Demo Run</Link>
          <Link to="/products">Dashboard</Link>
        </nav>
      </header>

      <main className="app-main">
        <Routes>
          <Route path="/" element={<HomePage />} />
          <Route path="/results" element={<ResultsPage />} />
          <Route path="/products" element={<ProductsPage />} />
          <Route path="/campaigns/new" element={<CreateCampaignPage />} />
          <Route path="/jobs/:jobId/preview" element={<PreviewPage />} />
          <Route path="/jobs/:jobId" element={<JobPage />} />
        </Routes>
      </main>
    </div>
  );
}

export default App;
