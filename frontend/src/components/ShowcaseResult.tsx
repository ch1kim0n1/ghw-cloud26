import { AnimatePresence, LayoutGroup, motion, useReducedMotion } from "framer-motion";
import { CSSProperties, startTransition, useMemo, useState } from "react";
import { contentSwapVariants, publicLayoutTransition, publicSwapTransition } from "./publicMotion";
import { demoExamples, featuredDemoExample } from "../content/demoContent";
import { publicCopy } from "../content/publicCopy";
import { HeartIcon, PlayIcon, SparkleIcon } from "./PinkIcons";

interface ShowcaseResultProps {
  compact?: boolean;
}

export function ShowcaseResult({ compact = false }: ShowcaseResultProps) {
  const [activeExampleId, setActiveExampleId] = useState(featuredDemoExample.id);
  const reducedMotion = useReducedMotion();
  const activeExample = useMemo(
    () => demoExamples.find((example) => example.id === activeExampleId) ?? featuredDemoExample,
    [activeExampleId],
  );

  const paletteStyle = {
    "--example-accent": activeExample.palette.accent,
    "--example-panel": activeExample.palette.panel,
    "--example-border": activeExample.palette.border,
    "--example-shadow": activeExample.palette.shadow,
    "--example-grass": activeExample.palette.grass,
    "--example-sky": activeExample.palette.sky,
  } as CSSProperties;

  return (
    <section className={`showcase-vault voxel-panel${compact ? " showcase-vault--compact" : ""}`}>
      <div className="section-heading section-heading--voxel">
        <p className="eyebrow">{publicCopy.landing.galleryEyebrow}</p>
        <h2>{publicCopy.landing.galleryTitle}</h2>
        <p>{compact ? "The hidden route still keeps the examples one click away." : publicCopy.landing.galleryLede}</p>
      </div>

      <LayoutGroup id="showcase-tabs">
        <div className="showcase-tabs" role="tablist" aria-label="Showcase examples">
          {demoExamples.map((example) => {
            const isActive = example.id === activeExample.id;

            return (
              <motion.button
                key={example.id}
                className={`showcase-tab${isActive ? " active" : ""}`}
                type="button"
                role="tab"
                aria-selected={isActive}
                whileHover={reducedMotion ? undefined : { y: -3 }}
                whileTap={reducedMotion ? undefined : { scale: 0.99 }}
                transition={publicLayoutTransition}
                onClick={() => {
                  startTransition(() => {
                    setActiveExampleId(example.id);
                  });
                }}
              >
                {isActive ? (
                  <motion.span
                    className="selection-pill selection-pill--soft"
                    layoutId="showcase-tab-indicator"
                    transition={publicLayoutTransition}
                  />
                ) : null}
                <span>{example.label}</span>
                <strong>{example.displayName}</strong>
              </motion.button>
            );
          })}
        </div>
      </LayoutGroup>

      <AnimatePresence mode="wait">
        <motion.article
          key={activeExample.id}
          className="showcase-focus"
          role="tabpanel"
          aria-label={activeExample.displayName}
          initial={reducedMotion ? false : "hidden"}
          animate="show"
          exit={reducedMotion ? undefined : "exit"}
          variants={contentSwapVariants}
          transition={publicSwapTransition}
          style={paletteStyle}
        >
          <div className="showcase-focus__copy">
            <span className="voxel-chip">
              <HeartIcon className="inline-icon" />
              {activeExample.shortTag}
            </span>

            <div className="showcase-focus__header">
              <h3>{activeExample.displayName}</h3>
              <p>{activeExample.heroBlurb}</p>
            </div>

            <div className="showcase-badges">
              <span>Processed source clip</span>
              <span>{activeExample.scene}</span>
              <span>{activeExample.proofLabels.window}</span>
              <span>{activeExample.selectedWindow}</span>
            </div>

            <div className="showcase-proof-stats">
              <div>
                <span>Anchor frames</span>
                <strong>{activeExample.anchorFrames}</strong>
              </div>
              <div>
                <span>Inserted bridge</span>
                <strong>{activeExample.insertedDurationSeconds.toFixed(1)}s</strong>
              </div>
              <div>
                <span>Preview length</span>
                <strong>{activeExample.previewDurationSeconds.toFixed(1)}s</strong>
              </div>
            </div>

            {!compact ? (
              <div className="showcase-callout">
                <strong>
                  <SparkleIcon className="inline-icon" />
                  Why this one lands
                </strong>
                <p>{activeExample.summary}</p>
              </div>
            ) : null}
          </div>

          <div className="showcase-focus__media">
            <div className="showcase-video-frame">
              <video controls playsInline poster={activeExample.finalPoster}>
                <source src={activeExample.finalVideo} type="video/mp4" />
              </video>
              <div className="showcase-video-frame__tag">
                <PlayIcon className="inline-icon" />
                {activeExample.proofLabels.final}
              </div>
            </div>

            <div className="showcase-media-strip">
              <figure>
                <img src={activeExample.generatedPreview} alt={`${activeExample.displayName} generated bridge`} />
                <figcaption>{activeExample.proofLabels.bridge}</figcaption>
              </figure>
              <figure>
                <img src={activeExample.startFrame} alt={`${activeExample.displayName} start anchor`} />
                <figcaption>{activeExample.proofLabels.original}</figcaption>
              </figure>
              <figure>
                <img src={activeExample.stopFrame} alt={`${activeExample.displayName} end anchor`} />
                <figcaption>{activeExample.proofLabels.window}</figcaption>
              </figure>
            </div>
          </div>
        </motion.article>
      </AnimatePresence>
    </section>
  );
}
