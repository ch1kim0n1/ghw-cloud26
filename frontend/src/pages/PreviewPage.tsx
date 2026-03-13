import { useParams } from "react-router-dom";
import { PreviewPlayer } from "../components/PreviewPlayer";
import { usePreview } from "../hooks/usePreview";

export function PreviewPage() {
  const { jobId } = useParams();
  const { error, loading } = usePreview(jobId);

  return (
    <div className="page-grid">
      <PreviewPlayer />
      <section className="panel">
        <p className="eyebrow">Render state</p>
        <h2>Preview scaffold {jobId ?? "demo-job"}</h2>
        <p>{loading ? "Loading preview placeholder…" : error ?? "Preview route scaffold ready."}</p>
      </section>
    </div>
  );
}
