import { Link } from "react-router-dom";

type DemoExample = {
  id: string;
  label: string;
  jobId: string;
  summary: string;
  sourceDurationSeconds: number;
  insertStartSeconds: number;
  insertedDurationSeconds: number;
  previewDurationSeconds: number;
  anchorFrames: string;
  selectedWindow: string;
  finalVideo: string;
  finalPoster: string;
  generatedPreview: string;
  startFrame: string;
  stopFrame: string;
};

const demoExamples: DemoExample[] = [
  {
    id: "example1",
    label: "Example1",
    jobId: "job_6678aff9-e05f-49c9-b4ee-1ffd0a9a0863",
    summary: "Outdoor bicycle moment with a later insertion point and a shorter bridge clip.",
    sourceDurationSeconds: 59.526,
    insertStartSeconds: 41.708,
    insertedDurationSeconds: 4.972,
    previewDurationSeconds: 64.498,
    anchorFrames: "1250 -> 1300",
    selectedWindow: "41.708s -> 43.377s",
    finalVideo: "/demo/example1-final.mp4",
    finalPoster: "/demo/example1-final-poster.png",
    generatedPreview: "/demo/example1-generated.gif",
    startFrame: "/demo/example1-start-frame.png",
    stopFrame: "/demo/example1-stop-frame.png",
  },
  {
    id: "example2",
    label: "Example2",
    jobId: "job_9de1cbb7-ec84-4e2c-99f7-0d2dc6f21e0a",
    summary: "Talking-head desk setup with a manual-import bridge clip and local render fallback.",
    sourceDurationSeconds: 59.002,
    insertStartSeconds: 20.5,
    insertedDurationSeconds: 6.535,
    previewDurationSeconds: 65.537,
    anchorFrames: "615 -> 630",
    selectedWindow: "20.5s -> 21.0s",
    finalVideo: "/demo/example2-final.mp4",
    finalPoster: "/demo/example2-final-poster.png",
    generatedPreview: "/demo/example2-generated.gif",
    startFrame: "/demo/start-frame.png",
    stopFrame: "/demo/stop-frame.png",
  },
];

function buildTimelineSegments(example: DemoExample) {
  return [
    {
      label: "Original video before insertion",
      seconds: example.insertStartSeconds,
      tone: "base",
    },
    {
      label: "Generated bridge clip",
      seconds: example.insertedDurationSeconds,
      tone: "inserted",
    },
    {
      label: "Original video after insertion",
      seconds: example.sourceDurationSeconds - example.insertStartSeconds,
      tone: "base",
    },
  ];
}

function buildTimelineScale(sourceDurationSeconds: number) {
  const steps = [0, 10, 20, 30, 40, 50, Math.round(sourceDurationSeconds)];
  return steps.map((step) => `${step}s`);
}

export function ResultsPage() {
  return (
    <div className="results-page">
      <section className="results-header">
        <div>
          <p className="eyebrow">Produced results</p>
          <h1>Both stitched previews, shown clearly.</h1>
          <p>
            This page collects both completed demo outputs. Each section shows the exported video, the selected
            insertion window, and a timeline that highlights exactly where the generated clip sits inside the original
            one-minute source.
          </p>
        </div>
        <div className="hero-actions">
          <Link className="button-link" to={`/jobs/${demoExamples[1].jobId}/preview`}>
            Open latest live preview
          </Link>
          <a className="button-link button-link--secondary" href={demoExamples[1].finalVideo} download>
            Download latest MP4
          </a>
        </div>
      </section>

      <section className="results-overview-grid">
        {demoExamples.map((example) => (
          <article key={example.id} className="results-overview-card">
            <div className="results-overview-card__top">
              <p className="eyebrow">{example.label}</p>
              <h2>{example.selectedWindow}</h2>
              <p>{example.summary}</p>
            </div>

            <video controls playsInline poster={example.finalPoster}>
              <source src={example.finalVideo} type="video/mp4" />
            </video>

            <div className="results-overview-card__stats">
              <div className="results-stat">
                <span>Final preview</span>
                <strong>{example.previewDurationSeconds.toFixed(3)}s</strong>
              </div>
              <div className="results-stat">
                <span>Generated section</span>
                <strong>{example.insertedDurationSeconds.toFixed(3)}s</strong>
              </div>
              <div className="results-stat">
                <span>Anchor frames</span>
                <strong>{example.anchorFrames}</strong>
              </div>
            </div>
          </article>
        ))}
      </section>

      {demoExamples.map((example) => {
        const insertStartPercent = (example.insertStartSeconds / example.sourceDurationSeconds) * 100;
        const insertWidthPercent = (example.insertedDurationSeconds / example.sourceDurationSeconds) * 100;
        const timelineSegments = buildTimelineSegments(example);
        const timelineScale = buildTimelineScale(example.sourceDurationSeconds);

        return (
          <section key={example.id} className="results-example-section">
            <div className="section-heading">
              <p className="eyebrow">{example.label} timeline</p>
              <h2>Generated segment inside the original minute</h2>
              <p>
                Neutral slate represents untouched source footage. Amber marks the generated bridge clip for {example.label}.
              </p>
            </div>

            <div className="results-layout">
              <article className="results-video-card">
                <video controls playsInline poster={example.finalPoster}>
                  <source src={example.finalVideo} type="video/mp4" />
                </video>
              </article>

              <aside className="results-meta-card">
                <div className="results-stat">
                  <span>Source clip</span>
                  <strong>{example.sourceDurationSeconds.toFixed(3)}s</strong>
                </div>
                <div className="results-stat">
                  <span>Generated section</span>
                  <strong>{example.insertedDurationSeconds.toFixed(3)}s</strong>
                </div>
                <div className="results-stat">
                  <span>Final preview</span>
                  <strong>{example.previewDurationSeconds.toFixed(3)}s</strong>
                </div>
                <div className="results-stat">
                  <span>Anchor frames</span>
                  <strong>{example.anchorFrames}</strong>
                </div>
                <div className="results-stat">
                  <span>Selected time</span>
                  <strong>{example.selectedWindow}</strong>
                </div>
                <div className="results-stat">
                  <span>Job view</span>
                  <strong>{example.jobId}</strong>
                </div>
              </aside>
            </div>

            <article className="results-timeline-card">
              <div className="timeline-scale">
                {timelineScale.map((tick) => (
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

              <div className="results-legend">
                <span className="legend-item">
                  <i className="legend-item__swatch legend-item__swatch--base" />
                  Original source
                </span>
                <span className="legend-item">
                  <i className="legend-item__swatch legend-item__swatch--insert" />
                  Generated bridge clip
                </span>
              </div>

              <div className="results-breakdown">
                {timelineSegments.map((segment) => (
                  <article key={`${example.id}-${segment.label}`} className="results-breakdown__item">
                    <span className={`segment-tag segment-tag--${segment.tone}`}>{segment.label}</span>
                    <strong>{segment.seconds.toFixed(3)}s</strong>
                  </article>
                ))}
              </div>
            </article>

            <div className="results-supporting">
              <article className="results-supporting__card">
                <p className="eyebrow">Generated clip</p>
                <img src={example.generatedPreview} alt={`${example.label} generated bridge clip preview`} />
              </article>

              <article className="results-supporting__card">
                <p className="eyebrow">Anchor frames</p>
                <div className="frame-pair">
                  <img src={example.startFrame} alt={`${example.label} start anchor frame`} />
                  <img src={example.stopFrame} alt={`${example.label} stop anchor frame`} />
                </div>
              </article>
            </div>
          </section>
        );
      })}
    </div>
  );
}
