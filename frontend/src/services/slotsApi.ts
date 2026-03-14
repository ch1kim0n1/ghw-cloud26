import { request } from "./apiClient";
import type { Slot } from "../types/Slot";

export function listSlots(jobId: string): Promise<{ job_id: string; slots: Slot[] }> {
  return request<{ job_id: string; slots: Slot[] }>(`/api/jobs/${jobId}/slots`);
}

export function getSlot(jobId: string, slotId: string): Promise<Slot> {
  return request<Slot>(`/api/jobs/${jobId}/slots/${slotId}`);
}

export function selectSlot(jobId: string, slotId: string): Promise<Record<string, unknown>> {
  return request(`/api/jobs/${jobId}/slots/${slotId}/select`, {
    method: "POST",
  });
}

export function manualSelectSlot(
  jobId: string,
  payload: { start_seconds: number; end_seconds: number },
): Promise<Record<string, unknown>> {
  return request(`/api/jobs/${jobId}/slots/manual-select`, {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
    },
    body: JSON.stringify(payload),
  });
}

export function manualImportSlot(
  jobId: string,
  payload: {
    slot_id?: string;
    start_seconds?: number;
    end_seconds?: number;
    generated_clip_path: string;
    generated_audio_path?: string;
  },
): Promise<Record<string, unknown>> {
  return request(`/api/jobs/${jobId}/slots/manual-import`, {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
    },
    body: JSON.stringify(payload),
  });
}

export function rejectSlot(jobId: string, slotId: string, note?: string): Promise<Record<string, unknown>> {
  return request(`/api/jobs/${jobId}/slots/${slotId}/reject`, {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
    },
    body: JSON.stringify({ note }),
  });
}

export function repickSlots(jobId: string): Promise<Record<string, unknown>> {
  return request(`/api/jobs/${jobId}/slots/re-pick`, {
    method: "POST",
  });
}

export function generateSlot(jobId: string, slotId: string, payload: Record<string, unknown>): Promise<Record<string, unknown>> {
  return request(`/api/jobs/${jobId}/slots/${slotId}/generate`, {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
    },
    body: JSON.stringify(payload),
  });
}
