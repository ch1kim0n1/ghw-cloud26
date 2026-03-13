import { CampaignForm } from "../components/CampaignForm";

export function CreateCampaignPage() {
  return (
    <div className="page-grid">
      <CampaignForm />
      <section className="panel">
        <p className="eyebrow">Phase 1 contract</p>
        <h2>Campaigns stop before analysis</h2>
        <p>
          Campaign creation validates the uploaded H.264 MP4, stores the measured FPS and duration, links the chosen product, and creates a job in <code>queued</code> with <code>ready_for_analysis</code>.
        </p>
      </section>
    </div>
  );
}
