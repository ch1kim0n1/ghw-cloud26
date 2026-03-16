import { AnimatePresence, LayoutGroup, motion, useReducedMotion } from "framer-motion";
import { useEffect, useState } from "react";
import { Navigate, NavLink, Route, Routes, useLocation } from "react-router-dom";
import { BrandMark } from "./components/BrandMark";
import { HeartIcon, PlayIcon, SparkleIcon, UploadIcon, UsersIcon } from "./components/PinkIcons";
import { pageShellVariants, publicLayoutTransition, publicQuickTransition, publicRoutePaths } from "./components/publicMotion";
import { runtimeConfig } from "./config/runtime";
import { publicCopy } from "./content/publicCopy";
import { AboutPage } from "./pages/AboutPage";
import { CreateCampaignPage } from "./pages/CreateCampaignPage";
import { HomePage } from "./pages/HomePage";
import { JobPage } from "./pages/JobPage";
import { PreviewPage } from "./pages/PreviewPage";
import { ProductsPage } from "./pages/ProductsPage";
import { ResultsPage } from "./pages/ResultsPage";
import { StudioPage } from "./pages/StudioPage";
import { UploadPage } from "./pages/UploadPage";
import { WebsiteAdsPage } from "./pages/WebsiteAdsPage";

function App() {
  const location = useLocation();
  const reducedMotion = useReducedMotion();
  const showcaseMode = runtimeConfig.showcaseMode;
  const [isScrolled, setIsScrolled] = useState(false);
  const isPublicRoute = publicRoutePaths.has(location.pathname);

  useEffect(() => {
    const updateScrolled = () => {
      setIsScrolled(window.scrollY > 28);
    };

    updateScrolled();
    window.addEventListener("scroll", updateScrolled, { passive: true });

    return () => {
      window.removeEventListener("scroll", updateScrolled);
    };
  }, [location.pathname]);

  const navItems: Array<{ to: string; label: string; icon: typeof HeartIcon; end?: boolean }> = [
    { to: "/", label: publicCopy.nav.home, icon: HeartIcon, end: true },
    { to: "/upload", label: publicCopy.nav.upload, icon: UploadIcon },
    { to: "/studio", label: publicCopy.nav.studio, icon: SparkleIcon },
    { to: "/gallery", label: publicCopy.nav.gallery, icon: PlayIcon },
    { to: "/about", label: publicCopy.nav.about, icon: UsersIcon },
  ];

  return (
    <div className="app-shell app-shell--voxel">
      <header className={`app-header app-header--voxel${isScrolled ? " app-header--scrolled" : ""}`}>
        <NavLink className="brand-lockup brand-lockup--voxel" to="/" aria-label={`${publicCopy.brand.name} home`}>
          <BrandMark />
        </NavLink>

        <LayoutGroup id="public-nav">
          <nav className="tab-nav" aria-label="Primary">
            {navItems.map((item) => (
              <NavLink
                key={item.to}
                className={({ isActive }) => `tab-nav__link${isActive ? " active" : ""}`}
                to={item.to}
                end={item.end}
              >
                {({ isActive }) => (
                  <>
                    {isActive ? (
                      <motion.span
                        layoutId="tab-nav-indicator"
                        className="tab-nav__indicator"
                        transition={publicLayoutTransition}
                      />
                    ) : null}
                    <item.icon className="tab-nav__icon" />
                    <span>{item.label}</span>
                  </>
                )}
              </NavLink>
            ))}
          </nav>
        </LayoutGroup>

        <div className="header-actions">
          <span className="header-badge">
            <HeartIcon className="tab-nav__icon" />
            {publicCopy.brand.badge}
          </span>
        </div>
      </header>

      <main className="app-main app-main--voxel">
        <AnimatePresence mode={reducedMotion ? "sync" : "wait"} initial={false}>
          <motion.div
            key={isPublicRoute ? location.pathname : "studio-shell"}
            className="app-route-shell"
            initial={reducedMotion ? false : "hidden"}
            animate="show"
            exit={reducedMotion ? undefined : "exit"}
            variants={pageShellVariants}
            transition={publicQuickTransition}
          >
            <Routes location={location}>
              <Route path="/" element={<HomePage />} />
              <Route path="/studio" element={<StudioPage />} />
              <Route path="/gallery" element={<ResultsPage />} />
              <Route path="/results" element={<Navigate to="/gallery" replace />} />
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
          </motion.div>
        </AnimatePresence>
      </main>
    </div>
  );
}

export default App;
