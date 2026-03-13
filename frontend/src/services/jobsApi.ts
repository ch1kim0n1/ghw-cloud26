import { request } from "./apiClient";
import type { Job, JobLog } from "../types/Job";

export function getJob(jobId: string): Promise<Job> {
  return request<Job>(`/api/jobs/${jobId}`);
}

export function getJobLogs(jobId: string): Promise<{ job_id: string; logs: JobLog[] }> {
  return request<{ job_id: string; logs: JobLog[] }>(`/api/jobs/${jobId}/logs`);
}
