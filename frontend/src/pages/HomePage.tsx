import { AnimatePresence, LayoutGroup, motion, useReducedMotion } from "framer-motion";
import { lazy, Suspense, startTransition, useMemo, useState } from "react";
import { Link } from "react-router-dom";
import { FloatingDecor } from "../components/FloatingDecor";
import { HeartIcon, PlayIcon, SparkleIcon, UploadIcon } from "../components/PinkIcons";
import { contentSwapVariants, publicLayoutTransition, publicSwapTransition } from "../components/publicMotion";
import { Reveal, StaggerItem, StaggerList } from "../components/Reveal";
import { WebsiteAdsShowcase } from "../components/WebsiteAdsShowcase";
import { demoExamples, demoSteps, featuredDemoExample, heroStats, proofPoints } from "../content/demoContent";
import { publicCopy } from "../content/publicCopy";

const AmbientParticles = lazy(() => import("../components/AmbientParticles").then((module) => ({ default: module.AmbientParticles })));

type ProofStageId = "original" | "window" | "bridge" | "final";

const proofStageOrder: ProofStageId[] = ["original", "window", "bridge", "final"];

export function HomePage() {
  const [activeExampleId, setActiveExampleId] = useState(featuredDemoExample.id);
  const [proofStage, setProofStage] = useState<ProofStageId>("final");
  const reducedMotion = useReducedMotion();
  const activeExample = useMemo(
    () => demoExamples.find((example) => example.id === activeExampleId) ?? featuredDemoExample,
    [activeExampleId],
  );

  return (
    <div className="public-page public-page--landing">
      <section className="landing-hero voxel-panel">
        <Suspense fallback={null}>
          <AmbientParticles />
        </Suspense>
        <FloatingDecor ids={activeExample.decorAssetIds} />

        <Reveal className="landing-hero__copy" delay={0.04}>
          <span className="voxel-chip voxel-chip--soft">{publicCopy.landing.eyebrow}</span>
          <h1>{publicCopy.landing.title}</h1>
          <p className="landing-hero__lede">{publicCopy.landing.lede}</p>

          <div className="hero-actions">
            <a className="cute-button" href="#proof-room">
              <PlayIcon className="inline-icon" />
              {publicCopy.landing.primaryCta}
            </a>
            <Link className="cute-button cute-button--secondary" to="/gallery">
              <HeartIcon className="inline-icon" />
              {publicCopy.landing.secondaryCta}
            </Link>
          </div>

          <div className="hero-stats">
            <span>{publicCopy.landing.heroStatsTitle}</span>
            <div className="hero-stats__grid">
              {heroStats.map((item) => (
                <div className="hero-stats__item" key={item.label}>
                  <strong>{item.value}</strong>
                  <small>{item.label}</small>
                </div>
              ))}
            </div>
          </div>

          <div className="hero-note">
            <strong>
              <SparkleIcon className="inline-icon" />
              {publicCopy.landing.heroNoteTitle}
            </strong>
            <p>{publicCopy.landing.heroNote}</p>
          </div>
        </Reveal>

        <Reveal className="landing-hero__stage" delay={0.12}>
          <div className="hero-preview-card">
            <AnimatePresence mode="wait" initial={false}>
              <motion.div
                key={activeExample.id}
                className="hero-preview-card__content"
                initial={reducedMotion ? false : "hidden"}
                animate="show"
                exit={reducedMotion ? undefined : "exit"}
                variants={contentSwapVariants}
                transition={publicSwapTransition}
              >
                <div className="hero-preview-card__top">
                  <span className="voxel-chip">
                    <HeartIcon className="inline-icon" />
                    {activeExample.label}
                  </span>
                  <p>{activeExample.scene}</p>
                </div>

                <div className="hero-preview-card__media">
                  <video controls playsInline poster={activeExample.finalPoster}>
                    <source src={activeExample.finalVideo} type="video/mp4" />
                  </video>
                </div>

                <div className="hero-preview-card__bottom">
                  <div>
                    <h2>{activeExample.displayName}</h2>
                    <p>{activeExample.heroBlurb}</p>
                  </div>
                  <div className="showcase-badges">
                    <span>Processed demo clip</span>
                    <span>{activeExample.selectedWindow}</span>
                    <span>{activeExample.anchorFrames}</span>
                  </div>
                </div>
              </motion.div>
            </AnimatePresence>
          </div>

          <LayoutGroup id="hero-scene-picker">
            <div className="hero-scene-picker" role="tablist" aria-label="Featured demo selector">
              {demoExamples.map((example) => {
                const isActive = activeExample.id === example.id;

                return (
                  <motion.button
                    className={`hero-scene-picker__button${isActive ? " active" : ""}`}
                    key={example.id}
                    type="button"
                    role="tab"
                    aria-selected={isActive}
                    whileHover={reducedMotion ? undefined : { y: -3 }}
                    whileTap={reducedMotion ? undefined : { scale: 0.985 }}
                    transition={publicLayoutTransition}
                    onClick={() => {
                      startTransition(() => {
                        setActiveExampleId(example.id);
                        setProofStage("final");
                      });
                    }}
                  >
                    {isActive ? (
                      <motion.span
                        className="selection-pill selection-pill--soft"
                        layoutId="hero-scene-indicator"
                        transition={publicLayoutTransition}
                      />
                    ) : null}
                    <strong>{example.label}</strong>
                    <span>{example.shortTag}</span>
                  </motion.button>
                );
              })}
            </div>
          </LayoutGroup>
        </Reveal>
      </section>

      <Reveal as="section" className="proof-room voxel-panel" delay={0.06}>
        <div className="section-heading section-heading--voxel" id="proof-room">
          <p className="eyebrow">{publicCopy.landing.proofEyebrow}</p>
          <h2>{publicCopy.landing.proofTitle}</h2>
          <p>{publicCopy.landing.proofLede}</p>
        </div>

        <div className="proof-room__layout">
          <LayoutGroup id="proof-rail">
            <div className="proof-rail" role="tablist" aria-label="Proof stages">
              {proofStageOrder.map((stage) => {
                const label = activeExample.proofLabels[stage];
                const isActive = proofStage === stage;

                return (
                  <motion.button
                    className={`proof-rail__button${isActive ? " active" : ""}`}
                    key={stage}
                    type="button"
                    role="tab"
                    aria-selected={isActive}
                    whileHover={reducedMotion ? undefined : { y: -2 }}
                    whileTap={reducedMotion ? undefined : { scale: 0.99 }}
                    transition={publicLayoutTransition}
                    onClick={() => {
                      startTransition(() => {
                        setProofStage(stage);
                      });
                    }}
                  >
                    {isActive ? (
                      <motion.span
                        className="selection-pill selection-pill--rose"
                        layoutId="proof-stage-indicator"
                        transition={publicLayoutTransition}
                      />
                    ) : null}
                    <small>{stage}</small>
                    <strong>{label}</strong>
                  </motion.button>
                );
              })}
            </div>
          </LayoutGroup>

          <AnimatePresence mode="wait">
            <motion.div
              key={`${activeExample.id}-${proofStage}`}
              className="proof-stage"
              initial={reducedMotion ? false : "hidden"}
              animate="show"
              exit={reducedMotion ? undefined : "exit"}
              variants={contentSwapVariants}
              transition={publicSwapTransition}
              role="tabpanel"
              aria-label={activeExample.proofLabels[proofStage]}
            >
              {proofStage === "final" ? (
                <div className="proof-stage__video">
                  <video controls playsInline poster={activeExample.finalPoster}>
                    <source src={activeExample.finalVideo} type="video/mp4" />
                  </video>
                </div>
              ) : null}

              {proofStage === "bridge" ? (
                <div className="proof-stage__single">
                  <img src={activeExample.generatedPreview} alt={`${activeExample.displayName} generated bridge preview`} />
                </div>
              ) : null}

              {proofStage === "original" ? (
                <div className="proof-stage__single">
                  <img src={activeExample.startFrame} alt={`${activeExample.displayName} original scene frame`} />
                </div>
              ) : null}

              {proofStage === "window" ? (
                <div className="proof-stage__window">
                  <figure>
                    <img src={activeExample.startFrame} alt={`${activeExample.displayName} insert window start frame`} />
                    <figcaption>Start anchor</figcaption>
                  </figure>
                  <figure>
                    <img src={activeExample.stopFrame} alt={`${activeExample.displayName} insert window stop frame`} />
                    <figcaption>Stop anchor</figcaption>
                  </figure>
                </div>
              ) : null}

              <div className="proof-stage__caption">
                <span className="voxel-chip voxel-chip--soft">{activeExample.proofLabels[proofStage]}</span>
                <p>{renderProofCopy(proofStage, activeExample)}</p>
              </div>
            </motion.div>
          </AnimatePresence>
        </div>

        <StaggerList className="proof-points">
          {proofPoints.map((point) => (
            <StaggerItem as="article" className="proof-point-card" key={point.title}>
              <h3>{point.title}</h3>
              <p>{point.body}</p>
            </StaggerItem>
          ))}
        </StaggerList>
      </Reveal>

      <Reveal as="section" className="landing-steps voxel-panel" delay={0.1}>
        <div className="section-heading section-heading--voxel">
          <p className="eyebrow">{publicCopy.landing.stepsEyebrow}</p>
          <h2>{publicCopy.landing.stepsTitle}</h2>
          <p>{publicCopy.landing.stepsLede}</p>
        </div>

        <StaggerList className="step-grid">
          {demoSteps.map((step, index) => (
            <StaggerItem as="article" className="step-card" key={step}>
              <span className="step-card__index">0{index + 1}</span>
              <p>{step}</p>
            </StaggerItem>
          ))}
        </StaggerList>
      </Reveal>

      <Reveal as="section" className="gallery-teaser voxel-panel" delay={0.12}>
        <div className="section-heading section-heading--voxel">
          <p className="eyebrow">{publicCopy.landing.teaserEyebrow}</p>
          <h2>{publicCopy.landing.teaserTitle}</h2>
          <p>{publicCopy.landing.teaserLede}</p>
        </div>

        <div className="gallery-teaser__grid">
          {demoExamples.map((example) => (
            <motion.article
              key={example.id}
              className={`gallery-teaser__card${example.featured ? " active" : ""}`}
              whileHover={reducedMotion ? undefined : { y: -5, scale: 1.01 }}
              whileTap={reducedMotion ? undefined : { scale: 0.992 }}
              transition={publicLayoutTransition}
            >
              <img src={example.finalPoster} alt={`${example.displayName} poster`} />
              <div className="gallery-teaser__card-copy">
                <span className="voxel-chip voxel-chip--soft">{example.featured ? "Featured on home" : "See in gallery"}</span>
                <h3>{example.displayName}</h3>
                <p>{example.heroBlurb}</p>
                <div className="teaser-card__meta">
                  <span>{example.scene}</span>
                  <span>{example.proofLabels.final}</span>
                </div>
              </div>
            </motion.article>
          ))}
        </div>

        <Link className="cute-button" to="/gallery">
          <PlayIcon className="inline-icon" />
          {publicCopy.landing.teaserCta}
        </Link>
      </Reveal>

      <Reveal delay={0.13}>
        <WebsiteAdsShowcase
          eyebrow="Static ad channel"
          title="CAFAI also shows off static website ads"
          lede="The same demo surface now includes three static-ad proofs, so the homepage shows both stitched video ads and injected website placements together."
        />
      </Reveal>

      <Reveal as="section" className="landing-cta voxel-panel" delay={0.14}>
        <FloatingDecor ids={["cloud", "bow", "heart"]} variant="upload" />
        <div className="section-heading section-heading--voxel">
          <p className="eyebrow">{publicCopy.landing.ctaEyebrow}</p>
          <h2>{publicCopy.landing.ctaTitle}</h2>
          <p>{publicCopy.landing.ctaLede}</p>
        </div>

        <div className="landing-cta__actions">
          <Link className="cute-button" to="/upload">
            <UploadIcon className="inline-icon" />
            {publicCopy.landing.ctaPrimary}
          </Link>
          <Link className="cute-button cute-button--secondary" to="/gallery">
            <PlayIcon className="inline-icon" />
            {publicCopy.landing.ctaSecondary}
          </Link>
        </div>
      </Reveal>

      <Reveal as="section" className="about-teaser voxel-panel" delay={0.16}>
        <div className="section-heading section-heading--voxel">
          <p className="eyebrow">{publicCopy.landing.aboutEyebrow}</p>
          <h2>{publicCopy.landing.aboutTitle}</h2>
          <p>{publicCopy.landing.aboutLede}</p>
        </div>

        <div className="about-teaser__grid">
          {publicCopy.about.cards.map((card) => (
            <motion.a
              className="founder-card founder-card--link founder-card--teaser voxel-panel"
              href={card.github}
              key={card.name}
              target="_blank"
              rel="noreferrer"
              aria-label={`${card.name} GitHub profile`}
              whileHover={reducedMotion ? undefined : { y: -5 }}
              whileTap={reducedMotion ? undefined : { scale: 0.992 }}
              transition={publicLayoutTransition}
            >
              <div className="founder-card__avatar-frame">
                <img className="founder-card__avatar" src={card.avatar} alt={`${card.name} profile meme`} />
              </div>
              <div className="founder-card__copy">
                <h3>{card.name}</h3>
                <p className="founder-card__role">{card.role}</p>
                <p>{card.bio}</p>
                <span className="founder-card__github">{card.githubLabel}</span>
              </div>
            </motion.a>
          ))}
        </div>

        <Link className="cute-link" to="/about">
          {publicCopy.landing.aboutCta}
        </Link>
      </Reveal>
    </div>
  );
}

function renderProofCopy(stage: ProofStageId, example: typeof featuredDemoExample) {
  switch (stage) {
    case "original":
      return `This is the source beat before the branded moment arrives, keeping the scene's original framing and tempo intact.`;
    case "window":
      return `The insert window sits between ${example.selectedWindow}, anchored by frames ${example.anchorFrames} so the transition stays believable.`;
    case "bridge":
      return `CAFAI generates a short bridge clip for the brand moment instead of throwing in a blunt cutaway.`;
    case "final":
      return `The stitched cut keeps the product beat inside the scene rhythm, which is exactly why the result feels polished instead of disruptive.`;
    default:
      return "";
  }
}
