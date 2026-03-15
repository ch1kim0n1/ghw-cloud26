import { Link } from "react-router-dom";
import { demoExamples } from "../content/demoContent";
import { HeartIcon, PlayIcon, SparkleIcon } from "./PinkIcons";

interface ShowcaseResultProps {
  compact?: boolean;
}

export function ShowcaseResult({ compact = false }: ShowcaseResultProps) {
  return (
    <section className={`showcase-card${compact ? " showcase-card--compact" : ""}`}>
      <div className="showcase-card__copy">
        <span className="showcase-pill">Showcase result</span>
        <h1>Ad insertion, but make it cute, seamless, and actually watchable.</h1>
        <p className="showcase-card__lede">
          Every available demo example is here, so you can show both finished cuts without digging through extra pages
          or hidden controls.
        </p>

        <div className="showcase-badges">
          <span>{demoExamples.length} polished examples</span>
          <span>Scene-aware insert windows</span>
          <span>Anchor-frame matched</span>
        </div>

        {!compact ? (
          <div className="showcase-card__note">
            <strong>
              <SparkleIcon className="inline-icon" />
              Why it works
            </strong>
            <p>
              Each cut keeps the brand moment inside the scene rhythm, so the result feels more like part of the story
              and less like an obvious interruption.
            </p>
          </div>
        ) : null}

        <div className="pixel-sticker-card">
          <div className="pixel-sticker-card__copy">
            <span className="pixel-label">
              <HeartIcon className="inline-icon" />
              pixel stickers
            </span>
            <p>Little pixel hearts keep the public demo playful without adding more UX noise.</p>
          </div>
          <img src="/pixel-hearts/pixel-hearts-preview.png" alt="Pixel hearts sticker pack preview" />
        </div>

        <Link className="cute-link" to="/upload">
          Upload your own video
        </Link>
      </div>

      <div className="showcase-gallery" aria-label="Available showcase examples">
        {demoExamples.map((example) => (
          <article className="showcase-example-card" key={example.id}>
            <div className="showcase-example-card__header">
              <span className="showcase-pill showcase-pill--pink">{example.label}</span>
              <div>
                <h2>
                  <PlayIcon className="inline-icon" />
                  {example.title}
                </h2>
                <p>{example.summary}</p>
              </div>
            </div>

            <div className="showcase-badges">
              <span>{example.scene}</span>
              <span>{example.selectedWindow}</span>
              <span>{example.anchorFrames}</span>
            </div>

            <div className="showcase-video-frame">
              <video controls playsInline poster={example.finalPoster}>
                <source src={example.finalVideo} type="video/mp4" />
              </video>
              <div className="showcase-video-frame__tag">{example.label}</div>
            </div>

            <div className="showcase-mini-strip">
              <img src={example.generatedPreview} alt={`${example.label} generated bridge clip preview`} />
              <div className="showcase-mini-strip__frames">
                <img src={example.startFrame} alt={`${example.label} start anchor frame`} />
                <img src={example.stopFrame} alt={`${example.label} stop anchor frame`} />
              </div>
            </div>
          </article>
        ))}
      </div>
    </section>
  );
}
