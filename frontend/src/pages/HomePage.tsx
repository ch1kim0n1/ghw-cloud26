import { Link } from "react-router-dom";
import { HealthStatusPanel } from "../components/HealthStatusPanel";

const demoJobId = "job_9de1cbb7-ec84-4e2c-99f7-0d2dc6f21e0a";
const exampleOneJobId = "job_6678aff9-e05f-49c9-b4ee-1ffd0a9a0863";

const simpleSteps = [
  "Upload one source clip",
  "Pick one insertion moment",
  "Generate one bridge clip",
  "Export one preview MP4",
];

const truthCards = [
  {
    title: "What is real",
    body: "Campaign intake, job orchestration, slot handling, preview stitching, download flow, and the live dashboard.",
  },
  {
    title: "What is demo-assisted",
    body: "If provider generation is blocked, one manually generated clip can be imported and stitched through the same backend path.",
  },
  {
    title: "Why it matters",
    body: "The demo shows that the software can place a branded moment inside the scene instead of cutting away to a traditional ad break.",
  },
];

const exampleCards = [
  {
    id: "example1",
    label: "Example1",
    finalPoster: "/demo/example1-final-poster.png",
    finalVideo: "/demo/example1-final.mp4",
    generatedPreview: "/demo/example1-generated.gif",
    startFrame: "/demo/example1-start-frame.png",
    stopFrame: "/demo/example1-stop-frame.png",
    finalDuration: "64.498s",
    window: "41.708s -> 43.377s",
    anchors: "1250 -> 1300",
    jobId: exampleOneJobId,
  },
  {
    id: "example2",
    label: "Example2",
    finalPoster: "/demo/example2-final-poster.png",
    finalVideo: "/demo/example2-final.mp4",
    generatedPreview: "/demo/example2-generated.gif",
    startFrame: "/demo/start-frame.png",
    stopFrame: "/demo/stop-frame.png",
    finalDuration: "65.537s",
    window: "20.5s -> 21.0s",
    anchors: "615 -> 630",
    jobId: demoJobId,
  },
];

export function HomePage() {
  return (
    <div className="demo-home">
      <section className="demo-hero">
        <div className="demo-hero__copy">
          <p className="eyebrow">Context-Aware Fused Ad Insertion</p>
          <h1>Insert ads that feel like part of the scene.</h1>
          <p className="demo-hero__support">
            CAFAI finds a believable moment inside a source video, inserts a short branded bridge clip, and exports one
            reviewable preview MP4.
          </p>

          <div className="demo-hero__chips">
            <span>2 demo examples</span>
            <span>1 inserted moment each</span>
            <span>reviewable preview MP4s</span>
          </div>

          <div className="hero-actions">
            <Link className="button-link" to={`/jobs/${demoJobId}`}>
              Open demo run
            </Link>
            <Link className="button-link button-link--secondary" to="/results">
              View result
            </Link>
            <Link className="button-link button-link--secondary" to={`/jobs/${demoJobId}/preview`}>
              Watch preview
            </Link>
          </div>
        </div>

        <div className="demo-stage">
          <div className="demo-stage__video">
            <video autoPlay muted loop playsInline poster="/demo/example2-final-poster.png">
              <source src="/demo/example2-final.mp4" type="video/mp4" />
            </video>
            <div className="demo-stage__badge">Latest completed demo output</div>
          </div>

          <div className="demo-stage__meta">
            <div className="demo-stage__stats">
              <div>
                <strong>2</strong>
                <span>example outputs</span>
              </div>
              <div>
                <strong>59s baseline</strong>
                <span>source profile</span>
              </div>
              <div>
                <strong>manual import ready</strong>
                <span>demo recovery path</span>
              </div>
            </div>
            <HealthStatusPanel />
          </div>
        </div>
      </section>

      <section className="demo-strip">
        {simpleSteps.map((step, index) => (
          <article key={step} className="demo-step">
            <span className="demo-step__index">0{index + 1}</span>
            <p>{step}</p>
          </article>
        ))}
      </section>

      <section className="demo-proof">
        <div className="section-heading">
          <p className="eyebrow">Hackathon demo</p>
          <h2>Short, clear, and easy to explain live</h2>
          <p>
            This presentation view is intentionally simple: one real run, one imported bridge clip, one stitched
            result, and direct links into the working operator dashboard.
          </p>
        </div>

        <div className="demo-proof__grid">
          {truthCards.map((card) => (
            <article key={card.title} className="demo-card">
              <h3>{card.title}</h3>
              <p>{card.body}</p>
            </article>
          ))}
        </div>
      </section>

      <section className="demo-output">
        <div className="section-heading">
          <p className="eyebrow">Output</p>
          <h2>The final stitched preview is the center of the pitch</h2>
        </div>

        <div className="demo-output__grid">
          <article className="demo-output__main">
            <video controls playsInline poster="/demo/example2-final-poster.png">
              <source src="/demo/example2-final.mp4" type="video/mp4" />
            </video>
          </article>

          <article className="demo-output__side">
            <img src="/demo/example2-generated.gif" alt="Generated bridge clip preview" />
            <div className="frame-pair">
              <img src="/demo/start-frame.png" alt="Start anchor frame" />
              <img src="/demo/stop-frame.png" alt="Stop anchor frame" />
            </div>
          </article>
        </div>
      </section>

      <section className="demo-examples">
        <div className="section-heading">
          <p className="eyebrow">Both outputs</p>
          <h2>Two completed examples are available in the demo</h2>
          <p>
            Both stitched outputs are now surfaced directly in the frontend so you can show the first example and the
            second example side by side during the presentation.
          </p>
        </div>

        <div className="demo-examples__grid">
          {exampleCards.map((example) => (
            <article key={example.id} className="demo-example-card">
              <div className="demo-example-card__media">
                <video controls playsInline poster={example.finalPoster}>
                  <source src={example.finalVideo} type="video/mp4" />
                </video>
                <span className="demo-example-card__label">{example.label}</span>
              </div>

              <div className="demo-example-card__meta">
                <div className="demo-example-card__stats">
                  <div>
                    <strong>{example.finalDuration}</strong>
                    <span>final preview</span>
                  </div>
                  <div>
                    <strong>{example.window}</strong>
                    <span>selected window</span>
                  </div>
                  <div>
                    <strong>{example.anchors}</strong>
                    <span>anchor frames</span>
                  </div>
                </div>

                <div className="frame-pair">
                  <img src={example.startFrame} alt={`${example.label} start anchor`} />
                  <img src={example.stopFrame} alt={`${example.label} stop anchor`} />
                </div>

                <div className="hero-actions">
                  <Link className="button-link button-link--secondary" to={`/jobs/${example.jobId}`}>
                    Open {example.label}
                  </Link>
                  <Link className="button-link button-link--secondary" to="/results">
                    View detailed results
                  </Link>
                </div>
              </div>
            </article>
          ))}
        </div>
      </section>
    </div>
  );
}
