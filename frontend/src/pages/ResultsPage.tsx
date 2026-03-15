import { ShowcaseResult } from "../components/ShowcaseResult";
import { WebsiteAdsShowcase } from "../components/WebsiteAdsShowcase";
import { publicCopy } from "../content/publicCopy";

export function ResultsPage() {
  return (
    <div className="public-page public-page--hidden-route">
      <section className="gallery-page-intro voxel-panel">
        <div className="section-heading section-heading--voxel">
          <p className="eyebrow">{publicCopy.landing.galleryEyebrow}</p>
          <h1>Gallery of processed videos</h1>
          <p>Every processed demo clip lives here, so the home page can stay focused while the full library stays easy to browse.</p>
        </div>
      </section>
      <ShowcaseResult />
      <WebsiteAdsShowcase
        eyebrow="Static ad gallery"
        title="Website ad previews live next to the video examples"
        lede="The gallery now carries both proof types: stitched video outputs and static website-ad placements captured on real pages."
      />
    </div>
  );
}
