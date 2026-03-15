import { Navigate, NavLink, Route, Routes } from "react-router-dom";
import { BrandMark } from "./components/BrandMark";
import { HeartIcon, PlayIcon, SparkleIcon, UploadIcon, UsersIcon } from "./components/PinkIcons";
import { runtimeConfig } from "./config/runtime";
import { publicCopy } from "./content/publicCopy";
import { AboutPage } from "./pages/AboutPage";
import { CreateCampaignPage } from "./pages/CreateCampaignPage";
import { HomePage } from "./pages/HomePage";
import { JobPage } from "./pages/JobPage";
import { PreviewPage } from "./pages/PreviewPage";
import { ProductsPage } from "./pages/ProductsPage";
import { ResultsPage } from "./pages/ResultsPage";
import { UploadPage } from "./pages/UploadPage";
import { WebsiteAdsPage } from "./pages/WebsiteAdsPage";

function App() {
  const navClassName = ({ isActive }: { isActive: boolean }) => (isActive ? "active" : undefined);
  const showcaseMode = runtimeConfig.showcaseMode;

  return (
    <div className="app-shell app-shell--voxel">
      <header className="app-header app-header--voxel">
        <NavLink className="brand-lockup brand-lockup--voxel" to="/" aria-label={`${publicCopy.brand.name} home`}>
          <BrandMark />
        </NavLink>

        <nav className="tab-nav" aria-label="Primary">
          <NavLink className={navClassName} to="/" end>
            <HeartIcon className="tab-nav__icon" />
            {publicCopy.nav.home}
          </NavLink>
          <NavLink className={navClassName} to="/gallery">
            <PlayIcon className="tab-nav__icon" />
            {publicCopy.nav.gallery}
          </NavLink>
          <NavLink className={navClassName} to="/website-ads">
            <SparkleIcon className="tab-nav__icon" />
            {publicCopy.nav.websiteAds}
          </NavLink>
          <NavLink className={navClassName} to="/upload">
            <UploadIcon className="tab-nav__icon" />
            {publicCopy.nav.upload}
          </NavLink>
          <NavLink className={navClassName} to="/about">
            <UsersIcon className="tab-nav__icon" />
            {publicCopy.nav.about}
          </NavLink>
        </nav>

        <div className="header-actions">
          <span className="header-badge">
            <HeartIcon className="tab-nav__icon" />
            {publicCopy.brand.badge}
          </span>
        </div>
      </header>

      <main className="app-main app-main--voxel">
        <Routes>
          <Route path="/" element={<HomePage />} />
          <Route path="/gallery" element={<ResultsPage />} />
          <Route path="/results" element={<ResultsPage />} />
          <Route path="/website-ads" element={<WebsiteAdsPage />} />
          <Route path="/upload" element={<UploadPage />} />
          <Route path="/about" element={<AboutPage />} />
          {showcaseMode ? null : (
            <>
              <Route path="/products" element={<ProductsPage />} />
              <Route path="/campaigns/new" element={<CreateCampaignPage />} />
              <Route path="/jobs/:jobId/preview" element={<PreviewPage />} />
              <Route path="/jobs/:jobId" element={<JobPage />} />
            </>
          )}
          <Route path="*" element={<Navigate to="/" replace />} />
        </Routes>
      </main>
    </div>
  );
}

export default App;
