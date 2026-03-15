import { ShowcaseResult } from "../components/ShowcaseResult";
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
    </div>
  );
}
