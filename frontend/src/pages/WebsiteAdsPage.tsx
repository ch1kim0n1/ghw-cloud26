import { FloatingDecor } from "../components/FloatingDecor";
import { Reveal } from "../components/Reveal";
import { WebsiteAdsShowcase } from "../components/WebsiteAdsShowcase";
import { publicCopy } from "../content/publicCopy";

export function WebsiteAdsPage() {
  return (
    <div className="studio-page">
      <Reveal as="section" className="website-ads-hero voxel-panel">
        <FloatingDecor ids={["bow", "cloud", "star"]} variant="about" />
        <div className="section-heading section-heading--voxel">
          <p className="eyebrow">{publicCopy.websiteAds.eyebrow}</p>
          <h1>{publicCopy.websiteAds.title}</h1>
          <p>{publicCopy.websiteAds.lede}</p>
        </div>

        <div className="status-strip">
          <span>experimental side lane</span>
          <span>3 injected screenshots</span>
          <span>real source pages</span>
          <span>not the core CAFAI MVP</span>
        </div>
      </Reveal>

      <WebsiteAdsShowcase
        eyebrow="Proof gallery"
        title="Injected ad placements on captured pages"
        lede="Each example below uses the real captured page image and a flattened preview with the horizontal and vertical ads already injected into the screenshot."
      />
    </div>
  );
}
