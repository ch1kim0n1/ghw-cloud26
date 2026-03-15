import { Reveal } from "../components/Reveal";
import { websiteAdsContent } from "../content/websiteAdsContent";
import { publicCopy } from "../content/publicCopy";

export function WebsiteAdsPage() {
  return (
    <div className="studio-page">
      <Reveal as="section" className="website-ads-hero voxel-panel">
        <div className="section-heading section-heading--voxel">
          <p className="eyebrow">{publicCopy.websiteAds.eyebrow}</p>
          <h1>{publicCopy.websiteAds.title}</h1>
          <p>{publicCopy.websiteAds.lede}</p>
        </div>

        <div className="status-strip">
          <span>3 injected screenshots</span>
          <span>real source pages</span>
          <span>no mock article shells</span>
        </div>
      </Reveal>

      <section className="voxel-panel website-ads-showcase">
        <div className="section-heading section-heading--voxel">
          <p className="eyebrow">Proof gallery</p>
          <h2>Injected ad placements on captured pages</h2>
          <p>
            Each example below uses the real captured page image and a flattened preview with the horizontal and
            vertical ads already injected into the screenshot.
          </p>
        </div>

        <div className="website-real-grid">
          {websiteAdsContent.examples.map((example) => (
            <article className="website-real-card" key={example.id}>
              <div className="website-real-card__header">
                <div>
                  <p className="eyebrow">{example.label}</p>
                  <h3>{example.title}</h3>
                </div>
                <a href={example.url} target="_blank" rel="noreferrer">
                  Open source
                </a>
              </div>

              <p className="website-real-card__note">{example.note}</p>

              <div className="website-real-card__preview">
                <img src={example.previewImage} alt={`${example.title} site preview with injected ads`} />
              </div>
            </article>
          ))}
        </div>
      </section>
    </div>
  );
}
