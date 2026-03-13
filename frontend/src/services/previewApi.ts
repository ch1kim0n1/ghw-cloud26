import { buildApiUrl, request } from "./apiClient";
import type { Preview } from "../types/Preview";

export function renderPreview(jobId: string, slotId: string): Promise<Record<string, unknown>> {
  return request(`/api/jobs/${jobId}/preview/render`, {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
    },
    body: JSON.stringify({ slot_id: slotId }),
  });
}

export function getPreview(jobId: string): Promise<Preview> {
  return request<Preview>(`/api/jobs/${jobId}/preview`);
}

export function getPreviewDownloadUrl(jobId: string): string {
  return buildApiUrl(`/api/jobs/${jobId}/preview/download`);
}

export function getPreviewStreamUrl(jobId: string): string {
  return buildApiUrl(`/api/jobs/${jobId}/preview/stream`);
}
