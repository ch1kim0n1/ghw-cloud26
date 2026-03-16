import { request } from "./apiClient";
import type { Job, JobLog, JobSummary } from "../types/Job";

export function listJobs(limit = 25): Promise<{ jobs: JobSummary[] }> {
  return request<{ jobs: JobSummary[] }>(`/api/jobs?limit=${limit}`);
}

export function getJob(jobId: string): Promise<Job> {
  return request<Job>(`/api/jobs/${jobId}`);
}

export function getJobLogs(jobId: string): Promise<{ job_id: string; logs: JobLog[] }> {
  return request<{ job_id: string; logs: JobLog[] }>(`/api/jobs/${jobId}/logs`);
}
