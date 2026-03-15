import { buildTimelineScale, buildTimelineSegments, type DemoExample } from "../content/demoContent";

interface TimelineBandProps {
  example: DemoExample;
  emphasized?: boolean;
}

export function TimelineBand({ example, emphasized = false }: TimelineBandProps) {
  const segments = buildTimelineSegments(example);
  const scale = buildTimelineScale(example.sourceDurationSeconds);
  const insertStartPercent = (example.insertStartSeconds / example.sourceDurationSeconds) * 100;
  const insertWidthPercent = (example.insertedDurationSeconds / example.sourceDurationSeconds) * 100;

  return (
    <article className={`timeline-card${emphasized ? " timeline-card--emphasized" : ""}`}>
      <div className="timeline-card__header">
        <div>
          <p className="eyebrow">{example.label}</p>
          <h3>{example.title}</h3>
        </div>
        <p>{example.selectedWindow}</p>
      </div>

      <div className="timeline-scale">
        {scale.map((tick) => (
          <span key={`${example.id}-${tick}`}>{tick}</span>
        ))}
      </div>

      <div className="results-timeline">
        <div className="results-timeline__base" />
        <div
          className="results-timeline__insert"
          style={{ left: `${insertStartPercent}%`, width: `${insertWidthPercent}%` }}
        />
      </div>

      <div className="results-breakdown">
        {segments.map((segment) => (
          <article key={`${example.id}-${segment.label}`} className="results-breakdown__item">
            <span className={`segment-tag segment-tag--${segment.tone}`}>{segment.label}</span>
            <strong>{segment.seconds.toFixed(3)}s</strong>
          </article>
        ))}
      </div>
    </article>
  );
}
