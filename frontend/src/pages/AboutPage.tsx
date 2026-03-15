import { publicCopy } from "../content/publicCopy";

export function AboutPage() {
  return (
    <div className="about-page about-page--profiles">
      <section className="about-hero voxel-panel">
        <span className="voxel-chip">{publicCopy.about.eyebrow}</span>
        <h1>{publicCopy.about.title}</h1>
        <p>{publicCopy.about.lede}</p>
      </section>

      <section className="team-grid" aria-label="CAFAI developers">
        {publicCopy.about.cards.map((card) => (
          <a
            className="founder-card founder-card--link voxel-panel"
            href={card.github}
            key={card.name}
            target="_blank"
            rel="noreferrer"
            aria-label={`${card.name} GitHub profile`}
          >
            <div className="founder-card__avatar-frame">
              <img className="founder-card__avatar" src={card.avatar} alt={`${card.name} profile meme`} />
            </div>
            <div className="founder-card__copy">
              <h2>{card.name}</h2>
              <p className="founder-card__role">{card.role}</p>
              <p>{card.bio}</p>
              <span className="founder-card__github">{card.githubLabel}</span>
            </div>
          </a>
        ))}
      </section>
    </div>
  );
}
