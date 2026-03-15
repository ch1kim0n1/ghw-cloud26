import { NavLink, Route, Routes } from "react-router-dom";
import { BrandMark } from "./components/BrandMark";
import { HeartIcon, PlayIcon, UploadIcon, UsersIcon } from "./components/PinkIcons";
import { AboutPage } from "./pages/AboutPage";
import { CreateCampaignPage } from "./pages/CreateCampaignPage";
import { HomePage } from "./pages/HomePage";
import { JobPage } from "./pages/JobPage";
import { PreviewPage } from "./pages/PreviewPage";
import { ProductsPage } from "./pages/ProductsPage";
import { ResultsPage } from "./pages/ResultsPage";
import { UploadPage } from "./pages/UploadPage";

function App() {
  const navClassName = ({ isActive }: { isActive: boolean }) => (isActive ? "active" : undefined);

  return (
    <div className="app-shell app-shell--cute">
      <header className="app-header app-header--cute">
        <NavLink className="brand-lockup brand-lockup--cute" to="/">
          <BrandMark />
          <span className="brand-lockup__text">
            <strong>PinkFrame</strong>
            <small>scene-aware ad magic</small>
          </span>
        </NavLink>

        <nav className="tab-nav" aria-label="Primary">
          <NavLink className={navClassName} to="/" end>
            <PlayIcon className="tab-nav__icon" />
            Showcase
          </NavLink>
          <NavLink className={navClassName} to="/upload">
            <UploadIcon className="tab-nav__icon" />
            Upload
          </NavLink>
        </nav>

        <div className="header-actions">
          <NavLink className={navClassName} to="/about">
            <UsersIcon className="tab-nav__icon" />
            About us
          </NavLink>
          <span className="header-badge">
            <HeartIcon className="tab-nav__icon" />
            girl-coded
          </span>
        </div>
      </header>

      <main className="app-main app-main--cute">
        <Routes>
          <Route path="/" element={<HomePage />} />
          <Route path="/upload" element={<UploadPage />} />
          <Route path="/about" element={<AboutPage />} />
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
