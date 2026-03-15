import { AnimatePresence, LayoutGroup, motion, useReducedMotion } from "framer-motion";
import { startTransition, useMemo, useState } from "react";
import { websiteAdsContent } from "../content/websiteAdsContent";
import { contentSwapVariants, publicLayoutTransition, publicSwapTransition } from "./publicMotion";

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
  const [activeExampleId, setActiveExampleId] = useState<string>(websiteAdsContent.examples[0]?.id ?? "");
  const reducedMotion = useReducedMotion();
  const activeExample = useMemo(
    () => websiteAdsContent.examples.find((example) => example.id === activeExampleId) ?? websiteAdsContent.examples[0],
    [activeExampleId],
  );

  if (!activeExample) {
    return null;
  }

  return (
    <section className="voxel-panel website-ads-showcase">
      <div className="section-heading section-heading--voxel">
        <p className="eyebrow">{eyebrow}</p>
        <h2>{title}</h2>
        <p>{lede}</p>
      </div>

      <LayoutGroup id="website-proof-switcher">
        <div className="website-proof-switcher" role="tablist" aria-label="Website ad proof examples">
          {websiteAdsContent.examples.map((example) => {
            const isActive = example.id === activeExample.id;

            return (
              <motion.button
                key={example.id}
                className={`website-proof-switcher__button${isActive ? " active" : ""}`}
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
                    className="selection-pill selection-pill--rose"
                    layoutId="website-proof-indicator"
                    transition={publicLayoutTransition}
                  />
                ) : null}
                <strong>{example.label}</strong>
                <span>{example.preview.publication}</span>
              </motion.button>
            );
          })}
        </div>
      </LayoutGroup>

      <AnimatePresence mode="wait" initial={false}>
        <motion.article
          key={activeExample.id}
          className="website-real-card"
          initial={reducedMotion ? false : "hidden"}
          animate="show"
          exit={reducedMotion ? undefined : "exit"}
          variants={contentSwapVariants}
          transition={publicSwapTransition}
          role="tabpanel"
          aria-label={activeExample.title}
        >
          <div className="website-real-card__header">
            <div>
              <p className="eyebrow">{activeExample.label}</p>
              <h3>{activeExample.title}</h3>
            </div>
            <a href={activeExample.url} target="_blank" rel="noreferrer">
              Open source
            </a>
          </div>

          <div className="website-real-card__meta">
            <span>{activeExample.preview.publication}</span>
            <span>{activeExample.preview.section}</span>
            <span>Captured page + injected placements</span>
          </div>

          <p className="website-real-card__note">{activeExample.note}</p>

          <div className="website-real-card__preview">
            <img src={activeExample.previewImage} alt={`${activeExample.title} site preview with injected ads`} />
          </div>

          <div className="website-real-card__story">
            <h4>{activeExample.preview.headline}</h4>
            <p>{activeExample.preview.dek}</p>
          </div>
        </motion.article>
      </AnimatePresence>

      <div className="website-real-grid">
        {websiteAdsContent.examples.map((example) => (
          <motion.article
            className={`website-real-card website-real-card--mini${example.id === activeExample.id ? " active" : ""}`}
            key={example.id}
            whileHover={reducedMotion ? undefined : { y: -4 }}
            whileTap={reducedMotion ? undefined : { scale: 0.994 }}
            transition={publicLayoutTransition}
          >
            <button
              className="website-real-card__mini-button"
              type="button"
              onClick={() => {
                startTransition(() => {
                  setActiveExampleId(example.id);
                });
              }}
            >
              <div className="website-real-card__preview">
                <img src={example.previewImage} alt={`${example.title} site preview with injected ads`} />
              </div>
              <div className="website-real-card__mini-copy">
                <strong>{example.label}</strong>
                <span>{example.title}</span>
              </div>
            </button>
          </motion.article>
        ))}
      </div>
    </section>
  );
}
