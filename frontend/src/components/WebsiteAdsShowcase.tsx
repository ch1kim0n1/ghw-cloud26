import { websiteAdsContent } from "../content/websiteAdsContent";

interface WebsiteAdsShowcaseProps {
  eyebrow?: string;
  title?: string;
  lede?: string;
}

export function WebsiteAdsShowcase({
  eyebrow = "Static ads proof",
  title = "Injected website ad placements on real captured pages",
  lede = "These examples show the static-ad side of the product: banner and vertical units composited onto real page captures so the site can demonstrate both channels together.",
}: WebsiteAdsShowcaseProps) {
  return (
    <section className="voxel-panel website-ads-showcase">
      <div className="section-heading section-heading--voxel">
        <p className="eyebrow">{eyebrow}</p>
        <h2>{title}</h2>
        <p>{lede}</p>
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
  );
}
