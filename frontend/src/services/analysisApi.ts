import { request } from "./apiClient";

export function startAnalysis(jobId: string): Promise<{ job_id: string; status: string; current_stage: string; message: string }> {
  return request(`/api/jobs/${jobId}/start-analysis`, {
    method: "POST",
  });
}
