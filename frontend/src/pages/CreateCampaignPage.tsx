import { CampaignForm } from "../components/CampaignForm";

export function CreateCampaignPage() {
  return (
    <div className="page-grid">
      <CampaignForm />
      <section className="panel">
        <p className="eyebrow">Expected flow</p>
        <h2>Campaign setup stays visible</h2>
        <p>
          This page reserves the campaign upload and inline-product path so
          Phase 1 can attach the real multipart workflow without reshaping the
          dashboard.
        </p>
      </section>
    </div>
  );
}
