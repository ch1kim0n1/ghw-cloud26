import { CampaignForm } from "../components/CampaignForm";
import { Reveal } from "../components/Reveal";

export function CreateCampaignPage() {
  return (
    <div className="studio-page">
      <Reveal as="section" className="studio-hero">
        <p className="eyebrow">Studio intake</p>
        <h1>Create a campaign, stage the asset, then hand off to analysis.</h1>
        <p>
          This view stays operational, but the framing is cleaner: intake first, operator context second, and the
          resulting job ready for a live demo handoff.
        </p>
      </Reveal>

      <div className="page-grid page-grid--studio">
        <CampaignForm />
        <section className="panel panel--supporting">
          <p className="eyebrow">Operator note</p>
          <h2>Campaigns stop before analysis</h2>
          <p>
            Campaign creation validates the uploaded H.264 MP4, stores measured FPS and duration, links the chosen
            product, and creates a job in <code>queued</code> with <code>ready_for_analysis</code>.
          </p>
        </section>
      </div>
    </div>
  );
}
