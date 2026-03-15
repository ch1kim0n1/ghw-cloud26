import { HeartIcon, SparkleIcon, UsersIcon } from "../components/PinkIcons";

const placeholderDevelopers = [
  {
    name: "Developer One",
    role: "Frontend + product vibes",
    bio: "Add your intro here later. This card is ready for a photo, links, and a real bio when you have them.",
    initials: "D1",
  },
  {
    name: "Developer Two",
    role: "ML + pipeline magic",
    bio: "Add your intro here later. This spot is ready for another teammate profile without needing a layout rewrite.",
    initials: "D2",
  },
];

export function AboutPage() {
  return (
    <div className="about-page">
      <section className="about-card">
        <span className="showcase-pill showcase-pill--pink">
          <UsersIcon className="inline-icon" />
          About us
        </span>
        <h1>Two builders, one cute little ad-insertion experiment.</h1>
        <p>
          This page is just a placeholder for now, but the layout is ready for two developer profiles, a short team
          story, and links when you want to fill it in properly.
        </p>

        <div className="about-stats">
          <span>
            <SparkleIcon className="inline-icon" />
            Demo-first frontend
          </span>
          <span>
            <HeartIcon className="inline-icon" />
            Placeholder team cards
          </span>
        </div>
      </section>

      <section className="team-grid" aria-label="Developer placeholders">
        {placeholderDevelopers.map((developer) => (
          <article className="developer-card" key={developer.name}>
            <div className="developer-card__avatar" aria-hidden="true">
              <span>{developer.initials}</span>
            </div>
            <div className="developer-card__copy">
              <h2>{developer.name}</h2>
              <p className="developer-card__role">{developer.role}</p>
              <p>{developer.bio}</p>
            </div>
          </article>
        ))}
      </section>

      <section className="pixel-pack-card">
        <div className="pixel-pack-card__copy">
          <span className="showcase-pill">
            <SparkleIcon className="inline-icon" />
            Pixel pack
          </span>
          <h2>Little pixel hearts are part of the public look now.</h2>
          <p>
            The sticker art comes from the free Pixel Hearts pack by TokyoGeisha on OpenGameArt, released under CC0.
          </p>
          <a
            className="cute-link"
            href="https://opengameart.org/content/pixel-hearts"
            target="_blank"
            rel="noreferrer"
          >
            View asset source
          </a>
        </div>

        <div className="pixel-pack-card__art">
          <img src="/pixel-hearts/pixel-hearts-preview.png" alt="Pixel hearts asset pack preview" />
          <img src="/pixel-hearts/sprite-sheet.png" alt="Pixel hearts sprite sheet" />
        </div>
      </section>
    </div>
  );
}
