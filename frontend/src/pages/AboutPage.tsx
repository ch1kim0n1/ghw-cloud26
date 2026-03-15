import { motion, useReducedMotion } from "framer-motion";
import { FloatingDecor } from "../components/FloatingDecor";
import { publicLayoutTransition, staggerItemVariants } from "../components/publicMotion";
import { publicCopy } from "../content/publicCopy";

export function AboutPage() {
  const reducedMotion = useReducedMotion();

  return (
    <div className="about-page about-page--profiles">
      <section className="about-hero voxel-panel">
        <FloatingDecor ids={["cloud", "heart", "star"]} variant="about" />
        <span className="voxel-chip">{publicCopy.about.eyebrow}</span>
        <h1>{publicCopy.about.title}</h1>
        <p>{publicCopy.about.lede}</p>
        <div className="about-stats">
          <span>2 builders</span>
          <span>frontend + pipeline</span>
          <span>real demo assets</span>
        </div>
      </section>

      <section className="team-grid" aria-label="CAFAI developers">
        {publicCopy.about.cards.map((card) => (
          <motion.a
            className="founder-card founder-card--link voxel-panel"
            href={card.github}
            key={card.name}
            target="_blank"
            rel="noreferrer"
            aria-label={`${card.name} GitHub profile`}
            initial={reducedMotion ? false : "hidden"}
            whileInView="show"
            viewport={{ once: true, amount: 0.2 }}
            variants={staggerItemVariants}
            whileHover={reducedMotion ? undefined : { y: -5 }}
            whileTap={reducedMotion ? undefined : { scale: 0.992 }}
            transition={publicLayoutTransition}
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
          </motion.a>
        ))}
      </section>
    </div>
  );
}
