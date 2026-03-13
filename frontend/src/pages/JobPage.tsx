import { useParams } from "react-router-dom";
import { JobStatusCard } from "../components/JobStatusCard";
import { ProductLineEditor } from "../components/ProductLineEditor";
import { SlotCard } from "../components/SlotCard";
import { useJob } from "../hooks/useJob";
import { useSlots } from "../hooks/useSlots";

export function JobPage() {
  const { jobId } = useParams();
  const { error: jobError, loading: jobLoading } = useJob(jobId);
  const { error: slotsError, loading: slotsLoading } = useSlots(jobId);

  return (
    <div className="page-grid">
      <section className="panel">
        <p className="eyebrow">Job workflow</p>
        <h2>Job placeholder {jobId ?? "demo-job"}</h2>
        <p>
          Job state, slot review, and product-line review are scaffolded now.
          The API calls are expected to return a Phase 0 placeholder until the
          later phases implement real logic.
        </p>
        <div className="status-strip">
          <span>{jobLoading ? "Loading job…" : jobError ?? "Job scaffold ready"}</span>
          <span>{slotsLoading ? "Loading slots…" : slotsError ?? "Slot scaffold ready"}</span>
        </div>
      </section>
      <div className="card-grid">
        <JobStatusCard
          title="Analysis stage"
          description="The backend job routes exist now and return the standard 501 placeholder envelope."
        />
        <SlotCard
          title="Slot proposals"
          description="Ranked slot cards will attach here in Phase 2 once analysis output is persisted."
        />
        <SlotCard
          title="Re-pick controls"
          description="Rejection and re-pick remain out of scope for the foundation pass, but the UI footprint is locked."
        />
      </div>
      <ProductLineEditor />
    </div>
  );
}
