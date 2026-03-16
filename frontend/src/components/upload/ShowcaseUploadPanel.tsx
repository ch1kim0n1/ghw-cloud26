import { motion, useReducedMotion } from "framer-motion";
import { Link } from "react-router-dom";
import { HeartIcon } from "../PinkIcons";
import { publicLayoutTransition } from "../publicMotion";
import { websiteAdsContent } from "../../content/websiteAdsContent";

interface ShowcaseUploadPanelProps {
  uploadMode: "video" | "website";
}

export function ShowcaseUploadPanel({ uploadMode }: ShowcaseUploadPanelProps) {
  const showcaseExamples = websiteAdsContent.examples;
  const reducedMotion = useReducedMotion();

  return (
    <div className="upload-showcase-card">
      <div className="upload-showcase-card__intro">
        <span className="status-pill status-pill--progress">showcase build</span>
        <h2>{uploadMode === "video" ? "Video pipeline preview" : "Experimental website-ad preview"}</h2>
        <p>
          This GitHub Pages deployment is a static showcase. The live generation backend is disabled here, so this page explains the
          inputs and points to completed examples instead of submitting real jobs.
        </p>
      </div>

      {uploadMode === "video" ? (
        <div className="upload-showcase-grid">
          <article className="upload-showcase-block">
            <h3>What the live video flow asks for</h3>
            <ul>
              <li>Campaign name</li>
              <li>Brand or product name</li>
              <li>One MP4 source clip</li>
            </ul>
          </article>

          <article className="upload-showcase-block">
            <h3>What happens in the real backend</h3>
            <ul>
              <li>Scene analysis and candidate slot detection</li>
              <li>Operator review or manual override</li>
              <li>Bridge generation and final preview stitching</li>
            </ul>
          </article>
        </div>
      ) : (
        <div className="upload-showcase-grid">
          <article className="upload-showcase-block">
            <h3>What the live website-ad flow asks for</h3>
            <ul>
              <li>Saved product or inline product info</li>
              <li>Article headline and article context</li>
              <li>Visual direction for the creative</li>
            </ul>
          </article>

          <article className="upload-showcase-block">
            <h3>What the backend generates</h3>
            <ul>
              <li>One horizontal banner at 1200x628</li>
              <li>One vertical sidebar ad at 300x600</li>
              <li>Stored assets and reviewable gallery output</li>
            </ul>
          </article>
        </div>
      )}

      {uploadMode === "website" ? (
        <div className="upload-showcase-examples">
          {showcaseExamples.map((example) => (
            <motion.article
              className="upload-showcase-example"
              key={example.id}
              whileHover={reducedMotion ? undefined : { y: -4 }}
              transition={publicLayoutTransition}
            >
              <img src={example.previewImage} alt={`${example.title} injected placement preview`} />
              <div>
                <strong>
                  {example.label}: {example.title}
                </strong>
                <p>{example.note}</p>
              </div>
            </motion.article>
          ))}
        </div>
      ) : null}

      <div className="upload-form__actions">
        <Link className="cute-button" to={uploadMode === "video" ? "/gallery" : "/website-ads"}>
          {uploadMode === "video" ? "Open video gallery" : "Open website ads gallery"}
        </Link>
        <Link className="cute-button cute-button--secondary" to="/">
          <HeartIcon className="inline-icon" />
          Back to home
        </Link>
      </div>
    </div>
  );
}
